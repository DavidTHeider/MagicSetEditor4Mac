package domain_test

// **Das Urteil über ein Deck** — Werkstück 5 von Phase 4.6
// (#1293),
// nach Vermerk 0061.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Deck, das jede Zahl einhält, hat keinen Größenbefund.**
func TestEinDeckInDerRichtigenGroesse(t *testing.T) {
	format := domain.Format{
		Name: "Konstruiert", Spiel: "magic",
		Mindestgroesse: 60, Sideboardgroesse: 15, HoechstzahlGleich: 4,
	}

	// **Fünfzehn verschiedene Karten und keine öfter als viermal** — ein
	// Deck, das die Höchstzahl einhält, muss aus verschiedenen Karten
	// bestehen. Der erste Anlauf legte 56 gleiche hinein und war damit ein
	// Verstoß, kein Vergleichsfall.
	plaetze := make([]domain.Deckplatz, 0, 19)
	for stelle := range 15 {
		plaetze = append(plaetze, domain.Deckplatz{
			Printing: fmt.Sprintf("haupt-%d", stelle),
			Menge:    4, Bereich: domain.BereichHauptdeck,
		})
	}

	for stelle := range 4 {
		plaetze = append(plaetze, domain.Deckplatz{
			Printing: fmt.Sprintf("side-%d", stelle),
			Menge:    4 - stelle, Bereich: domain.BereichSideboard,
		})
	}

	deck := domain.Deck{Spiel: "magic", Plaetze: plaetze}

	if befunde := domain.PruefeGroessen(domain.Posten(deck), format); len(befunde) != 0 {
		t.Errorf("%v", befunde)
	}
}

// **Jede Zahl bekommt ihren eigenen Befund** — und jeder nennt sein Feld.
func TestJedeGroessenregelMeldetSich(t *testing.T) {
	for _, fall := range []struct {
		name   string
		format domain.Format
		deck   domain.Deck
		feld   string
	}{
		{
			"zu wenige im Hauptdeck",
			domain.Format{Mindestgroesse: 60},
			domain.Deck{Plaetze: []domain.Deckplatz{
				{Printing: "a", Menge: 59, Bereich: domain.BereichHauptdeck},
			}},
			"hauptdeck",
		},
		{
			"zu viele im Hauptdeck",
			domain.Format{Mindestgroesse: 100, Hoechstgroesse: 100},
			domain.Deck{Plaetze: []domain.Deckplatz{
				{Printing: "a", Menge: 101, Bereich: domain.BereichHauptdeck},
			}},
			"hauptdeck",
		},
		{
			"zu viele im Sideboard",
			domain.Format{Sideboardgroesse: 15},
			domain.Deck{Plaetze: []domain.Deckplatz{
				{Printing: "a", Menge: 16, Bereich: domain.BereichSideboard},
			}},
			"sideboard",
		},
		{
			"ein Sideboard, wo keines vorgesehen ist",
			domain.Format{Sideboardgroesse: 0, Mindestgroesse: 100},
			domain.Deck{Plaetze: []domain.Deckplatz{
				{Printing: "a", Menge: 100, Bereich: domain.BereichHauptdeck},
				{Printing: "b", Menge: 1, Bereich: domain.BereichSideboard},
			}},
			"sideboard",
		},
		{
			"zu viele gleiche",
			domain.Format{HoechstzahlGleich: 4},
			domain.Deck{Plaetze: []domain.Deckplatz{
				{Printing: "a", Menge: 5, Bereich: domain.BereichHauptdeck},
			}},
			"gleiche",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			befunde := domain.PruefeGroessen(domain.Posten(fall.deck), fall.format)
			if len(befunde) != 1 {
				t.Fatalf("%d Befunde, erwartet 1: %v", len(befunde), befunde)
			}

			if befunde[0].Feld != fall.feld {
				t.Errorf("Feld %q, erwartet %q", befunde[0].Feld, fall.feld)
			}
		})
	}
}

// **Null ist keine Grenze** — bei drei der vier Zahlen.
//
// # Die vierte ist das Sideboard, und sie ist die Ausnahme
//
// `sideboardgroesse: 0` heißt **kein Sideboard**, nicht „unbegrenzt". So steht
// es seit Werkstück 4 an `domain.Format` und in der Wanderung 0029, und der
// Grund ist CR 903.5e: Commanderpartien kennen kein Sideboard. Ohne diese
// Lesart bräuchte Commander ein zusätzliches Feld, das nur „nein" sagt.
//
// **Der erste Anlauf dieses Laufs hat das übersehen** und ein Sideboard in ein
// Format ohne eines gelegt — er widersprach damit einem anderen Fall derselben
// Datei. Die Dokumentation entschied.
func TestNullIstKeineGrenze(t *testing.T) {
	offen := domain.Format{Name: "Offen", Spiel: "magic", Sideboardgroesse: 99}

	deck := domain.Deck{Plaetze: []domain.Deckplatz{
		{Printing: "a", Menge: 99, Bereich: domain.BereichHauptdeck},
		{Printing: "b", Menge: 99, Bereich: domain.BereichSideboard},
	}}

	if befunde := domain.PruefeGroessen(domain.Posten(deck), offen); len(befunde) != 0 {
		t.Errorf("null wurde als Grenze gelesen: %v", befunde)
	}
}

// **Hauptdeck und Sideboard werden getrennt gezählt** — dieselbe Karte darf
// in beiden stehen, und die Höchstzahl gilt je Bereich.
//
// # Warum je Bereich und nicht über beides
//
// **Weil es nachgelesen ist und die Regel das Gegenteil sagt.** CR 100.4a:
// die Vierergrenze aus CR 100.2a gilt für Deck und Sideboard **zusammen**.
// Genau deshalb steht hier ein Lauf, der es festhält — vier im
// Hauptdeck und eine im Sideboard sind **fünf** und ein Befund.
func TestDieHoechstzahlGiltUeberBeideBereiche(t *testing.T) {
	format := domain.Format{HoechstzahlGleich: 4, Sideboardgroesse: 15}

	deck := domain.Deck{Plaetze: []domain.Deckplatz{
		{Printing: "a", Menge: 4, Bereich: domain.BereichHauptdeck},
		{Printing: "a", Menge: 1, Bereich: domain.BereichSideboard},
	}}

	befunde := domain.PruefeGroessen(domain.Posten(deck), format)
	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %v", len(befunde), befunde)
	}

	if !strings.Contains(befunde[0].Meldung, "5") {
		t.Errorf("die Meldung nennt die Zahl nicht: %q", befunde[0].Meldung)
	}
}

// **Ein Urteil ist zulässig, wenn nichts dagegen steht.**
func TestWannEinUrteilZulaessigIst(t *testing.T) {
	for _, fall := range []struct {
		name     string
		urteil   domain.Deckurteil
		erwartet bool
	}{
		{"nichts dagegen", domain.Deckurteil{}, true},
		{"ein Größenbefund", domain.Deckurteil{
			Groessen: []domain.Groessenbefund{{Feld: "hauptdeck"}},
		}, false},
		{"eine verbotene Karte", domain.Deckurteil{
			Urteile: []domain.Kartenurteil{{Zulaessigkeit: domain.ZulaessigVerboten}},
		}, false},
		{"eine nicht gelistete Karte", domain.Deckurteil{
			Urteile: []domain.Kartenurteil{{Zulaessigkeit: domain.ZulaessigNichtGelistet}},
		}, false},
		// **`ungeprüft` ist kein `nicht zulässig`** — das ist die Zusage des
		// Vermerks. Eine eigene Karte ist in keinem Format der Welt gelistet;
		// sie deshalb rot zu färben hieße, jedem eigenen Deck zu sagen, es sei
		// kaputt.
		{"eine ungeprüfte Karte", domain.Deckurteil{
			Urteile: []domain.Kartenurteil{{Zulaessigkeit: domain.ZulaessigUngeprueft}},
		}, true},
		// **`restricted` ebenso** — wie oft eine eingeschränkte Karte
		// vorkommen darf, steht in den Turnierregeln und nicht in den
		// Comprehensive Rules. Dieses Programm meldet den Zustand und erfindet
		// keine Zahl.
		{"eine eingeschränkte Karte", domain.Deckurteil{
			Urteile: []domain.Kartenurteil{{Zulaessigkeit: domain.ZulaessigEingeschraenkt}},
		}, true},
	} {
		t.Run(fall.name, func(t *testing.T) {
			if got := fall.urteil.Zulaessig(); got != fall.erwartet {
				t.Errorf("Zulaessig() = %v, erwartet %v", got, fall.erwartet)
			}
		})
	}
}

// **Und es sagt, wie viel es nicht wissen konnte.**
func TestEinUrteilZaehltDasUngepruefte(t *testing.T) {
	urteil := domain.Deckurteil{Urteile: []domain.Kartenurteil{
		{Zulaessigkeit: domain.ZulaessigLegal, Menge: 4},
		{Zulaessigkeit: domain.ZulaessigUngeprueft, Menge: 3},
		{Zulaessigkeit: domain.ZulaessigUngeprueft, Menge: 2},
	}}

	// **Karten, nicht Plätze** — dieselbe Zählweise wie überall sonst im Deck.
	if got := urteil.Ungeprueft(); got != 5 {
		t.Errorf("Ungeprueft() = %d, erwartet 5", got)
	}
}

// **Eine eigene Karte wird über ihre Kennung gezählt**, nicht über einen
// Druck, den sie nicht hat.
//
// Ohne diesen Zweig zählten alle eigenen Karten als **dieselbe** — ihr
// `printing` ist leer, und ein leerer Schlüssel ist für jede derselbe.
func TestEigeneKartenWerdenEinzelnGezaehlt(t *testing.T) {
	format := domain.Format{HoechstzahlGleich: 4, Sideboardgroesse: 15}

	verschieden := domain.Deck{Plaetze: []domain.Deckplatz{
		{KarteID: 1, Menge: 4, Bereich: domain.BereichHauptdeck},
		{KarteID: 2, Menge: 4, Bereich: domain.BereichHauptdeck},
	}}

	if befunde := domain.PruefeGroessen(domain.Posten(verschieden), format); len(befunde) != 0 {
		t.Errorf("zwei verschiedene eigene Karten wurden zusammengezählt: %v", befunde)
	}

	dieselbe := domain.Deck{Plaetze: []domain.Deckplatz{
		{KarteID: 1, Menge: 4, Bereich: domain.BereichHauptdeck},
		{KarteID: 1, Menge: 1, Bereich: domain.BereichSideboard},
	}}

	if befunde := domain.PruefeGroessen(domain.Posten(dieselbe), format); len(befunde) != 1 {
		t.Errorf("%d Befunde, erwartet 1: %v", len(befunde), befunde)
	}
}

// **`Zaehlen` füllt die beiden Zahlen aus den Plätzen.**
//
// # Warum es sie neben den Plätzen gibt
//
// **Weil die Übersicht keine Plätze lädt.** Der Speicher zählt sie dort in
// SQL; wer ein einzelnes Deck liest, hat sie ohnehin und rechnet hier.
//
// **Gesehen bei der Sichtprüfung am 2026-09-01**: ein Deck mit vier Karten
// stand als „0 im Hauptdeck" da, weil die Übersicht über eine leere Liste
// rechnete.
func TestZaehlenFuelltDieBeidenZahlen(t *testing.T) {
	deck := domain.Deck{Plaetze: []domain.Deckplatz{
		{Printing: "a", Menge: 4, Bereich: domain.BereichHauptdeck},
		{Printing: "b", Menge: 2, Bereich: domain.BereichHauptdeck},
		{Printing: "c", Menge: 3, Bereich: domain.BereichSideboard},
	}}

	deck.Zaehlen()

	if deck.Hauptdeck != 6 || deck.Sideboard != 3 {
		t.Errorf("Hauptdeck %d, Sideboard %d — erwartet 6 und 3",
			deck.Hauptdeck, deck.Sideboard)
	}
}

// **Ein Wort, das dieses Programm nicht kennt, ist ungeprüft** — Befund 3 des
// Zweitblicks über Phase 4.6.
//
// # Was der erste Anlauf tat
//
// Der Parser meldete einen Befund („dieses Format bleibt für diesen Druck
// ungeprüft") und reichte den **Rohwert** weiter. `Zulaessig()` wies nur
// `banned` und `not_legal` ab, `Ungeprueft()` zählte nur den genauen Wert
// `ungeprüft`. Ein neues Wort von Scryfall — etwa `suspended` — hätte damit
// ein Deck als **zulässig** und **null ungeprüft** gemeldet, während der
// Befund daneben das Gegenteil sagte.
//
// **Ein Wächter, der seine eigene Zusage nicht hält** (§10).
func TestEinUnbekanntesWortGiltAlsUngeprueft(t *testing.T) {
	urteil := domain.Deckurteil{Urteile: []domain.Kartenurteil{
		{Zulaessigkeit: domain.Zulaessigkeit("suspended"), Menge: 3},
	}}

	// **Es steht dem Deck nicht entgegen** — so wenig wie `ungeprüft`.
	if !urteil.Zulaessig() {
		t.Error("ein unbekanntes Wort machte das Deck unzulässig")
	}

	// **Aber es wird mitgezählt.** Sonst sagte der Kopf „null ungeprüft" über
	// drei Karten, über die nichts bekannt ist.
	if urteil.Ungeprueft() != 3 {
		t.Errorf("Ungeprueft() = %d, erwartet 3", urteil.Ungeprueft())
	}
}

// **Und die vier bekannten Wörter zählen weiter nicht als ungeprüft.**
func TestDieBekanntenWoerterZaehlenNichtAlsUngeprueft(t *testing.T) {
	urteil := domain.Deckurteil{Urteile: []domain.Kartenurteil{
		{Zulaessigkeit: domain.ZulaessigLegal, Menge: 1},
		{Zulaessigkeit: domain.ZulaessigNichtGelistet, Menge: 1},
		{Zulaessigkeit: domain.ZulaessigEingeschraenkt, Menge: 1},
		{Zulaessigkeit: domain.ZulaessigVerboten, Menge: 1},
	}}

	if urteil.Ungeprueft() != 0 {
		t.Errorf("Ungeprueft() = %d, erwartet 0", urteil.Ungeprueft())
	}
}
