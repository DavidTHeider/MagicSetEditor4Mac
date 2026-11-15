// **Die Sammlungssichten über MCP** — Werkstück 5 von Phase 4.5
// (#1292).
package mcp_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// sichtenprobe merkt sich, was gerufen wurde.
type sichtenprobe struct {
	printing string
	behalten int
	fehler   error
}

func (s *sichtenprobe) Besitz(_ context.Context, printing string) (any, error) {
	s.printing = printing

	if s.fehler != nil {
		return nil, s.fehler
	}

	return map[string]any{"stueck": 4, "gewuenscht": 2, "besitzt": true}, nil
}

func (s *sichtenprobe) Lagerorte(context.Context) (any, error) {
	if s.fehler != nil {
		return nil, s.fehler
	}

	return []any{map[string]any{"lagerort": "Kiste 4", "stueck": 4}}, nil
}

func (s *sichtenprobe) Dubletten(_ context.Context, behalten int) (any, error) {
	s.behalten = behalten

	if s.fehler != nil {
		return nil, s.fehler
	}

	return []any{}, nil
}

func (s *sichtenprobe) Wuensche(context.Context) (any, error) {
	if s.fehler != nil {
		return nil, s.fehler
	}

	return []any{}, nil
}

// mitSichten sucht ein Werkzeug in der Liste mit Sammlungssichten.
func mitSichten(t *testing.T, probe *sichtenprobe, name string) mcp.Werkzeug {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{}, Angaben: neueAngabenprobe(),
		Stile: neueStilprobe(), Bestand: neueBestandsprobeWeit(),
		Sammlungssichten: probe,
	}) {
		if werkzeug.Name == name {
			return werkzeug
		}
	}

	t.Fatalf("kein Werkzeug %q", name)

	return mcp.Werkzeug{}
}

// **Alle vier stehen in der Liste und tun, was sie sagen.**
func TestDieSammlungssichtenStehenUeberMCP(t *testing.T) {
	t.Run("besitz", func(t *testing.T) {
		probe := &sichtenprobe{}

		antwort := rufeRoh(t, mitSichten(t, probe, "lesen.sammlung.besitz"),
			`{"printing":"probe-druck"}`)

		if probe.printing != "probe-druck" {
			t.Errorf("gefragt wurde nach %q", probe.printing)
		}

		if !strings.Contains(antwort, "gewuenscht") {
			t.Errorf("die Antwort trennt Besitz und Wunsch nicht: %s", antwort)
		}
	})

	t.Run("lagerorte", func(t *testing.T) {
		antwort := rufeRoh(t, mitSichten(t, &sichtenprobe{},
			"lesen.sammlung.lagerorte"), `{}`)

		if !strings.Contains(antwort, "Kiste 4") {
			t.Errorf("die Antwort nennt keinen Lagerort: %s", antwort)
		}
	})

	t.Run("dubletten", func(t *testing.T) {
		probe := &sichtenprobe{}

		rufeRoh(t, mitSichten(t, probe, "lesen.sammlung.dubletten"),
			`{"behalten":4}`)

		// **Die Zahl kommt vom Rufenden** — eine Dublette ist kein Zustand,
		// sondern ein Verhältnis (Vermerk 0057).
		if probe.behalten != 4 {
			t.Errorf("behalten kam als %d an", probe.behalten)
		}
	})

	t.Run("wuensche", func(t *testing.T) {
		if antwort := rufeRoh(t, mitSichten(t, &sichtenprobe{},
			"lesen.sammlung.wuensche"), `{}`); antwort == "" {
			t.Error("die Wunschliste kam leer zurück")
		}
	})
}

// **Ein Fehler geht auf jedem der vier Wege als Fehler hinaus.**
func TestDieSammlungssichtenReichenFehlerDurch(t *testing.T) {
	gestolpert := errors.New("die Platte klemmt")

	for _, fall := range []struct {
		name      string
		argumente string
	}{
		{name: "lesen.sammlung.besitz", argumente: `{"printing":"abc"}`},
		{name: "lesen.sammlung.lagerorte", argumente: `{}`},
		{name: "lesen.sammlung.dubletten", argumente: `{"behalten":1}`},
		{name: "lesen.sammlung.wuensche", argumente: `{}`},
	} {
		t.Run(fall.name, func(t *testing.T) {
			werkzeug := mitSichten(t, &sichtenprobe{fehler: gestolpert}, fall.name)

			if _, err := werkzeug.Ruf(t.Context(),
				[]byte(fall.argumente)); !errors.Is(err, gestolpert) {
				t.Errorf("Fehler %v, erwartet %v", err, gestolpert)
			}
		})
	}
}

// **Unbrauchbare Argumente sind ein Fehler und kein Absturz.**
func TestDieSammlungssichtenMitUnbrauchbarenArgumenten(t *testing.T) {
	for _, name := range []string{
		"lesen.sammlung.besitz", "lesen.sammlung.dubletten",
	} {
		werkzeug := mitSichten(t, &sichtenprobe{}, name)

		if _, err := werkzeug.Ruf(t.Context(), []byte(`{`)); err == nil {
			t.Errorf("%s: kein Fehler bei kaputtem JSON", name)
		}
	}
}

// **Ohne Naht gibt es die vier Werkzeuge nicht.**
func TestOhneSichtennahtKeineWerkzeuge(t *testing.T) {
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{}, Angaben: neueAngabenprobe(),
		Stile: neueStilprobe(), Bestand: neueBestandsprobeWeit(),
	}) {
		for _, verboten := range []string{
			"lesen.sammlung.besitz", "lesen.sammlung.lagerorte",
			"lesen.sammlung.dubletten", "lesen.sammlung.wuensche",
		} {
			if werkzeug.Name == verboten {
				t.Errorf("ohne Naht steht %q in der Liste", werkzeug.Name)
			}
		}
	}
}
