package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Ein Textskript ist eine Zeichenkette mit eingebetteten Ausdrücken.**
//
// MSE nennt das `StringScript`; ein `reminder:` eines Schlüsselworts ist einer
// (`src/data/keyword.hpp`, Zeile 95, Fassung f9d9356 vom 2021-07-06).
func TestEinTextskriptSetztSeineAusdruckeEin(t *testing.T) {
	t.Parallel()

	umgebung := msescript.Bibliothek()
	umgebung.Setze("param1", msescript.Text("Rot"))

	for _, fall := range []struct{ quelle, erwartet string }{
		{"schützt vor {param1}", "schützt vor Rot"},
		{"ganz ohne Ausdruck", "ganz ohne Ausdruck"},
		{"", ""},
		// **Ein Anführungszeichen ist ein Zeichen**, kein Ende der
		// Zeichenkette — der Grund, aus dem es diese Funktion gibt.
		{`sagt "hallo" zu {param1}`, `sagt "hallo" zu Rot`},
	} {
		wert, fehler := msescript.AuswertenTextskript(t.Context(), fall.quelle, umgebung)
		if fehler != nil {
			t.Fatalf("%q: %v", fall.quelle, fehler)
		}

		if wert.AlsText() != fall.erwartet {
			t.Errorf("%q gab %q, erwartet %q", fall.quelle, wert.AlsText(), fall.erwartet)
		}
	}
}
