package service_test

// **Die Betriebsschau** — Werkstück O9 von Phase 5.2
// (#1433).
//
// # Was sie beantwortet
//
// Zwei Fragen, und sie sind **nicht** dieselbe: *lebt dieser Prozess* und
// *kann er arbeiten*. Die erste beantwortet schon die Antwort selbst; die
// zweite hängt am Bestand — und **nur** am Bestand.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bestandsdoppel antwortet auf die Probe, wie ihm gesagt wird.
type bestandsdoppel struct {
	fehler  error
	wandert bool
	gefragt int
}

func (b *bestandsdoppel) Probe(context.Context) error {
	b.gefragt++

	return b.fehler
}

func (b *bestandsdoppel) WanderungLaeuft() bool { return b.wandert }

// fremddoppel sagt, was über die fremden Dienste bekannt ist.
type fremddoppel struct{ zustaende []domain.Fremdzustand }

func (f fremddoppel) Fremdzustaende(context.Context) []domain.Fremdzustand {
	return f.zustaende
}

// **Ein Bestand, der antwortet, macht bereit.**
func TestEinAntwortenderBestandMachtBereit(t *testing.T) {
	t.Parallel()

	bestand := &bestandsdoppel{}
	schau := service.NeueBetriebsschau(bestand, fremddoppel{})

	zustand := schau.Zustand(t.Context())

	switch {
	case !zustand.Lebt:
		t.Error("der Prozess galt als tot")
	case !zustand.Bereit():
		t.Errorf("nicht bereit: %+v", zustand)
	case bestand.gefragt != 1:
		t.Errorf("der Bestand wurde %dmal gefragt, erwartet einmal", bestand.gefragt)
	}
}

// **Ohne Bestand ist er zu — und das ist kein Fehler.**
func TestOhneBestandIstErZu(t *testing.T) {
	t.Parallel()

	zustand := service.NeueBetriebsschau(nil, fremddoppel{}).Zustand(t.Context())

	if !zustand.Lebt || zustand.Grund != domain.GrundBestandZu {
		t.Errorf("%+v", zustand)
	}
}

// **Ein Bestand, der nicht antwortet, ist unlesbar.**
func TestEinKlemmenderBestandIstUnlesbar(t *testing.T) {
	t.Parallel()

	schau := service.NeueBetriebsschau(
		&bestandsdoppel{fehler: errors.New("zu")}, fremddoppel{})

	if grund := schau.Zustand(t.Context()).Grund; grund != domain.GrundBestandUnlesbar {
		t.Errorf("Grund %q, erwartet %q", grund, domain.GrundBestandUnlesbar)
	}
}

// **Eine laufende Wanderung geht vor** — sie ist der genauere Grund.
//
// Ein Bestand, der gerade wandert, antwortet auf die Probe womöglich; „bereit"
// wäre er trotzdem nicht, denn das Schema wechselt gerade.
func TestEineLaufendeWanderungGehtVor(t *testing.T) {
	t.Parallel()

	schau := service.NeueBetriebsschau(
		&bestandsdoppel{wandert: true}, fremddoppel{})

	if grund := schau.Zustand(t.Context()).Grund; grund != domain.GrundWanderungLaeuft {
		t.Errorf("Grund %q, erwartet %q", grund, domain.GrundWanderungLaeuft)
	}
}

// **Und die fremden Dienste stehen dabei, ohne etwas zu nehmen.**
//
// Das ist die Abnahme des Issues: „Scryfall-Ausfall macht lokalen Editor nicht
// unbereit."
func TestFremdeAusfaelleStehenDaUndNehmenNichts(t *testing.T) {
	t.Parallel()

	schau := service.NeueBetriebsschau(&bestandsdoppel{}, fremddoppel{
		zustaende: []domain.Fremdzustand{
			{Dienst: domain.DienstScryfall, Erreichbar: false, Grund: "kein Netz"},
		},
	})

	zustand := schau.Zustand(t.Context())

	switch {
	case !zustand.Bereit():
		t.Error("ein fremder Ausfall nahm die Bereitschaft")
	case len(zustand.Fremde) != 1 || zustand.Fremde[0].Erreichbar:
		t.Errorf("%+v", zustand.Fremde)
	}
}

// **Ohne Auskunft über die Fremden bleibt die Liste leer — nicht `nil`** (§3).
func TestOhneFremdauskunftBleibtDieListeLeer(t *testing.T) {
	t.Parallel()

	zustand := service.NeueBetriebsschau(&bestandsdoppel{}, nil).Zustand(t.Context())

	if zustand.Fremde == nil || len(zustand.Fremde) != 0 {
		t.Errorf("%+v", zustand.Fremde)
	}
}
