// **Was ein Ausdruck an dieser Stelle nennen darf** — Werkstück 1 von
// Phase 3.3 (#1164).
package msescript_test

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Alle Ebenen, sortiert, ohne Doppel.**
func TestAlleNamenGehenUeberJedeEbene(t *testing.T) {
	t.Parallel()

	// **Zwei Ebenen**: die untere ist die, in der ein Ausdruck läuft, die
	// obere die Bibliothek darunter in der Kette.
	oben := msescript.NeueUmgebung(map[string]msescript.Wert{
		"length": msescript.Text("die Bibliothek"),
		"set":    msescript.Text("ein Set"),
	})

	unten := oben.Darueber()
	unten.Setze("card", msescript.Text("eine Karte"))
	unten.Setze("length", msescript.Text("überschrieben"))

	namen := unten.AlleNamen()

	for _, erwartet := range []string{"card", "length", "set"} {
		if !slices.Contains(namen, erwartet) {
			t.Errorf("%q fehlt in %v", erwartet, namen)
		}
	}

	// **Ohne Doppel**: `length` steht auf beiden Ebenen und kommt einmal vor.
	var wieOft int

	for _, name := range namen {
		if name == "length" {
			wieOft++
		}
	}

	if wieOft != 1 {
		t.Errorf("%q kommt %dmal vor", "length", wieOft)
	}

	if !slices.IsSorted(namen) {
		t.Errorf("die Namen sind nicht sortiert: %v", namen)
	}
}

// **Die Bibliothek ist darin enthalten.**
//
// Der Lauf, der die Zahl belegt: 90 Namen führt sie heute, und wer
// vervollständigt, soll sie alle sehen.
func TestAlleNamenEnthaltenDieBibliothek(t *testing.T) {
	t.Parallel()

	namen := msescript.Bibliothek().Darueber().AlleNamen()

	// **`if` steht nicht dabei, und das ist richtig**: es ist ein Wort des
	// Zerteilers und kein Name der Umgebung. Wer es in die Vervollständigung
	// nähme, nähme es von hier — und dann stünde dort, die Umgebung kenne
	// etwas, das sie nicht kennt (§10).
	for _, erwartet := range []string{"length", "to_upper", "sort_text", "true"} {
		if !slices.Contains(namen, erwartet) {
			t.Errorf("%q fehlt", erwartet)
		}
	}

	if len(namen) < 80 {
		t.Errorf("es sind nur %d Namen — die Bibliothek führt rund 90", len(namen))
	}
}
