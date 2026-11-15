package msescript_test

// **Die zwei Zweitnamen aus Phase 4.9** —
// #1324.
//
// Am Orakel gemessen (Fassung f9d9356 vom 2021-07-06, am 2026-09-02):
//
//	match_text(match: "a+", input: "xaay")  → true
//	sanitize("<b>x</b>")                    → "<b>x</b>"

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestZweitnamen(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, quelle, erwartet string }{
		// **`match_text` ist `match`** — dieselbe Funktion, und zwar so
		// herum: `match` ist bei MSE2 der **alte** Name.
		{"match_text trifft", `match_text(match: "a+", input: "xaay")`, "true"},
		{"match_text trifft nicht", `match_text(match: "z+", input: "xaay")`, "false"},
		{"match_text stellungsweise", `match_text("xaay", match: "a+")`, "true"},
		{"match gibt dasselbe", `match(match: "a+", input: "xaay")`, "true"},

		// **`sanitize` tut nichts** — im Original steht `//TODO` und
		// `SCRIPT_RETURN(input)`. Der Name ist hier, damit eine Vorlage, die
		// ihn ruft, durchläuft; eine Absicherung ist er nicht.
		{"sanitize lässt Auszeichnung stehen", `sanitize("<b>x</b>")`, "<b>x</b>"},
		{"sanitize lässt alles stehen", `sanitize("a & b < c")`, "a & b < c"},
		{"sanitize einer leeren Zeichenkette", `sanitize("")`, ""},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Und der Fehlerweg** (§3).
func TestSanitizeOhneEingabe(t *testing.T) {
	t.Parallel()

	_, err := msescript.AuswertenQuelle(t.Context(), `sanitize()`,
		msescript.Bibliothek())
	if err == nil {
		t.Fatal("der Aufruf glückte, erwartet ein Fehler")
	}

	if !strings.Contains(err.Error(), "sanitize") {
		t.Errorf("die Meldung %q nennt sanitize nicht", err)
	}
}
