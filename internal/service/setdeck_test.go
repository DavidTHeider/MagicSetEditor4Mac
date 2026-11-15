package service_test

// **Aus einem Set ein Deck** — FMT-04 (#1693, R52): die Karten eines Sets
// werden Plätze eines neuen Decks, jede mit derselben Menge, in **einem** Zug
// und mit der Herkunft „aus dem Set".

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

type setkartenprobe struct {
	kopf         domain.SetHead
	karten       []*domain.CardRow
	fehler       error
	kartenfehler error
	gefragt      int64
}

func (s *setkartenprobe) Setzeile(_ context.Context, setID int64) (domain.SetHead, error) {
	s.gefragt = setID

	return s.kopf, s.fehler
}

func (s *setkartenprobe) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return s.karten, s.kartenfehler
}

type deckaufnahmeprobe struct {
	deck     domain.Deck
	plaetze  []domain.Deckplatz
	herkunft domain.Herkunft
	fehler   error
}

func (d *deckaufnahmeprobe) Uebernehmen(
	_ context.Context, deck domain.Deck, plaetze []domain.Deckplatz, herkunft domain.Herkunft, _ int64,
) (domain.Deck, error) {
	d.deck, d.plaetze, d.herkunft = deck, plaetze, herkunft
	deck.ID = 9

	return deck, d.fehler
}

func TestAusEinemSetWirdEinDeck(t *testing.T) {
	t.Parallel()

	karten := &setkartenprobe{
		kopf:   domain.SetHead{ID: 4, Name: "Werkstatt", Game: "magic"},
		karten: []*domain.CardRow{{ID: 11}, {ID: 12}, {ID: 13}},
	}
	decks := &deckaufnahmeprobe{}
	dienst := service.NeuesSetdeck(karten, decks)

	deck, err := dienst.Anlegen(t.Context(), 4, "  ", 2)
	if err != nil || deck.ID != 9 {
		t.Fatalf("%+v %v", deck, err)
	}

	// **Ohne Namen heißt das Deck wie das Set**, und es gehört zu dessen Spiel.
	if decks.deck.Name != "Werkstatt" || decks.deck.Spiel != "magic" || len(decks.plaetze) != 3 {
		t.Errorf("Deck %+v mit %d Plätzen", decks.deck, len(decks.plaetze))
	}

	for i, platz := range decks.plaetze {
		if platz.KarteID != int64(11+i) || platz.Menge != 2 || platz.Bereich != domain.BereichHauptdeck {
			t.Errorf("Platz %d: %+v", i, platz)
		}
	}

	if decks.herkunft.Art != domain.HerkunftEinfuhr || decks.herkunft.Quelle != "Werkstatt" {
		t.Errorf("Herkunft %+v", decks.herkunft)
	}

	if _, err := dienst.Anlegen(t.Context(), 4, "Mein Deck", 1); err != nil || decks.deck.Name != "Mein Deck" {
		t.Errorf("mit Namen: %q, %v", decks.deck.Name, err)
	}
}

func TestEinDeckAusEinemSetSagtNein(t *testing.T) {
	t.Parallel()

	nicht := errors.New("gibt es nicht")

	for name, fall := range map[string]struct {
		karten *setkartenprobe
		menge  int
		fehler error
		pruefe func(error) bool
	}{
		"eine Menge unter eins": {
			karten: &setkartenprobe{karten: []*domain.CardRow{{ID: 1}}}, menge: 0,
			pruefe: func(err error) bool { return errors.Is(err, domain.ErrValidation) },
		},
		"ein Set ohne Karten": {
			karten: &setkartenprobe{}, menge: 1,
			pruefe: func(err error) bool { return errors.Is(err, domain.ErrValidation) },
		},
		"ein Set, das es nicht gibt": {
			karten: &setkartenprobe{fehler: nicht}, menge: 1,
			pruefe: func(err error) bool { return errors.Is(err, nicht) },
		},
		"Karten, die sich nicht lesen lassen": {
			karten: &setkartenprobe{kartenfehler: nicht}, menge: 1,
			pruefe: func(err error) bool { return errors.Is(err, nicht) },
		},
		"ein Schreiben, das scheitert": {
			karten: &setkartenprobe{karten: []*domain.CardRow{{ID: 1}}}, menge: 1, fehler: domain.ErrConflict,
			pruefe: func(err error) bool { return errors.Is(err, domain.ErrConflict) },
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			decks := &deckaufnahmeprobe{fehler: fall.fehler}
			if _, err := service.NeuesSetdeck(fall.karten, decks).Anlegen(t.Context(), 1, "x", fall.menge); !fall.pruefe(err) {
				t.Errorf("%v", err)
			}
		})
	}
}
