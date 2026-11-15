package render_test

import (
	"image"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Ein Symbolzug wird im Feld ausgerichtet, nicht an seine linke Kante gelegt.**
//
// Die Manakosten von `magic-m15` tragen `alignment: middle right`, und ihr
// Kasten ist breiter als die Kosten. Bis zum 2026-08-17 rechnete
// `internal/service` diesen Versatz selbst aus — waagerecht, und senkrecht gar
// nicht. Beides gehört dorthin, wo die Symbole gemessen werden
// (#330).
func TestEinSymbolzugSitztDortWoDasFeldIhnHinstellt(t *testing.T) {
	quelle := bildmasse{da: map[string]image.Point{"/r.png": {X: 20, Y: 10}}}

	plan := &render.Plan{Breite: 100, Hoehe: 40, Ebenen: []render.Ebene{{
		Art:    render.ArtSymbolzug,
		Rahmen: render.Rechteck{Breite: 100, Hoehe: 40},
		Abschnitte: []render.Abschnitt{{
			Text:    "R",
			Symbole: []render.Symbol{{Bild: render.Blatt("/r.png")}},
			Symbolmasse: render.Symbolmasse{
				Groesse: 5, Bildgroesse: 10,
				AbstandWaagerecht: 1, Ausrichtung: render.MitteMitte,
			},
		}},
		Satz:    render.Satz{Ausrichtung: render.RechtsMitte},
		Deckung: 1,
	}}}

	bild, befunde, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	// Der Kasten misst 11 × 5 und sitzt rechts (links 89) und mittig
	// (oben 17,5). Das Bild darin ist 10 × 5 und mittig im Kasten.
	if _, _, _, a := bild.At(94, 19).RGBA(); a == 0 {
		t.Error("bei (94,19) steht nichts — der Zug sitzt nicht rechts in der Mitte")
	}

	if _, _, _, a := bild.At(50, 19).RGBA(); a != 0 {
		t.Error("bei (50,19) steht etwas — der Zug klebt an der linken Kante")
	}

	if _, _, _, a := bild.At(94, 2).RGBA(); a != 0 {
		t.Error("bei (94,2) steht etwas — der Zug sitzt an der Oberkante")
	}
}

// **Ein Symbol mitten im Text sitzt in der Zeile, nicht auf ihrer Oberkante.**
//
// `SymbolFont::drawSymbol` richtet das Bild in seinem Kasten aus
// (`align_in_rect(align, bmp_size, sym_rect)`), und der Kasten ist die ganze
// Zeilenhöhe (`split_left` erhält die Höhe des Zeilenrechtecks). Ohne
// Angabe gilt `ALIGN_MIDDLE_CENTER` (`SymbolFontRef::SymbolFontRef`, Zeile
// 556) — das Symbol steht also **mittig**.
//
// Bis zum 2026-08-17 stand hier „auf der Zeile und nicht auf der Grundlinie,
// eine Grundlinie hat ein Bild nicht" — eine Setzung, keine gelesene Regel,
// und sie war der Anlass für #330.
func TestEinSymbolImTextSitztMittigInDerZeile(t *testing.T) {
	quelle := bildmasse{da: map[string]image.Point{"/r.png": {X: 10, Y: 10}}}

	plan := &render.Plan{Breite: 40, Hoehe: 40, Ebenen: []render.Ebene{{
		Art:    render.ArtText,
		Rahmen: render.Rechteck{Breite: 40, Hoehe: 40},
		Abschnitte: []render.Abschnitt{{
			Text:        "R",
			Symbole:     []render.Symbol{{Bild: render.Blatt("/r.png")}},
			Symbolmasse: render.Symbolmasse{Groesse: 4, Bildgroesse: 10},
		}},
		Farbe:   render.Farbe{A: 255},
		Satz:    render.Satz{Schrift: "Aufrecht", Groesse: 20, Ausrichtung: render.LinksOben},
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

	// Die Zeile ist 20 hoch, das Symbol 4 — mittig heißt von 8 bis 12.
	if _, _, _, a := bild.At(2, 10).RGBA(); a == 0 {
		t.Error("bei (2,10) steht nichts — das Symbol sitzt nicht in der Zeilenmitte")
	}

	if _, _, _, a := bild.At(2, 1).RGBA(); a != 0 {
		t.Error("bei (2,1) steht etwas — das Symbol hängt an der Zeilenoberkante")
	}
}

// **Ein Symbolzug ohne Symbole ist ein Mangel und kein leeres Bild.**
//
// Dieselbe Zusage wie bei einer Bildebene ohne Bild: eine Ebene, die nichts
// zeichnen kann, wird benannt, statt still zu verschwinden (§8).
func TestEinSymbolzugOhneSymboleWirdBenannt(t *testing.T) {
	plan := &render.Plan{Breite: 40, Hoehe: 40, Ebenen: []render.Ebene{{
		Art:     render.ArtSymbolzug,
		Rahmen:  render.Rechteck{Breite: 40, Hoehe: 40},
		Deckung: 1,
	}}}

	_, befunde, err := render.NeuerRasterer(bildmasse{da: map[string]image.Point{}}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 1 {
		t.Fatalf("die Befunde sind %+v, erwartet einer", befunde)
	}
}

// **Die `alignment:` der Symbolschrift sagt, wo das Bild in seinem Kasten sitzt.**
//
// Nicht wo der Zug im Feld sitzt — das ist die `alignment:` des Feldes. Zwei
// Ausrichtungen aus zwei Blöcken, und `magic-m15` schreibt für die Manakosten
// beide hin (`align_in_rect(align, bmp_size, sym_rect)`, `symbol_font.cpp`,
// Zeilen 279–294, Fassung f9d9356 vom 2021-07-06).
//
// Der Kasten ist hier 20 × 20, das Bild darin 10 × 10: der Abstand von 10
// waagerecht wie senkrecht macht den Kasten doppelt so groß wie das Bild.
func TestDasBildSitztImKastenDortWoDieSymbolschriftEsHinstellt(t *testing.T) {
	faelle := []struct {
		name         string
		ausrichtung  render.Ausrichtung
		x, y         int
		leerX, leerY int
	}{
		{"links oben", render.LinksOben, 2, 2, 17, 17},
		{"rechts unten", render.RechtsUnten, 17, 17, 2, 2},
		{"mitte mitte", render.MitteMitte, 7, 7, 1, 1},
		{"ohne Angabe — mittig", "", 7, 7, 1, 1},
		{"links unten", render.LinksUnten, 2, 17, 17, 2},
		{"rechts oben", render.RechtsOben, 17, 2, 2, 17},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			quelle := bildmasse{da: map[string]image.Point{"/r.png": {X: 10, Y: 10}}}

			plan := &render.Plan{Breite: 20, Hoehe: 20, Ebenen: []render.Ebene{{
				Art:    render.ArtSymbolzug,
				Rahmen: render.Rechteck{Breite: 20, Hoehe: 20},
				Abschnitte: []render.Abschnitt{{
					Text:    "R",
					Symbole: []render.Symbol{{Bild: render.Blatt("/r.png")}},
					Symbolmasse: render.Symbolmasse{
						Groesse: 10, Bildgroesse: 10,
						AbstandWaagerecht: 10, AbstandSenkrecht: 10,
						Ausrichtung: fall.ausrichtung,
					},
				}},
				Satz:    render.Satz{Ausrichtung: render.LinksOben},
				Deckung: 1,
			}}}

			bild, befunde, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
			if err != nil {
				t.Fatalf("zeichnen: %v", err)
			}

			if len(befunde) > 0 {
				t.Fatalf("Befunde: %v", befunde)
			}

			if _, _, _, a := bild.At(fall.x, fall.y).RGBA(); a == 0 {
				t.Errorf("bei (%d,%d) steht nichts", fall.x, fall.y)
			}

			// **Und an der Stelle, die frei bleiben muss, nichts.** Ohne diese
			// Probe bestünde jeder Fall, der das ganze Feld füllte.
			if _, _, _, a := bild.At(fall.leerX, fall.leerY).RGBA(); a != 0 {
				t.Errorf("bei (%d,%d) steht etwas", fall.leerX, fall.leerY)
			}
		})
	}
}

// **Ein fehlendes Bild in einem Symbolzug wird benannt.**
//
// Dieselbe Zusage wie im Text: eine Karte, der ein Manasymbol fehlt, sieht aus
// wie eine billigere (§8). Und der **Platz bleibt stehen** — das Symbol daneben
// rutscht nicht nach links, sonst zeigte der Befund auf eine Stelle, an der
// nichts mehr steht.
func TestEinFehlendesSymbolbildImZugWirdBenannt(t *testing.T) {
	quelle := bildmasse{da: map[string]image.Point{"/da.png": {X: 10, Y: 10}}}

	plan := &render.Plan{Breite: 40, Hoehe: 20, Ebenen: []render.Ebene{{
		Art:    render.ArtSymbolzug,
		Rahmen: render.Rechteck{Breite: 40, Hoehe: 20},
		Abschnitte: []render.Abschnitt{
			// **Ein Abschnitt ohne Symbole steht dazwischen** — ein Symbolfeld
			// mit Auszeichnungen zerfällt in Abschnitte, und die ohne Symbole
			// tragen nichts zum Zug bei.
			{Text: "unsichtbar"},
			{
				Text: "XR",
				Symbole: []render.Symbol{
					{Bild: render.Blatt("/fehlt.png")}, {Bild: render.Blatt("/da.png")},
				},
				Symbolmasse: render.Symbolmasse{Groesse: 10, Bildgroesse: 10},
			},
		},
		Satz:    render.Satz{Ausrichtung: render.LinksOben},
		Deckung: 1,
	}}}

	bild, befunde, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "fehlt.png") {
		t.Fatalf("die Befunde sind %+v, erwartet einer über /fehlt.png", befunde)
	}

	// Das zweite Symbol steht an seiner Stelle: von 10 bis 20, nicht von 0 an.
	if _, _, _, a := bild.At(15, 5).RGBA(); a == 0 {
		t.Error("bei (15,5) steht nichts — das zweite Symbol ist nachgerückt")
	}

	if _, _, _, a := bild.At(5, 5).RGBA(); a != 0 {
		t.Error("bei (5,5) steht etwas — der Platz des fehlenden Bildes ist belegt")
	}
}
