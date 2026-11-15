package mcp_test

// **Das Werkzeug für den Modellkatalog** — Werkstück A6 von Phase 5.3
// (#1445).

import (
	"errors"
	"strings"
	"testing"
)

// **Der Katalog kommt an — samt Beleg, Herkunft und Regel.**
func TestDasWerkzeugKIModelleGibtSeinenKatalog(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"lesen.ki.modelle", `{}`)
	if err != nil {
		t.Fatalf("lesen.ki.modelle: %v", err)
	}

	for _, teil := range []string{
		`"nurLokal": true`, `"alias": "beispiel:3b"`, `"ort": "lokal"`,
		`"faehigkeit": "einbettung"`, `"datum": "2026-09-11"`,
		`"art": "artefakt"`, `"revision": "sha256:dddd"`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("die Antwort nennt %s nicht:\n%s", teil, antwort)
		}
	}
}

// **Eine leere Liste bleibt eine Liste.**
//
// Ein `null` sähe für ein Sprachmodell aus wie „konnte nicht beantwortet
// werden" — und das ist das Gegenteil der Auskunft „es steht keines fest".
func TestEinLeererModellkatalogBleibtEineListe(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, nilprobe{},
		"lesen.ki.modelle", `{}`)
	if err != nil {
		t.Fatalf("lesen.ki.modelle: %v", err)
	}

	if !strings.Contains(antwort, `"modelle": []`) {
		t.Errorf("die Antwort ist %s", antwort)
	}
}

// **Und ein Fehler der Naht kommt durch.**
func TestEinFehlerDesModellkatalogsKommtDurch(t *testing.T) {
	t.Parallel()

	if _, err := rufeSammlungswerkzeug(t, nilprobe{}, stolpernderPreis{},
		"lesen.ki.modelle", `{}`); !errors.Is(err, errGestolpertMCP) {
		t.Errorf("der Fehler ist %v", err)
	}
}

// **Der Vorschlag kommt an — samt Kennung, Feldern und Frist.**
func TestDasWerkzeugKIVorschlagGibtSeinenVorschlag(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"schreiben.ki.vorschlag",
		`{"karteId": 4, "alias": "gemma4:8b", "auftrag": "Schreibe."}`)
	if err != nil {
		t.Fatalf("schreiben.ki.vorschlag: %v", err)
	}

	for _, teil := range []string{
		`"kennung": "ki-4-1"`, `"modell": "gemma4:8b"`, `"feld": "name"`,
		`"ablauf": "2026-09-12T12:00:00Z"`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("die Antwort nennt %s nicht:\n%s", teil, antwort)
		}
	}

	// **Ohne »uebernehmen« wurde nichts geschrieben** — und das steht in der
	// Antwort und nicht zwischen den Zeilen.
	if !strings.Contains(antwort, `"uebernommen": 0`) {
		t.Errorf("die Antwort sagt nicht, dass nichts geschrieben wurde:\n%s", antwort)
	}
}

// **Mit »uebernehmen« kommt die Zahl der geschriebenen Felder.**
//
// **Die Gegenprobe zur Zeile darüber**: ohne sie sagte der Lauf nur, dass die
// Flagge nichts schadet — nicht, dass sie ankommt.
func TestDasWerkzeugKIVorschlagReichtUebernehmenDurch(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"schreiben.ki.vorschlag",
		`{"karteId": 4, "alias": "gemma4:8b", "auftrag": "x", "uebernehmen": true}`)
	if err != nil {
		t.Fatalf("schreiben.ki.vorschlag: %v", err)
	}

	if !strings.Contains(antwort, `"uebernommen": 1`) {
		t.Errorf("die Flagge kam nicht an:\n%s", antwort)
	}
}

// **Leere Listen bleiben Listen.**
func TestEinLeererKIVorschlagBleibtEineListe(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, nilprobe{},
		"schreiben.ki.vorschlag",
		`{"karteId": 4, "alias": "gemma4:8b", "auftrag": "x"}`)
	if err != nil {
		t.Fatalf("schreiben.ki.vorschlag: %v", err)
	}

	for _, teil := range []string{`"felder": []`, `"befunde": []`} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("die Antwort nennt %s nicht:\n%s", teil, antwort)
		}
	}
}

// **Ein Fehler der Werkstatt kommt als Fehler durch.**
func TestEinFehlerBeimKIVorschlagKommtDurch(t *testing.T) {
	t.Parallel()

	_, err := rufeSammlungswerkzeug(t, nilprobe{}, stolpernderPreis{},
		"schreiben.ki.vorschlag",
		`{"karteId": 4, "alias": "gemma4:8b", "auftrag": "x"}`)

	if !errors.Is(err, errGestolpertMCP) {
		t.Fatalf("der Fehler ist %v", err)
	}
}

// **Und ein unlesbares Argument ebenso.**
func TestEinUnlesbaresArgumentBeimKIVorschlagKommtDurch(t *testing.T) {
	t.Parallel()

	if _, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"schreiben.ki.vorschlag", `{"karteId": "vier"}`); err == nil {
		t.Fatal("das unlesbare Argument kam durch")
	}
}
