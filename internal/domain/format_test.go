package domain_test

// **Das Format** — Werkstück 4 von Phase 4.6
// (#1293),
// nach Vermerk 0061: „Daten, nicht Code".

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Format braucht Namen und Spiel.**
func TestEinFormatBrauchtNamenUndSpiel(t *testing.T) {
	for _, fall := range []struct {
		name   string
		format domain.Format
		feld   string
	}{
		{"ohne Namen", domain.Format{Spiel: "magic"}, "name"},
		{"ohne Spiel", domain.Format{Name: "Standard"}, "spiel"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			befunde := fall.format.Validate()
			if len(befunde) == 0 {
				t.Fatal("kein Befund")
			}

			if befunde[0].Field != fall.feld {
				t.Errorf("Feld %q, erwartet %q", befunde[0].Field, fall.feld)
			}
		})
	}
}

// **Eine Obergrenze unter der Untergrenze ist ein Widerspruch.**
//
// **Null heißt „keine Grenze"** — und darf deshalb nicht als Widerspruch
// gelesen werden.
func TestEineObergrenzeUnterDerUntergrenze(t *testing.T) {
	widerspruch := domain.Format{
		Name: "Krumm", Spiel: "magic", Mindestgroesse: 60, Hoechstgroesse: 40,
	}

	if len(widerspruch.Validate()) == 0 {
		t.Error("der Widerspruch kam durch")
	}

	ohneGrenze := domain.Format{
		Name: "Offen", Spiel: "magic", Mindestgroesse: 60, Hoechstgroesse: 0,
	}

	if befunde := ohneGrenze.Validate(); len(befunde) != 0 {
		t.Errorf("null wurde als Grenze gelesen: %v", befunde[0])
	}
}

// **Negative Zahlen gibt es nicht.**
func TestNegativeZahlenImFormat(t *testing.T) {
	for _, fall := range []struct {
		name   string
		format domain.Format
		feld   string
	}{
		{"Untergrenze", domain.Format{
			Name: "F", Spiel: "magic", Mindestgroesse: -1,
		}, "mindestgroesse"},
		{"Obergrenze", domain.Format{
			Name: "F", Spiel: "magic", Hoechstgroesse: -1,
		}, "hoechstgroesse"},
		{"Sideboard", domain.Format{
			Name: "F", Spiel: "magic", Sideboardgroesse: -1,
		}, "sideboardgroesse"},
		{"gleiche Karten", domain.Format{
			Name: "F", Spiel: "magic", HoechstzahlGleich: -1,
		}, "hoechstzahlGleich"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			befunde := fall.format.Validate()
			if len(befunde) == 0 {
				t.Fatal("kein Befund")
			}

			if befunde[0].Field != fall.feld {
				t.Errorf("Feld %q, erwartet %q", befunde[0].Field, fall.feld)
			}
		})
	}
}

// **Ein Scryfallname ist ein Schlüssel aus `legalities`** — Kleinbuchstaben
// und sonst nichts. Leer heißt: Scryfall kennt dieses Format nicht.
func TestDerScryfallnameHatEineGestalt(t *testing.T) {
	for _, fall := range []struct {
		name    string
		wert    string
		befunde int
	}{
		{"leer ist erlaubt", "", 0},
		{"standard", "standard", 0},
		{"paupercommander", "paupercommander", 0},
		{"mit Großbuchstaben", "Standard", 1},
		{"mit Bindestrich", "pauper-commander", 1},
		{"mit Leerzeichen", "pauper commander", 1},
	} {
		t.Run(fall.name, func(t *testing.T) {
			format := domain.Format{Name: "F", Spiel: "magic", Scryfallname: fall.wert}

			if got := len(format.Validate()); got != fall.befunde {
				t.Errorf("%d Befunde, erwartet %d", got, fall.befunde)
			}
		})
	}
}

// **Normalisieren glättet, bevor geprüft wird.**
func TestEinFormatWirdGeglaettet(t *testing.T) {
	format := domain.Format{
		Name: "  Standard  ", Spiel: " magic ", Scryfallname: " STANDARD ",
	}

	format.Normalisieren()

	if format.Name != "Standard" || format.Spiel != "magic" {
		t.Errorf("%+v", format)
	}

	// **Der Scryfallname wird kleingeschrieben**, nicht abgewiesen: er kommt
	// aus einer Auswahl, und ein „Standard" von dort ist gemeint, nicht falsch.
	if format.Scryfallname != "standard" {
		t.Errorf("Scryfallname %q", format.Scryfallname)
	}
}

// **Die drei Bauarten tragen die Zahlen aus den Comprehensive Rules.**
//
// # Warum sie hier stehen und nicht in einer Saatdatei
//
// Weil jede Zahl **nachgelesen** ist und ihre Regel nennt (§10). Eine Saat
// ohne Herkunft wäre eine Erfindung mit Datenbankzeile.
func TestDieBauartenTragenNachgeleseneZahlen(t *testing.T) {
	for _, fall := range []struct {
		bauart                                   domain.Bauart
		mindest, hoechst, sideboard, hoechstzahl int
	}{
		// CR 100.2a: mindestens sechzig Karten, höchstens vier je Name;
		// CR 100.4a: höchstens fünfzehn im Sideboard.
		{domain.BauartKonstruiert, 60, 0, 15, 4},

		// CR 903.5a: genau hundert, Mindest- und Höchstzahl fallen zusammen;
		// 903.5b: jede Karte außer Standardländern ein anderer Name;
		// 903.5e: Commanderpartien kennen kein Sideboard.
		{domain.BauartCommander, 100, 100, 0, 1},

		// CR 100.2b: mindestens vierzig Karten, Dubletten so oft, wie das
		// Produkt sie mitbringt.
		{domain.BauartLimitiert, 40, 0, 0, 0},
	} {
		t.Run(string(fall.bauart), func(t *testing.T) {
			format, da := domain.VonBauart(fall.bauart)
			if !da {
				t.Fatal("die Bauart gibt es nicht")
			}

			if format.Mindestgroesse != fall.mindest ||
				format.Hoechstgroesse != fall.hoechst ||
				format.Sideboardgroesse != fall.sideboard ||
				format.HoechstzahlGleich != fall.hoechstzahl {
				t.Errorf("%+v", format)
			}
		})
	}
}

// **Eine Bauart, die es nicht gibt, sagt das.**
func TestEineUnbekannteBauart(t *testing.T) {
	if _, da := domain.VonBauart(domain.Bauart("turnier")); da {
		t.Error("eine erfundene Bauart kam durch")
	}
}

// **Die Aufzählung der Bauarten nennt jede, die es gibt.**
//
// **Beide Richtungen**: was `Bauarten` nennt, muss `VonBauart` kennen — sonst
// böte eine Fläche eine Wahl an, die ins Leere geht.
func TestDieAufzaehlungDerBauartenIstVollstaendig(t *testing.T) {
	genannt := domain.Bauarten()

	if len(genannt) != 3 {
		t.Fatalf("%d Bauarten, erwartet 3", len(genannt))
	}

	for _, bauart := range genannt {
		if _, da := domain.VonBauart(bauart); !da {
			t.Errorf("%q wird genannt, ist aber nicht zu haben", bauart)
		}
	}
}

// **Ein Befund kann Werte für seine Platzhalter tragen** —
// #1369.
//
// # Warum sie nicht im Satz stehen
//
// Weil der Satz übersetzt wird: der Katalog des Fensters führt für jeden
// Schlüssel einen eigenen Text mit Platzhaltern. Stünden die Namen nur im
// deutschen Satz aus Go, ginge die englische Fassung ohne sie hinaus.
func TestEinBefundKannWerteTragen(t *testing.T) {
	befund := domain.NewFieldError("karteID", domain.SchluesselKarteInDeck,
		"Diese Karte liegt in %d Deck(s) — %s.", 1, "Aggro").
		MitWerten(map[string]string{"anzahl": "1", "decks": "Aggro"})

	if befund.Werte["decks"] != "Aggro" {
		t.Errorf("%+v", befund.Werte)
	}

	// **Und er ist ein Konflikt, keine Validierung.** Die Eingabe ist in
	// Ordnung; der Bestand steht ihr entgegen.
	if !errors.Is(befund, domain.ErrConflict) {
		t.Error("der Befund meldet sich nicht als Konflikt")
	}
}

// **Limited hat ein Sideboard ohne Obergrenze** — Befund 2 des Zweitblicks
// über Phase 4.6.
//
// # Die Regel, wörtlich
//
// CR 100.4b: im Limited für einzelne Spieler ist **alles Sideboard, was
// vom Kartenvorrat nicht im Deck steckt**. Es gibt dort keine Höchstzahl — und keine Null.
//
// # Was der erste Anlauf tat
//
// **Er setzte `Sideboardgroesse: 0`** — und null heißt in diesem Modell
// **kein** Sideboard (CR 903.5e, Commander). Jedes Limited-Deck mit einer
// übrigen Poolkarte galt damit als unzulässig. Der Kommentar daneben sagte das
// Richtige und die Zahl das Gegenteil.
func TestLimitedHatEinSideboardOhneObergrenze(t *testing.T) {
	format, da := domain.VonBauart(domain.BauartLimitiert)
	if !da {
		t.Fatal("die Bauart gibt es nicht")
	}

	if !format.SideboardOffen {
		t.Error("Limited verbietet sein Sideboard")
	}

	// **Und Commander verbietet es weiterhin** — CR 903.5e.
	commander, _ := domain.VonBauart(domain.BauartCommander)
	if commander.SideboardOffen {
		t.Error("Commander erlaubt plötzlich ein Sideboard")
	}
}

// **Ein offenes Sideboard beanstandet nichts.**
func TestEinOffenesSideboardBeanstandetNichts(t *testing.T) {
	format := domain.Format{Name: "Limitiert", Mindestgroesse: 40, SideboardOffen: true}

	posten := []domain.Deckposten{
		{Platz: domain.Deckplatz{
			ID: 1, Printing: "a", Menge: 40,
			Bereich: domain.BereichHauptdeck,
		}},
		{Platz: domain.Deckplatz{
			ID: 2, Printing: "b", Menge: 45,
			Bereich: domain.BereichSideboard,
		}},
	}

	if befunde := domain.PruefeGroessen(posten, format); len(befunde) != 0 {
		t.Errorf("das offene Sideboard wurde beanstandet: %v", befunde)
	}
}
