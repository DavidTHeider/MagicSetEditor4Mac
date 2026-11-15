package render_test

// **`<bullet>` legt seine Zeichen in den linken Rand**
// (#431).
//
//	if (i < elements.paragraphs[i_para].margin_end_char) {
//	  line.margin_left += c.size.width; // character in left margin
//	}
//
// `src/render/text/viewer.cpp`, Zeilen 611–613, Fassung f9d9356 vom
// 2021-07-06; `margin_end_char` setzt das schließende `</bullet`
// (`element.cpp:176`). Die Zeichen werden gezeichnet, wo sie stehen — aber die
// **folgenden** Zeilen des Absatzes beginnen hinter ihnen: ein hängender
// Einzug.
//
// `magic.mse-game` klammert damit jede Zeile, die mit „• " beginnt
// (`script`, Zeilen 1943–1946).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestEinAufzaehlungspunktHaengt(t *testing.T) {
	satz := render.Satz{Groesse: 10, Ausrichtung: render.LinksOben}
	// Zeichen sind 5 breit; „• " misst 10.
	rahmen := render.Rechteck{Breite: 45, Hoehe: 100}

	gesetzt := render.SetzeAbschnitte(satz, rahmen, []render.Abschnitt{
		{Text: "• ", ImRand: true},
		{Text: "ab cd ef"},
	}, render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

	if len(gesetzt.Zeilen) < 2 {
		t.Fatalf("%d Zeilen, erwartet mindestens zwei: %+v", len(gesetzt.Zeilen), gesetzt.Zeilen)
	}

	// **Die erste Zeile beginnt am Rand des Feldes** — der Punkt steht, wo er
	// steht.
	if gesetzt.Zeilen[0].Links != 0 {
		t.Errorf("erste Zeile bei %g, erwartet 0", gesetzt.Zeilen[0].Links)
	}

	// **Die zweite hinter ihm.**
	if gesetzt.Zeilen[1].Links != 10 {
		t.Errorf("zweite Zeile bei %g, erwartet 10", gesetzt.Zeilen[1].Links)
	}
}

// **Ohne `<bullet>` hängt nichts** — die Gegenprobe: ohne sie bestünde der
// Test auch, wenn jede Folgezeile einrückte.
func TestOhneAufzaehlungspunktHaengtNichts(t *testing.T) {
	satz := render.Satz{Groesse: 10, Ausrichtung: render.LinksOben}
	rahmen := render.Rechteck{Breite: 45, Hoehe: 100}

	gesetzt := render.SetzeAbschnitte(satz, rahmen,
		[]render.Abschnitt{{Text: "• "}, {Text: "ab cd ef"}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

	if len(gesetzt.Zeilen) < 2 {
		t.Fatalf("%d Zeilen, erwartet mindestens zwei", len(gesetzt.Zeilen))
	}

	if gesetzt.Zeilen[1].Links != 0 {
		t.Errorf("zweite Zeile bei %g, erwartet 0", gesetzt.Zeilen[1].Links)
	}
}
