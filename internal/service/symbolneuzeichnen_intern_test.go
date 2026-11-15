package service

// **Gezielt neu zeichnen** — DOM-02.AK-03
// (#1697):
// „In Karten und Stilen eingesetzte Symbole werden nach Änderung gezielt neu
// gerendert; unbeteiligte Objekte werden nicht unnötig invalidiert."
//
// **Ein innerer Lauf**, weil er den Weg einer Karte nachgeht: der Bildknoten
// `symbol_variation` liest die Datei über die Ablage des Sets
// ([SymbolquelleAusAblage]) und trägt ihre Teile in den Plan, und der
// Schlüssel eines gerasterten Bildes geht über den Plan ([Texturschluessel]).
// Nachziehen schreibt die Datei **unter ihrem Namen** neu — genau das tut
// dieser Lauf.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// feste Abdrücke — die Dateien, die der Plan nur nennt, ändern sich hier nicht.
type festeRessourcen struct{}

func (festeRessourcen) Bildabdruck(string) string    { return "bild" }
func (festeRessourcen) Schriftabdruck(string) string { return "schrift" }
func (festeRessourcen) Rueckenbild(string) string    { return "" }

// kartenschluessel gibt den Schlüssel des Bildes einer Karte, die das
// Setsymbol in der Fassung `common` zeigt.
func kartenschluessel(bilder *Bildablage, setID int64, datei string) string {
	baum, _ := fassungsbild(datei, msepaket.Symbolfassung{Name: "common", Fuellfarbe: "black"}).AlsBild()
	plan := &render.Plan{
		Breite: 20, Hoehe: 20,
		Ebenen: []render.Ebene{{
			Art:     render.ArtSymbol,
			Rahmen:  render.Rechteck{Breite: 20, Hoehe: 20},
			Bild:    alsZeichenbaum(baum, "", SymbolquelleAusAblage(bilder.Verzeichnis(setID), nil, "")),
			Deckung: 1,
		}},
	}

	return Texturschluessel(plan, festeRessourcen{}, Texturwunsch{Seite: "vorne", Guete: GueteTisch, Gesichter: 1})
}

func TestNachziehenZeichnetNurDieBetroffenenNeu(t *testing.T) {
	bilder := NeueBildablage(t.TempDir())
	ablage := NeueSymbolablage(bilder)

	eins, err := ablage.Ablegen(1, msepaket.Standardsymbol(), "")
	if err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	anders := msepaket.Standardsymbol()
	anders.Teile[0].Punkte[1].Stelle.Y += 0.2

	zwei, err := ablage.Ablegen(2, anders, "")
	if err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	vorherEins, vorherZwei := kartenschluessel(bilder, 1, eins), kartenschluessel(bilder, 2, zwei)

	// **Dasselbe ohne Änderung gibt denselben Schlüssel** — sonst wäre jedes
	// Bild bei jedem Blick neu zu zeichnen.
	if kartenschluessel(bilder, 1, eins) != vorherEins {
		t.Fatal("derselbe Stand gab einen anderen Schlüssel")
	}

	neu := msepaket.Standardsymbol()
	neu.Teile[0].Punkte[0].Stelle.X += 0.1

	if _, err := ablage.Ablegen(1, neu, eins); err != nil {
		t.Fatalf("nachziehen: %v", err)
	}

	if kartenschluessel(bilder, 1, eins) == vorherEins {
		t.Error("die Karte des nachgezogenen Sets behielt ihren Schlüssel")
	}

	if kartenschluessel(bilder, 2, zwei) != vorherZwei {
		t.Error("die Karte des unbeteiligten Sets bekam einen neuen Schlüssel")
	}
}
