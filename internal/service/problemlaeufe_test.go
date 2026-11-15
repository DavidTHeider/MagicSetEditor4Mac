package service

// **Ein Lauf je Set, und eine Antwort gehört zu ihrem Stand** — Befunde 3, 4
// und 15 des Zweitblicks über Phase 3 von 0.9.0 (PERF02, #2352).

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Erst der Abdruck, dann die Karten** (Befund 3): ändert sich das Set,
// nachdem die Karten gelesen sind, trägt der Stand den alten Abdruck — und
// heißt veraltet. Andersherum trüge er den neuen über alte Karten.
func TestDerAbdruckStehtVorDenKarten(t *testing.T) {
	t.Parallel()

	stand := "a"
	pruefung := pruefungMit(probekarten(3), nil, "a")
	pruefung.abdruck = func(context.Context, int64) (string, error) { return stand, nil }
	pruefung.karten = func(context.Context, int64) ([]*domain.CardRow, error) {
		karten := probekarten(3)
		stand = "b" // eine Karte kommt dazu, eben nachdem gelesen ist

		return karten, nil
	}

	if _, err := pruefung.Pruefen(t.Context(), 1, nil); err != nil {
		t.Fatal(err)
	}

	gelesen, heute, err := pruefung.Stand(t.Context(), 1)
	if err != nil || gelesen.Urteil(heute) != domain.ProblemeVeraltet {
		t.Errorf("Urteil %q über %q gegen %q (%v) — alte Karten mit neuem Abdruck", gelesen.Urteil(heute), gelesen.Abdruck, heute, err)
	}
}

// **Ein neuerer Lauf löst den älteren ab** (Befund 4): der ältere bricht ab
// und legt nichts ab, auch wenn er nach dem neueren endet.
func TestEinNeuererLaufLoestDenAelterenAb(t *testing.T) {
	t.Parallel()

	pruefung := pruefungMit(probekarten(2), nil, "a")
	pruefung.arbeiter = 1

	alterLaeuft := make(chan struct{})
	pruefung.pruefen = func(ctx context.Context, karte *domain.CardRow) ([]string, error) {
		if karte.ID == 1 && ctx.Value(alterSchluessel{}) != nil {
			close(alterLaeuft)
			<-ctx.Done()

			return nil, ctx.Err()
		}

		return []string{"neu"}, nil
	}

	alt := make(chan error, 1)

	go func() {
		_, err := pruefung.Pruefen(context.WithValue(t.Context(), alterSchluessel{}, true), 1, nil)
		alt <- err
	}()

	<-alterLaeuft

	neu, err := pruefung.Pruefen(t.Context(), 1, nil)
	if err != nil || !neu.Vollstaendig || len(neu.Karten) != 2 {
		t.Fatalf("der neue Lauf: %+v, %v", neu, err)
	}

	if abgeloest := <-alt; !errors.Is(abgeloest, domain.ErrConflict) {
		t.Errorf("der alte Lauf endete mit %v — erwartet: abgelöst", abgeloest)
	}

	gelesen, _, err := pruefung.Stand(t.Context(), 1)
	if err != nil || !gelesen.Vollstaendig || len(gelesen.Karten) != 2 {
		t.Errorf("abgelegt ist %+v (%v) — der abgelöste Lauf hat den neuen verdrängt", gelesen, err)
	}
}

type alterSchluessel struct{}

// **Ein Abbruch verdrängt keinen vollständigen Stand desselben Abdrucks** —
// der gilt weiter; einen anderen Abdrucks schon, denn der ist veraltet.
func TestEinAbbruchVerdraengtKeinenGueltigenStand(t *testing.T) {
	t.Parallel()

	pruefung := pruefungMit(probekarten(4), map[int64][]string{1: {"alt"}}, "a")
	pruefung.arbeiter = 1

	if _, err := pruefung.Pruefen(t.Context(), 1, nil); err != nil {
		t.Fatal(err)
	}

	ctx, abbrechen := context.WithCancel(t.Context())
	if _, err := pruefung.Pruefen(ctx, 1, func(geprueft, _ int) {
		if geprueft == 1 {
			abbrechen()
		}
	}); !errors.Is(err, context.Canceled) {
		t.Fatalf("Abbruch: %v", err)
	}

	if gelesen, _, _ := pruefung.Stand(t.Context(), 1); !gelesen.Vollstaendig { //nolint:errcheck // geprüft wird der Stand
		t.Errorf("der vollständige Stand wich einem Abbruch: %+v", gelesen)
	}
}

// **Die Stände sind begrenzt** (Befund 15): über der Grenze geht der älteste,
// und ein Set, das es nicht mehr gibt, nimmt seinen mit.
func TestDieStaendeSindBegrenzt(t *testing.T) {
	t.Parallel()

	pruefung := pruefungMit(probekarten(1), nil, "a")
	uhr := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)
	pruefung.uhr = func() time.Time { uhr = uhr.Add(time.Second); return uhr }

	for setID := int64(1); setID <= ProblemstaendeHoechstens+1; setID++ {
		if _, err := pruefung.Pruefen(t.Context(), setID, nil); err != nil {
			t.Fatal(err)
		}
	}

	if len(pruefung.staende) != ProblemstaendeHoechstens {
		t.Errorf("%d Stände, erwartet %d", len(pruefung.staende), ProblemstaendeHoechstens)
	}

	if _, da := pruefung.staende[1]; da {
		t.Error("der älteste Stand blieb")
	}

	pruefung.abdruck = func(context.Context, int64) (string, error) { return "", domain.ErrNotFound }

	if _, _, err := pruefung.Stand(t.Context(), 2); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ein gelöschtes Set: %v", err)
	}

	if _, da := pruefung.staende[2]; da {
		t.Error("der Stand eines gelöschten Sets blieb")
	}
}

// **Der Fortschritt reist auch über den Kontext** (Befund 11): ein Auftrag
// des Dienstes und ein Werkzeug über MCP hören zu, ohne einen eigenen Melder
// hineinzureichen.
func TestDiePruefungMeldetImKontext(t *testing.T) {
	t.Parallel()

	var gehoert []domain.Fortschritt

	ctx := MitFortschritt(t.Context(), func(stand domain.Fortschritt) { gehoert = append(gehoert, stand) })

	if _, err := pruefungMit(probekarten(3), nil, "a").Pruefen(ctx, 1, nil); err != nil {
		t.Fatal(err)
	}

	if len(gehoert) != 3 || gehoert[2] != (domain.Fortschritt{Erledigt: 3, Gesamt: 3, Einheit: domain.EinheitKarte}) {
		t.Errorf("gehört %+v", gehoert)
	}
}
