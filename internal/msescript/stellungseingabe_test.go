// **Ein Stellungsargument bindet `input` — und sonst nichts**
// (#555).
package msescript_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// TestEinStellungsargumentBindetNurInput hält MSEs Zerteiler fest.
//
// # Es gibt gar keine Stellungsargumente
//
//	void parseCallArguments(TokenIterator& input, Script& script,
//	                        vector<Variable>& arguments) {
//	  …
//	    } else {
//	      // implicit "input" argument
//	      arguments.push_back(SCRIPT_VAR_input);
//	      parseOper(input, script, PREC_SEQ);
//	    }
//
// `src/script/parser.cpp`, Zeilen 881–902, Fassung f9d9356 vom 2021-07-06.
// **Jedes unbenannte Argument wird zu `input:`** — auch das dritte. Was hier
// „Stellung" heißt, ist eine Bequemlichkeit dieses Programms und keine Form
// der Sprache.
//
// Und was eine Bildfunktion liest, liest sie **beim Namen**:
//
//	SCRIPT_FUNCTION(linear_blend) {
//	  SCRIPT_PARAM(GeneratedImageP, image1);
//	  SCRIPT_PARAM(GeneratedImageP, image2);
//
//	#define SCRIPT_PARAM_N(Type, str, name) \
//	    Type name = from_script<Type>(ctx.getVariable(str), str)
//
// `src/script/functions/image.cpp`, Zeilen 33–39, und
// `src/script/functions/util.hpp`, Zeilen 98–101.
//
// # Woran es aufgefallen ist
//
// `magic-m15-split-fusable.mse-style` schreibt in `bridge_colored`
//
//	image1 := mask_bridge(color: color1)
//	image2 := mask_bridge(color: color2)
//
//	linear_blend(x1:0, y1:0, x2:1, y2:0)
//
// — beide Bilder aus dem umgebenden Bereich, kein Argument. Dieses Programm
// schob dort das umgebende `input` als erstes Stellungsargument nach und nahm
// es als `image1`; auf `Alive` war das die Zeichenkette `"true"`, und die
// Brücke der linken Kartenhälfte fehlte. Über die Stichprobe des Kanonlaufs
// **zehnmal**.
//
// # Was das Pack dazu sagt
//
// Gezählt über `style`, `script` und die Blendmodule: `set_mask` 27-mal
// benannt und **null**-mal unbenannt, `combine_blend` 63/0, `linear_blend`
// 535 benannt. Einzig `recolor_image` steht durchweg unbenannt — 64-mal —,
// und **sein** Parameter heißt bei MSE `input`
// (`SCRIPT_PARAM_C(GeneratedImageP, input)`).
func TestEinStellungsargumentBindetNurInput(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name, quelle, erwartet string
	}{
		{
			// Der Fall aus dem Pack: `image1` und `image2` stehen im
			// umgebenden Bereich, `input` daneben.
			name: "linear_blend nimmt image1 aus der Umgebung, nicht input",
			quelle: `f := {
				image1 := "links.png"
				image2 := "rechts.png"
				linear_blend(x1: 0, y1: 0, x2: 1, y2: 0)
			}
			f("true")`,
			erwartet: "links.png",
		},
		{
			name:     "recolor_image nimmt das Stellungsargument als input",
			quelle:   `recolor_image("bild.png", red: rgb(255,0,0))`,
			erwartet: "bild.png",
		},
		{
			name:     "und input: schreibt dasselbe",
			quelle:   `recolor_image(input: "bild.png", red: rgb(255,0,0))`,
			erwartet: "bild.png",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("%v", err)
			}

			baum, istBild := wert.AlsBild()
			if !istBild {
				t.Fatalf("ergab %s %q, erwartet ein Bild", wert.Art, wert.AlsText())
			}

			if !strings.Contains(baumtext(baum), fall.erwartet) {
				t.Errorf("Baum %s enthält %q nicht", baumtext(baum), fall.erwartet)
			}
		})
	}
}

// TestEineBildangabeOhneBildIstEinFehler prüft die Gegenprobe.
//
// **Ohne diesen Fall wäre die Änderung oben nicht belegt**: ein `image1`, das
// weder im Aufruf noch in der Umgebung steht, muss ein Fehler sein und darf
// nicht still das umgebende `input` nehmen.
func TestEineBildangabeOhneBildIstEinFehler(t *testing.T) {
	t.Parallel()

	_, err := msescript.AuswertenQuelle(t.Context(),
		`f := { linear_blend(x1: 0, y1: 0, x2: 1, y2: 0) }
		f("true")`, msescript.Bibliothek())
	if err == nil {
		t.Fatal("kein Fehler — das umgebende input wurde als image1 genommen")
	}

	if !strings.Contains(err.Error(), "image1") {
		t.Errorf("Fehler %q nennt image1 nicht", err.Error())
	}
}

// baumtext macht einen Bildbaum flach lesbar.
func baumtext(baum msescript.Bildbaum) string {
	teile := make([]string, 0, 2+len(baum.Kinder))
	teile = append(teile, baum.Knotenart, baum.Pfad)

	for _, kind := range baum.Kinder {
		teile = append(teile, kind.Name+"="+baumtext(kind.Baum))
	}

	return strings.Join(teile, " ")
}
