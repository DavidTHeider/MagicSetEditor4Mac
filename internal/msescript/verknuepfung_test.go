package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`+` fällt auf die Aneinandersetzung zurück — für jede Kombination.**
//
//	case I_ADD: // add is quite overloaded
//	  if (at == SCRIPT_NIL) { a = b; }
//	  else if (bt == SCRIPT_NIL) { }
//	  else if (at == SCRIPT_FUNCTION && bt == SCRIPT_FUNCTION) { … compose … }
//	  else if (at == SCRIPT_COLLECTION && bt == SCRIPT_COLLECTION) { … concat … }
//	  else if (at == SCRIPT_INT && bt == SCRIPT_INT) { … }
//	  else if (… numbers …) { … }
//	  else { a = to_script(a->toString() + b->toString()); }
//
// `src/script/context.cpp`, Zeilen 496–513, Fassung f9d9356 vom 2021-07-06.
//
// **Der letzte Zweig fängt alles.** Eine Abbildung mit einer Zahl, ein
// Wahrheitswert mit einem Text — MSE setzt die Zeichenketten aneinander. Hier
// stand für zwei dieser Fälle ein Fehler, und die Begründung war „nirgends
// belegt": **47 Befunde** über die 190 Stile von `M15-Magic-Pack`
// (#270).
func TestPlusFaelltAufDieAneinandersetzungZurueck(t *testing.T) {
	faelle := []struct{ name, quelle, erwartet string }{
		{"eine Abbildung und eine Zahl", `to_string([1, 2] + 3)`, "123"},
		{"eine Zahl und eine Abbildung", `to_string(3 + [1, 2])`, "312"},
		{"ein Wahrheitswert und ein Text", `to_string(true + "x")`, "truex"},
		{"eine Abbildung und ein Wahrheitswert", `to_string([1] + false)`, "1false"},

		// **Die gebauten Fälle bleiben, wie sie sind.**
		{"zwei Abbildungen werden vereinigt", `to_string([a: 1] + [b: 2])`, "12"},
		{"zwei Zahlen werden addiert", `to_string(1 + 2)`, "3"},
		{"nichts links gibt rechts", `to_string(nil + "x")`, "x"},
		{"nichts rechts gibt links", `to_string("x" + nil)`, "x"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("ausrechnen: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}
