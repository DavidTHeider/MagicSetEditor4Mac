package service

// **Die Funktion `keyword_usage` in der Umgebung** —
// #1354.
//
// **Getrennt vom Vermerk daneben**, weil sie eine andere Frage beantwortet:
// dort geht es um das, was aufgezeichnet wird, hier um das, was ein Skript
// davon sieht.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// frage baut die Umgebung mit einem gefüllten Vermerk.
func frage(t *testing.T) *msescript.Umgebung {
	t.Helper()

	vermerk := &Schluesselwortverwendung{}
	vermerk.Vermerke("rule text", "Flying")
	vermerk.Vermerke("rule text", "Flying")
	vermerk.Vermerke("rule text 2", "Trample")

	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("keyword_usage", schluesselwortverwendungsfrage(vermerk))

	return umgebung
}

// TestKeywordUsageAntwortetImSkript ist die Grundzusage.
func TestKeywordUsageAntwortetImSkript(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		quelle   string
		erwartet string
	}{
		// **Ohne `unique:` steht jede Verwendung da** — MSEs Vorgabe ist
		// `false` (`SCRIPT_OPTIONAL_PARAM_(bool, unique)`).
		{"keyword_usage()", "Flying, Flying, Trample"},
		{"keyword_usage(unique: true)", "Flying, Trample"},
		{"keyword_usage(unique: false)", "Flying, Flying, Trample"},

		// **Und was sich nicht als Wahrheit lesen lässt, zählt nicht.**
		{`keyword_usage(unique: "vielleicht")`, "Flying, Flying, Trample"},
	} {
		t.Run(fall.quelle, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, frage(t))
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if gefunden := wert.AlsText(); gefunden != fall.erwartet {
				t.Errorf("%q, erwartet %q", gefunden, fall.erwartet)
			}
		})
	}
}
