package domain_test

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Manasymbole zerlegt einen Kosten in seine Symbole.**
//
// # Warum das neben [domain.ManakostenOhneKlammern] steht
//
// Die beantwortet eine andere Frage: *wie schreibt MSE diesen Kosten?* Diese
// hier fragt: *aus welchen Symbolen besteht er?* — und die Antwort behält die
// Klammern, denn so führt Scryfall sie in der Symbologie.
func TestManasymbole(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		kosten   string
		erwartet []string
	}{
		"schlicht":        {"{2}{R}", []string{"{2}", "{R}"}},
		"einfarbig":       {"{W}", []string{"{W}"}},
		"hybrid":          {"{2/U}", []string{"{2/U}"}},
		"phyrexianisch":   {"{U/P}", []string{"{U/P}"}},
		"leer":            {"", nil},
		"nur Leerzeichen": {"   ", nil},

		// **Ein Kosten ohne Klammern ist kein Kosten**, sondern die Form, in
		// der MSE ihn ablegt. Wer ihn hier hineingäbe, bekäme nichts — und das
		// ist richtiger, als jedes Zeichen einzeln als Symbol auszugeben.
		"ohne Klammern": {"2R", nil},

		// **Und ein halb geschriebener wird nicht geraten.**
		"offene Klammer": {"{2}{R", []string{"{2}"}},

		// **Nicht-ASCII bleibt** — `{∞}` ist ein echtes Symbol.
		"unendlich": {"{∞}", []string{"{∞}"}},

		"mit Leerzeichen dazwischen": {"{2} {R}", []string{"{2}", "{R}"}},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			bekommen := domain.Manasymbole(fall.kosten)
			if !slices.Equal(bekommen, fall.erwartet) {
				t.Errorf("Manasymbole(%q) = %v, erwartet %v",
					fall.kosten, bekommen, fall.erwartet)
			}
		})
	}
}
