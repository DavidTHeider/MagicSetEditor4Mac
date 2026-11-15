package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// verlaufsdouble ist ein Verlauf im Arbeitsspeicher.
type verlaufsdouble struct {
	// vorgang hält fest, was der Dienst durchgereicht hat (#1393).
	vorgang string

	eintraege []domain.Verlaufseintrag
	fehler    error

	// gefragt hält fest, wonach gefragt wurde — damit ein Test die durchgereichte
	// Kennung prüfen kann statt nur das Ergebnis.
	gefragt int64
}

func (d *verlaufsdouble) Verlauf(
	_ context.Context, setID int64,
) ([]domain.Verlaufseintrag, error) {
	d.gefragt = setID

	return d.eintraege, d.fehler
}

func (d *verlaufsdouble) VorgangZuruecknehmen(
	_ context.Context, setID int64, vorgang string,
) (domain.Verlaufsschritt, bool, error) {
	d.gefragt, d.vorgang = setID, vorgang

	return d.bewege(true)
}

func (d *verlaufsdouble) VorgangWiederherstellen(
	_ context.Context, setID int64, vorgang string,
) (domain.Verlaufsschritt, bool, error) {
	d.gefragt, d.vorgang = setID, vorgang

	return d.bewege(false)
}

func (d *verlaufsdouble) bewege(zurueck bool) (domain.Verlaufsschritt, bool, error) {
	if d.fehler != nil {
		return domain.Verlaufsschritt{}, false, d.fehler
	}

	if len(d.eintraege) == 0 {
		return domain.Verlaufsschritt{}, false, nil
	}

	eintrag := d.eintraege[len(d.eintraege)-1]
	eintrag.Zurueckgenommen = zurueck

	return domain.Verlaufsschritt{
		Verlaufseintrag: eintrag,
		Umfang:          len(d.eintraege),
	}, true, nil
}

// einEintrag ist eine geschriebene Feldänderung.
func einEintrag() domain.Verlaufseintrag {
	return domain.Verlaufseintrag{
		ID: 7, KarteID: 42, Handlung: "feld", Feld: "name",
		Vorher: "Feuerball", GabVorher: true,
		Nachher: "Eisball", GabNachher: true,
		Urheber: domain.UrheberServer,
	}
}

// **Eine Bewegung nennt die betroffene Karte und ihren Urheber.**
//
// Die Karte, damit die Oberfläche **nur sie** nachlädt (AGENTS.md §7). Den
// Urheber, weil im Zwei-Schreiber-Betrieb sonst niemand sagen kann, wessen
// Änderung eben zurückgenommen wurde (#52).
func TestEineBewegungNenntKarteUndUrheber(t *testing.T) {
	for name, bewegen := range map[string]func(
		*service.Verlaufsdienst, context.Context,
	) (service.Bewegung, error){
		"zurücknehmen": func(d *service.Verlaufsdienst, ctx context.Context) (service.Bewegung, error) {
			return d.Zuruecknehmen(ctx, 3, "")
		},
		"wiederherstellen": func(d *service.Verlaufsdienst, ctx context.Context) (service.Bewegung, error) {
			return d.Wiederherstellen(ctx, 3, "")
		},
	} {
		t.Run(name, func(t *testing.T) {
			double := &verlaufsdouble{eintraege: []domain.Verlaufseintrag{einEintrag()}}

			bewegung, err := bewegen(service.NeuerVerlaufsdienst(double), t.Context())
			if err != nil {
				t.Fatalf("bewegen: %v", err)
			}

			if !bewegung.Bewegt {
				t.Fatal("nichts bewegt")
			}

			if bewegung.KarteID != 42 || bewegung.Feld != "name" {
				t.Errorf("%+v", bewegung)
			}

			if bewegung.Urheber != domain.UrheberServer {
				t.Errorf("Urheber %q", bewegung.Urheber)
			}

			// **Der Umfang kommt mit** — eine Massenänderung ist ein Schritt
			// über vielen Karten, und die Oberfläche soll sagen können, wie
			// viele es waren.
			if bewegung.Umfang != 1 {
				t.Errorf("Umfang %d", bewegung.Umfang)
			}

			if double.gefragt != 3 {
				t.Errorf("gefragt nach Set %d, erwartet 3", double.gefragt)
			}
		})
	}
}

// **Ein leerer Verlauf ist kein Fehler, sondern eine Antwort.**
//
// ⌘Z am Anfang ist eine gewöhnliche Lage; wer daraus einen Fehler machte, zeigte
// eine Meldung für etwas, das niemand falsch gemacht hat (§8).
func TestEinLeererVerlaufBewegtNichtsUndMeldetNichts(t *testing.T) {
	dienst := service.NeuerVerlaufsdienst(&verlaufsdouble{})

	bewegung, err := dienst.Zuruecknehmen(t.Context(), 1, "")
	if err != nil {
		t.Fatalf("zurücknehmen: %v", err)
	}

	if bewegung.Bewegt || bewegung.KarteID != 0 {
		t.Errorf("%+v — erwartet war eine leere Bewegung", bewegung)
	}

	if bewegung.Urheber != "" {
		t.Errorf("Urheber %q — eine Bewegung, die es nicht gab, hat keinen", bewegung.Urheber)
	}
}

// **Der Verlauf kommt so heraus, wie er dasteht.**
func TestDerVerlaufKommtHeraus(t *testing.T) {
	double := &verlaufsdouble{eintraege: []domain.Verlaufseintrag{einEintrag()}}

	verlauf, err := service.NeuerVerlaufsdienst(double).Verlauf(t.Context(), 3)
	if err != nil {
		t.Fatalf("Verlauf: %v", err)
	}

	if len(verlauf) != 1 || verlauf[0].Feld != "name" {
		t.Errorf("%+v", verlauf)
	}
}

// **Jeder der drei Wege meldet seinen Fehler.**
func TestDerVerlaufsdienstMeldetJedenFehler(t *testing.T) {
	kaputt := errors.New("die Platte ist voll")

	for name, rufen := range map[string]func(*service.Verlaufsdienst, context.Context) error{
		"Verlauf": func(d *service.Verlaufsdienst, ctx context.Context) error {
			_, err := d.Verlauf(ctx, 1)

			return err
		},
		"zurücknehmen": func(d *service.Verlaufsdienst, ctx context.Context) error {
			_, err := d.Zuruecknehmen(ctx, 1, "")

			return err
		},
		"wiederherstellen": func(d *service.Verlaufsdienst, ctx context.Context) error {
			_, err := d.Wiederherstellen(ctx, 1, "")

			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			dienst := service.NeuerVerlaufsdienst(&verlaufsdouble{fehler: kaputt})

			if err := rufen(dienst, t.Context()); !errors.Is(err, kaputt) {
				t.Errorf("Fehler %v, erwartet war der durchgereichte", err)
			}
		})
	}
}
