// **Dateien hin und her über den gebauten Dienst** — FMT-03.AK-03
// (#1692).
//
// Ein Aufrufer ohne Pfad auf diesem Rechner schickt das Set selbst, sieht es
// vorab an, fährt es ein, legt einen Bogen ohne Ziel und holt ihn ab — und
// danach ist die Ausgabe fort.
package flaechen_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// gezipptesProbeset packt `testdaten/probe.mse-set` so, wie MSE es schreibt.
func gezipptesProbeset(t *testing.T) []byte {
	t.Helper()

	inhalt, err := os.ReadFile(filepath.Join(probeset(t), "set")) //nolint:gosec // eine Datei im Repositorium
	if err != nil {
		t.Fatal(err)
	}

	var puffer bytes.Buffer
	archiv := zip.NewWriter(&puffer)

	eintrag, err := archiv.Create("set")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := eintrag.Write(inhalt); err != nil {
		t.Fatal(err)
	}

	if err := archiv.Close(); err != nil {
		t.Fatal(err)
	}

	return puffer.Bytes()
}

// uebertrage schickt einen Rumpf mit einer Art und gibt Status, Kopf und Rumpf.
func uebertrage(
	t *testing.T, an dienst, methode, weg, art string, rumpf []byte,
) (status int, kopf http.Header, antwort []byte) {
	t.Helper()

	anfrage, err := http.NewRequestWithContext(t.Context(), methode, an.adresse+weg, bytes.NewReader(rumpf))
	if err != nil {
		t.Fatal(err)
	}

	anfrage.Header.Set("Authorization", "Bearer "+an.marke)

	if art != "" {
		anfrage.Header.Set("Content-Type", art)
	}

	kam, err := an.klient.Do(anfrage)
	if err != nil {
		t.Fatalf("%s %s: %v", methode, weg, err)
	}

	defer func() { _ = kam.Body.Close() }() //nolint:errcheck // nur gelesen

	roh, err := io.ReadAll(kam.Body)
	if err != nil {
		t.Fatal(err)
	}

	return kam.StatusCode, kam.Header, roh
}

func TestEinSetGehtOhnePfadHinUndZurueck(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)
	set := gezipptesProbeset(t)

	status, _, rumpf := uebertrage(t, an, http.MethodPost, "/sets/vorschau", "application/zip", set)
	if status != http.StatusOK || !strings.Contains(string(rumpf), `"name":"Mein Probeset"`) {
		t.Fatalf("Vorschau: %d %s", status, rumpf)
	}

	status, _, rumpf = uebertrage(t, an, http.MethodPost, "/sets", "application/zip", set)

	var eingefahren struct {
		SetID string `json:"setID"`
	}

	if status != http.StatusCreated || json.Unmarshal(rumpf, &eingefahren) != nil || eingefahren.SetID == "" {
		t.Fatalf("Einfahren: %d %s", status, rumpf)
	}

	// **Dieselbe Übertragung ein zweites Mal legt kein zweites Set an.**
	if _, _, zweiter := uebertrage(t, an, http.MethodPost, "/sets", "application/zip", set); !strings.Contains(string(zweiter), `"setID":"`+eingefahren.SetID+`"`) {
		t.Errorf("zweites Einfahren: %s", zweiter)
	}

	status, kopf, rumpf := uebertrage(t, an, http.MethodPost, "/sets/"+eingefahren.SetID+"/bogen", "application/json", []byte(`{}`))
	ort := kopf.Get("Location")

	if status != http.StatusCreated || !strings.HasPrefix(ort, "/ausgang/") {
		t.Fatalf("Bogen: %d %v %s", status, kopf, rumpf)
	}

	status, kopf, rumpf = uebertrage(t, an, http.MethodGet, ort, "", nil)
	if status != http.StatusOK || kopf.Get("Content-Type") != "application/pdf" || !bytes.HasPrefix(rumpf, []byte("%PDF-")) {
		t.Fatalf("Abholen: %d %v %.20q", status, kopf, rumpf)
	}

	if status, _, _ := uebertrage(t, an, http.MethodDelete, ort, "", nil); status != http.StatusNoContent {
		t.Errorf("Vergessen: %d", status)
	}

	if status, _, _ := uebertrage(t, an, http.MethodGet, ort, "", nil); status != http.StatusNotFound {
		t.Errorf("danach: %d", status)
	}
}

// langsamerLeser gibt seinen Inhalt in Stücken und wartet dazwischen — eine
// Leitung, über die eine Übertragung länger dauert als die Lesefrist.
type langsamerLeser struct {
	rest   []byte
	stueck int
	pause  time.Duration
}

func (l *langsamerLeser) Read(ziel []byte) (int, error) {
	if len(l.rest) == 0 {
		return 0, io.EOF
	}

	time.Sleep(l.pause)

	n := copy(ziel[:min(len(ziel), l.stueck)], l.rest)
	l.rest = l.rest[n:]

	return n, nil
}

// **Eine Übertragung darf länger dauern als die Lesefrist** — gemessen und
// nicht angenommen: der Dienst liest den Kopf in zehn Sekunden, und bis
// FMT-03 galt dieselbe Frist für den ganzen Rumpf. Ein Set über eine
// langsame Leitung brach dann mitten im Lesen ab.
func TestEineLangsameUebertragungKommtAn(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)
	set := gezipptesProbeset(t)

	anfrage, err := http.NewRequestWithContext(t.Context(), http.MethodPost, an.adresse+"/sets/vorschau",
		&langsamerLeser{rest: set, stueck: 1 + len(set)/12, pause: time.Second})
	if err != nil {
		t.Fatal(err)
	}

	anfrage.ContentLength = int64(len(set))
	anfrage.Header.Set("Authorization", "Bearer "+an.marke)
	anfrage.Header.Set("Content-Type", "application/zip")

	kam, err := an.klient.Do(anfrage)
	if err != nil {
		t.Fatalf("die Übertragung brach ab: %v", err)
	}

	defer func() { _ = kam.Body.Close() }() //nolint:errcheck // nur gelesen

	if kam.StatusCode != http.StatusOK {
		roh, _ := io.ReadAll(kam.Body) //nolint:errcheck // nur für die Meldung
		t.Errorf("%d %s", kam.StatusCode, roh)
	}
}
