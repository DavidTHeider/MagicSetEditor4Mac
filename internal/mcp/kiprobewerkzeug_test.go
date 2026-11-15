package mcp_test

// **Die zwei Werkzeuge der Messung** — Werkstück A9 von Phase 5.3
// (#1448).
//
// **Eine eigene Datei**, weil `kiwerkzeug_test.go` mit ihr über die
// Zeilengrenze aus §2 ginge.

import (
	"errors"
	"strings"
	"testing"
)

// **Eine Messung kommt mit ihren Belegen zurück.**
func TestEineMessungKommtMitIhrenBelegen(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"schreiben.ki.modell.pruefen",
		`{"anbieter": "openai", "alias": "gpt-probe"}`)
	if err != nil {
		t.Fatalf("schreiben.ki.modell.pruefen: %v", err)
	}

	if !strings.Contains(antwort, "gpt-probe") ||
		!strings.Contains(antwort, "strukturiert") {
		t.Errorf("die Antwort ist %s", antwort)
	}
}

// **Eine leere Liste bleibt eine Liste.**
//
// Ein `null` sähe für ein Sprachmodell aus wie „konnte nicht beantwortet
// werden" — und das ist etwas anderes als „nichts gemessen".
func TestEineLeereMessungBleibtEineListe(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, nilprobe{},
		"schreiben.ki.modell.pruefen",
		`{"anbieter": "openai", "alias": "gpt-probe"}`)
	if err != nil {
		t.Fatalf("schreiben.ki.modell.pruefen: %v", err)
	}

	if !strings.Contains(antwort, `"proben": []`) {
		t.Errorf("die Antwort ist %s", antwort)
	}
}

// **Eine vergessene Messung sagt, dass sie fort ist.**
func TestEineVergesseneMessungSagtEs(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"schreiben.ki.modell.vergessen",
		`{"anbieter": "openai", "alias": "gpt-probe"}`)
	if err != nil {
		t.Fatalf("schreiben.ki.modell.vergessen: %v", err)
	}

	if !strings.Contains(antwort, `"vergessen": true`) {
		t.Errorf("die Antwort ist %s", antwort)
	}
}

// **Und ein Fehler der Naht kommt durch — an beiden Werkzeugen.**
func TestEinFehlerDerMessungKommtDurch(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range []string{
		"schreiben.ki.modell.pruefen", "schreiben.ki.modell.vergessen",
	} {
		t.Run(werkzeug, func(t *testing.T) {
			t.Parallel()

			if _, err := rufeSammlungswerkzeug(t, nilprobe{}, stolpernderPreis{},
				werkzeug, `{"anbieter": "openai", "alias": "gpt-probe"}`,
			); !errors.Is(err, errGestolpertMCP) {
				t.Errorf("der Fehler ist %v", err)
			}
		})
	}
}

// **Ein unlesbarer Rumpf wird abgewiesen — an beiden.**
func TestEinUnlesbarerRumpfBeiDerMessungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range []string{
		"schreiben.ki.modell.pruefen", "schreiben.ki.modell.vergessen",
	} {
		t.Run(werkzeug, func(t *testing.T) {
			t.Parallel()

			if _, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
				werkzeug, `{kaputt`); err == nil {
				t.Error("der unlesbare Rumpf kam durch")
			}
		})
	}
}
