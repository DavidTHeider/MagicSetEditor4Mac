package render

import "image"

// Textschatten ist der Schatten einer Schrift — `shadow color:`,
// `shadow displacement x:`/`y:` und `shadow blur:` aus ihrem `font:`-Block.
//
// # Wie MSE ihn zeichnet
//
//	void RotatedDC::DrawTextWithShadow(const String& text, const Font& font, const RealPoint& pos, double scale, double stretch) {
//	  DrawText(text, pos + font.shadow_displacement * scale, font.shadow_color, font.shadow_blur * scale, 1, stretch);
//	  DrawText(text, pos, font.color, 0, 1, stretch);
//	}
//
// `src/util/rotation.cpp`, Zeilen 221–224, Fassung f9d9356 vom 2021-07-06 —
// je Textelement erst der Schatten, dann der Text. **Ohne Abfrage, ob es einen
// gibt**: ein durchsichtiger fällt in `DrawText` heraus (Zeile 191), und die
// Vorgabe ist durchsichtig (`Color()`, `src/gfx/color.hpp`, Zeile 32).
//
// **Der Versatz geht durch die Vergrößerung, die Weichheit nicht.** Der Punkt
// `pos + shadow_displacement` wird wie jeder andere ins Bild gerechnet; die
// Weichheit kommt als ganze Zahl in `draw_resampled_text` an und wirkt dort
// auf das fertige, kleine Bild — in Bildpunkten des Ziels.
type Textschatten struct {
	Farbe Farbe
	// X und Y sind der Versatz in Vorlagenpunkten.
	X, Y float64
	// Weichheit ist `shadow blur` — wie oft geweicht wird, in Bildpunkten des
	// Ziels und **abgeschnitten**: `DrawText` nimmt sie als `int`.
	Weichheit float64
}

// schattenlauf legt den Lauf des Schattens an — oder keinen, wenn er
// durchsichtig ist.
func schattenlauf(
	schatten Textschatten, links, oben, breite, hoehe, grundlinie, stauchung, vergroesserung float64,
) *vierfachlauf {
	if schatten.Farbe.A == 0 {
		return nil
	}

	return vierfachlaufMitRand(links+schatten.X*vergroesserung, oben+schatten.Y*vergroesserung,
		breite, hoehe, grundlinie, stauchung, int(schatten.Weichheit))
}

// legeSchatten rechnet den Schattenlauf herunter, nimmt die Deckung der Farbe
// hinein, weicht und legt ihn auf — die Reihenfolge von
// `draw_resampled_text`: `downsample_to_alpha`, `set_alpha`, dann
// `blur_image_alpha` so oft wie die Weichheit
// (`src/gfx/resample_text.cpp`, Zeilen 185–200, Fassung f9d9356 vom
// 2021-07-06).
func (r *Rasterer) legeSchatten(
	ziel *image.NRGBA, ebene Ebene, maske *maskenbild, rahmen image.Rectangle,
	lauf *vierfachlauf, schatten Textschatten,
) {
	farbe := schatten.Farbe
	farbe.A = 255

	r.legeMit(ziel, ebene, farbe, maske, rahmen, lauf, func(deckung []uint8) {
		// `Byte b_alpha = Byte(alpha * 255);` und `im[i] = (im[i] * b_alpha) / 255;`
		// — `set_alpha`, `src/gfx/blend_image.cpp`, Zeilen 107–118, Fassung
		// f9d9356 vom 2021-07-06.
		if schatten.Farbe.A != 255 {
			anteil := int(uint8(float64(schatten.Farbe.A) / 255 * 255))
			for i := range deckung {
				// #nosec G115 -- beide Faktoren sind höchstens 255, das Ergebnis also auch
				deckung[i] = uint8(int(deckung[i]) * anteil / 255) //nolint:gosec // dito
			}
		}

		for range int(schatten.Weichheit) {
			weichzeichne(deckung, lauf.ausgabebreite, lauf.hoehe)
		}
	})
}

// weichzeichne ist `blur_image_alpha`: jeder Punkt wird
// `(2·Mitte + links + oben + rechts + unten) / 6`, am Rand die Mitte statt des
// fehlenden Nachbarn — **an Ort und Stelle**, sodass links und oben schon
// geweicht sind (`src/gfx/resample_text.cpp`, Zeilen 144–162, Fassung f9d9356
// vom 2021-07-06). Das ist nicht symmetrisch, und genau so steht es dort.
func weichzeichne(deckung []uint8, breite, hoehe int) {
	for y := range hoehe {
		for x := range breite {
			i := y*breite + x
			mitte := int(deckung[i])
			links, oben, rechts, unten := mitte, mitte, mitte, mitte

			if x > 0 {
				links = int(deckung[i-1])
			}

			if y > 0 {
				oben = int(deckung[i-breite])
			}

			if x < breite-1 {
				rechts = int(deckung[i+1])
			}

			if y < hoehe-1 {
				unten = int(deckung[i+breite])
			}

			// #nosec G115 -- ein gewichtetes Mittel aus sechs Bytes ist höchstens 255
			deckung[i] = uint8((2*mitte + links + oben + rechts + unten) / 6) //nolint:gosec // dito
		}
	}
}
