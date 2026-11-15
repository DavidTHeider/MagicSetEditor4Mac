package app_test

// **Die Naht zwischen Pflege und Fläche** — Werkstück E von Phase 4.11
// (#1341).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// regelweg baut den Weg über einer Ablage und einem Spiel.
func regelweg(regeln []domain.Autoersetzungsregel, an bool) *app.Autoersetzungsweg {
	return app.NeuerAutoersetzungsweg(service.NeueAutoersetzungspflege(
		&ersetzablageprobe{regeln: regeln, an: an},
		&spielprobeErsetzungApp{spiel: "probe"}))
}

// **Die Liste kommt als DTO — mit Schalter.**
func TestDerAutoersetzungswegGibtDieListeAlsDTO(t *testing.T) {
	t.Parallel()

	regeln, an, err := regelweg([]domain.Autoersetzungsregel{
		{Muster: "(W)", Ersatz: "weiss", GanzesWort: true, Aktiv: true, Eigen: true},
	}, true).Liste(probekontext(t), 1)
	if err != nil {
		t.Fatalf("Liste: %v", err)
	}

	if !an || len(regeln) != 1 {
		t.Fatalf("die Liste lautet %+v / %v", regeln, an)
	}

	if regeln[0] != (app.AutoersetzungsregelDTO{
		Muster: "(W)", Ersatz: "weiss", GanzesWort: true, Aktiv: true, Eigen: true,
	}) {
		t.Errorf("die Regel lautet %+v", regeln[0])
	}
}

// **Und die vier Schreibwege gehen durch.**
func TestDerAutoersetzungswegReichtDieSchreibwegeDurch(t *testing.T) {
	t.Parallel()

	weg := regelweg([]domain.Autoersetzungsregel{{Muster: "(W)", Aktiv: true}}, true)

	if err := weg.Setzen(probekontext(t), 1, app.AutoersetzungsregelDTO{
		Muster: "(U)", Ersatz: "blau", GanzesWort: true, Aktiv: true,
	}); err != nil {
		t.Errorf("Setzen: %v", err)
	}

	if err := weg.Loeschen(probekontext(t), 1, "(W)"); err != nil {
		t.Errorf("Loeschen: %v", err)
	}

	if err := weg.Zuruecksetzen(probekontext(t), 1); err != nil {
		t.Errorf("Zuruecksetzen: %v", err)
	}

	if err := weg.SchalterSetzen(probekontext(t), 1, false); err != nil {
		t.Errorf("SchalterSetzen: %v", err)
	}
}

// **Und die wirksamen Regeln lassen die abgeschalteten aus.**
//
//	if (!ar->enabled) continue;
//
// `src/gui/value/text.cpp`, Zeile 1038, Fassung f9d9356 vom 2021-07-06.
func TestDerAutoersetzungswegLaesstAbgeschalteteAus(t *testing.T) {
	t.Parallel()

	wirksam, err := regelweg([]domain.Autoersetzungsregel{
		{Muster: "(W)", Ersatz: "weiss", Aktiv: true},
		{Muster: "(U)", Ersatz: "blau", Aktiv: false},
	}, true).Autoersetzungen(probekontext(t), 1)
	if err != nil {
		t.Fatalf("Autoersetzungen: %v", err)
	}

	if len(wirksam) != 1 || wirksam[0].Muster != "(W)" {
		t.Errorf("wirksam sind %+v", wirksam)
	}
}

// **Ein Fehler der Pflege kommt durch — an beiden Lesewegen.**
func TestEinFehlerDerPflegeKommtDurchDenAutoersetzungsweg(t *testing.T) {
	t.Parallel()

	eigener := errors.New("kein Set")

	weg := app.NeuerAutoersetzungsweg(service.NeueAutoersetzungspflege(
		&ersetzablageprobe{}, &spielfehlerprobe{fehler: eigener}))

	if _, _, err := weg.Liste(probekontext(t), 1); !errors.Is(err, eigener) {
		t.Errorf("Liste gab %v", err)
	}

	if _, err := weg.Autoersetzungen(probekontext(t), 1); !errors.Is(err, eigener) {
		t.Errorf("Autoersetzungen gab %v", err)
	}
}

// spielfehlerprobe scheitert am Spielnamen.
type spielfehlerprobe struct{ fehler error }

func (s *spielfehlerprobe) SpielVon(context.Context, int64) (string, error) {
	return "", s.fehler
}

func (s *spielfehlerprobe) Paketregeln(context.Context, int64) (
	[]msepaket.AutoErsetzung, error,
) {
	return nil, nil
}
