package msepaket

import (
	"strconv"
	"strings"
)

// **Ein Symbol darf Text auf sein Bild zeichnen — `draw text:`.**
//
// Die Zahl ist die Gruppennummer des Musters; ihr Treffer wird in das Bild
// geschrieben, in `text font:`, innerhalb der vier `text margin`-Ränder.
//
// `SymbolFont::drawSymbol`, `src/data/symbol_font.cpp`, Zeilen 287–329, Fassung
// f9d9356 vom 2021-07-06.
//
// **Daran hängt jede Zahl in den Manakosten** — ohne sie steht auf der Karte
// der leere Kreis (#336).

// Symboltextschrift ist das `text font:` eines Symbols.
type Symboltextschrift struct {
	Name  string
	Farbe string

	// Groesse ist ein **Anteil** der Symbolgröße, nicht eine Punktzahl:
	// `size: .6` heißt sechs Zehntel (`double size = font_size * text_font->size`).
	Groesse float64

	// Dehnung ist `max stretch` — wie weit der Text gestaucht werden darf,
	// bevor MSE die Schriftgröße verkleinert.
	Dehnung float64
}

// symbolMitText liest einen `symbol:`-Block samt seiner Textangaben.
func symbolMitText(knoten *Knoten, code, bild string) Symbol {
	symbol := Symbol{
		Code: code, Bild: bild, Stelle: knoten.Stelle,
		Bildgroesse: zahlOderNull(knoten.Ersten("image font size")),
		Bedingung:   strings.TrimSpace(wertOderLeer(knoten.Ersten("enabled"))),
		Ausdruck:    jaOderNein(knoten.Ersten("regex")),

		// **Minus eins heißt „kein Text"** — MSEs Vorgabe.
		Textgruppe:      textgruppe(knoten.Ersten("draw text")),
		Textausrichtung: strings.TrimSpace(wertOderLeer(knoten.Ersten("text alignment"))),
		TextrandLinks:   zahlOderNull(knoten.Ersten("text margin left")),
		TextrandRechts:  zahlOderNull(knoten.Ersten("text margin right")),
		TextrandOben:    zahlOderNull(knoten.Ersten("text margin top")),
		TextrandUnten:   zahlOderNull(knoten.Ersten("text margin bottom")),
	}

	if schrift := knoten.Ersten("text font"); schrift != nil {
		symbol.Textschrift = Symboltextschrift{
			Name:    strings.TrimSpace(wertOderLeer(schrift.Ersten("name"))),
			Farbe:   strings.TrimSpace(wertOderLeer(schrift.Ersten("color"))),
			Groesse: zahlOderNull(schrift.Ersten("size")),
			Dehnung: zahlOderNull(schrift.Ersten("max stretch")),
		}
	}

	return symbol
}

// textgruppe liest `draw text:`; ohne Angabe gilt minus eins.
func textgruppe(knoten *Knoten) int {
	if knoten == nil {
		return -1
	}

	nummer, err := strconv.Atoi(strings.TrimSpace(knoten.Wert))
	if err != nil {
		return -1
	}

	return nummer
}
