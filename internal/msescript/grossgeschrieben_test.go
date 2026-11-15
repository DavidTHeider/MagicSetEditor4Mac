package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Jedes Wort groß, wie MSE2 es schreibt** — `capitalize_in_place`,
// `src/util/string.cpp`, Zeilen 95–113, Fassung f9d9356 vom 2021-07-06. Die
// Fälle, auf die sich die Namen von Paketen und die Symbolmenüs verlassen
// (UX-05, #1677).
func TestGrossgeschriebenFolgtMSE(t *testing.T) {
	faelle := []struct {
		name, ein, aus string
	}{
		{"ein Wort", "normal", "Normal"},
		{"nach Leerzeichen und Schrägstrich", "m15 fan/made", "M15 Fan/Made"},
		{"die kleinen Wörter mitten im Namen", "magic of the gathering", "Magic of the Gathering"},
		{"alle sieben", "x is the in of to at a y", "X is the in of to at a Y"},
		{"ein kleines Wort am Anfang", "the nexus", "The Nexus"},
		{"ein kleines Wort am Ende wird groß", "set of", "Set Of"},
		{"nach einem Tabulator nicht", "eins\tzwei", "Eins\tzwei"},
		{"ein Umlaut", "übersicht", "Übersicht"},
		{"was schon groß ist, bleibt", "M15 Normal", "M15 Normal"},
		{"leer", "", ""},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if got := msescript.Grossgeschrieben(fall.ein); got != fall.aus {
				t.Errorf("Grossgeschrieben(%q) = %q, erwartet %q", fall.ein, got, fall.aus)
			}
		})
	}
}
