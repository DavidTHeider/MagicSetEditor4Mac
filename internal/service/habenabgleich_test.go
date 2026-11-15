// **„Habe ich das?"** — Werkstück 6 von Phase 4.6
// (#1293).
package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Was da ist, ist da; was fehlt, wird genannt.**
func TestDerAbgleichNenntWasFehlt(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.besitze(druckA, 2)

	werk.legeDruck(t, druckA, 4)

	abgleich, err := werk.abgleich.Fuer(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("abgleichen: %v", err)
	}

	if len(abgleich.Zeilen) != 1 {
		t.Fatalf("%d Zeilen, erwartet 1", len(abgleich.Zeilen))
	}

	zeile := abgleich.Zeilen[0]

	if zeile.Gebraucht != 4 || zeile.Vorhanden != 2 || zeile.Fehlend != 2 {
		t.Errorf("%+v", zeile)
	}

	if abgleich.Fehlend != 2 {
		t.Errorf("Fehlend %d, erwartet 2", abgleich.Fehlend)
	}
}

// **Ein Wunsch ist kein Besitz** (0057).
//
// Wer eine Karte auf die Wunschliste gesetzt hat, hat sie nicht — und ein
// Abgleich, der sie mitzählte, sagte „vollständig" über ein Deck, das niemand
// spielen kann.
func TestEinWunschZaehltNichtAlsBesitz(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.wuensche(druckA, 4)

	werk.legeDruck(t, druckA, 4)

	abgleich, err := werk.abgleich.Fuer(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("abgleichen: %v", err)
	}

	if abgleich.Zeilen[0].Vorhanden != 0 || abgleich.Fehlend != 4 {
		t.Errorf("%+v", abgleich)
	}
}

// **Derselbe Druck in Hauptdeck und Sideboard wird einmal gezählt — und der
// Bedarf addiert.**
//
// # Warum je Druck und nicht je Platz
//
// **Weil es dieselben Karten sind.** Vier im Hauptdeck und eine im Sideboard
// verlangen fünf Stück Papier — CR 100.4a rechnet die Höchstzahl ebenso über
// beide Bereiche. Ein Abgleich je Platz gäbe den Vorrat zweimal aus und
// meldete „alles da", wo eine fehlt.
func TestDerselbeDruckWirdEinmalGezaehlt(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.besitze(druckA, 4)

	werk.legeDruck(t, druckA, 4)
	werk.legeDruckIns(t, druckA, 1, domain.BereichSideboard)

	abgleich, err := werk.abgleich.Fuer(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("abgleichen: %v", err)
	}

	if len(abgleich.Zeilen) != 1 {
		t.Fatalf("%d Zeilen, erwartet 1: %+v", len(abgleich.Zeilen), abgleich.Zeilen)
	}

	if abgleich.Zeilen[0].Gebraucht != 5 || abgleich.Fehlend != 1 {
		t.Errorf("%+v", abgleich.Zeilen[0])
	}
}

// **Ein eigener Entwurf ohne Verweis hat kein Exemplar** — und das ist kein
// Fehlbestand.
//
// **Dieselbe Haltung wie bei der Formatprüfung**: das Programm sagt, was es
// weiß und was nicht. „Vier Wald fehlen" wäre falsch; „über vier Karten ist
// nichts bekannt" ist richtig.
func TestEinEigenerEntwurfIstKeinFehlbestand(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.legeKarte(t, 3)

	abgleich, err := werk.abgleich.Fuer(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("abgleichen: %v", err)
	}

	if abgleich.Fehlend != 0 {
		t.Errorf("Fehlend %d — ein Entwurf wurde als Fehlbestand gezählt", abgleich.Fehlend)
	}

	if abgleich.Unbekannt != 3 {
		t.Errorf("Unbekannt %d, erwartet 3", abgleich.Unbekannt)
	}

	if len(abgleich.Zeilen) != 1 || abgleich.Zeilen[0].Grund == "" {
		t.Errorf("%+v", abgleich.Zeilen)
	}
}

// **Eine eigene Karte mit Verweis zählt über ihren Druck.**
func TestEineEigeneKarteMitVerweisZaehltUeberIhrenDruck(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.verweise.verweise[eigeneKarte] = domain.Kartenverweis{Printing: druckA}
	werk.besitze(druckA, 1)

	werk.legeKarte(t, 3)

	abgleich, err := werk.abgleich.Fuer(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("abgleichen: %v", err)
	}

	if abgleich.Unbekannt != 0 {
		t.Errorf("Unbekannt %d — der Verweis wurde nicht gelesen", abgleich.Unbekannt)
	}

	if abgleich.Fehlend != 2 {
		t.Errorf("Fehlend %d, erwartet 2", abgleich.Fehlend)
	}
}

// **Ein Deck, das es nicht gibt, ist ein fachliches Nein.**
func TestEinFehlendesDeckLaesstSichNichtAbgleichen(t *testing.T) {
	werk := neuesPruefwerk(t)

	if _, err := werk.abgleich.Fuer(t.Context(), 999); !errors.Is(
		err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", err)
	}
}

// **Eine schweigende Sammlung hält den Abgleich an.**
//
// **Nicht durchwinken.** „Ich konnte nicht nachsehen" ist nicht „du hast
// nichts" — und der Unterschied ist der ganze Zweck dieses Wegs.
func TestEineSchweigendeSammlungHaeltDenAbgleichAn(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.legeDruck(t, druckA, 1)

	kaputt := errors.New("die Sammlung schweigt")
	werk.sammlung.fehler = kaputt

	if _, err := werk.abgleich.Fuer(t.Context(), werk.deck.ID); !errors.Is(
		err, kaputt) {
		t.Errorf("Fehler %v, erwartet den Speicherfehler", err)
	}
}

// **Und ein unlesbarer Verweis ebenso.**
func TestEinUnlesbarerVerweisHaeltDenAbgleichAn(t *testing.T) {
	werk := neuesPruefwerk(t)

	werk.legeKarte(t, 1)

	kaputt := errors.New("die Verweisablage schweigt")
	werk.verweise.fehler = kaputt

	if _, err := werk.abgleich.Fuer(t.Context(), werk.deck.ID); !errors.Is(
		err, kaputt) {
		t.Errorf("Fehler %v, erwartet den Speicherfehler", err)
	}
}
