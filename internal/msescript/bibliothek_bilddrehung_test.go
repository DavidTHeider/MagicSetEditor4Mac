package msescript_test

// **`rotate`, `rotate_image` und `built_in_image`** — Werkstück A von
// Phase 4.9 (#1324).
//
// # Warum das kein Sonderfall ist
//
// `msescript` kennt keinen Bildtyp und darf `internal/render` nicht kennen
// (§2). Es trägt einen **Baum** und deutet ihn nicht — dieselbe Bauform, die
// `recolor_image` und die anderen fünfzehn Umformungen schon benutzen. Damit
// sind diese drei keine Entscheidung, sondern drei Zeilen.
//
//	SCRIPT_FUNCTION(rotate) {
//	  SCRIPT_PARAM_C(GeneratedImageP, input);
//	  SCRIPT_PARAM(Degrees, angle);
//	  return make_intrusive<RotateImage>(input,deg_to_rad(angle));
//	}
//
//	SCRIPT_FUNCTION(built_in_image) {
//	  SCRIPT_PARAM_C(String, input);
//	  return make_intrusive<BuiltInImage>(input);
//	}
//
// `src/script/functions/image.cpp`, Zeilen 126–130 und 204–207, Fassung
// f9d9356 vom 2021-07-06.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestDrehungBautEinenBildknoten(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, quelle, knotenart string }{
		{"rotate", `rotate(to_image("a.png"), angle: 90)`, "rotate"},

		// **`rotate_image` ist derselbe Zeiger** — im Original:
		//
		//	ctx.setVariable(_("rotate_image"), script_rotate);
		//
		// `image.cpp`, Zeile 227. Dieselbe Gestalt wie `match_text` neben
		// `match` und `replace_text` neben `replace`.
		{"rotate_image", `rotate_image(to_image("a.png"), angle: 90)`, "rotate"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			baum, istBild := wert.AlsBild()
			if !istBild {
				t.Fatalf("herauskam %s, erwartet ein Bild", wert.Art)
			}

			if baum.Knotenart != fall.knotenart {
				t.Errorf("Knotenart %q, erwartet %q",
					baum.Knotenart, fall.knotenart)
			}
		})
	}
}

// **`built_in_image` ist ein Erzeuger** — es macht aus einem Namen ein Bild.
//
// **Der Name reist als benannter Wert mit**, nicht als Kind: es gibt kein
// Eingangsbild. Dieselbe Stelle, an der bei `rotate` der Winkel steht — der
// Baum trennt Kinder von Werten, und ein Name ist ein Wert.
func TestEingebautesBildTraegtSeinenNamen(t *testing.T) {
	t.Parallel()

	wert, err := msescript.AuswertenQuelle(t.Context(),
		`built_in_image("card_shadow")`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	baum, istBild := wert.AlsBild()
	if !istBild {
		t.Fatalf("herauskam %s, erwartet ein Bild", wert.Art)
	}

	if baum.Knotenart != "built_in_image" {
		t.Errorf("Knotenart %q, erwartet %q", baum.Knotenart, "built_in_image")
	}

	if len(baum.Werte) != 1 || baum.Werte[0].Name != "name" {
		t.Fatalf("die Werte sind %v, erwartet einen namens »name«", baum.Werte)
	}

	if baum.Werte[0].Wert.AlsText() != "card_shadow" {
		t.Errorf("der Name ist %q, erwartet %q",
			baum.Werte[0].Wert.AlsText(), "card_shadow")
	}

	if len(baum.Kinder) != 0 {
		t.Errorf("%d Kinder, erwartet keine — es gibt kein Eingangsbild",
			len(baum.Kinder))
	}
}

// **Und die Fehlerwege** (§3).
func TestBilddrehungMeldetWasFehlt(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, quelle, nennt string }{
		{"rotate ohne Bild", `rotate(3, angle: 90)`, "rotate"},
		{"built_in_image ohne Namen", `built_in_image()`, "built_in_image"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err == nil {
				t.Fatal("der Aufruf glückte, erwartet ein Fehler")
			}

			if !strings.Contains(err.Error(), fall.nennt) {
				t.Errorf("die Meldung %q nennt %q nicht", err, fall.nennt)
			}
		})
	}
}
