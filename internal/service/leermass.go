package service

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// leermass ist das Maß eines **leeren** Felds, wie MSE2 es meldet.
//
// # Ein leerer Text ist eine leere Zeile
//
// MSE2 nimmt die drei Maße aus dem gesetzten Text (`reflect_layout`,
// `src/data/field/text.cpp`, Zeilen 114–121), und auch ein leerer Text wird
// gesetzt: eine Zeile ohne Breite, mit der Höhe der Schrift
// (`update_size`, `src/render/text/viewer.cpp`, Zeilen 350–353; Fassung
// f9d9356 vom 2021-07-06). Gefragt am 2026-10-11 über einen Orakelstil
// (`leerinhalt_test.go`).
//
// # Eine leere Wahl misst ihr Bild wie eine gefüllte
//
// Sie kommt hier gar nicht an ([wahlmass]): für den leeren Wert ergibt die
// Nachschlagetafel der Wahlbilder `nil`, und das ist bei MSE2 ein leeres Bild
// in der Größe des Kastens — gefragt am 2026-10-11 über einen Orakelstil
// (#2303). Hier stand bis dahin, eine leere Wahl habe Nullmaße.
//
// **Für jede andere Art bleibt es beim Kasten**, wie bisher: ein Bild- oder
// Symbolfeld führt bei MSE2 gar kein `content_width`, und eine Vorrichtung
// ohne Spiel sagt nicht, welche Art ein Feld hat.
func leermass(
	ctx context.Context, feld *msepaket.Feldstil, vorlage Vorlage,
	schriften render.Schriftquelle, bilder render.Bildquelle, umgebung *msescript.Umgebung,
) (Inhaltsmass, bool) {
	switch feldart(vorlage, feld.Feld) {
	case domain.KindText:
		// **Derselbe Schnitt wie beim gefüllten Feld** ([inhaltsmasse]): ein
		// Symbolzug ist es nur, wo der Stil `always symbol:` sagt.
		if name := strings.TrimSpace(feld.Symbolschrift.Name); name != "" && feld.ImmerSymbole {
			return leerSymbolmass(ctx, feld, name, vorlage, bilder, umgebung)
		}

		gesetzt, da := textsatz(ctx, feld, []rohabschnitt{{}}, vorlage, schriften, bilder, umgebung)
		if !da {
			return Inhaltsmass{}, false
		}

		return Inhaltsmass{Hoehe: gesetzt.Hoehe, Zeilen: len(gesetzt.Zeilen), Satz: gesetzt}, true

	default:
		return Inhaltsmass{}, false
	}
}

// leerSymbolmass ist das Maß eines leeren Symbolzugs — die Manakosten eines
// Lands.
//
// MSE2 gibt der leeren Zeile dann die Höhe seines **Ersatzsymbols**: des
// ersten Musters, das „0" trifft, samt Abstand; gibt es keines, einen
// Bildpunkt samt Abstand (`TextViewer::prepareLines`,
// `src/render/text/viewer.cpp`, Zeilen 397–401; `SymbolFont::defaultSymbol`
// und `defaultSymbolSize`, `src/data/symbol_font.cpp`, Zeilen 263–268 und
// 403–410; Fassung f9d9356 vom 2021-07-06).
//
// **Portierungsmarke adapted**: MSE2 prüft dabei auch `enabled:`; das
// Zerlegen hier kennt die Bedingung eines Symbols an keiner Stelle
// ([ersterTreffer]), und eine zweite Antwort nur für das Ersatzsymbol wäre
// eine Abweichung zwischen Messen und Zeichnen.
func leerSymbolmass(
	ctx context.Context, feld *msepaket.Feldstil, name string, vorlage Vorlage,
	bilder render.Bildquelle, umgebung *msescript.Umgebung,
) (Inhaltsmass, bool) {
	schrift, da := vorlage.Symbolschriften[name]
	if !da {
		return Inhaltsmass{}, false
	}

	seite := symbolseiteVorSatz(ctx, feld, umgebung)
	if seite <= 0 {
		return Inhaltsmass{}, false
	}

	symbol, da := ersatzsymbol(schrift)
	if !da {
		return Inhaltsmass{Hoehe: 1 + schrift.AbstandSenkrecht, Zeilen: 1}, true
	}

	_, hoehe := render.Symbolzugmass(
		symbolbilder(ctx, []Symboltreffer{{Symbol: symbol, Text: "0"}}, name, umgebung),
		render.Symbolmasse{
			Groesse:           seite,
			Bildgroesse:       schrift.Bildgroesse,
			AbstandWaagerecht: schrift.AbstandWaagerecht,
			AbstandSenkrecht:  schrift.AbstandSenkrecht,
		},
		bilder)

	return Inhaltsmass{Hoehe: hoehe, Zeilen: 1}, true
}

// ersatzsymbol ist das erste Symbol mit Muster, das „0" trifft —
// `SymbolFont::defaultSymbol`.
func ersatzsymbol(schrift msepaket.Symbolschrift) (msepaket.Symbol, bool) {
	for i := range schrift.Symbole {
		symbol := schrift.Symbole[i]
		if !symbol.Ausdruck {
			continue
		}

		gelesen, err := muster(symbol.Code)
		if err != nil {
			continue
		}

		if _, _, passt := gelesen.treffer("0"); passt {
			return symbol, true
		}
	}

	return msepaket.Symbol{}, false
}
