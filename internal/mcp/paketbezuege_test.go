package mcp_test

// **Die Stände der bezogenen Bäume über MCP** — G04
// (#2251).

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// bezugspflegeprobe ist eine Pflege, die auch die Bezüge kennt.
type bezugspflegeprobe struct {
	*pflegeprobe

	gefragt    bool
	bestaetigt bool
	fehler     error
}

func (p *bezugspflegeprobe) PaketBezuege(_ context.Context, fragen bool) ([]mcp.Bezugsstand, error) {
	p.gefragt = fragen

	return []mcp.Bezugsstand{{Name: "M15-Magic-Pack", Lage: "neuer", Spitze: "b"}}, p.fehler
}

func (p *bezugspflegeprobe) PaketBezugVorschau(_ context.Context, name, stand string) (mcp.Bezugsvorschau, error) {
	return mcp.Bezugsvorschau{Name: name, Zu: stand, Gruende: []string{"gekuerzt"}}, p.fehler
}

func (p *bezugspflegeprobe) PaketBezugAktualisieren(_ context.Context, name, stand string, bestaetigt bool) (mcp.Paketbezug, error) {
	p.bestaetigt = bestaetigt

	return mcp.Paketbezug{Ort: "/pakete/" + name, Stand: stand}, p.fehler
}

func (p *bezugspflegeprobe) PaketBezugZurueck(_ context.Context, name string) (mcp.Paketbezug, error) {
	return mcp.Paketbezug{Ort: "/pakete/" + name, Stand: "a"}, p.fehler
}

func bezugswerkzeug(t *testing.T, probe *bezugspflegeprobe, name string) mcp.Werkzeug {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Ausfuhr: probe.pflegeprobe, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{}, Angaben: neueAngabenprobe(),
		Stile: neueStilprobe(), Bestand: neueBestandsprobeWeit(),
		Pflege: probe, Kartenbilder: probe.pflegeprobe, Vorlagen: probe.pflegeprobe,
	}) {
		if werkzeug.Name == name {
			return werkzeug
		}
	}

	t.Fatalf("kein Werkzeug %q", name)

	return mcp.Werkzeug{}
}

// dicht nimmt den Leerraum aus einer Antwort — sie kommt eingerückt.
func dicht(antwort string) string { return strings.Join(strings.Fields(antwort), "") }

func TestDieBezugswerkzeugeTunWasSieSagen(t *testing.T) {
	probe := &bezugspflegeprobe{pflegeprobe: &pflegeprobe{}}

	if antwort := rufeRoh(t, bezugswerkzeug(t, probe, "lesen.paket.bezuege"), `{"fragen":true}`); !probe.gefragt ||
		!strings.Contains(dicht(antwort), `"lage":"neuer"`) {
		t.Errorf("Bezüge: %s, gefragt %v", antwort, probe.gefragt)
	}

	if antwort := rufeRoh(t, bezugswerkzeug(t, probe, "lesen.paket.bezug.vorschau"),
		`{"name":"M15-Magic-Pack","stand":"b"}`); !strings.Contains(dicht(antwort), `"gruende":["gekuerzt"]`) {
		t.Errorf("Vorschau: %s", antwort)
	}

	if antwort := rufeRoh(t, bezugswerkzeug(t, probe, "schreiben.paket.bezug.aktualisieren"),
		`{"name":"M15-Magic-Pack","stand":"b","bestaetigt":true}`); !probe.bestaetigt || !strings.Contains(dicht(antwort), `"stand":"b"`) {
		t.Errorf("Aktualisieren: %s", antwort)
	}

	if antwort := rufeRoh(t, bezugswerkzeug(t, probe, "schreiben.paket.bezug.zurueck"),
		`{"name":"M15-Magic-Pack"}`); !strings.Contains(dicht(antwort), `"stand":"a"`) {
		t.Errorf("Zurück: %s", antwort)
	}
}

func TestDieBezugswerkzeugeMeldenFehlerUndUnlesbares(t *testing.T) {
	probe := &bezugspflegeprobe{pflegeprobe: &pflegeprobe{}, fehler: errors.New("weg")}

	for _, name := range []string{
		"lesen.paket.bezuege", "lesen.paket.bezug.vorschau",
		"schreiben.paket.bezug.aktualisieren", "schreiben.paket.bezug.zurueck",
	} {
		werkzeug := bezugswerkzeug(t, probe, name)

		if _, err := werkzeug.Ruf(t.Context(), json.RawMessage(`{"name":"x","stand":"b"}`)); err == nil {
			t.Errorf("%s: kein Fehler", name)
		}

		if _, err := werkzeug.Ruf(t.Context(), json.RawMessage(`[`)); err == nil {
			t.Errorf("%s: unlesbare Argumente ohne Fehler", name)
		}
	}
}

// **Eine Pflege ohne Bezüge bekommt die vier Werkzeuge nicht** — sie blieben
// sonst Werkzeuge, die bei jedem Ruf scheitern.
func TestOhneBezugspflegeKeineBezugswerkzeuge(t *testing.T) {
	probe := &pflegeprobe{}

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Ausfuhr: probe, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{}, Angaben: neueAngabenprobe(),
		Stile: neueStilprobe(), Bestand: neueBestandsprobeWeit(),
		Pflege: probe, Kartenbilder: probe, Vorlagen: probe,
	}) {
		if strings.Contains(werkzeug.Name, "bezug") && werkzeug.Name != "schreiben.paket.beziehen" {
			t.Errorf("ohne Bezugspflege steht %q da", werkzeug.Name)
		}
	}
}
