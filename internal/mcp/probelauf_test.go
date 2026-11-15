package mcp_test

// **Der Probelauf über MCP** — UXR-F7
// (#2015).
//
// Zwei **lesende** Werkzeuge: sie sagen, was der Druckbogen und die
// Exportvorlage ergäben, und schreiben nichts. Einordnung nach §16:
// **erweiternd** — zwei Werkzeuge kommen hinzu, kein vorhandenes ändert seine
// Bedeutung.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// probelaufprobe merkt sich, was gefragt wurde.
type probelaufprobe struct {
	bogen    mcp.Bogenprobelauf
	vorlage  mcp.Vorlagenprobelauf
	fehler   error
	gefragt  []string
	wahl     map[string]string
	ziel     string
	kennung  int64
	paketCmd string
}

func (p *probelaufprobe) BogenProbelauf(
	_ context.Context, setID int64, ziel string,
) (mcp.Bogenprobelauf, error) {
	p.gefragt = append(p.gefragt, "bogen")
	p.kennung, p.ziel = setID, ziel

	return p.bogen, p.fehler
}

func (p *probelaufprobe) VorlageProbelauf(
	_ context.Context, setID int64, paket, ziel string, wahl map[string]string,
) (mcp.Vorlagenprobelauf, error) {
	p.gefragt = append(p.gefragt, "vorlage")
	p.kennung, p.ziel, p.paketCmd, p.wahl = setID, ziel, paket, wahl

	return p.vorlage, p.fehler
}

// mitProbelauf sucht ein Werkzeug in der Liste mit einem Probelauf.
func mitProbelauf(t *testing.T, probe *probelaufprobe, name string) mcp.Werkzeug {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{}, Ausfuhr: &ausfuhrprobe{},
		Auskunft: stilleAuskunft{}, Sammeldaten: stilleSammeldaten{},
		Angaben: neueAngabenprobe(), Stile: neueStilprobe(), Bestand: neueBestandsprobeWeit(),
		Probelauf: probe,
	}) {
		if werkzeug.Name == name {
			return werkzeug
		}
	}

	t.Fatalf("kein Werkzeug %q", name)

	return mcp.Werkzeug{}
}

// **Beide Werkzeuge lesen — und werden als lesend angeboten.**
func TestDieProbelaufwerkzeugeSindLesend(t *testing.T) {
	for _, name := range []string{"lesen.set.bogenprobelauf", "lesen.vorlage.probelauf"} {
		werkzeug := mitProbelauf(t, &probelaufprobe{}, name)

		if werkzeug.Art != mcp.Lesend || !werkzeug.Idempotent {
			t.Errorf("%s: Art %v, idempotent %v", name, werkzeug.Art, werkzeug.Idempotent)
		}

		if !strings.Contains(werkzeug.Erklaerung, "NICHTS") {
			t.Errorf("%s sagt nicht, dass es nichts schreibt: %s", name, werkzeug.Erklaerung)
		}
	}
}

// **Der Bogen: die Zahlen kommen an, das Ziel ist freiwillig.**
func TestDerBogenprobelaufNenntDieZahlen(t *testing.T) {
	probe := &probelaufprobe{bogen: mcp.Bogenprobelauf{
		Karten: 10, Seiten: 2, JeSeite: 9, Spalten: 3, Zeilen: 3, AufLetzterSeite: 1,
		Seitenbreite: 210, Seitenhoehe: 297, DPI: 300,
	}}

	antwort := rufeRoh(t, mitProbelauf(t, probe, "lesen.set.bogenprobelauf"), `{"setID":7}`)

	var gelesen map[string]any
	if err := json.Unmarshal([]byte(antwort), &gelesen); err != nil {
		t.Fatalf("Antwort lesen: %v (%s)", err, antwort)
	}

	if gelesen["probelauf"] != true || gelesen["karten"] != float64(10) ||
		gelesen["seiten"] != float64(2) || gelesen["jeSeite"] != float64(9) {
		t.Errorf("Antwort: %s", antwort)
	}

	if probe.kennung != 7 || probe.ziel != "" {
		t.Errorf("gefragt wurde nach %d und %q", probe.kennung, probe.ziel)
	}

	rufeRoh(t, mitProbelauf(t, probe, "lesen.set.bogenprobelauf"), `{"setID":7,"ziel":"/tmp/b.pdf"}`)

	if probe.ziel != "/tmp/b.pdf" {
		t.Errorf("das Ziel kam als %q an", probe.ziel)
	}
}

// **Die Vorlage: Größe, Namen und die Wahl für diesen Lauf.**
func TestDerVorlagenprobelaufNenntDieGroesse(t *testing.T) {
	probe := &probelaufprobe{vorlage: mcp.Vorlagenprobelauf{
		Karten: 41, Dateiname: "ausfuhr.xml", Bytes: 1200, Beigaben: 3,
		BeigabenBytes: 90, Verzeichnis: "ausfuhr-files", Geschaetzt: true,
		Befunde: []string{"eine Meldung"},
	}}

	antwort := rufeRoh(t, mitProbelauf(t, probe, "lesen.vorlage.probelauf"),
		`{"setID":7,"paket":"p.mse-export-template","wahl":{"mit namen":"no"}}`)

	var gelesen map[string]any
	if err := json.Unmarshal([]byte(antwort), &gelesen); err != nil {
		t.Fatalf("Antwort lesen: %v (%s)", err, antwort)
	}

	if gelesen["probelauf"] != true || gelesen["bytes"] != float64(1200) ||
		gelesen["dateiname"] != "ausfuhr.xml" || gelesen["geschaetzt"] != true {
		t.Errorf("Antwort: %s", antwort)
	}

	if probe.paketCmd != "p.mse-export-template" || probe.wahl["mit namen"] != "no" ||
		probe.kennung != 7 {
		t.Errorf("gefragt wurde nach %q, %v, %d", probe.paketCmd, probe.wahl, probe.kennung)
	}
}

// **Eine Antwort ohne Befunde trägt eine leere Liste, kein `null`** (§3).
func TestDerVorlagenprobelaufTraegtNieNullBefunde(t *testing.T) {
	antwort := rufeRoh(t, mitProbelauf(t, &probelaufprobe{}, "lesen.vorlage.probelauf"),
		`{"setID":7,"paket":"p"}`)

	if !strings.Contains(antwort, `"befunde": []`) && !strings.Contains(antwort, `"befunde":[]`) {
		t.Errorf("Antwort: %s", antwort)
	}
}

// **Fehler, unlesbare Angaben und falsche Kennungen gehen als Fehler hinaus.**
func TestDieProbelaufwerkzeugeReichenFehlerDurch(t *testing.T) {
	gestolpert := errors.New("die Platte klemmt")

	faelle := map[string]struct {
		werkzeug  string
		argumente string
		fehler    error
	}{
		"Bogen, Fehler":         {"lesen.set.bogenprobelauf", `{"setID":7}`, gestolpert},
		"Bogen, kein JSON":      {"lesen.set.bogenprobelauf", `kein JSON`, nil},
		"Bogen, Kennung null":   {"lesen.set.bogenprobelauf", `{"setID":0}`, nil},
		"Vorlage, Fehler":       {"lesen.vorlage.probelauf", `{"setID":7,"paket":"p"}`, gestolpert},
		"Vorlage, kein JSON":    {"lesen.vorlage.probelauf", `kein JSON`, nil},
		"Vorlage, Kennung null": {"lesen.vorlage.probelauf", `{"setID":0,"paket":"p"}`, nil},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			if !rufeMitFehler(t, mitProbelauf(t, &probelaufprobe{fehler: fall.fehler}, fall.werkzeug),
				fall.argumente) {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Ohne Naht gibt es die Werkzeuge nicht** — wie bei jeder anderen.
func TestOhneProbelaufGibtEsDieWerkzeugeNicht(t *testing.T) {
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{}, Ausfuhr: &ausfuhrprobe{},
		Auskunft: stilleAuskunft{}, Sammeldaten: stilleSammeldaten{},
		Angaben: neueAngabenprobe(), Stile: neueStilprobe(), Bestand: neueBestandsprobeWeit(),
	}) {
		if strings.Contains(werkzeug.Name, "probelauf") {
			t.Errorf("%s steht da, obwohl niemand einen Probelauf eingesetzt hat", werkzeug.Name)
		}
	}
}
