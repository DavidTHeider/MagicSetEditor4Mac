package service_test

// **Eine Symbolfassung gehört zu ihrem Setfeld, nicht zum Set** —
// #548,
// Befund B.
//
//	SymbolStyleP style = dynamic_pointer_cast<SymbolStyle>(
//	    set->stylesheetForP(card)->styleFor(value->fieldP));
//	…
//	FOR_EACH(v, style->variations) {
//	  if (v->name == variation) { … }
//	}
//
// `src/script/functions/image.cpp`, Zeilen 161–170, Fassung f9d9356 vom
// 2021-07-06. **`value->fieldP`**: gesucht wird am Stil des Feldes, aus dem
// der Wert stammt — nicht in einem Topf über alle Setfelder.
//
// # Wo das auseinandergeht
//
// `magic.mse-game/set_fields` erklärt `symbol:` (sieben Fassungen) **und**
// `masterpiece symbol:` (sechs). Sie unterscheiden sich an genau einer Stelle:
// `watermark` ist unter `symbol:` `fill color: rgb(0,0,0)`, unter
// `masterpiece symbol:` `rgb(255,255,255)`.
//
// Verflacht man beide Listen, entscheidet die Reihenfolge in der Datei — und
// `symbol_variation(symbol: set.symbol, variation: "watermark")` gäbe weiß
// statt schwarz, sobald jemand die Blöcke tauscht.
//
// **Heute verdeckt die Erste-gewinnt-Regel es zufällig**, weil `symbol:` vorn
// steht. Dieser Test hält die Regel fest, statt sich auf die Reihenfolge zu
// verlassen.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEineSymbolfassungGehoertZuIhremFeld(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		angabe   string
		erwartet string
	}{
		// Schwarz — die Fassung unter `symbol`.
		"das gewöhnliche Symbolfeld": {
			`{ symbol_variation(symbol: set.symbol, variation: "watermark") }`,
			"rgb(0,0,0)",
		},

		// Weiß — die gleichnamige Fassung unter `masterpiece symbol`.
		"das Meisterwerksymbol": {
			`{ symbol_variation(symbol: set.masterpiece_symbol, variation: "watermark") }`,
			"rgb(255,255,255)",
		},
	}

	fassungen := []msepaket.Symbolfassung{
		{Feld: "symbol", Name: "watermark", Fuellfarbe: "rgb(0,0,0)"},
		{Feld: "masterpiece symbol", Name: "watermark", Fuellfarbe: "rgb(255,255,255)"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stil := msepaket.Stil{
				Breite: 375, Hoehe: 523,
				Symbolfassungen: fassungen,
				Felder: []msepaket.Feldstil{{
					Feld: "text", Links: "0", Oben: "0", Breite: "100", Hoehe: "20",
					Bild:    fall.angabe,
					Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
				}},
			}

			karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
				{Field: "text", Text: "x"},
			}}

			plan, befunde, err := service.NeuerPlanbauer().
				MitSchriften(schriftprobe{breite: 0.5}).
				Stilplan(t.Context(), karte, service.Vorlage{
					Stil: stil,
					Setfelder: []domain.Field{
						{Name: "symbol", Kind: domain.KindImage},
						{Name: "masterpiece symbol", Kind: domain.KindImage},
					},
					Setangaben: []domain.Value{
						{Field: "symbol", Text: "sym.mse-symbol"},
						{Field: "masterpiece symbol", Text: "meister.mse-symbol"},
					},
				})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			for _, befund := range befunde {
				t.Logf("Befund %s: %s", befund.Ort, befund.Meldung)
			}

			if gab := fuellfarbeIm(plan); !strings.Contains(gab, fall.erwartet) {
				t.Errorf("gezeichnet wird %q, erwartet %q darin", gab, fall.erwartet)
			}
		})
	}
}

// fuellfarbeIm gibt die Füllfarbe des ersten Symbolknotens im Plan.
//
// **Aus dem Bildbaum und nicht aus dem gerasterten Bild**: die Fassung
// entscheidet über die Farben, und die stehen im Plan (Vermerk 0022).
func fuellfarbeIm(plan *render.Plan) string {
	for i := range plan.Ebenen {
		farbe, da := plan.Ebenen[i].Bild.Farben["fill_color"]
		if da {
			return fmt.Sprintf("rgb(%d,%d,%d)", farbe.R, farbe.G, farbe.B)
		}
	}

	return "(keine Füllfarbe)"
}
