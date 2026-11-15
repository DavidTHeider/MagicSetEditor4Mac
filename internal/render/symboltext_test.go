package render_test

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Ein Symbol schreibt Text auf sein Bild — `draw text:`.**
//
//	// 2. draw text
//	if (text.empty() || !sym.text_font) return;
//	sym_rect = RealRect(bmp_pos, bmp_size);
//	sym_rect.x += font_size * sym.text_margin_left;
//	…
//	double size = font_size * sym.text_font->size;
//
// `SymbolFont::drawSymbol`, `src/data/symbol_font.cpp`, Zeilen 287–329, Fassung
// f9d9356 vom 2021-07-06.
//
// **Daran hängt jede Zahl in den Manakosten.** Ohne sie steht auf der Karte der
// leere Kreis, und die Eins in `1R` fehlt
// (#336).
func TestEinSymbolSchreibtTextAufSeinBild(t *testing.T) {
	schrift := echteSchrift(t)

	ohne := gezeichneterSymbolzug(t, schrift, "")
	mit := gezeichneterSymbolzug(t, schrift, "8")

	// **Der Text macht dunkle Fläche dazu**, wo vorher nur das Bild war.
	if dunkle(mit) <= dunkle(ohne) {
		t.Errorf("mit Text %d dunkle Punkte, ohne %d — der Text fehlt",
			dunkle(mit), dunkle(ohne))
	}
}

// **Ohne Textschrift wird nichts geschrieben.**
//
// `if (text.empty() || !sym.text_font) return;` — beides ist eine Bedingung,
// und eine Vorlage, die `draw text:` ohne `text font:` schreibt, meint keinen
// Text.
func TestOhneTextschriftWirdNichtsGeschrieben(t *testing.T) {
	schrift := echteSchrift(t)

	ohne := gezeichneterSymbolzug(t, schrift, "")

	stumm := symbolplan("8")
	stumm.Ebenen[1].Abschnitte[0].Symbole[0].Textsatz.Schrift = ""

	if dunkle(gerastertMitSchrift(t, stumm, schrift)) != dunkle(ohne) {
		t.Error("ohne `text font:` wurde trotzdem geschrieben")
	}
}

// gezeichneterSymbolzug rastert einen Symbolzug mit dem gegebenen Text.
func gezeichneterSymbolzug(t *testing.T, schrift render.Glyphenschrift, text string) *image.NRGBA {
	t.Helper()

	return gerastertMitSchrift(t, symbolplan(text), schrift)
}

func gerastertMitSchrift(
	t *testing.T, plan *render.Plan, schrift render.Glyphenschrift,
) *image.NRGBA {
	t.Helper()

	bild, _, err := render.NeuerRasterer(bilder{"/probe.mse-symbol-font/kreis.png": weissesFeld()}).
		MitSchriften(schriften{"Andale": schrift}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	return bild
}

// symbolplan ist ein Plan aus einem einzigen Symbol.
func symbolplan(text string) *render.Plan {
	return &render.Plan{Breite: 60, Hoehe: 60, Ebenen: []render.Ebene{
		{
			Art:     render.ArtFlaeche,
			Rahmen:  render.Rechteck{Breite: 60, Hoehe: 60},
			Farbe:   render.Farbe{R: 255, G: 255, B: 255, A: 255},
			Deckung: 1,
		},
		{
			Art:     render.ArtSymbolzug,
			Rahmen:  render.Rechteck{Breite: 60, Hoehe: 60},
			Deckung: 1,
			Satz:    render.Satz{Groesse: 40},
			Abschnitte: []render.Abschnitt{{
				Symbolmasse: render.Symbolmasse{Groesse: 40},
				Symbole: []render.Symbol{{
					Bild: render.Blatt("/probe.mse-symbol-font/kreis.png"),
					Text: text,
					Textsatz: render.Symboltextsatz{
						Schrift: "Andale", Groesse: 0.6, Dehnung: 0.5,
						Farbe: render.Farbe{A: 255},
					},
				}},
			}},
		},
	}}
}

// weissesFeld ist ein deckend weißes Bild — dunkle Punkte darauf sind Text.
func weissesFeld() image.Image {
	const seite = 40

	breite, hoehe := seite, seite

	bild := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))
	for y := range hoehe {
		for x := range breite {
			bild.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	return bild
}

// dunkle zählt die Bildpunkte, die deutlich dunkler als weiß sind.
func dunkle(bild *image.NRGBA) int {
	gezaehlt := 0

	for y := bild.Bounds().Min.Y; y < bild.Bounds().Max.Y; y++ {
		for x := bild.Bounds().Min.X; x < bild.Bounds().Max.X; x++ {
			if farbe := bild.NRGBAAt(x, y); farbe.A > 0 && farbe.R < 128 {
				gezaehlt++
			}
		}
	}

	return gezaehlt
}

// **Ein Text, der nicht passt, wird kleiner — und irgendwann weggelassen.**
//
//	while (true) {
//	  if (size <= 0) return; // text too small
//	  …
//	  size -= dc.getFontSizeStep();
//	}
//
// `SymbolFont::drawSymbol`, `src/data/symbol_font.cpp`, Zeilen 307–322, Fassung
// f9d9356 vom 2021-07-06. **`max stretch` ist eine Schwelle**: erst wenn der
// Text auch gestaucht nicht mehr hineinpasst, wird der Grad kleiner.
func TestEinTextDerNichtPasstWirdKleiner(t *testing.T) {
	schrift := echteSchrift(t)

	kurz := gezeichneterSymbolzug(t, schrift, "8")

	// **Ein langer Text in denselben Kreis**: er wird gestaucht und
	// verkleinert, steht aber immer noch da.
	lang := gezeichneterSymbolzug(t, schrift, "888")
	if dunkle(lang) == 0 {
		t.Error("ein langer Text wurde gar nicht gezeichnet")
	}

	// **Und einer, für den kein Grad mehr reicht, wird weggelassen.** Der
	// Kasten ist hier durch die Ränder auf nichts geschrumpft.
	eng := symbolplan("8")
	eng.Ebenen[1].Abschnitte[0].Symbole[0].Textsatz.RandOben = 0.5
	eng.Ebenen[1].Abschnitte[0].Symbole[0].Textsatz.RandUnten = 0.5

	if dunkle(gerastertMitSchrift(t, eng, schrift)) != 0 {
		t.Error("ein Text ohne Platz wurde trotzdem gezeichnet")
	}

	_ = kurz
}

// **Eine Textschrift, die es nicht gibt, wird benannt.**
//
// Der Kreis steht trotzdem da — ein Manasymbol ohne Ziffer ist eine halbe
// Auskunft, ein fehlendes gar keine (§8).
func TestEineFehlendeTextschriftWirdBenannt(t *testing.T) {
	plan := symbolplan("8")
	plan.Ebenen[1].Abschnitte[0].Symbole[0].Textsatz.Schrift = "GibtsNicht"

	_, befunde, err := render.NeuerRasterer(
		bilder{"/probe.mse-symbol-font/kreis.png": weissesFeld()}).
		MitSchriften(schriften{}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "GibtsNicht") {
		t.Errorf("Befunde: %v", befunde)
	}
}

// **Ohne Schriftsuche bleibt der Text weg, ohne Aufhebens.**
//
// Ein Rasterer ohne Schriften zeichnet keine Texte — das steht schon an der
// Textebene, und ein zweiter Befund je Symbol machte die Liste unlesbar.
func TestOhneSchriftsucheBleibtDerSymboltextWeg(t *testing.T) {
	plan := symbolplan("8")

	bild, befunde, err := render.NeuerRasterer(
		bilder{"/probe.mse-symbol-font/kreis.png": weissesFeld()}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if dunkle(bild) != 0 {
		t.Error("ohne Schriftsuche wurde Text gezeichnet")
	}
}

// **Eine Schrift ohne Glyphen wird benannt.**
//
// Sie kann Breiten und damit den Satz, aber keine Bilder — dann steht der
// Kreis da und die Ziffer nicht, und der Befund sagt welche Schrift.
func TestEineSchriftOhneGlyphenWirdBenannt(t *testing.T) {
	plan := symbolplan("8")

	_, befunde, err := render.NeuerRasterer(
		bilder{"/probe.mse-symbol-font/kreis.png": weissesFeld()}).
		MitSchriften(nurbreiten{}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "Andale") {
		t.Errorf("Befunde: %v", befunde)
	}
}

// **Ein Zeichen, das die Schrift nicht führt, rückt trotzdem vor.**
//
// Sonst stünde der Rest des Textes übereinander. Gemeldet wird es nicht: ein
// Symboltext ist ein Zeichen lang, und ein Befund je Karte und Symbol stünde in
// einer Liste, die niemand zu Ende liest.
func TestEinFehlendesZeichenImSymboltextRuecktVor(t *testing.T) {
	schrift := echteSchrift(t)

	// Ein Zeichen, das keine gewöhnliche Schrift führt.
	mit := gezeichneterSymbolzug(t, schrift, "\U0010FFFD")
	if dunkle(mit) != 0 {
		t.Error("für ein fehlendes Zeichen wurde etwas gezeichnet")
	}
}

// nurbreiten ist eine Schrift, die messen kann und nicht zeichnen.
type nurbreiten struct{}

func (nurbreiten) Breite(rune) float64  { return 0.5 }
func (nurbreiten) Zeilenhoehe() float64 { return 1.2 }

func (nurbreiten) Suche(string, bool) (render.Schrift, error) { return nurbreiten{}, nil }
