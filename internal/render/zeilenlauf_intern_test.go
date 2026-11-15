package render

import (
	"image"
	"testing"

	"golang.org/x/image/font/gofont/goregular"
)

// **Eine Zeile wird je Abschnitt in einem Zug gezeichnet** —
// #2302.
//
// MSE2 zeichnet eine Zeile, die nicht im Blocksatz steht, je Element mit
// **einem** Ruf: `elements.draw(dc, scale, rect, …, l.start, l.end())`
// (`TextViewer::draw`, `src/render/text/viewer.cpp`, Zeilen 93–96), und
// `FontTextElement::draw` gibt den ganzen Ausschnitt an `DrawTextWithShadow`
// (`src/render/text/font.cpp`, Zeilen 15–25), Fassung f9d9356 vom
// 2021-07-06. Wo die Wörter darin stehen, bestimmen die Vorschübe der
// Schrift — nicht die Stellen, die der Umbruch auf das Viertelpunktraster
// gelegt hat. Hier wurde jedes Wort für sich an seine gerasterte Stelle
// gezeichnet.
//
// **Gemessen hat es der Kanon**: seit ein gestauchtes Feld umbricht (#2308),
// geht auch der Name diesen Weg, und die Namen der Vanguard-Avatare lagen
// danach um bis zu sechs weiter von MSE2 weg — im Namensstreifen 38 statt 1,2.

// zeilenbild zeichnet einen gesetzten Zug auf ein leeres Bild.
func zeilenbild(t *testing.T, ebene Ebene, gesetzt Satzergebnis, schriften Schriftsatz) *image.NRGBA {
	t.Helper()

	ziel := image.NewNRGBA(image.Rect(0, 0, 400, 60))
	r := NeuerRasterer(nil)

	if fehlend, _, _ := r.zeichneZeilen(ziel, ebene, nil, 1, gesetzt, schriften, GueteMittel); fehlend != "" {
		t.Fatalf("es fehlen Zeichen: %s", fehlend)
	}

	return ziel
}

// inEinemZug fasst die Stücke einer Zeile zusammen, die laut `zug` zusammen
// gezeichnet werden — die Stelle ist die des ersten.
func inEinemZug(gesetzt Satzergebnis, zug func(stelle int) int) Satzergebnis {
	zusammen := gesetzt
	zusammen.Zeilen = make([]Zeile, len(gesetzt.Zeilen))

	for nummer, zeile := range gesetzt.Zeilen {
		var (
			stuecke []Stueck
			vorige  int
		)

		for stelle, stueck := range zeile.Stuecke {
			if letztes := len(stuecke) - 1; letztes >= 0 && vorige == zug(stelle) {
				stuecke[letztes].Text += stueck.Text
				stuecke[letztes].Breite = stueck.Links + stueck.Breite - stuecke[letztes].Links

				continue
			}

			stuecke = append(stuecke, stueck)
			vorige = zug(stelle)
		}

		zeile.Stuecke = stuecke
		zusammen.Zeilen[nummer] = zeile
	}

	return zusammen
}

func TestEineZeileWirdJeAbschnittInEinemZugGezeichnet(t *testing.T) {
	t.Parallel()

	schrift, err := SchriftAusDatei(goregular.TTF)
	if err != nil {
		t.Fatalf("Go Regular: %v", err)
	}

	schriften := Schriftsatz{Aufrecht: schrift}

	for _, fall := range []struct {
		name       string
		abschnitte []Abschnitt
		// teil gibt den Abschnitt des Stücks an einer Stelle der Zeile — von
		// Hand, denn die Erwartung darf nicht aus dem rechnen, was sie prüft.
		teil func(stelle int) int
	}{
		{
			"ein Abschnitt ist ein Zug",
			Aufrecht("Iiii Wavy iii Lilly"),
			func(int) int { return 0 },
		},
		{
			"zwei Abschnitte sind zwei Züge, auch in derselben Schrift",
			// „Iiii", „ ", „Wavy", „ " aus dem ersten, der Rest aus dem zweiten.
			[]Abschnitt{{Text: "Iiii Wavy "}, {Text: "iii Lilly"}},
			func(stelle int) int { return min(stelle/4, 1) },
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			// **Eine Größe, deren Vorschübe nicht auf dem Raster liegen** —
			// sonst fielen gerasterte und laufende Stellen zusammen, und der
			// Fall prüfte nichts.
			ebene := Ebene{
				Art: ArtText, Rahmen: Rechteck{Links: 4, Oben: 4, Breite: 390, Hoehe: 50},
				Abschnitte: fall.abschnitte, Farbe: Farbe{A: 255}, Deckung: 1,
				Satz: Satz{Schrift: "Go", Groesse: 13.3, Ausrichtung: LinksOben},
			}

			gesetzt := SetzeAbschnitte(ebene.Satz, ebene.Rahmen, ebene.Abschnitte, schriften, nil)
			if len(gesetzt.Zeilen) != 1 || len(gesetzt.Zeilen[0].Stuecke) < 3 {
				t.Fatalf("erwartet eine Zeile aus mehreren Stücken, gesetzt %+v", gesetzt.Zeilen)
			}

			erwartet := zeilenbild(t, ebene, inEinemZug(gesetzt, fall.teil), schriften)
			if gleicheDeckung(zeilenbild(t, ebene, gesetzt, schriften), erwartet) {
				return
			}

			t.Error("die Zeile wurde nicht je Abschnitt in einem Zug gezeichnet")
		})
	}
}

// **Die Gegenprobe der Vorrichtung**: Wort für Wort gezeichnet sieht diese
// Zeile anders aus als in einem Zug — sonst bewiese der Fall oben nichts.
func TestWortFuerWortSiehtDieZeileAndersAus(t *testing.T) {
	t.Parallel()

	schrift, err := SchriftAusDatei(goregular.TTF)
	if err != nil {
		t.Fatalf("Go Regular: %v", err)
	}

	schriften := Schriftsatz{Aufrecht: schrift}
	ebene := Ebene{
		Art: ArtText, Rahmen: Rechteck{Links: 4, Oben: 4, Breite: 390, Hoehe: 50},
		Abschnitte: Aufrecht("Iiii Wavy iii Lilly"), Farbe: Farbe{A: 255}, Deckung: 1,
		Satz: Satz{Schrift: "Go", Groesse: 13.3, Ausrichtung: LinksOben},
	}

	gesetzt := SetzeAbschnitte(ebene.Satz, ebene.Rahmen, ebene.Abschnitte, schriften, nil)
	einzeln := gesetzt
	einzeln.Zeilen = []Zeile{gesetzt.Zeilen[0]}

	// **Jedes Stück als eigener Abschnitt** — so zeichnet der Rasterer Wort
	// für Wort, wie vor #2302.
	stuecke := append([]Stueck{}, gesetzt.Zeilen[0].Stuecke...)
	for nummer := range stuecke {
		stuecke[nummer].Lauf = nummer + 1
	}

	einzeln.Zeilen[0].Stuecke = stuecke

	if gleicheDeckung(zeilenbild(t, ebene, einzeln, schriften),
		zeilenbild(t, ebene, inEinemZug(gesetzt, func(int) int { return 0 }), schriften)) {
		t.Error("Wort für Wort und in einem Zug sehen gleich aus — die Größe prüft nichts")
	}
}

// **Der Puffer ist so breit wie der gezeichnete Text** (Nachlauf über die
// Folgen vor 0.8.0, Befund 3). MSE2 misst ihn in `RotatedDC::DrawText` neu —
// `RealRect r(pos, GetTextExtent(text))` (`src/util/rotation.cpp`, Zeilen
// 189–207, Fassung f9d9356 vom 2021-07-06) —, und aus dieser Breite rechnet
// `draw_resampled_text` den gestauchten Puffer. Die Breite aus dem Umbruch
// hängt dagegen am Text davor: an der Vorsilbe des Abschnitts und an der
// Kernung über die Bruchkante.
func TestDerPufferMisstDenGezeichnetenText(t *testing.T) {
	t.Parallel()

	schrift, err := SchriftAusDatei(goregular.TTF)
	if err != nil {
		t.Fatalf("Go Regular: %v", err)
	}

	schriften := Schriftsatz{Aufrecht: schrift}
	ebene := Ebene{
		Art: ArtText, Rahmen: Rechteck{Links: 4, Oben: 4, Breite: 390, Hoehe: 50},
		Farbe: Farbe{A: 255}, Deckung: 1,
		Satz: Satz{Schrift: "Go", Groesse: 13.3, Ausrichtung: LinksOben},
	}

	zug := func(breite float64) Satzergebnis {
		return Satzergebnis{
			Groesse: 13.3, Stauchung: 0.8,
			Zeilen: []Zeile{{Oben: 4, Hoehe: 16, Stuecke: []Stueck{
				{Text: "Wavy Lilly", Links: 4, Breite: breite, Lauf: 1},
			}}},
		}
	}

	// **Zwei Breiten aus dem Umbruch, ein gezeichneter Text** — gestaucht,
	// denn nur dort hängt das Bild an der Breite des Puffers.
	if !gleicheDeckung(zeilenbild(t, ebene, zug(61), schriften), zeilenbild(t, ebene, zug(68.5), schriften)) {
		t.Error("das Bild hängt an der Breite aus dem Umbruch statt am gezeichneten Text")
	}
}
