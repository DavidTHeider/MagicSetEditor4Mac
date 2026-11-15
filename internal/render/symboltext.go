package render

import (
	"image"
	"strings"
)

// **Ein Symbol schreibt Text auf sein Bild — `draw text:`.**
//
//	// 2. draw text
//	if (text.empty() || !sym.text_font) return;
//	sym_rect = RealRect(bmp_pos, bmp_size);
//	sym_rect.x      += font_size * sym.text_margin_left;
//	sym_rect.y      += font_size * sym.text_margin_top;
//	sym_rect.width  -= font_size * (sym.text_margin_left + sym.text_margin_right);
//	sym_rect.height -= font_size * (sym.text_margin_top  + sym.text_margin_bottom);
//	double size = font_size * sym.text_font->size;
//
// `SymbolFont::drawSymbol`, `src/data/symbol_font.cpp`, Zeilen 287–329, Fassung
// f9d9356 vom 2021-07-06.
//
// **Daran hängt jede Zahl in den Manakosten.** Ohne sie steht auf der Karte der
// leere Kreis, und die Eins in `1R` fehlt
// (#336).

// Schriftgroessenschritt ist MSEs `getFontSizeStep`.
//
// **Ein Punkt.** MSE verkleinert die Schrift in ganzen Punkten, bis der Text
// passt; bei einer Vergrößerung rechnet es in Bildpunkten des Ziels. Hier
// steht die Zahl in Kartenpunkten — dieselbe Einheit wie `size`.
const Schriftgroessenschritt = 1.0

// symboltext schreibt den Text eines Symbols in dessen Bildkasten und gibt
// zurück, was ihm dabei fehlte — **als ganze Meldungen**.
//
// **Der Kasten ist der des Bildes, nicht der des Symbols.** MSE setzt
// `sym_rect = RealRect(bmp_pos, bmp_size)` — der Abstand der Symbolschrift
// gehört nicht dazu.
func (r *Rasterer) symboltext(
	ziel *image.NRGBA, ebene Ebene, maske *maskenbild, rahmen image.Rectangle,
	symbol *Symbol, kasten Symbolkasten, groesse float64,
	bildLinks, bildOben, vergroesserung float64, gestaucht stauchflaeche,
) []string {
	if symbol.Text == "" || strings.TrimSpace(symbol.Textsatz.Schrift) == "" || r.schriften == nil {
		return nil
	}

	// **Ein ganzer Satz und kein Name**, und nicht als fehlendes Bild. Hier
	// kam bis QA-03 der bloße Schriftname zurück, und der Aufrufer schrieb
	// »die Symbolbilder fehlen:« davor — ein Mangel der Umgebung zählte als
	// Portierung: ohne `MSE_SCHRIFTWURZEL` 426 der 679 Befunde aus #1656.
	schrift, err := r.schriften.Suche(symbol.Textsatz.Schrift, false)
	if err != nil {
		return []string{"die Schrift " + symbol.Textsatz.Schrift +
			" für den Text auf einem Symbol ist nicht zu finden: " + err.Error()}
	}

	glyphen, kann := schrift.(Glyphenschrift)
	if !kann {
		return []string{"die Schrift " + symbol.Textsatz.Schrift +
			" gibt für den Text auf einem Symbol keine Glyphen her"}
	}

	satz := symbol.Textsatz

	feld := Rechteck{
		Links:  bildLinks + groesse*satz.RandLinks,
		Oben:   bildOben + groesse*satz.RandOben,
		Breite: kasten.Bildbreite - groesse*(satz.RandLinks+satz.RandRechts),
		Hoehe:  kasten.Bildhoehe - groesse*(satz.RandOben+satz.RandUnten),
	}

	grad, dehnung, breite, hoehe := eingepasst(symbol.Text, groesse*satz.Groesse, feld, satz, schrift)
	if grad <= 0 {
		return nil
	}

	// **Ausgerichtet im Kasten**, mit `middle center` als Vorgabe — MSEs
	// `text_alignment(ALIGN_MIDDLE_CENTER)`.
	ausrichtung := satz.Ausrichtung
	if ausrichtung == "" {
		ausrichtung = MitteMitte
	}

	links := feld.Links + waagerechterVersatz(ausrichtung, feld.Breite, breite)
	oben := feld.Oben + senkrechterVersatz(ausrichtung, feld.Hoehe, hoehe)

	r.zeichneSymbolschrift(ziel, ebene, maske, rahmen, symbol, glyphen,
		grad, dehnung, gestaucht.stelle(links), oben, vergroesserung, gestaucht)

	return nil
}

// eingepasst verkleinert den Grad, bis der Text in den Kasten passt.
//
//	while (true) {
//	  if (size <= 0) return; // text too small
//	  ts = dc.GetTextExtent(text);
//	  if (ts.height <= sym_rect.height) {
//	    if (ts.width <= sym_rect.width) break;
//	    else if (ts.width * sym.text_font->max_stretch <= sym_rect.width) { … }
//	  }
//	  size -= dc.getFontSizeStep();
//	}
//
// Zeilen 308–322.
//
// **`max_stretch` ist eine Schwelle, kein Faktor.** Erst wenn der Text auch
// gestaucht nicht mehr hineinpasst, wird der Grad kleiner; die Stauchung selbst
// ist dann `Kastenbreite / Textbreite`.
func eingepasst(
	text string, grad float64, feld Rechteck, satz Symboltextsatz, schrift Schrift,
) (grad2, dehnung, breite, hoehe float64) {
	for size := grad; size > 0; size -= Schriftgroessenschritt {
		gemessen := Textbreite(text, size, schrift)

		if size > feld.Hoehe {
			continue
		}

		if gemessen <= feld.Breite {
			return size, 1, gemessen, size
		}

		if gemessen*satz.Dehnung <= feld.Breite && gemessen > 0 {
			return size, feld.Breite / gemessen, feld.Breite, size
		}
	}

	return 0, 0, 0, 0
}

// zeichneSymbolschrift malt die Zeichen des Symboltexts.
func (r *Rasterer) zeichneSymbolschrift(
	ziel *image.NRGBA, ebene Ebene, maske *maskenbild, rahmen image.Rectangle,
	symbol *Symbol, schrift Glyphenschrift,
	grad, dehnung, links, oben, vergroesserung float64, gestaucht stauchflaeche,
) {
	// **Die Farbe des Symbols, nicht die der Ebene.** Ein Manakreis ist weiß,
	// und die Ziffer darin ist schwarz — die Ebene weiß von beidem nichts.
	eigene := ebene
	if symbol.Textsatz.Farbe.A > 0 {
		eigene.Farbe = symbol.Textsatz.Farbe
	}

	// **Vierfach und heruntergerechnet, wie jeder Text** — MSE zeichnet den
	// Symboltext mit `DrawTextWithShadow` und damit über
	// `draw_resampled_text` (`SymbolFont::drawWithText`,
	// `src/data/symbol_font.cpp`, Zeile 328, Fassung f9d9356 vom 2021-07-06).
	// Die Dehnung aus [eingepasst] wirkt dort erst beim Herunterrechnen.
	satz := Satz{Groesse: grad}
	lauf := neuerVierfachlauf(links*vergroesserung, oben*vergroesserung,
		Textbreite(symbol.Text, grad, schrift)*vergroesserung,
		schrift.Zeilenhoehe()*grad*vergroesserung,
		schrift.Grundlinie()*grad*vergroesserung, dehnung*gestaucht.faktor)

	versatz := 0.0
	vorher := rune(0)

	for _, zeichen := range gesetzt(symbol.Text, satz, schrift) {
		korrektur := kernungZwischen(vorher, zeichen, satz, schrift) * vergroesserung

		if glyph, ecke, da := schrift.Maske(zeichen, grad*vergroesserung*Textraster); da {
			lauf.setze(glyph, ecke, versatz+korrektur)
		}

		versatz += korrektur + schrift.Breite(zeichen)*grad*vergroesserung
		vorher = zeichen
	}

	r.lege(ziel, eigene, maske, rahmen, lauf)
}
