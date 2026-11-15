// **Die Sammlungsausfuhr über MCP** — Werkstück 4 von Phase 4.5
// (#1292).
package mcp_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// sammlungsausfuhrprobe merkt sich, was gerufen wurde.
type sammlungsausfuhrprobe struct {
	form      string
	gelesen   string
	fehler    error
	ausgeben  string
	angelegte []int64

	// zusammenfuehrung und vorgeschaut halten fest, was DOM-03.AK-01 (#1698)
	// durchreicht.
	zusammenfuehrung string
	vorgeschaut      string
}

func (a *sammlungsausfuhrprobe) CSVVorschauen(_ context.Context, text string) (
	mcp.Sammelvorschau, error,
) {
	a.vorgeschaut = text

	if a.fehler != nil {
		return mcp.Sammelvorschau{}, a.fehler
	}

	return mcp.Sammelvorschau{Posten: []mcp.Vorschauposten{{Zeile: 2, Art: "neu", Nachher: 1}}, Befunde: []string{}}, nil
}

func (a *sammlungsausfuhrprobe) Ausgeben(_ context.Context, form string) (string, error) {
	a.form = form

	if a.fehler != nil {
		return "", a.fehler
	}

	return a.ausgeben, nil
}

func (a *sammlungsausfuhrprobe) CSVEinlesen(_ context.Context, text, _, zusammenfuehrung string) (
	mcp.Einfuhrbericht, error,
) {
	a.gelesen, a.zusammenfuehrung = text, zusammenfuehrung

	if a.fehler != nil {
		return mcp.Einfuhrbericht{}, a.fehler
	}

	return mcp.Einfuhrbericht{
		Angelegt: a.angelegte, Stueck: len(a.angelegte), Befunde: []string{},
	}, nil
}

// mitSammlungsausfuhr sucht ein Werkzeug in der Liste mit Sammlungsausfuhr.
func mitSammlungsausfuhr(t *testing.T, probe *sammlungsausfuhrprobe, name string) mcp.Werkzeug {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{}, Angaben: neueAngabenprobe(),
		Stile: neueStilprobe(), Bestand: neueBestandsprobeWeit(),
		Sammlungsausfuhr: probe,
	}) {
		if werkzeug.Name == name {
			return werkzeug
		}
	}

	t.Fatalf("kein Werkzeug %q", name)

	return mcp.Werkzeug{}
}

// **Beide stehen in der Liste und tun, was sie sagen.**
func TestDieSammlungsausfuhrStehtUeberMCP(t *testing.T) {
	t.Run("ausgeben", func(t *testing.T) {
		probe := &sammlungsausfuhrprobe{ausgeben: "printing,menge\nabc,1\n"}

		antwort := rufeRoh(t, mitSammlungsausfuhr(t, probe, "lesen.sammlung.ausgeben"),
			`{"form":"csv"}`)

		if probe.form != "csv" {
			t.Errorf("gefragt wurde nach %q", probe.form)
		}

		// **Der Text geht roh hinaus und nicht als JSON.** Ein CSV in einen
		// JSON-String zu wickeln machte aus jedem Zeilenumbruch ein `\n`, das
		// der Rufende erst wieder auspacken müsste.
		if !strings.Contains(antwort, "printing,menge") {
			t.Errorf("die Antwort ist %q", antwort)
		}
	})

	t.Run("einlesen", func(t *testing.T) {
		probe := &sammlungsausfuhrprobe{angelegte: []int64{1, 2}}

		antwort := rufeRoh(t, mitSammlungsausfuhr(t, probe, "schreiben.sammlung.einlesen"),
			`{"text":"printing,menge\nabc,1\n"}`)

		if !strings.Contains(probe.gelesen, "printing,menge") {
			t.Errorf("eingelesen wurde %q", probe.gelesen)
		}

		if !strings.Contains(antwort, "stueck") {
			t.Errorf("die Antwort nennt die Stückzahl nicht: %s", antwort)
		}
	})
}

// **Ein Fehler geht als Fehler hinaus.**
func TestDieSammlungsausfuhrReichtFehlerDurch(t *testing.T) {
	gestolpert := errors.New("die Platte klemmt")

	for _, fall := range []struct {
		name      string
		werkzeug  string
		argumente string
	}{
		{name: "ausgeben", werkzeug: "lesen.sammlung.ausgeben", argumente: `{"form":"csv"}`},
		{
			name: "einlesen", werkzeug: "schreiben.sammlung.einlesen",
			argumente: `{"text":"x"}`,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			werkzeug := mitSammlungsausfuhr(t, &sammlungsausfuhrprobe{fehler: gestolpert}, fall.werkzeug)

			if _, err := werkzeug.Ruf(t.Context(),
				[]byte(fall.argumente)); !errors.Is(err, gestolpert) {
				t.Errorf("Fehler %v, erwartet %v", err, gestolpert)
			}
		})
	}
}

// **Unbrauchbare Argumente sind ein Fehler und kein Absturz.**
func TestDieSammlungsausfuhrMitUnbrauchbarenArgumenten(t *testing.T) {
	for _, name := range []string{
		"lesen.sammlung.ausgeben", "schreiben.sammlung.einlesen",
	} {
		werkzeug := mitSammlungsausfuhr(t, &sammlungsausfuhrprobe{}, name)

		if _, err := werkzeug.Ruf(t.Context(), []byte(`{`)); err == nil {
			t.Errorf("%s: kein Fehler bei kaputtem JSON", name)
		}
	}
}

// **Ohne Naht gibt es die Werkzeuge nicht.**
func TestOhneAusfuhrnahtKeineWerkzeuge(t *testing.T) {
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{}, Angaben: neueAngabenprobe(),
		Stile: neueStilprobe(), Bestand: neueBestandsprobeWeit(),
	}) {
		if strings.Contains(werkzeug.Name, "sammlung.") {
			t.Errorf("ohne Naht steht %q in der Liste", werkzeug.Name)
		}
	}
}
