// **Ein Dienst, der mitten in einem Auftrag endet** — BK-07b.AK-04
// (#2116).
//
// Ein echter `mse-rest` über einem eigenen Bestand nimmt einen Bogen als
// Auftrag an und bekommt gleich danach `SIGTERM`. Gemessen wird nicht die
// Rückgabe von `Shutdown`, sondern was danach dasteht: der Prozess ist in
// seiner Frist weg, er endet ohne Fehler, die Zieldatei ist ganz oder gar
// nicht da, und im Protokoll steht kein Schreiben in einen geschlossenen
// Bestand.
package flaechen_test

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestEinDienstendeMittenImAuftragLaesstNichtsHalbesZurueck(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := setkennungAus(t, mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json"))
	pfad := filepath.Join(t.TempDir(), "abgebrochen.pdf")

	// **Ohne den Kontext des Tests**: das Signal schickt dieser Lauf selbst,
	// und ein Abbruch durch den Test wäre ein `SIGKILL` an seiner Stelle.
	// #nosec G204 -- gebautes Programm aus TestMain
	lauf := exec.CommandContext(context.WithoutCancel(t.Context()), gebaut.rest, "-port", "0")
	lauf.Env = umgebung

	angesagt := &mitschrift{}
	lauf.Stdout, lauf.Stderr = angesagt, angesagt

	if fehler := lauf.Start(); fehler != nil {
		t.Fatalf("den Dienst starten: %v", fehler)
	}

	geendet := make(chan error, 1)
	zu := make(chan struct{})

	go func() {
		geendet <- lauf.Wait()
		close(zu)
	}()

	t.Cleanup(func() {
		_ = lauf.Process.Kill() //nolint:errcheck // er ist dann meist schon weg
		<-zu
	})

	adresse, marke, _ := ansageLesen(t, angesagt, zu)
	an := dienst{adresse: adresse, marke: marke, klient: &http.Client{Timeout: 30 * time.Second}}

	anfrage, fehler := http.NewRequestWithContext(t.Context(), http.MethodPost,
		an.adresse+"/sets/"+setID+"/bogen", strings.NewReader(`{"ziel":`+alsJSONText(t, pfad)+`}`))
	if fehler != nil {
		t.Fatal(fehler)
	}

	anfrage.Header.Set("Authorization", "Bearer "+an.marke)
	anfrage.Header.Set("Prefer", "respond-async")

	kam, fehler := an.klient.Do(anfrage)
	if fehler != nil {
		t.Fatal(fehler)
	}

	_ = kam.Body.Close() //nolint:errcheck // nur der Kopf zählt

	if kam.StatusCode != http.StatusAccepted {
		t.Fatalf("POST: %d", kam.StatusCode)
	}

	// **Läuft er noch?** Sonst prüfte dieser Lauf ein Ende ohne Auftrag.
	status, rumpf := auftragsweg(t, an, "GET /auftraege/{auftragID}",
		strings.TrimPrefix(kam.Header.Get("Location"), "/auftraege/"))
	if status != http.StatusOK || !strings.Contains(rumpf, `"stand":"laeuft"`) {
		t.Fatalf("der Auftrag lief beim Signal nicht mehr: %d %s", status, rumpf)
	}

	angefangen := time.Now()

	if fehler := lauf.Process.Signal(syscall.SIGTERM); fehler != nil {
		t.Fatalf("SIGTERM: %v", fehler)
	}

	select {
	case fehler := <-geendet:
		// **Ohne Fehler**: ein Auftrag, der in der Frist nicht endete, käme
		// als Fehler zurück (`rest.ErrAuftraegeLaufenNoch`).
		if fehler != nil {
			t.Errorf("der Dienst endete mit %v:\n%s", fehler, angesagt.text())
		}
	case <-time.After(15 * time.Second):
		t.Fatalf("der Dienst lief 15 s nach SIGTERM noch:\n%s", angesagt.text())
	}

	t.Logf("zu Ende nach %v", time.Since(angefangen))

	// **Ganz oder gar nicht**: eine halbe Datei wäre ein Bogen, den niemand
	// als kaputt erkennt.
	roh, lesefehler := os.ReadFile(pfad) // #nosec G304 -- Testpfad aus t.TempDir()
	if lesefehler == nil && !strings.HasSuffix(strings.TrimSpace(string(roh)), "%%EOF") {
		t.Errorf("eine halbe Datei blieb zurück: %d Bytes", len(roh))
	}

	if protokoll := angesagt.text(); strings.Contains(protokoll, "database is closed") {
		t.Errorf("der Auftrag schrieb in den geschlossenen Bestand:\n%s", protokoll)
	}
}
