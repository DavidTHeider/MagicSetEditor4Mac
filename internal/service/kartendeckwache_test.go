package service_test

// **Eine Karte, die in einem Deck liegt, verschwindet nicht unbemerkt** —
// #1369.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// deckwachedoppel sagt, in welchen Decks eine Karte liegt.
type deckwachedoppel struct {
	je     map[int64][]domain.Deckbezug
	fehler error
}

func (d *deckwachedoppel) DecksMitKarte(_ context.Context, karteID int64) (
	[]domain.Deckbezug, error,
) {
	return d.je[karteID], d.fehler
}

// **Eine Karte in einem Deck wird nicht gelöscht — und die Meldung nennt es.**
func TestEineKarteInEinemDeckWirdNichtStillGeloescht(t *testing.T) {
	aufsicht, zyklus := aufsichtMitDeckwache(t, map[int64][]domain.Deckbezug{
		7: {{ID: 1, Name: "Aggro"}, {ID: 2, Name: "Midrange"}},
	})

	err := aufsicht.KarteLoeschen(t.Context(), 1, 7, false)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("Fehler %v, erwartet ErrConflict", err)
	}

	// **Mit ihren Namen.** Eine Rückfrage, die nur Zahlen nennt, hilft
	// niemandem beim Entscheiden.
	if !strings.Contains(err.Error(), "Aggro") ||
		!strings.Contains(err.Error(), "Midrange") {
		t.Errorf("die Meldung nennt die Decks nicht: %v", err)
	}

	if len(zyklus.geloescht) != 0 {
		t.Errorf("es wurde %d-mal gelöscht", len(zyklus.geloescht))
	}
}

// **Wer bestätigt, verliert den Platz — und weiß es.**
func TestMitBestaetigungWirdGeloescht(t *testing.T) {
	aufsicht, zyklus := aufsichtMitDeckwache(t, map[int64][]domain.Deckbezug{
		7: {{ID: 1, Name: "Aggro"}},
	})

	if err := aufsicht.KarteLoeschen(t.Context(), 1, 7, true); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	if len(zyklus.geloescht) != 1 {
		t.Errorf("es wurde %d-mal gelöscht", len(zyklus.geloescht))
	}
}

// **Eine Karte, die in keinem Deck liegt, wird ohne Rückfrage gelöscht.**
func TestEineKarteOhneDeckWirdOhneRueckfrageGeloescht(t *testing.T) {
	aufsicht, zyklus := aufsichtMitDeckwache(t, nil)

	if err := aufsicht.KarteLoeschen(t.Context(), 1, 7, false); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	if len(zyklus.geloescht) != 1 {
		t.Errorf("es wurde %d-mal gelöscht", len(zyklus.geloescht))
	}
}

// **Ohne Deckwache wird gelöscht wie bisher.**
//
// **Der MCP-Prozess hängt sie an, das Fenster auch** — aber ein Aufrufer, der
// keine Decks kennt, soll nicht an einer fehlenden Abhängigkeit scheitern.
// Was er verliert, ist die Rückfrage, nicht die Funktion.
func TestOhneDeckwacheWirdGeloescht(t *testing.T) {
	zyklus := &lebenszyklusprotokoll{}
	aufsicht := service.NeueKartenaufsicht(&kartensicht{karten: map[int64]int64{7: 1, 8: 1}}, zyklus)

	if err := aufsicht.KarteLoeschen(t.Context(), 1, 7, false); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	if len(zyklus.geloescht) != 1 {
		t.Errorf("es wurde %d-mal gelöscht", len(zyklus.geloescht))
	}
}

// **Auch das Löschen mehrerer fragt zurück** — und nennt jede betroffene
// Karte.
//
// **Vor der ersten**, wie die Setgrenze: sonst stünde die Zusage „ganz oder
// gar nicht" auf dem Kopf.
func TestAuchMehrereKartenFragenZurueck(t *testing.T) {
	aufsicht, zyklus := aufsichtMitDeckwache(t, map[int64][]domain.Deckbezug{
		8: {{ID: 1, Name: "Aggro"}},
	})

	err := aufsicht.KartenLoeschen(t.Context(), 1, []int64{7, 8}, false)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("Fehler %v, erwartet ErrConflict", err)
	}

	if len(zyklus.geloescht) != 0 {
		t.Errorf("es wurde %d-mal gelöscht", len(zyklus.geloescht))
	}

	if err := aufsicht.KartenLoeschen(t.Context(), 1, []int64{7, 8}, true); err != nil {
		t.Fatalf("mit Bestätigung: %v", err)
	}
}

// **Eine Deckwache, die schweigt, hält das Löschen an.**
//
// **Nicht durchwinken.** „Ich konnte nicht nachsehen" ist nicht dasselbe wie
// „sie liegt in keinem Deck" — und der Unterschied ist genau der, den diese
// Wache halten soll.
func TestEineSchweigendeDeckwacheHaeltDasLoeschenAn(t *testing.T) {
	zyklus := &lebenszyklusprotokoll{}
	kaputt := errors.New("die Deckablage schweigt")

	aufsicht := service.NeueKartenaufsicht(&kartensicht{karten: map[int64]int64{7: 1, 8: 1}}, zyklus).
		MitDeckwache(&deckwachedoppel{fehler: kaputt})

	if err := aufsicht.KarteLoeschen(t.Context(), 1, 7, false); !errors.Is(
		err, kaputt) {
		t.Errorf("Fehler %v, erwartet den Speicherfehler", err)
	}

	if len(zyklus.geloescht) != 0 {
		t.Error("es wurde gelöscht, obwohl niemand nachsehen konnte")
	}
}

// aufsichtMitDeckwache baut eine Aufsicht, die Decks kennt.
func aufsichtMitDeckwache(t *testing.T, je map[int64][]domain.Deckbezug) (
	*service.Kartenaufsicht, *lebenszyklusprotokoll,
) {
	t.Helper()

	zyklus := &lebenszyklusprotokoll{}

	return service.NeueKartenaufsicht(&kartensicht{karten: map[int64]int64{7: 1, 8: 1}}, zyklus).
		MitDeckwache(&deckwachedoppel{je: je}), zyklus
}
