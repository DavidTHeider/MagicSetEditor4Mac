package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Marken an Karten erreichen das Fenster** —
// #1208.

// markenpflegeprobe merkt sich die Aufrufe.
type markenpflegeprobe struct {
	setID, karteID int64
	marke          string
	fehler         error
}

func (m *markenpflegeprobe) Setzen(_ context.Context, setID, karteID int64, marke string) error {
	m.setID, m.karteID, m.marke = setID, karteID, marke

	return m.fehler
}

func (m *markenpflegeprobe) Entfernen(
	_ context.Context, setID, karteID int64, marke string,
) error {
	m.setID, m.karteID, m.marke = setID, karteID, marke

	return m.fehler
}

// markenanwendung gibt eine Anwendung mit dieser Pflege.
func markenanwendung(pflege app.Markenpflege) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitMarken(pflege)

	return anwendung
}

// **Setzen reicht Set, Karte und Marke durch.**
func TestMarkeSetzenReichtDurch(t *testing.T) {
	t.Parallel()

	probe := &markenpflegeprobe{}

	if antwort := markenanwendung(probe).
		MarkeSetzen(probekontext(t), "1", "7", "fertig"); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.setID != 1 || probe.karteID != 7 || probe.marke != "fertig" {
		t.Errorf("durchgereicht: %+v", probe)
	}
}

// **Entfernen ebenso.**
func TestMarkeEntfernenReichtDurch(t *testing.T) {
	t.Parallel()

	probe := &markenpflegeprobe{}

	if antwort := markenanwendung(probe).
		MarkeEntfernen(probekontext(t), "1", "7", "fertig"); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.marke != "fertig" {
		t.Errorf("durchgereicht: %+v", probe)
	}
}

// **Eine Kennung, die keine ist, wird an beiden Wegen abgewiesen — und zwar
// die Set- wie die Kartenkennung.**
func TestMarkenwegeWeisenEineUngueltigeKennungAb(t *testing.T) {
	t.Parallel()

	anwendung := markenanwendung(&markenpflegeprobe{})

	for name, fehler := range map[string]*app.Fehler{
		"MarkeSetzen ohne Set":      anwendung.MarkeSetzen(t.Context(), "keine", "7", "a").Fehler,
		"MarkeSetzen ohne Karte":    anwendung.MarkeSetzen(t.Context(), "1", "keine", "a").Fehler,
		"MarkeEntfernen ohne Set":   anwendung.MarkeEntfernen(t.Context(), "keine", "7", "a").Fehler,
		"MarkeEntfernen ohne Karte": anwendung.MarkeEntfernen(t.Context(), "1", "keine", "a").Fehler,
	} {
		if fehler == nil {
			t.Errorf("%s ließ eine unbrauchbare Kennung durch", name)
		}
	}
}

// **Und ein Fehler des Dienstes wird übersetzt.**
func TestEinMarkenfehlerWirdUebersetzt(t *testing.T) {
	t.Parallel()

	probe := &markenpflegeprobe{fehler: errors.New("planmäßig gescheitert")}
	anwendung := markenanwendung(probe)

	if antwort := anwendung.MarkeSetzen(t.Context(), "1", "7", "a"); antwort.Fehler == nil {
		t.Error("der Fehler ging verloren")
	}

	probe.fehler = domain.NewFieldError("marke", domain.SchluesselPflichtfeld, "leer")

	antwort := anwendung.MarkeSetzen(t.Context(), "1", "7", "")
	if antwort.Fehler == nil || antwort.Fehler.Feld != "marke" {
		t.Errorf("kein Feldbezug: %+v", antwort.Fehler)
	}
}
