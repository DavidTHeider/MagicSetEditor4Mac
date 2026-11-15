package mcp_test

// **Das Werkzeug der atomaren Gruppe** — Werkstück A13.1b von Phase 5.3
// (#1452).

import (
	"errors"
	"strings"
	"testing"
)

// **Die Gruppe kommt an — samt den Namen je Karte.**
func TestDasWerkzeugKIGruppeGibtDieGeschriebenenFelder(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"schreiben.ki.gruppe",
		`{"wahl": [{"kennung": "a1", "felder": ["name"]}]}`)
	if err != nil {
		t.Fatalf("schreiben.ki.gruppe: %v", err)
	}

	for _, teil := range []string{
		`"karten": 1`, `"felder": 1`, `"karteId": 42`, `"name"`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("die Antwort nennt %s nicht:\n%s", teil, antwort)
		}
	}
}

// **Leere Listen bleiben Listen.**
//
// **Eine leere `wahl` weist `service` ab** — hier geht sie durch, weil das
// Doppel die Entscheidung nicht trifft. Geprüft wird die **Gestalt** der
// Antwort: `null` sieht für ein Sprachmodell aus wie „konnte nicht
// beantwortet werden".
func TestEineLeereGruppeGibtListenUndKeinNull(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, nilprobe{},
		"schreiben.ki.gruppe", `{"wahl": []}`)
	if err != nil {
		t.Fatalf("schreiben.ki.gruppe: %v", err)
	}

	if strings.Contains(antwort, "null") {
		t.Errorf("die Antwort trägt ein null:\n%s", antwort)
	}

	if !strings.Contains(antwort, `"namen": []`) {
		t.Errorf("die Antwort nennt namen nicht als Liste:\n%s", antwort)
	}
}

// **Ein Fehler der Werkstatt kommt als Fehler durch.**
func TestEinFehlerDerGruppeKommtAlsFehlerDurch(t *testing.T) {
	t.Parallel()

	_, err := rufeSammlungswerkzeug(t, nilprobe{}, stolpernderPreis{},
		"schreiben.ki.gruppe",
		`{"wahl": [{"kennung": "a1", "felder": ["name"]}]}`)
	if !errors.Is(err, errGestolpertMCP) {
		t.Fatalf("der Fehler ist %v", err)
	}
}

// **Und ein kaputtes Argument ebenso.**
func TestEinKaputtesGruppenargumentWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"schreiben.ki.gruppe", `{"wahl": "keine Liste"}`); err == nil {
		t.Fatal("das kaputte Argument kam durch")
	}
}

// **Und die Listen *innerhalb* der Namen bleiben Listen.**
//
// **Der Lauf darüber greift eine Ebene zu hoch**: er prüft `namen`, und die
// Felder **darin** bleiben dabei unberührt.
func TestEineGruppeOhneFelderGibtListenUndKeinNull(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, nilprobe{},
		"schreiben.ki.gruppe",
		`{"wahl": [{"kennung": "a1", "felder": ["name"]}]}`)
	if err != nil {
		t.Fatalf("schreiben.ki.gruppe: %v", err)
	}

	if strings.Contains(antwort, "null") {
		t.Errorf("die Antwort trägt ein null:\n%s", antwort)
	}
}
