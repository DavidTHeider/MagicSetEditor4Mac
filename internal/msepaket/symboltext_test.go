package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein Symbol darf Text auf sein Bild zeichnen — `draw text:`.**
//
//	int draw_text; ///< The index of the captured regex expression to draw,
//	               ///  or -1 to not draw text
//
// `SymbolInFont`, `src/data/symbol_font.cpp`, Zeile 84, Fassung f9d9356 vom
// 2021-07-06. Die Zahl ist die **Gruppennummer** des Musters:
//
//	code: ([XYZ0-9])/W
//	regex: yes
//	draw text: 1
//	text margin left: -.25
//	text font:
//		name: MPlantin
//		size: .6
//		max stretch: 0.5
//
// **Daran hängt jede Zahl in den Manakosten.** Ohne sie zeichnet dieses
// Programm den leeren Kreis und lässt die Ziffer weg — auf jeder Karte mit
// generischem Mana (#336).
func TestEinSymbolDarfTextAufSeinBildZeichnen(t *testing.T) {
	pfad := probesymbolschrift(t, `mse version: 2.0.0
image font size: 135
symbol:
	code: ([XYZ0-9])/W
	regex: yes
	image: mana_nw.png
	draw text: 1
	text margin left: -.25
	text margin right: .2
	text margin top: -.14
	text margin bottom: .4
	text alignment: top left
	text font:
		name: MPlantin
		size: .6
		color: black
		max stretch: 0.5
symbol:
	code: W
	image: mana_w.png
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	schrift, befunde := paket.Symbolmodell()
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(schrift.Symbole) != 2 {
		t.Fatalf("%d Symbole, erwartet zwei", len(schrift.Symbole))
	}

	mit := schrift.Symbole[0]
	if mit.Textgruppe != 1 {
		t.Errorf("Textgruppe %d, erwartet 1", mit.Textgruppe)
	}

	if mit.Textschrift.Name != "MPlantin" || mit.Textschrift.Groesse != 0.6 {
		t.Errorf("Textschrift %+v", mit.Textschrift)
	}

	if mit.Textschrift.Dehnung != 0.5 {
		t.Errorf("max stretch %g, erwartet 0,5", mit.Textschrift.Dehnung)
	}

	if mit.TextrandLinks != -0.25 || mit.TextrandUnten != 0.4 {
		t.Errorf("Ränder %+v", mit)
	}

	if mit.Textausrichtung != "top left" {
		t.Errorf("Ausrichtung %q", mit.Textausrichtung)
	}

	// **Ohne `draw text:` ist die Gruppe minus eins** — MSEs Vorgabe, und sie
	// heißt „kein Text". Null wäre die **ganze** Trefferzeichenkette und damit
	// etwas anderes.
	if ohne := schrift.Symbole[1]; ohne.Textgruppe != -1 {
		t.Errorf("Textgruppe ohne `draw text:` ist %d, erwartet -1", ohne.Textgruppe)
	}
}

// **Ein `draw text:`, das keine Zahl ist, heißt „kein Text".**
//
// MSE liest die Angabe als `int`; was sich nicht lesen lässt, bleibt bei der
// Vorgabe −1. Eine Vorlage mit `draw text: ja` ist falsch geschrieben, und ein
// erfundener Text verdeckte das (§8).
func TestEinDrawTextOhneZahlHeisstKeinText(t *testing.T) {
	pfad := probesymbolschrift(t, `mse version: 2.0.0
symbol:
	code: W
	image: mana_w.png
	draw text: ja
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	schrift, _ := paket.Symbolmodell()
	if len(schrift.Symbole) != 1 || schrift.Symbole[0].Textgruppe != -1 {
		t.Errorf("Textgruppe: %+v", schrift.Symbole)
	}
}
