// **Ein Textskript wird streng geprüft** —
// #1934.
//
// [msescript.AuswertenTextskript] lässt einen Ausdruck, der sich nicht lesen
// lässt, als Text stehen — richtig für einen Erinnerungstext, der ein Muster
// wie `\x{20AC}` trägt. Ein Bildwert wie `border-mask{pt()}.png` braucht die
// andere Antwort: MSE2 liest ihn im Zeichenkettenmodus (`parse(*this, true)`,
// `src/script/image.cpp` Zeilen 67–78, Fassung f9d9356 vom 2021-07-06), und
// ein kaputter Ausdruck darin ist ein Fehler an seiner Stelle.
package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Die Stelle ist die in der Quelle**: der Zerteiler meldet die offene `(`
// in `f(1,` an Spalte 2 seines Ausdrucks, und der beginnt hinter der `{`.
func TestEinTextskriptWirdStrengGeprueft(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		quelle        string
		zeile, spalte int
		schluessel    msescript.Meldungsschluessel
	}{
		"ohne Ausdruck":                   {quelle: "rahmen.png"},
		"ein Ausdruck in der Mitte":       {quelle: "border-mask{pt()}.png"},
		"eine geschützte Klammer":         {quelle: `a\{b.png`},
		"Anführungszeichen im Text":       {quelle: `sagt "{card.name}"`},
		"eine Musterflucht":               {quelle: `\x{20AC}`},
		"ein kaputter Ausdruck":           {quelle: "a{f(1,}.png", zeile: 1, spalte: 4},
		"eine offene Klammer":             {quelle: "rahmen{pt(.png", zeile: 1, spalte: 7, schluessel: msescript.SchluesselKlammerOffen},
		"eine offene Klammer vorn":        {quelle: "{ if card.farbe", zeile: 1, spalte: 1, schluessel: msescript.SchluesselKlammerOffen},
		"ein Fehler in Zeile zwei":        {quelle: "a.png\nb{f(}", zeile: 2, spalte: 4},
		"ein mehrzeiliger Ausdruck":       {quelle: "a{\n f(1,\n}", zeile: 2, spalte: 3},
		"der zweite Ausdruck ist kaputt":  {quelle: "{1}x{f(}", zeile: 1, spalte: 7},
		"nach einem Ausdruck über Zeilen": {quelle: "a{\n1\n}b{f(}", zeile: 3, spalte: 5},
	} {
		err := msescript.TextskriptFehler(fall.quelle)

		if fall.zeile == 0 {
			if err != nil {
				t.Errorf("%s: %v, erwartet kein Fehler", name, err)
			}

			continue
		}

		befund := msescript.Befund(err)
		if err == nil || befund.Zeile != fall.zeile || befund.Spalte != fall.spalte {
			t.Errorf("%s: %v, erwartet Zeile %d, Spalte %d", name, err, fall.zeile, fall.spalte)
		}

		if fall.schluessel != "" && befund.Schluessel != fall.schluessel {
			t.Errorf("%s: Schlüssel %q, erwartet %q", name, befund.Schluessel, fall.schluessel)
		}
	}
}
