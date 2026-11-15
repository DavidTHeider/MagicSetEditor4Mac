// **Lange Datei-Läufe mit Fortschritt, an den gebauten Programmen** —
// FMT-03.AK-02 (#1692).
//
// Die Pakete prüfen Buch, Melder und Protokoll je für sich; hier geht es um
// die **Naht**: meldet der Fachkern wirklich bis zum Werkzeugserver durch,
// und nimmt `mse-rest` einen Bogen als Auftrag an, dessen Ergebnis danach das
// PDF nennt?
package flaechen_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEinBogenMeldetSeinenFortschrittUeberMCP(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := setkennungAus(t, mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json"))
	pfad := filepath.Join(t.TempDir(), "mcp.pdf")

	ruf := mussJSON(t, map[string]any{
		"jsonrpc": "2.0", "id": 2, "method": "tools/call",
		"params": map[string]any{
			"name":      "schreiben.set.bogen",
			"arguments": map[string]any{"setID": zahl(t, setID), "ziel": pfad},
			"_meta":     map[string]any{"progressToken": "bogen"},
		},
	})

	// #nosec G204 -- gebautes Programm aus TestMain
	lauf := exec.CommandContext(t.Context(), gebaut.mcp)
	lauf.Env = umgebung
	lauf.Stdin = strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n" + string(ruf) + "\n")

	var aus bytes.Buffer
	lauf.Stdout = &aus

	if err := lauf.Run(); err != nil {
		t.Fatalf("mse-mcp: %v", err)
	}

	zeilen := strings.Split(strings.TrimSpace(aus.String()), "\n")
	meldungen := 0

	for _, zeile := range zeilen {
		if strings.Contains(zeile, `"notifications/progress"`) && strings.Contains(zeile, `"progressToken":"bogen"`) {
			meldungen++
		}
	}

	if meldungen < 2 {
		t.Errorf("%d Meldungen, erwartet mindestens die erste und die letzte:\n%s", meldungen, aus.String())
	}

	letzte := zeilen[len(zeilen)-1]
	if !strings.Contains(letzte, `"id":2`) || strings.Contains(letzte, `"isError":true`) {
		t.Fatalf("die Antwort: %s", letzte)
	}

	istEinPDF(t, "MCP", pfad, letzte)
}

func TestEinBogenLaeuftAlsAuftragDesDienstes(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := setkennungAus(t, mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json"))
	pfad := filepath.Join(t.TempDir(), "rest.pdf")
	an := starteDienst(t, umgebung)

	anfrage, err := http.NewRequestWithContext(t.Context(), http.MethodPost,
		an.adresse+"/sets/"+setID+"/bogen", strings.NewReader(`{"ziel":`+alsJSONText(t, pfad)+`}`))
	if err != nil {
		t.Fatal(err)
	}

	anfrage.Header.Set("Authorization", "Bearer "+an.marke)
	anfrage.Header.Set("Prefer", "respond-async")

	kam, err := an.klient.Do(anfrage)
	if err != nil {
		t.Fatal(err)
	}

	_ = kam.Body.Close() //nolint:errcheck // nur der Kopf zählt

	ort := kam.Header.Get("Location")
	if kam.StatusCode != http.StatusAccepted || !strings.HasPrefix(ort, "/auftraege/") {
		t.Fatalf("POST: %d, Ort %q", kam.StatusCode, ort)
	}

	id := strings.TrimPrefix(ort, "/auftraege/")

	// **Auf die Wirkung gewartet** — der Stand, nicht der Ruf (§3).
	var auftrag struct {
		Auftrag struct {
			Stand  string `json:"stand"`
			Status int    `json:"status"`
		} `json:"auftrag"`
	}

	for frist := time.Now().Add(60 * time.Second); time.Now().Before(frist); time.Sleep(50 * time.Millisecond) {
		status, rumpf := auftragsweg(t, an, "GET /auftraege/{auftragID}", id)
		if status != http.StatusOK || json.Unmarshal([]byte(rumpf), &auftrag) != nil {
			t.Fatalf("GET %s: %d %s", ort, status, rumpf)
		}

		if auftrag.Auftrag.Stand != "laeuft" {
			break
		}
	}

	if auftrag.Auftrag.Stand != "fertig" || auftrag.Auftrag.Status != http.StatusCreated {
		t.Fatalf("Auftrag: %+v", auftrag.Auftrag)
	}

	status, ergebnis := auftragsweg(t, an, "GET /auftraege/{auftragID}/ergebnis", id)
	if status != http.StatusCreated {
		t.Fatalf("Ergebnis: %d %s", status, ergebnis)
	}

	istEinPDF(t, "REST", pfad, ergebnis)

	if status, liste := auftragsweg(t, an, "GET /auftraege", id); status != http.StatusOK || !strings.Contains(liste, id) {
		t.Errorf("Liste: %d %s", status, liste)
	}

	// **Ein beendeter Auftrag wird vergessen** — danach kennt ihn der Dienst nicht mehr.
	if status, _ := auftragsweg(t, an, "DELETE /auftraege/{auftragID}", id); status != http.StatusNoContent {
		t.Errorf("vergessen: %d", status)
	}

	if status, _ := auftragsweg(t, an, "GET /auftraege/{auftragID}", id); status != http.StatusNotFound {
		t.Errorf("nach dem Vergessen: %d", status)
	}
}

// auftragsweg ruft einen Weg in der Gestalt, in der der Katalog ihn nennt.
func auftragsweg(t *testing.T, an dienst, weg, id string) (status int, rumpf string) {
	t.Helper()

	methode, pfad, _ := strings.Cut(weg, " ")

	return an.ruf(t, methode, strings.ReplaceAll(pfad, "{auftragID}", id), "")
}
