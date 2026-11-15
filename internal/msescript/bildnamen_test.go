package msescript_test

// **Das Ausgangsbild einer Bildverknüpfung heißt `input`**
// (#483).
//
//	SCRIPT_FUNCTION(set_combine) {
//	  SCRIPT_PARAM(String, combine);
//	  SCRIPT_PARAM_C(GeneratedImageP, input);
//
// `src/script/functions/image.cpp`, Zeilen 69–75, Fassung f9d9356 vom
// 2021-07-06. `SCRIPT_PARAM_C` heißt: benannt **oder** stellungsweise.
//
// **Nur `set_mask` heißt wirklich `image`**; die Mischungen heißen
// `light`/`dark`/`mask` beziehungsweise `image1`/`image2`.
//
// Über `M15-Magic-Pack` schreiben **643** Aufrufe `input:` und **null**
// `image:`.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestEineBildverknuepfungNimmtInput(t *testing.T) {
	faelle := []struct {
		name, quelle string
		knotenart    string
		kind         string
	}{
		{
			name:      "set_combine benannt",
			quelle:    `set_combine(input: "a.png", combine: "multiply")`,
			knotenart: "set_combine", kind: "input",
		},
		{
			name:      "set_combine stellungsweise",
			quelle:    `set_combine("a.png", combine: "multiply")`,
			knotenart: "set_combine", kind: "input",
		},
		{
			name:      "recolor_image benannt",
			quelle:    `recolor_image(input: "a.png", color: rgb(1,2,3))`,
			knotenart: "recolor_image", kind: "input",
		},
		{
			name:      "saturate benannt",
			quelle:    `saturate(input: "a.png", amount: 0.5)`,
			knotenart: "saturate", kind: "input",
		},
		{
			name:      "invert_image stellungsweise",
			quelle:    `invert_image("a.png")`,
			knotenart: "invert_image", kind: "input",
		},
		{
			// **Die Ausnahme, und sie ist nachgelesen**: `set_mask` nennt sein
			// Ausgangsbild wirklich `image`.
			name:      "set_mask heißt image",
			quelle:    `set_mask(image: "a.png", mask: "m.png")`,
			knotenart: "set_mask", kind: "image",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(
				t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("ausrechnen: %v", err)
			}

			baum, istBild := wert.AlsBild()
			if !istBild {
				t.Fatalf("kein Bildwert, sondern %v", wert.Art)
			}

			if baum.Knotenart != fall.knotenart {
				t.Fatalf("Knotenart %q, erwartet %q", baum.Knotenart, fall.knotenart)
			}

			var gefunden bool
			for _, kind := range baum.Kinder {
				if kind.Name == fall.kind {
					gefunden = true
				}
			}

			if !gefunden {
				var namen []string
				for _, kind := range baum.Kinder {
					namen = append(namen, kind.Name)
				}

				t.Errorf("kein Kind %q — der Baum führt %v", fall.kind, namen)
			}
		})
	}
}
