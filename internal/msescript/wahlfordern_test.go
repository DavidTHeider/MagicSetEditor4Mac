package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`require_choice` verlangt mindestens eine der genannten Wahlen.**
//
//	SCRIPT_FUNCTION(require_choice) {
//	  SCRIPT_PARAM_C(String,input);
//	  SCRIPT_OPTIONAL_PARAM_(String, last_change);
//	  vector<String> choices;
//	  read_choices_param(ctx, choices);
//	  SCRIPT_RETURN(filter_choices(input, choices, 1, (int)choices.size(), last_change));
//	}
//
// `src/script/functions/editor.cpp`, Zeilen 331–337, Fassung f9d9356 vom
// 2021-07-06. **`min` ist eins, `max` die Zahl der Wahlen** — es dürfen alle
// stehen bleiben, aber nicht keine; das ist der Unterschied zu
// `exclusive_choice` (0…1) und `require_exclusive_choice` (1…1).
//
// # Was das gekostet hat
//
// `magic.mse-game/script` ruft es an der Kartenfarbe:
//
//	input := require_choice(choices: "multicolor, hybrid, land, artifact")
//
// Auf der selbst geschriebenen Demokarte *Zwietracht und Zwiespalt*
// (`black, red, multicolor`) fielen damit **zwei** Wertskripte aus — `card_color`
// und `indicator`
// (#270).
func TestRequireChoiceVerlangtEineWahl(t *testing.T) {
	faelle := []struct {
		name, quelle, erwartet string
	}{
		{
			name:     "eine genannte Wahl bleibt stehen",
			quelle:   `require_choice("red, multicolor", choices: "multicolor, hybrid, land")`,
			erwartet: "red, multicolor",
		},
		{
			// **Alle dürfen bleiben** — `max` ist die Zahl der Wahlen.
			name:     "mehrere genannte Wahlen bleiben",
			quelle:   `require_choice("multicolor, land", choices: "multicolor, hybrid, land")`,
			erwartet: "multicolor, land",
		},
		{
			// **Und ohne eine wird die erste ergänzt** — `min` ist eins.
			name:     "ohne Wahl kommt die erste dazu",
			quelle:   `require_choice("red", choices: "multicolor, hybrid, land")`,
			erwartet: "red, multicolor",
		},
	}

	// **Und die zwei Wege daneben.** `last_change` bevorzugt eine Wahl, und ein
	// Ruf ohne Eingabe wird abgewiesen statt geraten.
	t.Run("last_change bevorzugt eine Wahl", func(t *testing.T) {
		wert, err := msescript.AuswertenQuelle(t.Context(),
			`require_choice("red", choices: "multicolor, hybrid, land", last_change: "land")`,
			msescript.Bibliothek().Darueber())
		if err != nil {
			t.Fatalf("auswerten: %v", err)
		}

		if wert.AlsText() != "red, land" {
			t.Errorf("ergibt %q, erwartet \"red, land\"", wert.AlsText())
		}
	})

	t.Run("ohne Eingabe abgewiesen", func(t *testing.T) {
		if _, err := msescript.AuswertenQuelle(t.Context(),
			`require_choice(choices: "multicolor")`,
			msescript.Bibliothek().Darueber()); err == nil {
			t.Error("ein Ruf ohne Eingabe wurde angenommen")
		}
	})

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(
				t.Context(), fall.quelle, msescript.Bibliothek().Darueber())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}
