package app

// **Abschalten beendet den laufenden Dienst** — OPS-10.AK-03
// (#1908).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// merkendeEinstellungen nimmt jeden Wert an und gibt ihn wieder.
type merkendeEinstellungen struct{ werte map[string]string }

func (m *merkendeEinstellungen) Alle(context.Context) (map[string]string, error) {
	return m.werte, nil
}

func (m *merkendeEinstellungen) Setzen(_ context.Context, schluessel, wert string) error {
	m.werte[schluessel] = wert

	return nil
}

func (m *merkendeEinstellungen) Stand(context.Context) (map[string]service.Wahl, error) {
	return map[string]service.Wahl{}, nil
}

// **Ein laufender Dienst endet mit seinem Schalter** — und kommt mit ihm
// nicht von selbst wieder; ein anderer Schalter lässt ihn laufen.
func TestAbschaltenBeendetDenLaufendenDienst(t *testing.T) {
	t.Parallel()

	anwendung, gestartet, _ := einePostenhalterin(t,
		"http://127.0.0.1:8123", "geheim", "", nil)
	anwendung.MitEinstellungen(&merkendeEinstellungen{werte: map[string]string{}})

	gebunden := NeueBindungen(anwendung)
	if antwort := gebunden.RESTStarten(t.Context(), "127.0.0.1", 0); antwort.Fehler != nil {
		t.Fatalf("starten: %+v", antwort.Fehler)
	}

	ctx := context.Background()

	anwendung.EinstellungSetzen(ctx, domain.SchalterEinstellung("spielen"), domain.SchalterAus)

	if !anwendung.dienststand(domain.DienstREST).Laeuft {
		t.Fatal("ein fremder Schalter beendete den Dienst")
	}

	if antwort := anwendung.EinstellungSetzen(ctx,
		domain.SchalterEinstellung("dienst"), domain.SchalterAus); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	if anwendung.dienststand(domain.DienstREST).Laeuft {
		t.Error("der Dienst läuft nach dem Abschalten weiter")
	}

	anwendung.EinstellungSetzen(ctx, domain.SchalterEinstellung("dienst"), domain.SchalterAn)

	if anwendung.dienststand(domain.DienstREST).Laeuft || *gestartet != 1 {
		t.Errorf("der Dienst kam von selbst wieder: %d Starts", *gestartet)
	}
}

// **Eine Fläche ohne Fenster endet mit ihrem eigenen Schalter** — der Dienst
// mit `dienst`, der Werkzeugserver über HTTP mit `werkzeugserver`
// (OPS-10.AK-03). Ein Prozess, den niemand beendet, käme beim Einschalten
// von selbst wieder; deshalb hört er auf, statt nur abzuweisen.
func TestEineFlaecheEndetMitIhremEigenenSchalter(t *testing.T) {
	t.Parallel()

	einstellungen := &merkendeEinstellungen{werte: map[string]string{}}
	beendet := 0
	tor := Schalterwache{einstellungen: einstellungen}.Tor(
		domain.FlaecheREST, "dienst", func() { beendet++ })

	if fehler := tor(context.Background(), "GET /grenzen"); fehler != nil || beendet != 0 {
		t.Fatalf("an: %v, %d", fehler, beendet)
	}

	einstellungen.werte[domain.SchalterEinstellung("dienst")] = domain.SchalterAus

	var abgeschaltet *domain.Abschaltung
	if fehler := tor(context.Background(), "GET /grenzen"); !errors.As(fehler, &abgeschaltet) ||
		abgeschaltet.Schalter != "dienst" || beendet != 1 {
		t.Errorf("aus: %v, %d", fehler, beendet)
	}
}

// **Ein Weg ohne Schalter und ohne eigenen fragt den Bestand nicht.**
func TestEinWegDesKernsFragtDenBestandNicht(t *testing.T) {
	t.Parallel()

	tor := Schalterwache{einstellungen: zaehlendeEinstellungen{}}.Tor(domain.FlaecheREST, "", nil)

	if fehler := tor(context.Background(), "GET /grenzen"); fehler != nil {
		t.Errorf("%v", fehler)
	}
}

// zaehlendeEinstellungen schlägt fehl, sobald jemand liest.
type zaehlendeEinstellungen struct{}

func (zaehlendeEinstellungen) Alle(context.Context) (map[string]string, error) {
	panic("der Bestand wurde gefragt")
}
