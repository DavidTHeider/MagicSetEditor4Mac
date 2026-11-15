package render_test

import (
	"image/color"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Ein Symbol mitten im Text steht an seiner Stelle im Zug.**
//
// # Warum das nicht als eigene Ebene geht
//
// Ein Symbolfeld — die Manakosten — wird in `internal/service` in eigene Ebenen
// zerlegt; dort steht der Zug allein im Feld, und seine Stelle ist
// ausgerechnet. **Mitten im Regeltext geht das nicht**: wo `<sym>T</sym>`
// landet, weiß erst der Umbruch
// (#312).
func TestEinSymbolMittenImTextStehtAnSeinerStelle(t *testing.T) {
	quelle := bildquelle{da: map[string]color.RGBA{"/t.png": {R: 255, A: 255}}}

	plan := &render.Plan{Breite: 60, Hoehe: 20, Ebenen: []render.Ebene{{
		Art:    render.ArtText,
		Rahmen: render.Rechteck{Breite: 60, Hoehe: 20},
		Abschnitte: []render.Abschnitt{
			{Text: "a"},
			{
				Text: "T", Symbole: []render.Symbol{{Bild: render.Blatt("/t.png")}},
				Symbolmasse: render.Symbolmasse{Groesse: 6},
			},
		},
		Farbe:   render.Farbe{A: 255},
		Satz:    render.Satz{Schrift: "Aufrecht", Groesse: 10},
		Deckung: 1,
	}}}

	bild, befunde, err := render.NeuerRasterer(quelle).
		MitSchriften(schriften{"Aufrecht": einheitsschrift{}}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	// Der Buchstabe belegt 5 Punkte, das Symbol beginnt danach. Bei x = 7
	// steht also das Symbol — und nur, wenn es überhaupt gezeichnet wurde.
	//
	// **Senkrecht in der Zeilenmitte**, seit die Lage an MSE nachgelesen ist:
	// die Zeile ist 12 hoch, das Symbol 6, es steht also von 3 bis 9 (#330).
	if _, _, _, a := bild.At(7, 5).RGBA(); a == 0 {
		t.Error("bei (7,5) steht nichts — das Symbol wurde nicht gezeichnet")
	}

	// **Und links davon der Buchstabe.** Ein Symbol, das den Zug verdrängte,
	// stünde an seiner Stelle.
	if _, _, _, a := bild.At(1, 9).RGBA(); a == 0 {
		t.Error("bei (1,9) steht nichts — der Buchstabe ist verloren gegangen")
	}
}

// **Ein Symbolbild, das fehlt, wird benannt.**
//
// Es ist derselbe Satz wie bei einer fehlenden Glyphe: eine Karte, der ein
// Manasymbol fehlt, sieht aus wie eine billigere (§8).
func TestEinFehlendesSymbolbildWirdBenannt(t *testing.T) {
	plan := &render.Plan{Breite: 60, Hoehe: 20, Ebenen: []render.Ebene{{
		Art:    render.ArtText,
		Rahmen: render.Rechteck{Breite: 60, Hoehe: 20},
		Abschnitte: []render.Abschnitt{
			{
				Text: "T", Symbole: []render.Symbol{{Bild: render.Blatt("/fehlt.png")}},
				Symbolmasse: render.Symbolmasse{Groesse: 6},
			},
		},
		Farbe:   render.Farbe{A: 255},
		Satz:    render.Satz{Schrift: "Aufrecht", Groesse: 10},
		Deckung: 1,
	}}}

	_, befunde, err := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{}}).
		MitSchriften(schriften{"Aufrecht": einheitsschrift{}}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "fehlt.png") {
		t.Errorf("die Befunde sind %+v, erwartet einer über /fehlt.png", befunde)
	}
}

// **Ein Symbol, das über den Bildrand hinausragt, wird beschnitten.**
//
// Ein Regeltext, der bis an die Kante läuft, setzt sein letztes Symbol
// teilweise außerhalb — geschrieben wird nur, was auf der Karte liegt.
func TestEinSymbolUeberDemBildrandWirdBeschnitten(t *testing.T) {
	quelle := bildquelle{da: map[string]color.RGBA{"/t.png": {R: 255, A: 255}}}

	plan := &render.Plan{Breite: 8, Hoehe: 8, Ebenen: []render.Ebene{{
		Art:    render.ArtText,
		Rahmen: render.Rechteck{Breite: 40, Hoehe: 20},
		Abschnitte: []render.Abschnitt{
			{
				Text: "T", Symbole: []render.Symbol{{Bild: render.Blatt("/t.png")}},
				Symbolmasse: render.Symbolmasse{Groesse: 20},
			},
		},
		Farbe:   render.Farbe{A: 255},
		Satz:    render.Satz{Schrift: "Aufrecht", Groesse: 10},
		Deckung: 1,
	}}}

	bild, befunde, err := render.NeuerRasterer(quelle).
		MitSchriften(schriften{"Aufrecht": einheitsschrift{}}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if bild.Bounds().Dx() != 8 || bild.Bounds().Dy() != 8 {
		t.Errorf("das Bild ist %v — es ist mitgewachsen", bild.Bounds())
	}
}
