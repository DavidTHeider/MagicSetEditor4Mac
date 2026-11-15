package mcp_test

// **Die eigenen Übersetzungen über MCP** — LOC-02
// (#1688).

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

type sprachprobe struct {
	ergebnis any
	fehler   error

	gerufen              string
	kennung, name, basis string
	inhalt               string
	ersetzen             bool
	angaben, texte       map[string]string
}

func (p *sprachprobe) antwort(name string) (any, error) {
	p.gerufen = name

	return p.ergebnis, p.fehler
}

func (p *sprachprobe) Sprachpakete(context.Context) (any, error) { return p.antwort("Sprachpakete") }

func (p *sprachprobe) Sprachpaket(_ context.Context, kennung string) (any, error) {
	p.kennung = kennung

	return p.antwort("Sprachpaket")
}

func (p *sprachprobe) SprachpaketAnlegen(_ context.Context, kennung, name, basis string) (any, error) {
	p.kennung, p.name, p.basis = kennung, name, basis

	return p.antwort("Anlegen")
}

func (p *sprachprobe) SprachpaketAendern(
	_ context.Context, kennung string, angaben, texte map[string]string,
) (any, error) {
	p.kennung, p.angaben, p.texte = kennung, angaben, texte

	return p.antwort("Aendern")
}

func (p *sprachprobe) SprachpaketLoeschen(_ context.Context, kennung string) (any, error) {
	p.kennung = kennung

	return p.antwort("Loeschen")
}

func (p *sprachprobe) SprachpaketAusgabe(_ context.Context, kennung string) (any, error) {
	p.kennung = kennung

	return p.antwort("Ausgabe")
}

func (p *sprachprobe) SprachpaketEinlesen(_ context.Context, inhalt string, ersetzen bool) (any, error) {
	p.inhalt, p.ersetzen = inhalt, ersetzen

	return p.antwort("Einlesen")
}

func rufeSprachwerkzeug(t *testing.T, probe *sprachprobe, name, argumente string) (string, error) {
	t.Helper()

	return werkzeugAus(t, mcp.Naehte{Sprachpakete: probe}, name).Ruf(t.Context(), json.RawMessage(argumente))
}

// alleSprachwerkzeuge sind die sieben — samt dem, was ankommen muss.
var alleSprachwerkzeuge = []struct {
	werkzeug, argumente string
	angekommen          sprachprobe
}{
	{"lesen.sprachpakete", `{}`, sprachprobe{gerufen: "Sprachpakete"}},
	{"lesen.sprachpaket", `{"kennung":"da"}`, sprachprobe{gerufen: "Sprachpaket", kennung: "da"}},
	{"lesen.sprachpaket.datei", `{"kennung":"da"}`, sprachprobe{gerufen: "Ausgabe", kennung: "da"}},
	{
		"schreiben.sprachpaket.anlegen", `{"kennung":"da","name":"Dansk","basis":"en"}`,
		sprachprobe{gerufen: "Anlegen", kennung: "da", name: "Dansk", basis: "en"},
	},
	{
		"schreiben.sprachpaket.aendern",
		`{"kennung":"da","stand":"geprueft","texte":{"allgemein.titel":"Sæt"}}`,
		sprachprobe{
			gerufen: "Aendern", kennung: "da", angaben: map[string]string{"stand": "geprueft"},
			texte: map[string]string{"allgemein.titel": "Sæt"},
		},
	},
	{"schreiben.sprachpaket.loeschen", `{"kennung":"da"}`, sprachprobe{gerufen: "Loeschen", kennung: "da"}},
	{
		"schreiben.sprachpaket.einlesen", `{"inhalt":"{}","ersetzen":true}`,
		sprachprobe{gerufen: "Einlesen", inhalt: "{}", ersetzen: true},
	},
}

func TestJedesSprachwerkzeugReichtSeineAngabenDurch(t *testing.T) {
	t.Parallel()

	for _, fall := range alleSprachwerkzeuge {
		t.Run(fall.werkzeug, func(t *testing.T) {
			t.Parallel()

			probe := &sprachprobe{ergebnis: map[string]string{"kennung": "da"}}

			ausgabe, err := rufeSprachwerkzeug(t, probe, fall.werkzeug, fall.argumente)
			if err != nil || !strings.Contains(ausgabe, `"kennung": "da"`) {
				t.Fatalf("%q, %v", ausgabe, err)
			}

			erwartet := fall.angekommen
			if probe.gerufen != erwartet.gerufen || probe.kennung != erwartet.kennung ||
				probe.name != erwartet.name || probe.basis != erwartet.basis ||
				probe.inhalt != erwartet.inhalt || probe.ersetzen != erwartet.ersetzen ||
				!maps.Equal(probe.angaben, erwartet.angaben) || !maps.Equal(probe.texte, erwartet.texte) {
				t.Errorf("angekommen %+v, erwartet %+v", *probe, erwartet)
			}
		})
	}
}

// **Und jedes sagt nein**: einen Fehler der Naht reicht es weiter, ein
// unlesbares Argument weist es ab, ohne zu rufen.
func TestJedesSprachwerkzeugSagtNein(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Naht schweigt")

	for _, fall := range alleSprachwerkzeuge {
		t.Run(fall.werkzeug, func(t *testing.T) {
			t.Parallel()

			if _, err := rufeSprachwerkzeug(t, &sprachprobe{fehler: kaputt}, fall.werkzeug,
				fall.argumente); !errors.Is(err, kaputt) {
				t.Errorf("der Fehler ging verloren: %v", err)
			}

			if fall.argumente == `{}` {
				return
			}

			probe := &sprachprobe{}
			if _, err := rufeSprachwerkzeug(t, probe, fall.werkzeug, `kein JSON`); err == nil ||
				probe.gerufen != "" {
				t.Errorf("der Unsinn kam durch: %v, gerufen %q", err, probe.gerufen)
			}
		})
	}
}

// **Ohne Naht gibt es die Werkzeuge nicht.**
func TestOhneSprachnahtGibtEsKeineSprachwerkzeuge(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if strings.Contains(werkzeug.Name, "sprachpaket") {
			t.Errorf("ohne Naht angeboten: %s", werkzeug.Name)
		}
	}
}
