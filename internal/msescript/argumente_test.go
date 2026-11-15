package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Benannte und Stellungsargumente stehen nebeneinander.
//
// Die Feststellung in `docs/FORMATE.md` nennt beide Formen; der zuvor genannte
// Testfall prüfte nur die benannte. Gefunden vom Nachlauf des Zweitblicks über
// Phase 3.
func TestBenannteUndStellungsargumenteNebeneinander(t *testing.T) {
	umgebung := msescript.Bibliothek()

	// `rgb(156, 130, 90)` — drei Stellungsargumente.
	wert, err := msescript.AuswertenQuelle(t.Context(), `rgb(156, 130, 90)`, umgebung)
	if err != nil || wert.String() != "rgba(156,130,90,255)" {
		t.Errorf("Stellung: %v / %v", wert, err)
	}

	// `contains(style, match:"pw")` — eines von jeder Sorte, in beiden Folgen.
	for _, quelle := range []string{
		`contains("m15-pw", match:"pw")`,
		`contains(match:"pw", "m15-pw")`,
	} {
		wert, err := msescript.AuswertenQuelle(t.Context(), quelle, umgebung)
		if err != nil || wert.String() != "true" {
			t.Errorf("%s: %v / %v", quelle, wert, err)
		}
	}
}
