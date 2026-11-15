package service

// **Die Problemprüfung über ein ganzes Set** — PERF02
// (#2352).
// Gegen Attrappen: Karten, Prüfer und Abdruck sind Nähte.

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// probekarten gibt `n` Karten mit den Kennungen 1 bis n.
func probekarten(n int) []*domain.CardRow {
	karten := make([]*domain.CardRow, n)
	for i := range karten {
		karten[i] = &domain.CardRow{ID: int64(i + 1)}
	}

	return karten
}

// pruefungMit baut eine Prüfung über `karten`; `befunde` sagt je Kennung,
// was der Prüfer findet, und `abdruecke` gibt der Reihe nach den Abdruck.
func pruefungMit(karten []*domain.CardRow, befunde map[int64][]string, abdruecke ...string) *Problempruefung {
	var (
		mu  sync.Mutex
		ruf int
		uhr = time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)
	)

	return NeueProblempruefung(
		func(context.Context, int64) ([]*domain.CardRow, error) { return karten, nil },
		func(_ context.Context, karte *domain.CardRow) ([]string, error) { return befunde[karte.ID], nil },
		func(context.Context, int64) (string, error) {
			mu.Lock()
			defer mu.Unlock()

			abdruck := abdruecke[min(ruf, len(abdruecke)-1)]
			ruf++

			return abdruck, nil
		},
		func() time.Time { return uhr },
	)
}

func TestDiePruefungFindetJedeKarteMitProblem(t *testing.T) {
	t.Parallel()

	pruefung := pruefungMit(probekarten(5), map[int64][]string{2: {"Schrift fehlt"}, 4: {"Bild fehlt", "Text zu lang"}}, "a")

	var gemeldet []int

	stand, err := pruefung.Pruefen(t.Context(), 1, func(geprueft, _ int) { gemeldet = append(gemeldet, geprueft) })
	if err != nil {
		t.Fatal(err)
	}

	if stand.Urteil("a") != domain.ProblemeAktuell || stand.Geprueft != 5 || stand.Gesamt != 5 {
		t.Errorf("Stand %+v", stand)
	}

	if len(stand.Karten) != 2 || stand.Karten[0].KarteID != 2 || stand.Karten[1].KarteID != 4 || len(stand.Karten[1].Befunde) != 2 {
		t.Errorf("Karten %+v", stand.Karten)
	}

	if len(gemeldet) != 5 || gemeldet[4] != 5 {
		t.Errorf("Fortschritt %v", gemeldet)
	}

	// **Der letzte Stand bleibt lesbar**, mit dem heutigen Abdruck daneben.
	gelesen, heute, err := pruefung.Stand(t.Context(), 1)
	if err != nil || gelesen.Geprueft != 5 || heute != "a" {
		t.Errorf("gelesen %+v, heute %q, %v", gelesen, heute, err)
	}
}

// **Eine Karte, die sich nicht prüfen lässt, hat ein Problem** — sie
// verschwindet nicht aus der Zählung, und die Prüfung geht weiter.
func TestEineUnpruefbareKarteZaehltAlsProblem(t *testing.T) {
	t.Parallel()

	pruefung := pruefungMit(probekarten(3), nil, "a")
	pruefung.pruefen = func(_ context.Context, karte *domain.CardRow) ([]string, error) {
		if karte.ID == 2 {
			return nil, errors.New("der Stil fehlt")
		}

		return nil, nil
	}

	stand, err := pruefung.Pruefen(t.Context(), 1, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !stand.Vollstaendig || len(stand.Karten) != 1 || stand.Karten[0].KarteID != 2 || stand.Karten[0].Befunde[0] != "der Stil fehlt" {
		t.Errorf("Stand %+v", stand)
	}
}

// **Ein Abbruch gibt keinen vollständigen Stand** — und das falsche Null aus
// dem Issue bleibt aus.
func TestEinAbbruchGibtKeinFalschesNull(t *testing.T) {
	t.Parallel()

	ctx, abbrechen := context.WithCancel(t.Context())
	pruefung := pruefungMit(probekarten(10), nil, "a")
	pruefung.arbeiter = 1

	stand, err := pruefung.Pruefen(ctx, 1, func(geprueft, _ int) {
		if geprueft == 3 {
			abbrechen()
		}
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Fehler %v", err)
	}

	zahl, sicher := stand.Zahl("a")
	if stand.Vollstaendig || stand.Geprueft >= 10 || sicher || zahl != 0 {
		t.Errorf("Stand %+v: %d, sicher %v", stand, zahl, sicher)
	}

	gelesen, _, err := pruefung.Stand(t.Context(), 1)
	if err != nil || gelesen.Urteil("a") != domain.ProblemeUnvollstaendig {
		t.Errorf("gelesen %+v, %v", gelesen, err)
	}
}

// **Ein Wechsel während der Prüfung macht den Stand veraltet** — er trägt
// den Abdruck, mit dem er begann.
func TestEinWechselWaehrendDerPruefungMachtDenStandVeraltet(t *testing.T) {
	t.Parallel()

	pruefung := pruefungMit(probekarten(3), nil, "a", "b")

	stand, err := pruefung.Pruefen(t.Context(), 1, nil)
	if err != nil {
		t.Fatal(err)
	}

	gelesen, heute, err := pruefung.Stand(t.Context(), 1)
	if err != nil || stand.Abdruck != "a" || heute != "b" || gelesen.Urteil(heute) != domain.ProblemeVeraltet {
		t.Errorf("Stand %+v, heute %q, %v", gelesen, heute, err)
	}
}

func TestOhneKartenOderAbdruckScheitertDiePruefung(t *testing.T) {
	t.Parallel()

	keineKarten := pruefungMit(nil, nil, "a")
	keineKarten.karten = func(context.Context, int64) ([]*domain.CardRow, error) { return nil, domain.ErrNotFound }

	if _, err := keineKarten.Pruefen(t.Context(), 9, nil); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ohne Set: %v", err)
	}

	keinAbdruck := pruefungMit(probekarten(1), nil, "a")
	keinAbdruck.abdruck = func(context.Context, int64) (string, error) { return "", errors.New("kein Stil") }

	if _, err := keinAbdruck.Pruefen(t.Context(), 1, nil); err == nil {
		t.Error("ohne Abdruck geprüft")
	}

	if _, _, err := keinAbdruck.Stand(t.Context(), 1); err == nil {
		t.Error("ohne Abdruck gelesen")
	}
}

// **Eine Karte, deren Prüfung der Abbruch beendete, hat kein Problem** — sie
// ist nicht geprüft und zählt nicht mit.
func TestEineAbgebrocheneKarteZaehltNichtAlsProblem(t *testing.T) {
	t.Parallel()

	ctx, abbrechen := context.WithCancel(t.Context())
	pruefung := pruefungMit(probekarten(2), nil, "a")
	pruefung.arbeiter = 1
	pruefung.pruefen = func(ctx context.Context, _ *domain.CardRow) ([]string, error) {
		abbrechen()
		<-ctx.Done()

		return nil, ctx.Err()
	}

	stand, err := pruefung.Pruefen(ctx, 1, nil)
	if !errors.Is(err, context.Canceled) || stand.Geprueft != 0 || len(stand.Karten) != 0 || stand.Vollstaendig {
		t.Errorf("Stand %+v, %v", stand, err)
	}
}

// **So viele Arbeiter, wie die Wurzel sagt** — der Plan einer Karte kostet so
// viel wie ihre Rasterung und steht nicht hinter der Pforte des Rasterers;
// mit zwei Arbeitern brauchte `Black` 131 s statt höchstens 60.
func TestDieWurzelBestimmtDieZahlDerArbeiter(t *testing.T) {
	t.Parallel()

	var (
		mu                   sync.Mutex
		zugleich, hoechstens int
		alleDa               = make(chan struct{})
		einmal               sync.Once
	)

	pruefung := pruefungMit(probekarten(8), nil, "a").MitArbeitern(4)
	pruefung.pruefen = func(context.Context, *domain.CardRow) ([]string, error) {
		mu.Lock()
		zugleich++
		if zugleich > hoechstens {
			hoechstens = zugleich
		}
		if zugleich == 4 {
			einmal.Do(func() { close(alleDa) })
		}
		mu.Unlock()

		// **Mit Frist**: zu wenige Arbeiter füllen die Schranke nie.
		select {
		case <-alleDa:
		case <-time.After(2 * time.Second):
		}

		mu.Lock()
		zugleich--
		mu.Unlock()

		return nil, nil
	}

	if _, err := pruefung.Pruefen(t.Context(), 1, nil); err != nil {
		t.Fatal(err)
	}

	if hoechstens != 4 {
		t.Errorf("zugleich geprüft: %d, erwartet 4", hoechstens)
	}

	// **Weniger als einer ist einer** — sonst prüfte niemand.
	if einer := pruefungMit(probekarten(2), nil, "a").MitArbeitern(0); einer.arbeiter != 1 {
		t.Errorf("Arbeiter %d", einer.arbeiter)
	}
}
