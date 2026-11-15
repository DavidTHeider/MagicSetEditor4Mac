package service_test

// **Das Einfahren eines Setpakets meldet sich und lässt sich anhalten** —
// FMT-03.AK-02 und AK-04
// (#1692).

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDasEinfahrenZaehltEinSet(t *testing.T) {
	var erledigt []int

	ctx := service.MitFortschritt(t.Context(), func(stand domain.Fortschritt) {
		if stand.Gesamt != 1 || stand.Einheit != domain.EinheitSet {
			t.Errorf("Stand: %+v", stand)
		}

		erledigt = append(erledigt, stand.Erledigt)
	})

	if _, err := service.NeuerImporteur(&speicherDouble{}).
		Set(ctx, filepath.Join(testdaten, "probe.mse-set")); err != nil {
		t.Fatalf("Import: %v", err)
	}

	if !slices.Equal(erledigt, []int{0, 1}) {
		t.Errorf("erledigt: %v", erledigt)
	}
}

// **Ein Abbruch vor dem Ablegen legt nichts an** — das Set entsteht in einer
// Transaktion, und bis zu ihr ist nichts geschrieben.
func TestEinAbgebrochenesEinfahrenLegtNichtsAn(t *testing.T) {
	ctx, abbrechen := context.WithCancel(t.Context())
	defer abbrechen()

	ctx = service.MitFortschritt(ctx, func(domain.Fortschritt) { abbrechen() })

	speicher := &speicherDouble{}

	_, err := service.NeuerImporteur(speicher).Set(ctx, filepath.Join(testdaten, "probe.mse-set"))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Fehler: %v", err)
	}

	if speicher.set != nil {
		t.Errorf("angelegt trotz Abbruch: %q", speicher.set.Name)
	}
}
