package render_test

import (
	"image"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Ein einzelnes Symbol als eigenes Bild** —
// #658.
//
// # Wofür
//
// `symbols_to_html` legt für jedes Manasymbol eine PNG-Datei neben die Ausgabe
// und verweist mit `<img>` darauf:
//
//	Image img = symbol_font.getImage(size, sym);
//	wxFileName fn; fn.SetPath(ei.directory_absolute); fn.SetFullName(filename);
//	img.SaveFile(fn.GetFullPath());
//
// `src/script/functions/export.cpp`, Zeilen 193–199, Fassung f9d9356 vom
// 2021-07-06.
//
// # Warum das hier steht und nicht in `service`
//
// Weil es Rasterung ist. Der Zeichenweg hatte das Symbol bisher nur **in**
// einer Karte — `zeichneSymbole` malt es an eine Stelle in einer Zeile. Für
// eine eigene Datei braucht es dieselbe Rechnung ohne Zeile und ohne Karte;
// zwei Rechnungen über dieselbe Sache liefen auseinander (§10).
//
// # Die Größe
//
//	Image resampled_image((int)(actual_size.GetWidth()  * size / img_size),
//	                      (int)(actual_size.GetHeight() * size / img_size), false);
//
// `SymbolInFont::getImage`, `src/data/symbol_font.cpp`, Zeilen 116–129 —
// dieselbe Formel, die [Symbolmasse.kasten] schon rechnet. **Ohne den
// Abstand**: `image font size` bestimmt das Bild, der Abstand nur den Platz
// daneben.
func TestSymbolAlsBild(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name                   string
		masse                  render.Symbolmasse
		bildgroesse            float64
		erwarteteBreite, hoehe int
	}{
		{
			name:            "in der Größe der Schrift",
			masse:           render.Symbolmasse{Groesse: 20, Bildgroesse: 40},
			erwarteteBreite: 20, hoehe: 10,
		},
		{
			// **Die eigene `image font size` des Symbols schlägt die der
			// Schrift** (`SymbolInFont::SymbolInFont`, Zeilen 111–113).
			name:            "mit eigener Bildgröße des Symbols",
			masse:           render.Symbolmasse{Groesse: 20, Bildgroesse: 40},
			bildgroesse:     20,
			erwarteteBreite: 40, hoehe: 20,
		},
		{
			// **Der Abstand gehört nicht zum Bild.**
			name: "der Abstand zählt nicht mit",
			masse: render.Symbolmasse{
				Groesse: 20, Bildgroesse: 40,
				AbstandWaagerecht: 6, AbstandSenkrecht: 4,
			},
			erwarteteBreite: 20, hoehe: 10,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			symbol := render.Symbol{
				Bild:        render.Blatt("mana.png"),
				Bildgroesse: fall.bildgroesse,
			}

			bild, befund := render.NeuerRasterer(bilder{"mana.png": vollflaeche(40, 20)}).
				SymbolAlsBild(symbol, fall.masse)
			if befund != "" {
				t.Fatalf("Befund: %s", befund)
			}

			if bild.Bounds().Dx() != fall.erwarteteBreite || bild.Bounds().Dy() != fall.hoehe {
				t.Errorf("%dx%d, erwartet %dx%d",
					bild.Bounds().Dx(), bild.Bounds().Dy(),
					fall.erwarteteBreite, fall.hoehe)
			}

			if gedeckteFlaeche(bild) == 0 {
				t.Error("das Bild ist leer")
			}
		})
	}
}

// **Ein fehlendes Bild wird benannt, nicht durch eine leere Datei ersetzt.**
//
// Dieselbe Regel wie im Zeichenweg (§8): eine Karte, der ein Manasymbol fehlt,
// sieht aus wie eine billigere. Eine **Datei**, die niemand füllen konnte, ist
// schlimmer — die Vorlage verwiese darauf, und im Browser stünde ein leerer
// Rahmen.
func TestEinFehlendesSymbolAlsBildWirdBenannt(t *testing.T) {
	t.Parallel()

	bild, befund := render.NeuerRasterer(bilder{}).
		SymbolAlsBild(render.Symbol{Bild: render.Blatt("fehlt.png")},
			render.Symbolmasse{Groesse: 20, Bildgroesse: 40})

	if befund == "" {
		t.Fatal("kein Befund")
	}

	if bild != nil {
		t.Error("trotz Befund kam ein Bild zurück")
	}
}

// **Der Text steht auf dem Bild** — die Ziffer im Manakreis.
//
//	if (sym.draw_text.empty() || !sym.symbol->text_font)
//	  return sym.symbol->getImage(*this, font_size);
//	// with text
//	Bitmap bmp(sym.symbol->getImage(*this, font_size));
//
// `SymbolFont::getImage`, `src/data/symbol_font.cpp`, Zeilen 331–335, Fassung
// f9d9356 vom 2021-07-06.
//
// Geprüft an der **Deckung**: das Bild ist hier durchsichtig, und jeder
// gedeckte Bildpunkt kann nur vom Text stammen.
func TestDerSymboltextStehtAufDemBild(t *testing.T) {
	t.Parallel()

	schrift := echteSchrift(t)

	symbol := render.Symbol{
		Bild: render.Blatt("leer.png"),
		Text: "7",
		Textsatz: render.Symboltextsatz{
			Schrift: "Andale", Groesse: 1, Farbe: render.Farbe{A: 255},
		},
	}

	bild, befund := render.NeuerRasterer(bilder{"leer.png": durchsichtig(40, 40)}).
		MitSchriften(schriften{"Andale": schrift}).
		SymbolAlsBild(symbol, render.Symbolmasse{Groesse: 40, Bildgroesse: 40})
	if befund != "" {
		t.Fatalf("Befund: %s", befund)
	}

	if gedeckteFlaeche(bild) == 0 {
		t.Error("die Ziffer steht nicht auf dem Bild")
	}
}

// durchsichtig gibt eine Fläche ohne Deckung.
func durchsichtig(breite, hoehe int) image.Image {
	return image.NewNRGBA(image.Rect(0, 0, breite, hoehe))
}

// **Eine fehlende Textschrift ist ein Befund, kein stummes Bild.**
//
// Auf einer Karte bleibt die Ziffer weg und die übrigen Ebenen stehen (§8).
// Hier entstünde eine **Datei**: ein leerer Manakreis, auf den die Vorlage
// verweist, und niemand sähe, dass die Sieben fehlt.
func TestEineFehlendeSymbolschriftIstEinBefund(t *testing.T) {
	t.Parallel()

	symbol := render.Symbol{
		Bild:     render.Blatt("leer.png"),
		Text:     "7",
		Textsatz: render.Symboltextsatz{Schrift: "GibtsNicht", Groesse: 1},
	}

	bild, befund := render.NeuerRasterer(bilder{"leer.png": durchsichtig(40, 40)}).
		MitSchriften(schriften{}).
		SymbolAlsBild(symbol, render.Symbolmasse{Groesse: 40, Bildgroesse: 40})

	if befund == "" || bild != nil {
		t.Errorf("Befund %q, Bild %v", befund, bild != nil)
	}
}

// **Und dasselbe für das Symbolbild.**
//
// Die Gegenrichtung zum Test in `bildausbaum_test.go`: derselbe Fehler, dieselbe
// Ursache, zwei Einstiege. Ein Manasymbol, dessen Bildbaum eine ungebaute
// Verknüpfung trägt, wird gezeichnet — und die Meldung darüber gehört in die
// Ausfuhr.
func TestEinTeilbefundAmSymbolGehtNichtVerloren(t *testing.T) {
	t.Parallel()

	symbol := render.Symbol{
		Bild: render.Bildbaum{
			Knotenart: "rotate",
			Kinder: []render.Bildkind{
				{Name: "input", Baum: render.Blatt("mana.png")},
			},
		},
	}

	bild, befund := render.NeuerRasterer(bilder{"mana.png": vollflaeche(40, 20)}).
		SymbolAlsBild(symbol, render.Symbolmasse{Groesse: 20, Bildgroesse: 40})

	if bild == nil {
		t.Fatal("kein Bild")
	}

	if befund == "" {
		t.Error("der Befund über die ungebaute Verknüpfung fehlt")
	}
}
