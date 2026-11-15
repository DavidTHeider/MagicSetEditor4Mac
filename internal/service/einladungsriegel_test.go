package service_test

// **Eine Annahme, die sich hinaus meldet, fragt nach `externSenden`** —
// SEC-P4 (#2000).
//
// # Warum hier und nicht an der Fläche
//
// **Ob gemeldet wird, steht in der Einladung** — sie ist einlösbar oder
// nicht —, und das ist ein Domänenzustand (§2): an der Fläche wäre es ein
// `if` über ihn. Die Annahme selbst bleibt `lokalSchreiben`; erst die Meldung
// schickt den eigenen Namen und die eigene Adresse zu der Adresse, die der
// Einladende **selbst** in seine Einladung geschrieben hat.

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bertMitRiegel baut Berts Werk mit einem Riegel.
func bertMitRiegel(t *testing.T, riegel service.Pruefer) (*service.Einladungswerk, *einladungsablage) {
	t.Helper()

	ablage := neueEinladungsablage()

	return service.NeuesEinladungswerk(festerAusweis{ich: mussIdentitaet(t)}, festeRotation{}, ablage,
		func() time.Time { return einladungsuhr }).MitRiegel(riegel), ablage
}

func TestEineEinloesbareAnnahmeFragtNachDemSenden(t *testing.T) {
	t.Parallel()

	_, _, _, code, _ := annaStelltAus(t)
	nein := &neinsager{}
	werk, ablage := bertMitRiegel(t, nein)

	if _, fehler := werk.Annehmen(t.Context(), service.Einladungsannahme{
		Code: code, MeinName: "Bert",
	}); !errors.Is(fehler, domain.ErrNurLesend) {
		t.Errorf("%v statt der Absage des Riegels", fehler)
	}

	if !slices.Equal(nein.gefragt, []domain.Wirkung{domain.WirkungExternSenden}) {
		t.Errorf("gefragt wurde nach %v", nein.gefragt)
	}

	// **Ganz oder gar nicht**: ohne Meldung keine Annahme.
	if len(ablage.kontakte) != 0 {
		t.Errorf("trotzdem angenommen: %+v", ablage.kontakte)
	}

	// **Und gemeldet wird erst recht nicht** — der Bote wählt nie.
	gewaehlt := false
	bote := func(context.Context, peer.Kennung) (service.Einloeseleitung, error) {
		gewaehlt = true

		return &probeleitung{}, nil
	}

	if fehler := werk.Melden(t.Context(), bote, code, domain.Einloesung{}); !errors.Is(fehler, domain.ErrNurLesend) {
		t.Errorf("Melden: %v statt der Absage des Riegels", fehler)
	}

	if gewaehlt {
		t.Error("der Bote wählte trotz Absage")
	}
}

// **Eine Einladung, die nicht einlösbar ist, meldet sich nicht** — und fragt
// deshalb nicht nach dem Senden: sie anzunehmen bleibt `lokalSchreiben`.
func TestEineStilleAnnahmeFragtNichtNachDemSenden(t *testing.T) {
	t.Parallel()

	_, anna, _, _, _ := annaStelltAus(t)

	ohneAdresse, fehler := anna.Erstellen(t.Context(), service.Einladungsbitte{
		Name: "Anna", Zweck: domain.ZweckHandel,
	})
	if fehler != nil {
		t.Fatal(fehler)
	}

	nein := &neinsager{}
	werk, ablage := bertMitRiegel(t, nein)

	if _, fehler := werk.Annehmen(t.Context(), service.Einladungsannahme{
		Code: ohneAdresse.Code,
	}); fehler != nil {
		t.Fatalf("Annehmen: %v", fehler)
	}

	if len(nein.gefragt) != 0 || len(ablage.kontakte) != 1 {
		t.Errorf("gefragt %v, angenommen %d", nein.gefragt, len(ablage.kontakte))
	}
}
