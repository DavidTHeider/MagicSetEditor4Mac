package storage

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **`feldwert` liest, was `mitFeld` schreibt** — jedes Feld eines
// Schlüsselworts, sonst prüfte die Bedingung aus U02b (#2179) ein anderes
// Feld als das, das geschrieben wird.
func TestFeldwertLiestWasMitFeldSchreibt(t *testing.T) {
	t.Parallel()

	for _, feld := range []string{"wort", "muster", "erinnerung", "regeln", "modus"} {
		stand := mitFeld(domain.Schluesselwortstand{}, feld, "Wert von "+feld)
		if gelesen := feldwert(stand, feld); gelesen != "Wert von "+feld {
			t.Errorf("%s: gelesen %q", feld, gelesen)
		}
	}
}
