package render_test

// **Ein weicher Umbruch beginnt keinen Absatz — und bekommt deshalb keinen
// oberen Rand** (#584).
//
//	void addParagraphs(const String& text, size_t start, size_t end) {
//	  if (line == 0 && soft_line > 0) return;
//	  for (size_t i = start; i < end; ++i) {
//	    if (text.GetChar(i) == '\n') {
//	      paragraphs.back().end = i + 1;
//	      paragraphs.emplace_back();
//	      …
//	      paragraphs.back().margin_top = margins.back().top;
//	    }
//	  }
//	}
//
// `src/render/text/element.cpp`, Zeilen 272–290, Fassung f9d9356 vom
// 2021-07-06. **Die erste Zeile ist die Regel**: steht der Umbruch in einem
// `<soft-line>` und **nicht** in einem `<line>`, entsteht gar kein Absatz —
// also auch kein neuer oberer Rand, kein neuer Einzug, keine neue Ausrichtung.
//
// # Was das gekostet hat
//
// `Ranger Class` trägt
//
//	<margin:0:0:30>You may look at the top card of your library any
//	time.<soft-line>\n</soft-line><soft-line>\n</soft-line>You may cast
//	creature spells from the top of your library.</margin:0:0:30>
//
// — zwei weiche Umbrüche innerhalb **eines** Rands von 30. Wir schnitten an
// beiden einen Absatz und legten den Rand **zweimal** neu auf: 60 Punkte, die
// MSE nicht setzt. Gemessen war der ganze Block 68 Punkte höher als seiner,
// bei übereinstimmender Mitte.
//
// **Die Regel stand längst im Quelltext dieses Projekts** — als Kommentar an
// der Stelle, die sie nur für „ist der Umbruch weich?" benutzte und nicht für
// „beginnt hier ein Absatz?".
//
// **Und der erste Absatz bekommt gar keinen oberen Rand** — MSE liest
// `margin_top` nur beim Wechsel (`viewer.cpp`, Zeile 694). Hier lag die erste
// Zeile bis zum 2026-10-11 bei 30; gefragt über einen Orakelstil, steht
// `<margin:0:0:30>XXX` bei MSE2 genau dort, wo `XXX` ohne Rand steht.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestEinWeicherUmbruchBeginntKeinenAbsatz(t *testing.T) {
	t.Parallel()

	rand := &render.Margen{Oben: 30}

	faelle := map[string]struct {
		abschnitte []render.Abschnitt
		erwartet   []float64
	}{
		// **Ein harter Umbruch beginnt einen Absatz** — der Rand kommt oben
		// drauf. Die Gegenprobe zum Fall darunter.
		"ein harter Umbruch": {
			abschnitte: []render.Abschnitt{
				{Text: "a\n", Margen: rand},
				{Text: "b", Margen: rand},
			},
			erwartet: []float64{0, 42},
		},

		// **Ein weicher nicht.** Die zweite Zeile rückt um die Zeilenhöhe vor,
		// und um sonst nichts.
		"ein weicher Umbruch": {
			abschnitte: []render.Abschnitt{
				{Text: "a", Margen: rand},
				{Text: "\n", Margen: rand, WeicheZeile: true},
				{Text: "b", Margen: rand},
			},
			erwartet: []float64{0, 12},
		},

		// **Zwei weiche hintereinander** — der Fall von `Ranger Class`. Ohne
		// die Regel lägen hier 60 Punkte zu viel.
		"zwei weiche Umbrüche": {
			abschnitte: []render.Abschnitt{
				{Text: "a", Margen: rand},
				{Text: "\n", Margen: rand, WeicheZeile: true},
				{Text: "\n", Margen: rand, WeicheZeile: true},
				{Text: "b", Margen: rand},
			},
			erwartet: []float64{0, 12, 24},
		},

		// **`<line>` schlägt `<soft-line>`** — genau die erste Zeile der
		// Regel: `line == 0 && soft_line > 0`. Steht beides, gilt der Umbruch
		// weiter als Absatzende.
		"weich und Blockende zugleich": {
			abschnitte: []render.Abschnitt{
				{Text: "a", Margen: rand},
				{Text: "\n", Margen: rand, WeicheZeile: true, Blockende: true},
				{Text: "b", Margen: rand},
			},
			erwartet: []float64{0, 42},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gesetzt := render.SetzeAbschnitte(
				render.Satz{Groesse: 10, Ausrichtung: render.LinksOben},
				render.Rechteck{Breite: 200, Hoehe: 400},
				fall.abschnitte,
				render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

			if len(gesetzt.Zeilen) != len(fall.erwartet) {
				t.Fatalf("%d Zeilen, erwartet %d", len(gesetzt.Zeilen), len(fall.erwartet))
			}

			for i, oben := range fall.erwartet {
				if gesetzt.Zeilen[i].Oben != oben {
					t.Errorf("Zeile %d oben bei %g, erwartet %g",
						i, gesetzt.Zeilen[i].Oben, oben)
				}
			}
		})
	}
}
