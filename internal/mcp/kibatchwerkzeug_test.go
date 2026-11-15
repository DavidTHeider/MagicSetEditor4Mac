package mcp_test

// **Das Werkzeug des Batches** — Werkstück A13.1 von Phase 5.3
// (#1452).

import (
	"errors"
	"strings"
	"testing"
)

// **Der Batch kommt an — je Ziel eine Lage.**
func TestDasWerkzeugKIBatchGibtJedeLage(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"schreiben.ki.batch",
		`{"karteIds": [4, 7], "alias": "gemma4:8b", "auftrag": "Schreibe."}`)
	if err != nil {
		t.Fatalf("schreiben.ki.batch: %v", err)
	}

	for _, teil := range []string{
		`"kennung": "kibatch-1"`, `"karteId": 4`, `"karteId": 7`,
		`"modell": "gemma4:8b"`, `"feld": "name"`,

		// **Auch das gescheiterte Ziel steht da** — mit leeren Listen und
		// nicht mit `null`: für ein Sprachmodell ist das der Unterschied
		// zwischen „nichts gefunden" und „nicht beantwortet".
		`"fehler": "die Karte gibt es nicht"`, `"felder": []`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("die Antwort nennt %s nicht:\n%s", teil, antwort)
		}
	}
}

// **Leere Listen bleiben Listen.**
//
// **Für ein Sprachmodell sieht `null` aus wie „konnte nicht beantwortet
// werden"** — und ein Batch ohne Ziele ist etwas anderes als ein Batch, der
// scheiterte.
func TestEinLeererBatchGibtListenUndKeinNull(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, nilprobe{},
		"schreiben.ki.batch",
		`{"karteIds": [4], "alias": "gemma4:8b", "auftrag": "Schreibe."}`)
	if err != nil {
		t.Fatalf("schreiben.ki.batch: %v", err)
	}

	if strings.Contains(antwort, "null") {
		t.Errorf("die Antwort trägt ein null:\n%s", antwort)
	}

	if !strings.Contains(antwort, `"ziele": []`) {
		t.Errorf("die Antwort nennt keine leere Zielliste:\n%s", antwort)
	}
}

// **Und ein Fehler der Naht kommt als Fehler an.**
func TestEinStolpernderBatchGibtDenFehlerWeiter(t *testing.T) {
	t.Parallel()

	if _, err := rufeSammlungswerkzeug(t, nilprobe{}, stolpernderPreis{},
		"schreiben.ki.batch",
		`{"karteIds": [4], "alias": "gemma4:8b", "auftrag": "Schreibe."}`,
	); !errors.Is(err, errGestolpertMCP) {
		t.Errorf("der Fehler ist %v", err)
	}
}

// **Ein unlesbarer Rumpf kommt als Fehler zurück.**
func TestEinUnlesbarerBatchrumpfWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"schreiben.ki.batch", `{"karteIds": "keine Liste"}`); err == nil {
		t.Fatal("der unlesbare Rumpf kam durch")
	}
}
