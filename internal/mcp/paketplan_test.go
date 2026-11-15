// **Plan, Ersetzen und Zurück über MCP** — DOM-01.AK-02 und AK-03
// (#1696).
package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

func (p *pflegeprobe) PaketPlan(_ context.Context, vorgang, art, name, quelle string) (mcp.Paketplan, error) {
	p.geplant = vorgang + " " + art + "/" + name + " <- " + quelle

	if p.fehler != nil {
		return mcp.Paketplan{}, p.fehler
	}

	return mcp.Paketplan{
		Vorgang: vorgang, Paket: name + "." + art, Von: "2026-01-01", Moeglich: true,
		Sets: []string{"Erstes"},
		Konflikte: []mcp.Paketkonflikt{{
			Paket: "probe-schlicht.mse-style", Braucht: name + "." + art, Art: "fehlt",
		}},
	}, nil
}

func (p *pflegeprobe) PaketErsetzen(_ context.Context, quelle string, bestaetigt bool) (mcp.Paketzeile, error) {
	p.ersetzt = quelle + " " + strconv.FormatBool(bestaetigt)

	if p.fehler != nil {
		return mcp.Paketzeile{}, p.fehler
	}

	return mcp.Paketzeile{Art: "mse-style", Name: "fremd", Version: "2026-02-02"}, nil
}

func (p *pflegeprobe) PaketBedarf(_ context.Context, setID int64) ([]mcp.Paketkonflikt, error) {
	p.geplant = "bedarf " + strconv.FormatInt(setID, 10)

	if p.fehler != nil {
		return nil, p.fehler
	}

	return []mcp.Paketkonflikt{{Paket: "Erstes", Braucht: "probe.mse-game", Art: "fehlt"}}, nil
}

func (p *pflegeprobe) PaketZurueck(_ context.Context, art, name string) (mcp.Paketzeile, error) {
	p.zurueck = art + "/" + name

	if p.fehler != nil {
		return mcp.Paketzeile{}, p.fehler
	}

	return mcp.Paketzeile{Art: art, Name: name, Version: "2026-01-01"}, nil
}

func TestDerPlanNenntSetsUndKonflikte(t *testing.T) {
	probe := &pflegeprobe{}
	werkzeug := mitPflege(t, probe, "lesen.paket.plan")

	if werkzeug.Art != mcp.Lesend || !werkzeug.Idempotent {
		t.Errorf("der Plan ändert nichts: %+v", werkzeug)
	}

	antwort, err := werkzeug.Ruf(t.Context(), json.RawMessage(
		`{"vorgang":"entfernen","art":"mse-game","name":"probe"}`))
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	var gelesen struct {
		Plan mcp.Paketplan `json:"plan"`
	}

	if err := json.Unmarshal([]byte(antwort), &gelesen); err != nil {
		t.Fatalf("Antwort: %v (%s)", err, antwort)
	}

	if probe.geplant != "entfernen mse-game/probe <- " || gelesen.Plan.Paket != "probe.mse-game" ||
		len(gelesen.Plan.Sets) != 1 || len(gelesen.Plan.Konflikte) != 1 ||
		gelesen.Plan.Konflikte[0].Art != "fehlt" {
		t.Errorf("der Plan: %q, %+v", probe.geplant, gelesen.Plan)
	}
}

// **Der Bedarf eines Sets** — DOM-01.AK-04: lesend, und mit einer leeren
// Liste statt `null`, wenn nichts fehlt.
func TestDerBedarfEinesSetsGehtDurchDieNaht(t *testing.T) {
	probe := &pflegeprobe{}
	werkzeug := mitPflege(t, probe, "lesen.set.paketbedarf")

	if werkzeug.Art != mcp.Lesend || !werkzeug.Idempotent {
		t.Errorf("der Bedarf ändert nichts: %+v", werkzeug)
	}

	antwort, err := werkzeug.Ruf(t.Context(), json.RawMessage(`{"setID": 7}`))
	if err != nil || probe.geplant != "bedarf 7" || !strings.Contains(antwort, `"braucht": "probe.mse-game"`) {
		t.Errorf("der Bedarf: %q, %s, %v", probe.geplant, antwort, err)
	}
}

func TestErsetzenUndZurueckGehenDurchDieNaht(t *testing.T) {
	probe := &pflegeprobe{}

	antwort, err := mitPflege(t, probe, "schreiben.paket.ersetzen").Ruf(t.Context(),
		json.RawMessage(`{"quelle":"/irgendwo/fremd.mse-style","bestaetigt":true}`))
	if err != nil || probe.ersetzt != "/irgendwo/fremd.mse-style true" ||
		!strings.Contains(antwort, `"version": "2026-02-02"`) {
		t.Errorf("ersetzen: %q, %s, %v", probe.ersetzt, antwort, err)
	}

	// **Ohne Angabe ist nichts bestätigt.**
	if _, ohne := mitPflege(t, probe, "schreiben.paket.ersetzen").Ruf(t.Context(),
		json.RawMessage(`{"quelle":"/irgendwo/fremd.mse-style"}`)); ohne != nil ||
		probe.ersetzt != "/irgendwo/fremd.mse-style false" {
		t.Errorf("ohne Bestätigung: %q, %v", probe.ersetzt, ohne)
	}

	antwort, err = mitPflege(t, probe, "schreiben.paket.zurueck").Ruf(t.Context(),
		json.RawMessage(`{"art":"mse-style","name":"fremd"}`))
	if err != nil || probe.zurueck != "mse-style/fremd" || !strings.Contains(antwort, `"version": "2026-01-01"`) {
		t.Errorf("zurück: %q, %s, %v", probe.zurueck, antwort, err)
	}
}

func TestPlanErsetzenUndZurueckMeldenIhreFehler(t *testing.T) {
	klemmt := errors.New("klemmt")

	for name, roh := range map[string]string{
		"lesen.paket.plan":         `{"vorgang":"zurueck","art":"mse-game","name":"probe"}`,
		"schreiben.paket.ersetzen": `{"quelle":"/q"}`,
		"schreiben.paket.zurueck":  `{"art":"mse-game","name":"probe"}`,
		"lesen.set.paketbedarf":    `{"setID":7}`,
	} {
		werkzeug := mitPflege(t, &pflegeprobe{fehler: klemmt}, name)

		if _, err := werkzeug.Ruf(t.Context(), json.RawMessage(roh)); !errors.Is(err, klemmt) {
			t.Errorf("%s: %v", name, err)
		}

		if _, err := werkzeug.Ruf(t.Context(), json.RawMessage(`{`)); err == nil {
			t.Errorf("%s: kaputte Angaben gingen durch", name)
		}
	}
}
