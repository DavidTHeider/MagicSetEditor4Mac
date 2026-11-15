package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// kartemit gibt eine Karte mit genau einem gefüllten Feld.
func kartemit(feld, wert string) *domain.CardRow {
	return &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{{Field: feld, Text: wert}}}
}

// **`right:` ohne `left:` verankert die Fläche rechts** — und `bottom:` ohne
// `top:` unten
// (#240).
//
// # Der Befund
//
// `magic-m15.mse-style` legt die Manakosten so an:
//
//	casting cost:
//	    right: 346
//	    top: 29
//	    width: { max(30, card_style.casting_cost.content_width) + 5 }
//
// Es gibt **kein `left:`**. Dieses Programm nahm dafür null und rechnete die
// Breite aus — die Kosten landeten am linken Kartenrand statt rechts oben, wo
// sie auf jeder Magic-Karte stehen. Am gerenderten Bild war es zu sehen.
//
// # Die Regel
//
// Zwei der drei Angaben bestimmen die dritte. Steht die Länge und die
// **Gegen**kante, ist der Anfang ihre Differenz.
func TestRechtsOhneLinksVerankertRechts(t *testing.T) {
	faelle := []struct {
		name                              string
		links, rechts, breite             string
		erwarteterAnfang, erwarteteBreite float64
	}{
		{"links und Breite", "10", "", "40", 10, 40},
		{"links und rechts", "10", "60", "", 10, 50},

		// **Der Fall aus dem Pack.**
		{"rechts und Breite", "", "346", "35", 311, 35},

		// Ohne alles bleibt es der alte, gutmütige Weg: ab null bis zum Rand.
		{"nichts davon", "", "", "", 0, 375},

		// **Eine ausdrückliche Null ist keine fehlende Angabe.** Wer `left: 0`
		// schreibt, meint den linken Rand — auch neben einem `right:`.
		{"ausdrücklich null", "0", "346", "", 0, 346},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			feld := msepaket.Feldstil{
				Feld: "casting cost", Oben: "0", Hoehe: "20",
				Links: fall.links, Rechts: fall.rechts, Breite: fall.breite,
				Schrift: msepaket.Schrift{Name: "Beleren", Groesse: "12"},
			}

			karte := kartemit("casting cost", "2R")

			plan, befunde := gedeutetMit(t, karte, bildstil(feld))
			if len(befunde) > 0 {
				t.Fatalf("Befunde: %v", befunde)
			}

			rahmen := ebenen(t, plan)[0].Rahmen

			if rahmen.Links != fall.erwarteterAnfang || rahmen.Breite != fall.erwarteteBreite {
				t.Errorf("links=%v breite=%v, erwartet waren %v und %v",
					rahmen.Links, rahmen.Breite, fall.erwarteterAnfang, fall.erwarteteBreite)
			}
		})
	}
}

// **Dasselbe senkrecht.** `bottom:` ohne `top:` verankert unten.
func TestUntenOhneObenVerankertUnten(t *testing.T) {
	feld := msepaket.Feldstil{
		Feld: "pt", Links: "0", Breite: "60",
		Unten: "500", Hoehe: "28",
		Schrift: msepaket.Schrift{Name: "Beleren", Groesse: "12"},
	}

	plan, befunde := gedeutetMit(t, kartemit("pt", "3/3"), bildstil(feld))
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if oben := ebenen(t, plan)[0].Rahmen.Oben; oben != 472 {
		t.Errorf("oben=%v, erwartet war 472", oben)
	}
}
