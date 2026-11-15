package domain_test

// **Was ein Standardland ist** —
// #1370.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Regel steht in den Comprehensive Rules und nirgends sonst.**
//
// CR 205.4c: ein Land mit dem Supertyp „basic" ist ein Standardland; jedes
// Land ohne diesen Supertyp ist keines — **auch dann nicht, wenn es einen
// Landgrundtyp trägt**. — der zweite Satz ist der Grund, warum „Dryad Arbor"
// hier durchfällt, obwohl `Forest` darin steht.
//
// CR 205.4a: die Supertypen stehen unmittelbar vor den Kartentypen, und es
// sind fünf — basic, legendary, ongoing, snow, world. — deshalb
// wird **vor** dem Gedankenstrich gesucht.
func TestWasEinStandardlandIst(t *testing.T) {
	for _, fall := range []struct {
		name     string
		typzeile string
		erwartet bool
	}{
		// **Gemessen am 2026-09-01** an `api.scryfall.com/cards/named`.
		{"Forest", "Basic Land — Forest", true},
		{"Wastes — ohne Untertyp", "Basic Land", true},
		{"Snow-Covered Forest", "Basic Snow Land — Forest", true},

		// **Dryad Arbor** trägt `Forest` als Untertyp und ist trotzdem keines
		// — CR 205.4c, zweiter Satz.
		{"Dryad Arbor", "Land Creature — Forest Dryad", false},

		{"ein gewöhnliches Land", "Land", false},
		{"eine Kreatur", "Creature — Human Soldier", false},

		// **Ein legendäres Land ist keines** — `legendary` ist ein anderer
		// Supertyp.
		{"ein legendäres Land", "Legendary Land", false},

		// **Ohne Typzeile ist nichts bekannt** — und Unbekanntes wird
		// gezählt, nicht ausgenommen.
		{"leer", "", false},

		// **`Basic` ohne `Land`** gibt es nicht, und wenn es käme, wäre es
		// kein Land.
		{"Basic ohne Land", "Basic Creature — Human", false},

		// **Nur nach dem Strich** zählt nicht: dort stehen Untertypen.
		{"Basic nur im Untertyp", "Land — Basic Forest", false},
	} {
		t.Run(fall.name, func(t *testing.T) {
			if got := domain.IstStandardland(fall.typzeile); got != fall.erwartet {
				t.Errorf("IstStandardland(%q) = %v, erwartet %v",
					fall.typzeile, got, fall.erwartet)
			}
		})
	}
}

// **Standardländer zählen nicht gegen die Höchstzahl gleicher Karten.**
//
// CR 100.2a: ein Constructed-Deck trägt **Standardländer beliebig oft** und
// jede andere Karte eines englischen Namens höchstens viermal — **die
// Ausnahme gilt nur für Standardländer**.
func TestStandardlaenderZaehlenNichtGegenDieHoechstzahl(t *testing.T) {
	format := domain.Format{
		Name: "Konstruiert", HoechstzahlGleich: 4, Sideboardgroesse: 15,
	}

	posten := []domain.Deckposten{
		{Platz: domain.Deckplatz{
			ID: 1, Printing: "wald", Menge: 24, Bereich: domain.BereichHauptdeck,
		}, Standardland: true},
		{Platz: domain.Deckplatz{
			ID: 2, Printing: "blitz", Menge: 4, Bereich: domain.BereichHauptdeck,
		}},
	}

	if befunde := domain.PruefeGroessen(posten, format); len(befunde) != 0 {
		t.Errorf("vierundzwanzig Wald wurden beanstandet: %v", befunde)
	}
}

// **Und alles andere zählt weiter.**
func TestNichtlaenderZaehlenWeiter(t *testing.T) {
	format := domain.Format{Name: "Konstruiert", HoechstzahlGleich: 4}

	posten := []domain.Deckposten{{Platz: domain.Deckplatz{
		ID: 1, Printing: "blitz", Menge: 5, Bereich: domain.BereichHauptdeck,
	}}}

	if befunde := domain.PruefeGroessen(posten, format); len(befunde) != 1 {
		t.Errorf("%d Befunde, erwartet 1: %v", len(befunde), befunde)
	}
}

// **Posten ohne Auskunft nimmt nichts aus** — und das ist der ehrliche
// Vorgabewert.
//
// Wer nichts über die Typen weiß, weiß auch nicht, was ein Standardland ist.
// Eine Ausnahme auf Verdacht ließe ein Deck mit fünf gleichen Nicht-Ländern
// durchgehen.
func TestPostenOhneAuskunftNimmtNichtsAus(t *testing.T) {
	format := domain.Format{Name: "Konstruiert", HoechstzahlGleich: 4}

	deck := domain.Deck{Plaetze: []domain.Deckplatz{
		{ID: 1, Printing: "wald", Menge: 24, Bereich: domain.BereichHauptdeck},
	}}

	if befunde := domain.PruefeGroessen(domain.Posten(deck), format); len(befunde) != 1 {
		t.Errorf("%d Befunde, erwartet 1", len(befunde))
	}
}

// **Vier verschiedene Drucke derselben Karte sind vier Karten desselben
// Namens** — Befund 1 des Zweitblicks über Phase 4.6.
//
// # Die Regel, wörtlich
//
// CR 100.2a: höchstens vier je **englischem Namen**, Standardländer
// ausgenommen — und für den Deckbau gelten austauschbare Namen als derselbe
// englische Name.
//
// CR 903.5b: außer Standardländern trägt jede Karte eines Commanderdecks
// einen **anderen englischen Namen**.
//
// # Was der erste Anlauf zählte
//
// **Den Druck.** Vier verschiedene Drucke von „Lightning Bolt" waren vier
// Gruppen zu je einem — und kamen durch. Im Commander umging dasselbe die
// Ein-Karten-Regel vollständig.
func TestVierDruckeDerselbenKarteSindVier(t *testing.T) {
	format := domain.Format{
		Name: "Konstruiert", HoechstzahlGleich: 4,
		Sideboardgroesse: 15,
	}

	posten := make([]domain.Deckposten, 0, 5)

	for stelle := range 5 {
		posten = append(posten, domain.Deckposten{
			Platz: domain.Deckplatz{
				ID:       int64(stelle + 1),
				Printing: fmt.Sprintf("druck-%d", stelle),
				Menge:    1, Bereich: domain.BereichHauptdeck,
			},
			// **Derselbe Name, fünf verschiedene Drucke.**
			Name: "Lightning Bolt",
		})
	}

	befunde := domain.PruefeGroessen(posten, format)
	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %v", len(befunde), befunde)
	}

	if !strings.Contains(befunde[0].Meldung, "Lightning Bolt") {
		t.Errorf("die Meldung nennt die Karte nicht: %q", befunde[0].Meldung)
	}
}

// **Und ohne bekannten Namen entscheidet der Druck** — die ehrliche Vorgabe.
//
// Wer den Namen nicht kennt, kann nicht nach ihm gruppieren. Zwei verschiedene
// Drucke ohne Namen als dieselbe Karte zu zählen wäre eine Erfindung; sie
// getrennt zu zählen ist das, was man weiß.
func TestOhneNamenEntscheidetDerDruck(t *testing.T) {
	format := domain.Format{
		Name: "Konstruiert", HoechstzahlGleich: 4,
		Sideboardgroesse: 15,
	}

	verschieden := []domain.Deckposten{
		{Platz: domain.Deckplatz{
			ID: 1, Printing: "a", Menge: 4,
			Bereich: domain.BereichHauptdeck,
		}},
		{Platz: domain.Deckplatz{
			ID: 2, Printing: "b", Menge: 4,
			Bereich: domain.BereichHauptdeck,
		}},
	}

	if befunde := domain.PruefeGroessen(verschieden, format); len(befunde) != 0 {
		t.Errorf("zwei namenlose Drucke wurden zusammengezählt: %v", befunde)
	}

	derselbe := []domain.Deckposten{
		{Platz: domain.Deckplatz{
			ID: 1, Printing: "a", Menge: 4,
			Bereich: domain.BereichHauptdeck,
		}},
		{Platz: domain.Deckplatz{
			ID: 2, Printing: "a", Menge: 1,
			Bereich: domain.BereichSideboard,
		}},
	}

	if befunde := domain.PruefeGroessen(derselbe, format); len(befunde) != 1 {
		t.Errorf("%d Befunde, erwartet 1", len(befunde))
	}
}
