package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Eine Zuweisung in einer Schleife gilt im nächsten Durchgang.**
//
// # Wie MSE seine Namen führt
//
// MSE hat **eine** Variablentabelle. `setVariable` schreibt in den Platz des
// Namens und legt nur dann eine Schattenkopie an, wenn der Name aus einer
// tieferen Ebene stammt; die Ebene wechselt allein beim **Aufruf**
// (`case I_CALL: new_scope.reset(new LocalScope(*this))`). Eine Schleife und
// eine geklammerte Gruppe öffnen **keine**.
//
// `Context::setVariable` und `Context::eval`, `src/script/context.cpp`, Zeilen
// 43–44, 168–170 und 268–280, Fassung f9d9356 vom 2021-07-06.
//
// # Was das gekostet hat
//
// `sub_type_filter` in `magic.mse-game` schält in einer Zählschleife Wort für
// Wort von `input` ab — `input := replace(input, …)`. Schrieb jeder Durchgang
// in seine eigene Ebene, sah der nächste den alten Text: aus `Goblin Mage`
// wurde `Goblin Goblin Goblin Goblin`
// (#345).
func TestEineZuweisungInEinerSchleifeGiltImNaechstenDurchgang(t *testing.T) {
	faelle := []struct{ name, quelle, erwartet string }{
		{
			name: "die Zählschleife",
			quelle: `text := "abc"
				for i from 0 to 2 do ( erster := text.0 ; text := replace(text, match: "^.", replace: "") ; erster )`,
			erwartet: "abc",
		},
		{
			name: "die Listenschleife",
			quelle: `zahl := 0
				for each x in [1,2,3] do ( zahl := zahl + x ; to_string(zahl) )`,
			erwartet: "136",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("ausrechnen: %v", err)
			}

			// **Der Text der Schleife, nicht ihre Glieder.** Seit #390
			// addiert eine Schleife ihre Ergebnisse — aus drei Texten wird
			// einer, und `Eintraege` ist leer. Geprüft ist hier ohnehin die
			// Zuweisung über die Durchgänge hinweg, und die steht im Text.
			if gebaut := wert.AlsText(); gebaut != fall.erwartet {
				t.Errorf("ergibt %q, erwartet %q", gebaut, fall.erwartet)
			}
		})
	}
}
