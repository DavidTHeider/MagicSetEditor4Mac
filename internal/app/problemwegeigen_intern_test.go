package app

// **Der Weg gibt den Stand seines eigenen Laufs** — Befund 4 des Zweitblicks
// über Phase 3 von 0.9.0 (PERF02, #2352): er las nach dem Lauf den
// gemeinsamen Speicher, und dort stand, was ein anderer Lauf ablegte.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDerProblemwegGibtDenEigenenStand(t *testing.T) {
	t.Parallel()

	karten := []*domain.CardRow{{ID: 1}, {ID: 2}, {ID: 3}}
	pruefung := service.NeueProblempruefung(
		func(context.Context, int64) ([]*domain.CardRow, error) { return karten, nil },
		func(context.Context, *domain.CardRow) ([]string, error) { return []string{"fehlt"}, nil },
		func(context.Context, int64) (string, error) { return "a", nil },
		time.Now,
	).MitArbeitern(1)
	weg := &Problemweg{pruefung: pruefung, uhr: time.Now}

	if voll, err := weg.Pruefen(t.Context(), 1, nil); err != nil || voll.Geprueft != 3 {
		t.Fatalf("der erste Lauf: %+v, %v", voll, err)
	}

	// **Der zweite bricht ab** — und legt nichts über den vollständigen
	// desselben Abdrucks. Seine Antwort ist trotzdem seine.
	ctx, abbrechen := context.WithCancel(t.Context())

	abgebrochen, err := weg.Pruefen(ctx, 1, func(geprueft, _ int) {
		if geprueft == 1 {
			abbrechen()
		}
	})
	if !errors.Is(err, context.Canceled) || abgebrochen.Geprueft == 3 || abgebrochen.Sicher {
		t.Errorf("der abgebrochene Lauf nannte %+v (%v) — den Stand des vorigen", abgebrochen, err)
	}
}
