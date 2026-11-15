package msescript

// **Die Fehlerwege von `in_context`** (#2315): ein Kontext, der sich nicht
// übersetzen lässt, und ein Lauf, der abbricht — im Kontext wie im Muster.
// Dieselbe Gestalt wie `TestJedeMusterfunktionReichtDenAbbruchDurch`.

import (
	"strings"
	"testing"
	"time"
)

var kontextfunktionen = map[string]string{
	"replace":           `replace(eingabe, match: muster, in_context: kontext, replace: "X")`,
	"replace mit Rumpf": `replace(eingabe, match: muster, in_context: kontext, replace: { "X" })`,
	"filter_text":       `filter_text(eingabe, match: muster, in_context: kontext)`,
	"break_text":        `break_text(eingabe, match: muster, in_context: kontext)`,
}

func werteMitKontextAus(t *testing.T, quelle, muster, kontext, eingabe string) error {
	t.Helper()

	umgebung := Bibliothek().Darueber()
	umgebung.Setze("muster", Text(muster))
	umgebung.Setze("kontext", Text(kontext))
	umgebung.Setze("eingabe", Text(eingabe))

	_, err := AuswertenQuelle(t.Context(), quelle, umgebung)

	return err
}

func TestEinUngueltigerKontextWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, quelle := range kontextfunktionen {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := werteMitKontextAus(t, quelle, "a", "(", "abc")
			if err == nil || !strings.Contains(err.Error(), "in_context") {
				t.Errorf("ein kaputter Kontext ging durch: %v", err)
			}
		})
	}
}

// **Ein Abbruch im Kontext** — beim ersten von zwei Treffern; der zweite wird
// dann gar nicht mehr gefragt. **Und einer im Muster selbst**, mit einem
// harmlosen Kontext.
func TestEinAbbruchImKontextWirdGemeldet(t *testing.T) {
	vorher := musterFrist
	musterFrist = time.Millisecond

	t.Cleanup(func() { musterFrist = vorher })

	verrennung := `(?=a)(a+)+b`
	lang := strings.Repeat("a", 40)

	for name, quelle := range kontextfunktionen {
		t.Run(name+" im Kontext", func(t *testing.T) {
			if err := werteMitKontextAus(t, quelle, "c", verrennung, lang+"c c"); err == nil {
				t.Error("der Abbruch im Kontext ging still durch")
			}
		})

		t.Run(name+" im Muster", func(t *testing.T) {
			if err := werteMitKontextAus(t, quelle, verrennung, "x", lang+"c"); err == nil {
				t.Error("der Abbruch im Muster ging still durch")
			}
		})
	}
}
