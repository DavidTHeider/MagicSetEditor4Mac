package app

// **Was die Prüfung einer Karte nennt** — PERF02
// (#2352):
// die Befunde des Plans vor denen der Rasterung, ein Fehler als Fehler.

import (
	"context"
	"errors"
	"image"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

type planprobe struct {
	befunde []domain.Befund
	fehler  error
}

func (p planprobe) Kartenplan(context.Context, *domain.CardRow) (*render.Plan, []domain.Befund, error) {
	return &render.Plan{}, p.befunde, p.fehler
}

type rasterprobe struct {
	befunde []render.Befund
	fehler  error
}

func (r rasterprobe) Zeichne(context.Context, *render.Plan, float64, render.Guete) (*image.NRGBA, []render.Befund, error) {
	return nil, r.befunde, r.fehler
}

func TestDieKartenpruefungNenntPlanVorRaster(t *testing.T) {
	t.Parallel()

	pruefen := kartenpruefer(
		planprobe{befunde: []domain.Befund{{Meldung: "der Stil fehlt"}}},
		rasterprobe{befunde: []render.Befund{{Meldung: "die Schrift fehlt"}}},
	)

	meldungen, err := pruefen(t.Context(), &domain.CardRow{ID: 1})
	if err != nil || len(meldungen) != 2 || meldungen[0] != "der Stil fehlt" || meldungen[1] == "" {
		t.Errorf("%v, %v", meldungen, err)
	}

	for name, fall := range map[string]func(context.Context, *domain.CardRow) ([]string, error){
		"Plan":   kartenpruefer(planprobe{fehler: errors.New("kaputt")}, rasterprobe{}),
		"Raster": kartenpruefer(planprobe{}, rasterprobe{fehler: errors.New("kaputt")}),
	} {
		if _, err := fall(t.Context(), &domain.CardRow{ID: 1}); err == nil {
			t.Errorf("%s: ein Fehler ging verloren", name)
		}
	}
}

func TestDerStandNenntDieKartenMitProblem(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)
	stand := domain.Problemstand{
		SetID: 1, Abdruck: "a", Geprueft: 2, Gesamt: 2, Vollstaendig: true, Beendet: jetzt.Add(-time.Minute),
		Karten: []domain.Kartenproblem{{KarteID: 7, Befunde: []string{"x"}}},
	}

	dto := problemstandAlsDTO(stand, "a", jetzt)
	if dto.Zahl != 1 || !dto.Sicher || dto.Alter != 60 || len(dto.Karten) != 1 || dto.Karten[0].KarteID != "7" || !slices.Equal(dto.Karten[0].Befunde, []string{"x"}) {
		t.Errorf("%+v", dto)
	}
}

// groessenprobe merkt sich, in welcher Größe gerastert wurde.
type groessenprobe struct{ groesse *float64 }

func (g groessenprobe) Zeichne(_ context.Context, _ *render.Plan, vergroesserung float64, _ render.Guete) (*image.NRGBA, []render.Befund, error) {
	*g.groesse = vergroesserung

	return nil, nil, nil
}

// **In halber Größe und mit der Hälfte der Kerne** — PERF02 (#2352): so hielt
// die Prüfung von `Black` die 60 Sekunden (51 s statt 131 s). Was ein Mangel
// ist — eine fehlende Schrift, ein Bild, eine Maske —, hängt nicht an der
// Größe, in der gerastert wird.
func TestDiePruefungRastertInHalberGroesse(t *testing.T) {
	t.Parallel()

	var groesse float64

	if _, err := kartenpruefer(planprobe{}, groessenprobe{&groesse})(t.Context(), &domain.CardRow{ID: 1}); err != nil {
		t.Fatal(err)
	}

	if groesse != 0.5 {
		t.Errorf("gerastert in %g, erwartet 0,5", groesse)
	}

	if n := problemarbeiter(); n < 2 || n > max(2, runtime.GOMAXPROCS(0)) {
		t.Errorf("%d Arbeiter bei %d Kernen", n, runtime.GOMAXPROCS(0))
	}
}
