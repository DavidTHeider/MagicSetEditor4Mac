package service_test

// **Eine Exportvorlage zählt ihre Bilder und lässt sich anhalten** —
// FMT-03.AK-02 (#1692).
//
// **Was schon geschrieben ist, bleibt liegen** — wie bei einem Fehler: MSE
// schreibt inkrementell, und #750 nennt die Zahl, statt aufzuräumen.

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// zaehlenderBildschreiber zählt, was er schreiben sollte.
type zaehlenderBildschreiber struct{ dateien []string }

func (z *zaehlenderBildschreiber) Schreibe(_ context.Context, _ int64,
	verzeichnis, datei string, _ service.Bildformat, _ int,
) (*service.Bildschreibung, error) {
	z.dateien = append(z.dateien, datei)

	return &service.Bildschreibung{Pfad: filepath.Join(verzeichnis, datei)}, nil
}

const zweiBilder = `write_image_file(cards.0, file: "a.png") + write_image_file(cards.0, file: "b.png")`

func TestEineVorlageZaehltIhreBilder(t *testing.T) {
	var erledigt []int

	ctx := service.MitFortschritt(t.Context(), func(stand domain.Fortschritt) {
		if stand.Einheit != domain.EinheitBild {
			t.Errorf("Stand: %+v", stand)
		}

		erledigt = append(erledigt, stand.Erledigt)
	})

	schreiber := &zaehlenderBildschreiber{}

	if _, _, err := probelauf(service.Ausfuhrmittel{
		Bilder: schreiber, Bildziel: "/ziel", Verzeichnis: "ausfuhr-files",
	}).Fuehre(ctx, vorlageMit(zweiBilder), eineKarte(), nil, nil); err != nil {
		t.Fatalf("ausführen: %v", err)
	}

	if !slices.Equal(erledigt, []int{0, 1}) || len(schreiber.dateien) != 2 {
		t.Errorf("erledigt %v, geschrieben %v", erledigt, schreiber.dateien)
	}
}

func TestEineAbgebrocheneVorlageSchreibtNichtWeiter(t *testing.T) {
	ctx, abbrechen := context.WithCancel(t.Context())
	defer abbrechen()

	ctx = service.MitFortschritt(ctx, func(stand domain.Fortschritt) {
		if stand.Erledigt == 1 {
			abbrechen()
		}
	})

	schreiber := &zaehlenderBildschreiber{}

	_, _, err := probelauf(service.Ausfuhrmittel{
		Bilder: schreiber, Bildziel: "/ziel", Verzeichnis: "ausfuhr-files",
	}).Fuehre(ctx, vorlageMit(zweiBilder), eineKarte(), nil, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Fehler: %v", err)
	}

	if !slices.Equal(schreiber.dateien, []string{"a.png"}) {
		t.Errorf("geschrieben: %v", schreiber.dateien)
	}
}
