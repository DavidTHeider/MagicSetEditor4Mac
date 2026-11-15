package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// verlaufsprobe antwortet auf beide Richtungen dasselbe.
type verlaufsprobe struct {
	bewegung service.Bewegung
	fehler   error

	// gefragt hält die durchgereichte Kennung fest.
	gefragt int64

	// eintraege ist der Verlauf, den `Verlauf` gibt (U6, #1417).
	eintraege []domain.Verlaufseintrag

	// vorgang ist die Kennung, die die letzte Bewegung nannte.
	vorgang string
}

func (p *verlaufsprobe) Verlauf(_ context.Context, setID int64) ([]domain.Verlaufseintrag, error) {
	p.gefragt = setID

	return p.eintraege, p.fehler
}

func (p *verlaufsprobe) Zuruecknehmen(
	_ context.Context, setID int64, vorgang string,
) (service.Bewegung, error) {
	p.gefragt = setID
	p.vorgang = vorgang

	return p.bewegung, p.fehler
}

func (p *verlaufsprobe) Wiederherstellen(
	_ context.Context, setID int64, vorgang string,
) (service.Bewegung, error) {
	p.gefragt = setID
	p.vorgang = vorgang

	return p.bewegung, p.fehler
}

// mitVerlauf baut eine Anwendung über einer Probe.
func mitVerlauf(probe *verlaufsprobe) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitVerlauf(probe)

	return anwendung
}

// **Eine Bewegung kommt als DTO an — Karte als Zeichenkette, Urheber als Text.**
func TestDieBewegungKommtAlsDTOAn(t *testing.T) {
	for name, bewegen := range map[string]func(*app.App, context.Context) app.BewegungAntwort{
		"zurücknehmen": func(a *app.App, ctx context.Context) app.BewegungAntwort {
			return a.Zuruecknehmen(ctx, "3", "")
		},
		"wiederherstellen": func(a *app.App, ctx context.Context) app.BewegungAntwort {
			return a.Wiederherstellen(ctx, "3", "")
		},
	} {
		t.Run(name, func(t *testing.T) {
			probe := &verlaufsprobe{bewegung: service.Bewegung{
				Bewegt: true, KarteID: 42, Feld: "name",
				Umfang: 3, Urheber: domain.UrheberServer,
			}}

			antwort := bewegen(mitVerlauf(probe), probekontext(t))
			if antwort.Fehler != nil {
				t.Fatalf("%+v", antwort.Fehler)
			}

			if !antwort.Bewegt || antwort.KarteID != "42" || antwort.Feld != "name" {
				t.Errorf("%+v", antwort)
			}

			// **Der Umfang kommt durch** — sonst ginge die Zusage „eine
			// Massenänderung ist **ein** Schritt" an dieser Naht verloren.
			if antwort.Umfang != 3 {
				t.Errorf("Umfang %d", antwort.Umfang)
			}

			if antwort.Urheber != string(domain.UrheberServer) {
				t.Errorf("Urheber %q", antwort.Urheber)
			}

			if probe.gefragt != 3 {
				t.Errorf("gefragt nach Set %d", probe.gefragt)
			}
		})
	}
}

// **Eine unlesbare Kennung erreicht den Bestand gar nicht.**
func TestEineUnlesbareKennungBewegtNichts(t *testing.T) {
	probe := &verlaufsprobe{}

	antwort := mitVerlauf(probe).Zuruecknehmen(probekontext(t), "keine Zahl", "")
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}

	if probe.gefragt != 0 {
		t.Errorf("der Bestand wurde nach Set %d gefragt", probe.gefragt)
	}
}

// **Ein Fehler aus dem Dienst kommt übersetzt heraus**, nicht roh (§8).
func TestEinFehlerDesVerlaufsKommtUebersetzt(t *testing.T) {
	probe := &verlaufsprobe{fehler: errors.New("die Platte ist voll")}

	antwort := mitVerlauf(probe).Wiederherstellen(probekontext(t), "1", "")
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}

	if antwort.Bewegt {
		t.Error("bewegt, obwohl es scheiterte")
	}
}

// **Eine Bewegung, die es nicht gab, ist kein Fehler** — und nennt Karte `"0"`.
//
// Nicht ein fehlendes Feld: `bewegt` trifft die Fallunterscheidung schon, und
// ein `undefined` daneben zwänge das Frontend zu einer zweiten.
func TestEineAusgebliebeneBewegungIstKeinFehler(t *testing.T) {
	antwort := mitVerlauf(&verlaufsprobe{}).Zuruecknehmen(probekontext(t), "1", "")

	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Bewegt || antwort.KarteID != "0" || antwort.Urheber != "" {
		t.Errorf("%+v", antwort)
	}
}

// **Das Fenster nimmt gezielt zurück** — Werkstück U6 von Phase 5.1 (#1417):
// aus der Verlaufsliste heraus genau den genannten Schritt, wie es die drei
// anderen Flächen seit #1393 können. Ohne Vorgang bleibt es ⌘Z: der jüngste.
func TestDieBewegungReichtDenVorgangDurch(t *testing.T) {
	probe := &verlaufsprobe{bewegung: service.Bewegung{Bewegt: true, KarteID: 1}}
	anwendung := mitVerlauf(probe)

	if antwort := anwendung.Zuruecknehmen(probekontext(t), "3", "g1"); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.vorgang != "g1" {
		t.Errorf("zurückgenommen wurde %q, erwartet war g1", probe.vorgang)
	}

	if antwort := anwendung.Wiederherstellen(probekontext(t), "3", ""); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.vorgang != "" {
		t.Errorf("wiederhergestellt wurde %q, erwartet war der jüngste", probe.vorgang)
	}
}
