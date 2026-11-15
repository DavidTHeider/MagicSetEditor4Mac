package storage_test

// **In welchen Decks liegt diese Karte?** —
// #1369.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Decks kommen mit ihren Namen** — eine Rückfrage, die nur Kennungen
// nennt, hilft niemandem.
func TestDecksMitKarteNennenIhreNamen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	karte := kartenkennungFuerDeck(t, db)

	for _, name := range []string{"Midrange", "Aggro"} {
		deck, err := ablage.Anlegen(probekontext(t), domain.Deck{
			Name: name, Spiel: "magic",
		})
		if err != nil {
			t.Fatalf("das Deck %s: %v", name, err)
		}

		if _, err := ablage.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
			KarteID: karte, Menge: 1, Bereich: domain.BereichHauptdeck,
		}); err != nil {
			t.Fatalf("legen: %v", err)
		}
	}

	bezuege, err := ablage.DecksMitKarte(probekontext(t), karte)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(bezuege) != 2 {
		t.Fatalf("%d Decks, erwartet 2", len(bezuege))
	}

	// **Geordnet vom Bestand**, sonst nennt jede Fläche sie anders.
	if bezuege[0].Name != "Aggro" || bezuege[1].Name != "Midrange" {
		t.Errorf("%+v", bezuege)
	}

	if bezuege[0].ID == 0 {
		t.Error("ein Bezug ohne Kennung")
	}
}

// **Dieselbe Karte in Hauptdeck und Sideboard nennt das Deck einmal.**
//
// Zweimal genannt sähe aus wie zwei Decks — und die Rückfrage zählte falsch.
func TestEinDeckWirdNurEinmalGenannt(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	karte := kartenkennungFuerDeck(t, db)

	deck, err := ablage.Anlegen(probekontext(t), domain.Deck{Name: "Aggro", Spiel: "magic"})
	if err != nil {
		t.Fatalf("das Deck: %v", err)
	}

	for _, bereich := range []domain.Deckbereich{
		domain.BereichHauptdeck, domain.BereichSideboard,
	} {
		if _, legefehler := ablage.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
			KarteID: karte, Menge: 1, Bereich: bereich,
		}); legefehler != nil {
			t.Fatalf("legen: %v", legefehler)
		}
	}

	bezuege, err := ablage.DecksMitKarte(probekontext(t), karte)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(bezuege) != 1 {
		t.Errorf("%d Decks, erwartet 1: %+v", len(bezuege), bezuege)
	}
}

// **Eine Karte, die in keinem Deck liegt, gibt eine leere Liste.**
func TestEineKarteInKeinemDeck(t *testing.T) {
	db := frischeDatenbank(t)

	bezuege, err := storage.NeueDeckablage(db).DecksMitKarte(probekontext(t), 999)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(bezuege) != 0 {
		t.Errorf("%+v", bezuege)
	}
}

// **Und ein kaputter Bestand meldet sich.**
func TestDecksMitKarteMeldetEinenKaputtenBestand(t *testing.T) {
	if _, err := storage.NeueDeckablage(kaputterBestand(t)).DecksMitKarte(
		probekontext(t), 1); err == nil {
		t.Error("die Abfrage schwieg")
	}
}
