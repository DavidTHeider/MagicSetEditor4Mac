package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Derselbe Druck zweimal ist eine Dublette, keine fremde Änderung** —
// nativer Befund F08, ENG-05.AK-03
// (#1730).
//
// # Was das Audit sah
//
// Ein Druck lag einmal in der Sammlung — englisch, nonfoil, Near Mint. Ihn
// noch einmal aufzunehmen ergab „Der Eintrag wurde zwischenzeitlich geändert.
// Bitte neu laden.", und das Protokoll nannte die verletzte Eindeutigkeit.
// Niemand hatte etwas geändert.
//
// # Was hier steht
//
// Der Weg durch die Bindung des Fensters, gegen den echten Bestand: der
// zweite Versuch sagt `doppeltesExemplar`, die Menge bleibt, und dieselbe
// Karte in einer anderen Sprache ist eine eigene Zeile (S04.NAT-02).
func TestDerselbeDruckZweimalIstEineDublette(t *testing.T) {
	anwendung := gestartet(t)
	ctx := t.Context()

	exemplar := app.ExemplarDTO{Printing: "druck-0001", Menge: 1, Zustand: "nearMint", Sprache: "en"}

	erstes := anwendung.ExemplarAnlegen(ctx, exemplar)
	if erstes.Fehler != nil {
		t.Fatalf("das erste: %+v", erstes.Fehler)
	}

	zweites := anwendung.ExemplarAnlegen(ctx, exemplar)
	if zweites.Fehler == nil {
		t.Fatal("derselbe Druck kam zweimal hinein")
	}

	if zweites.Fehler.Schluessel != domain.SchluesselDoppeltesExemplar ||
		zweites.Fehler.Code != app.CodeKonflikt {
		t.Errorf("Code %q, Schlüssel %q — erwartet die Dublette", zweites.Fehler.Code, zweites.Fehler.Schluessel)
	}

	// **Und die Zeile, die es schon gibt** — DOM-03, S04.NAT-01 und 03
	// (#1698): das Fenster bietet an, ihre Menge zu erhöhen oder zu ihr zu
	// springen, und braucht dafür Kennung und Fassung.
	if zweites.Vorhanden == nil || zweites.Vorhanden.ID != erstes.ID ||
		zweites.Vorhanden.Menge != 1 || zweites.Vorhanden.Fassung == 0 {
		t.Errorf("die vorhandene Zeile: %+v", zweites.Vorhanden)
	}

	if erstes.Vorhanden != nil {
		t.Errorf("eine geglückte Aufnahme nennt eine vorhandene Zeile: %+v", erstes.Vorhanden)
	}

	anders := exemplar
	anders.Sprache = "de"

	if antwort := anwendung.ExemplarAnlegen(ctx, anders); antwort.Fehler != nil {
		t.Fatalf("dieselbe Karte auf Deutsch: %+v", antwort.Fehler)
	}

	sammlung := anwendung.Sammlung(ctx)
	if sammlung.Fehler != nil || len(sammlung.Exemplare) != 2 {
		t.Fatalf("die Sammlung: %+v", sammlung)
	}

	for _, zeile := range sammlung.Exemplare {
		if zeile.Menge != 1 {
			t.Errorf("die Menge von %s/%s ist %d — der Fehlversuch hat gezählt", zeile.Printing, zeile.Sprache, zeile.Menge)
		}
	}
}

// **Ein zweites Deck gleichen Namens steht am Namen.**
func TestEinZweitesDeckGleichenNamensStehtAmNamen(t *testing.T) {
	anwendung := gestartet(t)
	ctx := t.Context()

	if antwort := anwendung.DeckAnlegen(ctx, "Aggro", "magic", ""); antwort.Fehler != nil {
		t.Fatalf("das erste: %+v", antwort.Fehler)
	}

	zweites := anwendung.DeckAnlegen(ctx, "Aggro", "magic", "")
	if zweites.Fehler == nil {
		t.Fatal("zwei Decks gleichen Namens")
	}

	if zweites.Fehler.Schluessel != domain.SchluesselDoppelterName || zweites.Fehler.Feld != "name" {
		t.Errorf("Schlüssel %q am Feld %q — erwartet doppelterName am Namen",
			zweites.Fehler.Schluessel, zweites.Fehler.Feld)
	}
}
