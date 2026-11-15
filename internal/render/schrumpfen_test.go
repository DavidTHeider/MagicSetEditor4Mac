package render_test

import (
	"image"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **`shrink`: ein zu breiter Zug wird gestaucht, nicht überstehen gelassen.**
//
// # Warum das zählt
//
// Die Angabe steht **676-mal** in den 190 Stilen von `M15-Magic-Pack`, fast
// immer am Namen und an der Typzeile. **Wie** MSE staucht, steht in
// `stauchen_test.go`; hier steht, was der Umbruch daraus macht.
//
// **Es ist zugleich das fehlende Glied von Phase B.** Der zweistufige Satz
// macht das Namensfeld schmaler, wenn die Manakosten breit sind
// (`right: { 341 - card_style.casting_cost.content_width }`) — sichtbar wird
// das aber erst, wenn ein zu breiter Name darauf **reagiert**
// (#291,
// #319).

// **Ein gestauchter Zug bricht nicht um.**
func TestEinGestauchterZugBrichtNichtUm(t *testing.T) {
	schrift := gleichbreit{breite: 0.5}
	satz := render.Satz{Groesse: 16, Stauchen: true, Einzeilig: true, NurBeiUeberlauf: true}

	ergebnis := render.Setze(satz, render.Rechteck{Breite: 40, Hoehe: 30}, "abcdefghij", schrift)

	if len(ergebnis.Zeilen) != 1 {
		t.Fatalf("%d Zeilen, erwartet eine", len(ergebnis.Zeilen))
	}

	// Zehn Zeichen à 0,5 bei Größe 16 ergeben 80 in einem Feld von 40.
	if ergebnis.Stauchung != 0.5 {
		t.Errorf("die Stauchung ist %g, erwartet 0,5", ergebnis.Stauchung)
	}

	// **Die Größe bleibt**, und die gemessene Breite bleibt die ungestauchte.
	if ergebnis.Groesse != 16 {
		t.Errorf("die Größe wurde auf %g geändert, erwartet 16", ergebnis.Groesse)
	}
}

// **Was passt, wird nicht angefasst.**
func TestEinPassenderZugBleibtUngestaucht(t *testing.T) {
	schrift := gleichbreit{breite: 0.5}
	satz := render.Satz{Groesse: 16, Stauchen: true, Einzeilig: true, NurBeiUeberlauf: true}

	ergebnis := render.Setze(satz, render.Rechteck{Breite: 200, Hoehe: 30}, "abcd", schrift)

	if ergebnis.Stauchung != 1 {
		t.Errorf("die Stauchung ist %g, erwartet 1", ergebnis.Stauchung)
	}
}

// **Ohne die Angabe bleibt es beim Umbruch.**
//
// `shrink` ist eine Ansage des Stils, keine Vorsichtsmaßnahme des Rasterers.
func TestOhneStauchenWirdWeiterUmgebrochen(t *testing.T) {
	schrift := gleichbreit{breite: 0.5}
	satz := render.Satz{Groesse: 16}

	ergebnis := render.Setze(satz, render.Rechteck{Breite: 40, Hoehe: 60}, "abcde fghij", schrift)

	if len(ergebnis.Zeilen) < 2 {
		t.Errorf("%d Zeilen, erwartet mindestens zwei", len(ergebnis.Zeilen))
	}

	if ergebnis.Stauchung != 1 {
		t.Errorf("die Stauchung ist %g, erwartet 1", ergebnis.Stauchung)
	}
}

// **Ein gestauchter Zug bleibt in seinem Feld — auf dem Bild.**
//
// # Der Befund des Zweitblicks
//
// Der erste Anlauf rasterte die Glyphen mit der gesetzten Größe und rechnete
// Grundlinie und Vorschub weiter mit der angeforderten; der Zug ragte trotz
// erfolgreicher Messung hinaus. **Der Wächter darüber endete bei
// [render.Setze]** und sah die gezeichneten Glyphen nie — genau die Lücke, die
// §3 „zwischen den Prüfebenen" nennt.
//
// Seit dem 2026-08-17 wird gestaucht statt verkleinert (#319). Die Frage
// bleibt dieselbe, und sie wird auf dem Bild gestellt.
func TestEinGestauchterZugBleibtInSeinemFeld(t *testing.T) {
	rasterer := render.NeuerRasterer(nil).MitSchriften(einheitsschrift{})

	// Zehn Zeichen à 0,5 bei Größe 16 wären 80 breit; das Feld ist 40.
	plan := &render.Plan{Breite: 60, Hoehe: 40, Ebenen: []render.Ebene{{
		Art:        render.ArtText,
		Rahmen:     render.Rechteck{Breite: 40, Hoehe: 30},
		Abschnitte: render.Aufrecht("abcdefghij"),
		Farbe:      render.Farbe{A: 255},
		Satz: render.Satz{
			Schrift: "Probe", Groesse: 16, Stauchen: true, Einzeilig: true, NurBeiUeberlauf: true,
		},
		Deckung: 1,
	}}}

	bild, befunde, err := rasterer.Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	// **Nichts steht rechts von der Feldkante.** Mit den Abständen der
	// ungeschrumpften Größe reichte der Zug bis 80.
	for y := range bild.Bounds().Dy() {
		for x := 41; x < bild.Bounds().Dx(); x++ {
			if _, _, _, a := bild.At(x, y).RGBA(); a > 0 {
				t.Fatalf("bei (%d,%d) steht noch Text — der Zug ragt über sein Feld hinaus", x, y)
			}
		}
	}
}

// einheitsschrift ist eine Schrift, deren Zeichen quadratische Blöcke sind.
//
// **Ein Block und kein Buchstabe**, damit der Wächter über die *Stelle* redet
// und nicht über die Gestalt: was hier zählt, ist, wo ein Zeichen endet.
type einheitsschrift struct{}

func (einheitsschrift) Breite(rune) float64  { return 0.5 }
func (einheitsschrift) Zeilenhoehe() float64 { return 1.2 }
func (einheitsschrift) Grundlinie() float64  { return 0.8 }

func (einheitsschrift) Suche(string, bool) (render.Schrift, error) { return einheitsschrift{}, nil }

func (einheitsschrift) Maske(_ rune, groesse float64) (image.Image, image.Point, bool) {
	seite := int(groesse * 0.5)
	if seite < 1 {
		seite = 1
	}

	maske := image.NewAlpha(image.Rect(0, 0, seite, seite))
	for i := range maske.Pix {
		maske.Pix[i] = 255
	}

	return maske, image.Point{}, true
}
