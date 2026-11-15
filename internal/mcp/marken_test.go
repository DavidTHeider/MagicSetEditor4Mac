// **Die Markenwerkzeuge** — Werkstück 3 von Phase 3.5
// (#1208).
package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// markenprobe hält Marken im Arbeitsspeicher.
type markenprobe struct {
	marken  map[int64][]string
	fehler  error
	gesetzt string
}

func neueMarkenprobe() *markenprobe {
	return &markenprobe{marken: map[int64][]string{}}
}

func (m *markenprobe) MarkeSetzen(
	_ context.Context, _, karteID int64, marke string,
) error {
	if m.fehler != nil {
		return m.fehler
	}

	m.gesetzt = domain.Markennormalform(marke)
	m.marken[karteID] = append(m.marken[karteID], m.gesetzt)

	return nil
}

func (m *markenprobe) MarkeEntfernen(
	_ context.Context, _, karteID int64, marke string,
) error {
	if m.fehler != nil {
		return m.fehler
	}

	gesucht := domain.Markennormalform(marke)
	behalten := m.marken[karteID][:0]

	for _, hat := range m.marken[karteID] {
		if hat != gesucht {
			behalten = append(behalten, hat)
		}
	}

	m.marken[karteID] = behalten

	return nil
}

func (m *markenprobe) KartenMarken(_ context.Context, _, karteID int64) ([]string, error) {
	return m.marken[karteID], m.fehler
}

func (m *markenprobe) Setmarken(context.Context, int64) ([]string, error) {
	return m.marken[7], m.fehler
}

// rufeMarke sucht ein Markenwerkzeug und ruft es.
func rufeMarke(
	t *testing.T, marken mcp.Fremdmarken, name, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: probenbestand(), Schreiber: probenbestand(),
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Angaben:     neueAngabenprobe(), Stile: neueStilprobe(),
		Bestand: neueBestandsprobeWeit(),
		Marken:  marken,
	}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatalf("das Werkzeug %s gibt es nicht", name)

	return "", nil
}

// **Setzen bringt die Marke in Normalform und gibt sie zurück.**
func TestDasMarkenwerkzeugSetztInNormalform(t *testing.T) {
	t.Parallel()

	probe := neueMarkenprobe()

	antwort, fehler := rufeMarke(t, probe, "schreiben.marke",
		`{"setID":7,"karteID":7,"marke":"  Fertig "}`)
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if !strings.Contains(antwort, `"fertig"`) || probe.gesetzt != "fertig" {
		t.Errorf("Antwort %s, gesetzt %q", antwort, probe.gesetzt)
	}
}

// **Lesen gibt die Marken einer Karte und die eines Sets.**
func TestDieMarkenwerkzeugeLesenBeideListen(t *testing.T) {
	t.Parallel()

	probe := neueMarkenprobe()
	probe.marken[7] = []string{"fertig"}

	anKarte, fehler := rufeMarke(t, probe, "lesen.kartenmarken",
		`{"setID":7,"karteID":7}`)
	if fehler != nil || !strings.Contains(anKarte, `"fertig"`) {
		t.Fatalf("lesen.kartenmarken: %s (%v)", anKarte, fehler)
	}

	imSet, fehler := rufeMarke(t, probe, "lesen.marken", `{"setID":7}`)
	if fehler != nil || !strings.Contains(imSet, `"fertig"`) {
		t.Fatalf("lesen.marken: %s (%v)", imSet, fehler)
	}
}

// **Eine Karte ohne Marken gibt eine leere Liste und kein `null`.**
func TestOhneMarkenKommtEineLeereListe(t *testing.T) {
	t.Parallel()

	antwort, fehler := rufeMarke(t, neueMarkenprobe(), "lesen.kartenmarken",
		`{"setID":7,"karteID":9}`)
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if !strings.Contains(antwort, `"marken": []`) {
		t.Errorf("Antwort: %s", antwort)
	}
}

// **Entfernen nimmt sie wieder ab.**
func TestDasMarkenwerkzeugEntferntWieder(t *testing.T) {
	t.Parallel()

	probe := neueMarkenprobe()
	probe.marken[7] = []string{"fertig"}

	if _, fehler := rufeMarke(t, probe, "schreiben.markeentfernen",
		`{"setID":7,"karteID":7,"marke":"FERTIG"}`); fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if len(probe.marken[7]) != 0 {
		t.Errorf("übrig: %v", probe.marken[7])
	}
}

// **Ohne Naht gibt es die Werkzeuge nicht.**
//
// Ein Werkzeug, das dasteht und bei jedem Ruf scheitert, ist schlimmer als
// keines — ein Modell versucht es dreimal (#591).
func TestOhneMarkennahtGibtEsKeineWerkzeuge(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: probenbestand(), Schreiber: probenbestand(),
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Angaben:     neueAngabenprobe(), Stile: neueStilprobe(),
		Bestand: neueBestandsprobeWeit(),
	}) {
		if strings.Contains(werkzeug.Name, "marke") {
			t.Errorf("ohne Naht gibt es %q", werkzeug.Name)
		}
	}
}

// **Jede Kennung wird geprüft, und ein Fehler kommt durch.**
func TestDieMarkenwerkzeugePruefenIhreEingaben(t *testing.T) {
	t.Parallel()

	probe := neueMarkenprobe()

	for name, fall := range map[string]struct {
		werkzeug, argumente string
	}{
		"Set fehlt":            {"lesen.marken", `{"setID":0}`},
		"Karte fehlt":          {"lesen.kartenmarken", `{"setID":7,"karteID":0}`},
		"Set beim Setzen":      {"schreiben.marke", `{"setID":0,"karteID":7,"marke":"a"}`},
		"unlesbare Argumente":  {"schreiben.marke", `{"setID":"sieben"}`},
		"unlesbar beim Lesen":  {"lesen.marken", `{"setID":"sieben"}`},
		"unlesbar an derKarte": {"lesen.kartenmarken", `{"setID":"sieben"}`},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := rufeMarke(t, probe, fall.werkzeug, fall.argumente); fehler == nil {
				t.Error("die Eingabe kam durch")
			}
		})
	}
}

// **Und ein Fehler des Dienstes reist mit.**
func TestEinMarkenfehlerKommtDurch(t *testing.T) {
	t.Parallel()

	probe := neueMarkenprobe()
	probe.fehler = errors.New("planmäßig gescheitert")

	for _, fall := range []struct{ werkzeug, argumente string }{
		{"lesen.marken", `{"setID":7}`},
		{"lesen.kartenmarken", `{"setID":7,"karteID":7}`},
		{"schreiben.marke", `{"setID":7,"karteID":7,"marke":"a"}`},
		{"schreiben.markeentfernen", `{"setID":7,"karteID":7,"marke":"a"}`},
	} {
		t.Run(fall.werkzeug, func(t *testing.T) {
			t.Parallel()

			if _, fehler := rufeMarke(t, probe, fall.werkzeug, fall.argumente); fehler == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Eine unlesbare Marke im Rumpf wird abgewiesen.**
func TestEineUnlesbareMarkeWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	// **Beide Schreibwerkzeuge**: sie lesen denselben Rumpf, und ein Zweig,
	// den nur eines von beiden durchläuft, ist geprüft und der andere nicht.
	for _, werkzeug := range []string{"schreiben.marke", "schreiben.markeentfernen"} {
		t.Run(werkzeug, func(t *testing.T) {
			t.Parallel()

			if _, fehler := rufeMarke(t, neueMarkenprobe(), werkzeug,
				`{"setID":7,"karteID":7,"marke":42}`); fehler == nil {
				t.Error("die unlesbare Marke kam durch")
			}
		})
	}
}
