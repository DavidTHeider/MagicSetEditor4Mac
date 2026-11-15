// **Der Deckbau** — Werkstück 3 von Phase 4.6
// (#1293).
package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Deck entsteht, wird umbenannt und wiedergefunden.**
func TestEinDeckEntstehtUndLaesstSichUmbenennen(t *testing.T) {
	speicher := &deckdoppel{}
	bau := service.NeuerDeckbau(speicher, &kartenspieldoppel{}, &formatdoppel{})

	angelegt, err := bau.Anlegen(t.Context(), domain.Deck{
		Name: "  Aggro  ", Spiel: " magic ",
	})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// **Geglättet, bevor es in den Bestand geht.**
	if angelegt.Name != "Aggro" || angelegt.Spiel != "magic" {
		t.Errorf("%+v", angelegt)
	}

	angelegt.Name = "Midrange"

	umbenannt, err := bau.Umbenennen(t.Context(), angelegt)
	if err != nil {
		t.Fatalf("umbenennen: %v", err)
	}

	if umbenannt.Fassung != 2 {
		t.Errorf("Fassung %d, erwartet 2", umbenannt.Fassung)
	}
}

// **Ein Deck ohne Namen kommt nicht durch** — und der Befund nennt das Feld.
func TestEinDeckOhneNamenWirdAbgewiesen(t *testing.T) {
	bau := service.NeuerDeckbau(&deckdoppel{}, &kartenspieldoppel{}, &formatdoppel{})

	_, err := bau.Anlegen(t.Context(), domain.Deck{Spiel: "magic"})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler %v, erwartet ErrValidation", err)
	}

	var feld *domain.FieldError
	if !errors.As(err, &feld) || feld.Field != "name" {
		t.Errorf("der Befund nennt das Feld nicht: %v", err)
	}
}

// **Ein Deck, das seine Plätze mitbringt, wird abgewiesen — nicht gekürzt.**
//
// # Warum das eine eigene Zusage ist
//
// Der Speicher legt die Deckzeile an und **sonst nichts**. Ein Deck mit drei
// Plätzen käme mit null zurück, und der Aufrufer sähe einen Erfolg. Das ist
// die stille Übernahme aus §8, nur andersherum: still **verloren**.
func TestEinDeckMitPlaetzenWirdAbgewiesen(t *testing.T) {
	bau := service.NeuerDeckbau(&deckdoppel{}, &kartenspieldoppel{}, &formatdoppel{})

	_, err := bau.Anlegen(t.Context(), domain.Deck{
		Name: "Aggro", Spiel: "magic",
		Plaetze: []domain.Deckplatz{{
			Printing: gueltigerDruck, Menge: 1, Bereich: domain.BereichHauptdeck,
		}},
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet ErrValidation", err)
	}
}

// **Ein Druck lässt sich legen** — der Weg aus der Sammlung.
func TestEinDruckLaesstSichLegen(t *testing.T) {
	speicher := &deckdoppel{}
	bau := service.NeuerDeckbau(speicher, &kartenspieldoppel{}, &formatdoppel{})

	deck := deckAnlegenImDoppel(t, bau)

	gelegt, err := bau.PlatzLegen(t.Context(), deck.ID, domain.Deckplatz{
		Printing: " " + gueltigerDruck + " ", Menge: 4,
		Bereich: domain.BereichHauptdeck,
	})
	if err != nil {
		t.Fatalf("legen: %v", err)
	}

	// **Geglättet**, sonst stünde ein Leerzeichen in der Kennung, an der
	// später „habe ich das?" hängt.
	if gelegt.Printing != gueltigerDruck {
		t.Errorf("Druck %q", gelegt.Printing)
	}
}

// **Eine eigene Karte lässt sich legen — wenn sie zum Spiel des Decks gehört.**
func TestEineEigeneKarteAusDemselbenSpielLaesstSichLegen(t *testing.T) {
	speicher := &deckdoppel{}
	bau := service.NeuerDeckbau(speicher, &kartenspieldoppel{spiel: "magic"}, &formatdoppel{})

	deck := deckAnlegenImDoppel(t, bau)

	if _, err := bau.PlatzLegen(t.Context(), deck.ID, domain.Deckplatz{
		KarteID: 7, Menge: 1, Bereich: domain.BereichSideboard,
	}); err != nil {
		t.Fatalf("legen: %v", err)
	}
}

// **Eine Karte aus einem anderen Spiel wird abgewiesen.**
//
// **Ohne Spiel gäbe es kein Format** — und ein Deck aus Karten zweier Spiele
// ließe sich gegen keines prüfen.
func TestEineKarteAusEinemAnderenSpielWirdAbgewiesen(t *testing.T) {
	speicher := &deckdoppel{}
	bau := service.NeuerDeckbau(speicher, &kartenspieldoppel{spiel: "vs"}, &formatdoppel{})

	deck := deckAnlegenImDoppel(t, bau)

	_, err := bau.PlatzLegen(t.Context(), deck.ID, domain.Deckplatz{
		KarteID: 7, Menge: 1, Bereich: domain.BereichHauptdeck,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler %v, erwartet ErrValidation", err)
	}

	var feld *domain.FieldError
	if !errors.As(err, &feld) || feld.Field != "karteID" {
		t.Errorf("der Befund nennt das Feld nicht: %v", err)
	}
}

// **Ein Platz ohne Bezug kommt gar nicht erst zum Speicher.**
func TestEinPlatzOhneBezugKommtNichtDurch(t *testing.T) {
	speicher := &deckdoppel{}
	bau := service.NeuerDeckbau(speicher, &kartenspieldoppel{}, &formatdoppel{})

	deck := deckAnlegenImDoppel(t, bau)

	if _, err := bau.PlatzLegen(t.Context(), deck.ID, domain.Deckplatz{
		Menge: 1, Bereich: domain.BereichHauptdeck,
	}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet ErrValidation", err)
	}

	if speicher.gelegt != 0 {
		t.Errorf("der Speicher wurde %d-mal gerufen", speicher.gelegt)
	}
}

// **In ein Deck, das es nicht gibt, wird nichts gelegt.**
func TestInEinFehlendesDeckWirdNichtsGelegt(t *testing.T) {
	speicher := &deckdoppel{}
	bau := service.NeuerDeckbau(speicher, &kartenspieldoppel{}, &formatdoppel{})

	if _, err := bau.PlatzLegen(t.Context(), 999, domain.Deckplatz{
		Printing: gueltigerDruck, Menge: 1, Bereich: domain.BereichHauptdeck,
	}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", err)
	}
}

// **Eine Karte, die es nicht gibt, ebenso.**
func TestEineFehlendeKarteWirdNichtGelegt(t *testing.T) {
	speicher := &deckdoppel{}
	bau := service.NeuerDeckbau(speicher, &kartenspieldoppel{
		fehler: domain.ErrNotFound,
	}, &formatdoppel{})

	deck := deckAnlegenImDoppel(t, bau)

	if _, err := bau.PlatzLegen(t.Context(), deck.ID, domain.Deckplatz{
		KarteID: 7, Menge: 1, Bereich: domain.BereichHauptdeck,
	}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", err)
	}
}

// **Nehmen, löschen und lesen reichen durch.**
func TestDieDurchreichendenWegeDesDeckbaus(t *testing.T) {
	speicher := &deckdoppel{}
	bau := service.NeuerDeckbau(speicher, &kartenspieldoppel{}, &formatdoppel{})

	deck := deckAnlegenImDoppel(t, bau)

	if _, err := bau.Deck(t.Context(), deck.ID); err != nil {
		t.Errorf("lesen: %v", err)
	}

	decks, err := bau.Decks(t.Context(), "magic")
	if err != nil || len(decks) != 1 {
		t.Errorf("die Liste: %v, %d Decks", err, len(decks))
	}

	if err := bau.PlatzNehmen(t.Context(), deck.ID, 1); err != nil {
		t.Errorf("nehmen: %v", err)
	}

	if err := bau.Loeschen(t.Context(), deck.ID); err != nil {
		t.Errorf("löschen: %v", err)
	}
}

// **Und jeder Weg meldet den Fehler seines Speichers.**
func TestJederDeckbauwegMeldetSeinenSpeicherfehler(t *testing.T) {
	kaputt := errors.New("der Bestand schweigt")
	speicher := &deckdoppel{fehler: kaputt}
	bau := service.NeuerDeckbau(speicher, &kartenspieldoppel{}, &formatdoppel{})

	if _, err := bau.Anlegen(t.Context(), domain.Deck{
		Name: "Aggro", Spiel: "magic",
	}); !errors.Is(err, kaputt) {
		t.Error("anlegen schwieg")
	}

	if _, err := bau.Umbenennen(t.Context(), domain.Deck{
		ID: 1, Fassung: 1, Name: "Aggro", Spiel: "magic",
	}); !errors.Is(err, kaputt) {
		t.Error("umbenennen schwieg")
	}

	if _, err := bau.Deck(t.Context(), 1); !errors.Is(err, kaputt) {
		t.Error("lesen schwieg")
	}

	if _, err := bau.Decks(t.Context(), ""); !errors.Is(err, kaputt) {
		t.Error("die Liste schwieg")
	}

	if _, err := bau.PlatzLegen(t.Context(), 1, domain.Deckplatz{
		Printing: gueltigerDruck, Menge: 1, Bereich: domain.BereichHauptdeck,
	}); !errors.Is(err, kaputt) {
		t.Error("legen schwieg")
	}

	if err := bau.PlatzNehmen(t.Context(), 1, 1); !errors.Is(err, kaputt) {
		t.Error("nehmen schwieg")
	}

	if err := bau.Loeschen(t.Context(), 1); !errors.Is(err, kaputt) {
		t.Error("löschen schwieg")
	}
}

// **Ein Umbenennen ohne Namen kommt nicht durch** — dieselbe Prüfung wie beim
// Anlegen, und sie muss auch hier stehen.
func TestEinUmbenennenOhneNamenWirdAbgewiesen(t *testing.T) {
	bau := service.NeuerDeckbau(&deckdoppel{}, &kartenspieldoppel{}, &formatdoppel{})

	if _, err := bau.Umbenennen(t.Context(), domain.Deck{
		ID: 1, Fassung: 1, Spiel: "magic",
	}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet ErrValidation", err)
	}
}

// deckAnlegenImDoppel legt ein Deck an und gibt es zurück.
func deckAnlegenImDoppel(t *testing.T, bau *service.Deckbau) domain.Deck {
	t.Helper()

	deck, err := bau.Anlegen(t.Context(), domain.Deck{Name: "Aggro", Spiel: "magic"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	return deck
}
