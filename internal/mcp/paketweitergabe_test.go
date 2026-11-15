// **Pakete als `.mse-installer` weitergeben und einlesen** — FMT-02.AK-04
// (#1691).
package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

func (p *pflegeprobe) PaketWeitergeben(_ context.Context, pakete []string, ziel string) (string, error) {
	p.weitergegeben = strings.Join(pakete, ",") + " → " + ziel

	return ziel, p.fehler
}

func (p *pflegeprobe) PaketEinlesen(_ context.Context, datei string) (mcp.Paketuebernahme, error) {
	p.eingelesen = datei

	// **Auch im Fehlerfall mit Teilerfolg** — wie bei der Übernahme.
	return mcp.Paketuebernahme{
		Uebernommen: []mcp.Paketzeile{{Art: "mse-include", Name: "mitgebracht"}},
		Uebergangen: []string{}, Befunde: []string{},
	}, p.fehler
}

func TestPaketeGehenAlsInstallerHinausUndHerein(t *testing.T) {
	probe := &pflegeprobe{}

	antwort := rufeRoh(t, mitPflege(t, probe, "schreiben.paket.weitergeben"),
		`{"pakete":["probe.mse-game","probe-schlicht.mse-style"],"ziel":"/Users/ich/probe.mse-installer"}`)
	if probe.weitergegeben != "probe.mse-game,probe-schlicht.mse-style → /Users/ich/probe.mse-installer" ||
		!strings.Contains(antwort, `"datei": "/Users/ich/probe.mse-installer"`) {
		t.Errorf("weitergegeben %q, Antwort %s", probe.weitergegeben, antwort)
	}

	antwort = rufeRoh(t, mitPflege(t, probe, "schreiben.paket.einlesen"), `{"datei":"/Users/ich/probe.mse-installer"}`)
	if probe.eingelesen != "/Users/ich/probe.mse-installer" || !strings.Contains(antwort, `"name": "mitgebracht"`) {
		t.Errorf("eingelesen %q, Antwort %s", probe.eingelesen, antwort)
	}
}

// **Ein Teilerfolg nennt, was schon eingelesen ist** — wie die Übernahme:
// die Pakete liegen da, und das Gegenüber muss es wissen.
func TestDieFehlerDerWeitergabeKommenAn(t *testing.T) {
	klagend := &pflegeprobe{fehler: errors.New("der Bestand ist gesperrt")}

	_, err := mitPflege(t, klagend, "schreiben.paket.einlesen").Ruf(t.Context(), json.RawMessage(`{"datei":"/x.mse-installer"}`))
	if err == nil || !strings.Contains(err.Error(), "mitgebracht") {
		t.Errorf("Fehler %v — der Teilerfolg fehlt", err)
	}

	for name, fall := range map[string]struct{ werkzeug, argumente string }{
		"weitergeben: krummer Rumpf":  {"schreiben.paket.weitergeben", `{`},
		"weitergeben: die Naht klagt": {"schreiben.paket.weitergeben", `{"pakete":["a.mse-game"],"ziel":"/x"}`},
		"einlesen: krummer Rumpf":     {"schreiben.paket.einlesen", `{`},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := mitPflege(t, klagend, fall.werkzeug).Ruf(t.Context(), json.RawMessage(fall.argumente)); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}
