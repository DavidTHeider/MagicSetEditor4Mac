package service_test

// **Eine Partie mit einem eingeladenen Sitz** — P9b
// (#2377).
// Der Wirt legt sie an, bevor der Gast da ist; dessen Sitz steht ohne Deck und
// ohne Bibliothek da, bis er sich mit seinem eigenen setzt.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinEingeladenerSitzEntstehtOhneDeck(t *testing.T) {
	t.Parallel()

	decks := &deckprobe{decks: map[int64]domain.Deck{1: einDeck(1, "magic")}}

	partie, fehler := service.NeuerPartieaufbau(decks).Aufbauen(context.Background(), "p1", "magic",
		einRegelwerk(), "sitz-1", []service.Sitzwunsch{
			{Kennung: "sitz-1", Name: "Ada", DeckID: 1},
			{Kennung: "sitz-2", Name: "Bo", Eingeladen: "peer-bo"},
		})
	if fehler != nil {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	bo := partie.SitzVon("sitz-2")
	if !bo.Wartet() || bo.Eingeladen != "peer-bo" {
		t.Errorf("Bos Sitz: %+v", bo)
	}

	for _, objekt := range partie.Objekte {
		if objekt.Eigner == "sitz-2" {
			t.Fatalf("in Bos Bibliothek liegt %+v", objekt)
		}
	}

	// **Gelesen wurde nur Adas Deck** — Bos liegt bei Bo.
	if decks.gelesen != 1 {
		t.Errorf("%d Decks gelesen", decks.gelesen)
	}
}

// **Was der Gast mitbringt, ist derselbe Abzug wie am eigenen Tisch** —
// eingefroren, angereichert, mit Seitendeck und dem Spiel des Decks.
func TestDerGastBringtSeinDeckAlsAbzugMit(t *testing.T) {
	t.Parallel()

	aufbau := service.NeuerPartieaufbau(&deckprobe{decks: map[int64]domain.Deck{3: einDeck(3, "magic")}})

	mit, fehler := aufbau.Mitbringen(context.Background(), 3)
	if fehler != nil {
		t.Fatalf("Mitbringen: %v", fehler)
	}

	if mit.Spiel != "magic" || mit.Deck == nil || mit.Deck.Abdruck == "" || mit.Deck.Zaehle() != 9 {
		t.Errorf("mitgebracht: %+v", mit)
	}

	if fehler := mit.Pruefe(); fehler != nil {
		t.Errorf("hält seine Gestalt nicht: %v", fehler)
	}
}

func TestEinDeckDasEsNichtGibtBringtNiemandMit(t *testing.T) {
	t.Parallel()

	if _, fehler := einPartieaufbau().Mitbringen(context.Background(), 9); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("es kam %v", fehler)
	}
}
