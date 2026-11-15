// **Der Druckbogen und MTG JSON über MCP** — Werkstück D von Phase 4.4
// (#1291).
//
// **Eine eigene Datei**, seit `ablauf_test.go` an der Zeilengrenze aus §2
// stand. Der Schnitt folgt der Frage: dort steht der Ablauf, den es seit #886
// gibt — einfahren, sichern, zeichnen; hier das, was die Restparität
// hinzugefügt hat.
package mcp_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Der Druckbogen steht über MCP** — Werkstück D von Phase 4.4
// (#1291).
//
// `docs/FAEHIGKEITEN.md` führte ihn mit drei Gedankenstrichen und einer
// erhobenen Hand: **nur am Fenster, weil es einen Dateidialog gibt**. Das ist
// kein Grund, sondern der Weg des Fensters zu einem Pfad — `App.BogenLegen`
// nimmt ihn längst entgegen.
func TestBogenStehtUeberMCP(t *testing.T) {
	probe := &ausfuhrprobe{}

	antwort := rufeRoh(t, mitAusfuhr(t, probe, "schreiben.set.bogen"),
		`{"setID":7,"ziel":"/tmp/bogen.pdf"}`)

	if probe.gelegt != "/tmp/bogen.pdf" {
		t.Errorf("gelegt wurde %q", probe.gelegt)
	}

	for _, erwartet := range []string{`"seiten": 1`, `"karten": 3`, `"pfad": "/tmp/bogen.pdf"`} {
		if !strings.Contains(antwort, erwartet) {
			t.Errorf("die Antwort nennt %s nicht: %s", erwartet, antwort)
		}
	}
}

// **Eine Kennung, die keine ist, wird vorher abgewiesen** — wie überall.
func TestBogenPrueftDieKennung(t *testing.T) {
	probe := &ausfuhrprobe{}

	_, err := mitAusfuhr(t, probe, "schreiben.set.bogen").
		Ruf(t.Context(), json.RawMessage(`{"setID":0,"ziel":"/tmp/b.pdf"}`))
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if probe.gelegt != "" {
		t.Error("es wurde trotzdem gelegt")
	}
}

// **Ein Rumpf, der kein JSON ist, wird abgewiesen.**
func TestBogenWeistEinenKrummenRumpfAb(t *testing.T) {
	probe := &ausfuhrprobe{}

	if _, err := mitAusfuhr(t, probe, "schreiben.set.bogen").
		Ruf(t.Context(), json.RawMessage(`{`)); err == nil {
		t.Fatal("kein Fehler")
	}

	if probe.gelegt != "" {
		t.Error("es wurde trotzdem gelegt")
	}
}

// **Und ein Fehler aus der Naht kommt an.**
func TestBogenReichtDenFehlerDurch(t *testing.T) {
	probe := &ausfuhrprobe{fehler: errors.New("kein Stil")}

	if _, err := mitAusfuhr(t, probe, "schreiben.set.bogen").
		Ruf(t.Context(), json.RawMessage(`{"setID":7,"ziel":"/tmp/b.pdf"}`)); err == nil {
		t.Fatal("kein Fehler")
	}
}

// **MTG JSON steht über MCP** — Werkstück D von Phase 4.4
// (#1291).
//
// Zwei Zeilen des Katalogs standen mit drei Gedankenstrichen und einer
// erhobenen Hand da. Die Hand meint einen Dateidialog — und `App.MTGJSONLesen`
// wie `App.MTGJSONSchreiben` nehmen längst einen Pfad entgegen.
func TestMTGJSONStehtUeberMCP(t *testing.T) {
	t.Run("lesen", func(t *testing.T) {
		probe := &ausfuhrprobe{}

		antwort := rufeRoh(t, mitAusfuhr(t, probe, "schreiben.mtgjson.lesen"),
			`{"pfad":"/tmp/alle.json","spiel":"magic"}`)

		if probe.gejsont != "/tmp/alle.json" {
			t.Errorf("gelesen wurde %q", probe.gejsont)
		}

		// **Alle angelegten Kennungen, nicht eine** (#1291, Befund 1).
		for _, erwartet := range []string{"7", "8"} {
			if !strings.Contains(antwort, erwartet) {
				t.Errorf("die Antwort nennt %s nicht: %s", erwartet, antwort)
			}
		}
	})

	t.Run("schreiben", func(t *testing.T) {
		probe := &ausfuhrprobe{}

		antwort := rufeRoh(t, mitAusfuhr(t, probe, "schreiben.mtgjson.schreiben"),
			`{"setID":7,"spiel":"magic","ziel":"/tmp/raus.json"}`)

		if probe.gejsont != "/tmp/raus.json" {
			t.Errorf("geschrieben wurde %q", probe.gejsont)
		}

		if !strings.Contains(antwort, `"karten": 3`) {
			t.Errorf("die Antwort nennt die Kartenzahl nicht: %s", antwort)
		}
	})

	t.Run("Spiele", func(t *testing.T) {
		antwort := rufeRoh(t, mitAusfuhr(t, &ausfuhrprobe{}, "lesen.mtgjson.spiele"), `{}`)

		if !strings.Contains(antwort, "magic") {
			t.Errorf("die Antwort nennt kein Spiel: %s", antwort)
		}
	})
}

// **Und beide Wege weisen ab, was sie nicht lesen können.**
func TestMTGJSONWeistAb(t *testing.T) {
	faelle := map[string]struct {
		werkzeug  string
		argumente string
		probe     *ausfuhrprobe
	}{
		"lesen: krummer Rumpf": {"schreiben.mtgjson.lesen", `{`, &ausfuhrprobe{}},
		"lesen: die Naht klagt": {
			"schreiben.mtgjson.lesen", `{"pfad":"/tmp/x.json","spiel":"magic"}`,
			&ausfuhrprobe{fehler: errors.New("unlesbar")},
		},
		"schreiben: krummer Rumpf": {"schreiben.mtgjson.schreiben", `{`, &ausfuhrprobe{}},
		"schreiben: Kennung null": {
			"schreiben.mtgjson.schreiben", `{"setID":0,"spiel":"magic","ziel":"/tmp/x.json"}`,
			&ausfuhrprobe{},
		},
		"schreiben: die Naht klagt": {
			"schreiben.mtgjson.schreiben",
			`{"setID":7,"spiel":"nichts","ziel":"/tmp/x.json"}`,
			&ausfuhrprobe{fehler: errors.New("kein Spiel")},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			if _, err := mitAusfuhr(t, fall.probe, fall.werkzeug).
				Ruf(t.Context(), json.RawMessage(fall.argumente)); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// **Ein Kartenbild von der Platte steht über MCP** (#1291).
func TestEinKartenbildVonDerPlatteStehtUeberMCP(t *testing.T) {
	probe := &pflegeprobe{}

	antwort := rufeRoh(t, mitPflege(t, probe, "schreiben.karte.bild"),
		`{"setID":1,"karteID":7,"feld":"kunst","pfad":"/tmp/quelle.png"}`)

	if probe.bildgesetzt != "kunst <- /tmp/quelle.png" {
		t.Errorf("gesetzt wurde %q", probe.bildgesetzt)
	}

	for _, erwartet := range []string{`"quelleBreite": 400`, `"zielBreite": 315`} {
		if !strings.Contains(antwort, erwartet) {
			t.Errorf("die Antwort nennt %s nicht: %s", erwartet, antwort)
		}
	}
}

// **Und es weist ab, was es nicht lesen kann.**
func TestDasKartenbildWeistAb(t *testing.T) {
	faelle := map[string]struct {
		argumente string
		probe     *pflegeprobe
	}{
		"krummer Rumpf": {`{`, &pflegeprobe{}},
		"Setkennung null": {
			`{"setID":0,"karteID":7,"feld":"kunst","pfad":"/tmp/a.png"}`, &pflegeprobe{},
		},
		"Kartenkennung null": {
			`{"setID":1,"karteID":0,"feld":"kunst","pfad":"/tmp/a.png"}`, &pflegeprobe{},
		},
		"die Naht klagt": {
			`{"setID":1,"karteID":7,"feld":"kunst","pfad":"/tmp/a.png"}`,
			&pflegeprobe{fehler: errors.New("kein Bild")},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			if _, err := mitPflege(t, fall.probe, "schreiben.karte.bild").
				Ruf(t.Context(), json.RawMessage(fall.argumente)); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// **Ohne Naht gibt es das Werkzeug nicht.**
func TestOhneBildnahtFehltDasWerkzeug(t *testing.T) {
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{}, Angaben: neueAngabenprobe(),
		Stile: neueStilprobe(), Bestand: neueBestandsprobeWeit(),
	}) {
		if werkzeug.Name == "schreiben.karte.bild" {
			t.Error("das Werkzeug steht da, obwohl keine Naht gesetzt ist")
		}
	}
}

// **Die Exportvorlagen stehen über MCP** (#1291).
func TestDieExportvorlagenStehenUeberMCP(t *testing.T) {
	t.Run("nennen", func(t *testing.T) {
		antwort := rufeRoh(t, mitPflege(t, &pflegeprobe{}, "lesen.vorlagen"), `{"setID":1}`)

		for _, erwartet := range []string{"probe-liste", `"trenner"`, `"dateityp"`} {
			if !strings.Contains(antwort, erwartet) {
				t.Errorf("die Antwort nennt %s nicht: %s", erwartet, antwort)
			}
		}
	})

	t.Run("ausführen", func(t *testing.T) {
		probe := &pflegeprobe{}

		rufeRoh(t, mitPflege(t, probe, "schreiben.vorlage.ausfuehren"),
			`{"setID":1,"paket":"probe-liste.mse-export-template","ziel":"/tmp/raus.txt"}`)

		if probe.vorlagelauf != "probe-liste.mse-export-template -> /tmp/raus.txt" {
			t.Errorf("ausgeführt wurde %q", probe.vorlagelauf)
		}
	})

	t.Run("Option setzen", func(t *testing.T) {
		probe := &pflegeprobe{}

		rufeRoh(t, mitPflege(t, probe, "schreiben.vorlagenoption"),
			`{"vorlage":"probe-liste.mse-export-template","feld":"trenner","wert":"; "}`)

		if probe.optiongesetzt != "probe-liste.mse-export-template.trenner=; " {
			t.Errorf("gesetzt wurde %q", probe.optiongesetzt)
		}
	})
}

// **Und alle drei weisen ab, was sie nicht lesen können.**
func TestDieExportvorlagenWeisenAb(t *testing.T) {
	faelle := map[string]struct {
		werkzeug  string
		argumente string
		probe     *pflegeprobe
	}{
		"nennen: krummer Rumpf": {"lesen.vorlagen", `{`, &pflegeprobe{}},
		"nennen: Kennung null":  {"lesen.vorlagen", `{"setID":0}`, &pflegeprobe{}},
		"nennen: die Naht klagt": {
			"lesen.vorlagen", `{"setID":1}`, &pflegeprobe{fehler: errors.New("kein Spiel")},
		},
		"ausführen: krummer Rumpf": {"schreiben.vorlage.ausfuehren", `{`, &pflegeprobe{}},
		"ausführen: Kennung null": {
			"schreiben.vorlage.ausfuehren",
			`{"setID":0,"paket":"p","ziel":"/tmp/a"}`, &pflegeprobe{},
		},
		"ausführen: die Naht klagt": {
			"schreiben.vorlage.ausfuehren",
			`{"setID":1,"paket":"p","ziel":"/tmp/a"}`,
			&pflegeprobe{fehler: errors.New("keine Vorlage")},
		},
		"Option: krummer Rumpf": {"schreiben.vorlagenoption", `{`, &pflegeprobe{}},
		"Option: die Naht klagt": {
			"schreiben.vorlagenoption", `{"vorlage":"p","feld":"f","wert":"w"}`,
			&pflegeprobe{fehler: errors.New("kein Feld")},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			if _, err := mitPflege(t, fall.probe, fall.werkzeug).
				Ruf(t.Context(), json.RawMessage(fall.argumente)); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// **Ohne Naht gibt es die drei Werkzeuge nicht.**
func TestOhneVorlagennahtFehlenDieWerkzeuge(t *testing.T) {
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{}, Angaben: neueAngabenprobe(),
		Stile: neueStilprobe(), Bestand: neueBestandsprobeWeit(),
	}) {
		if strings.Contains(werkzeug.Name, "vorlage") {
			t.Errorf("%q steht da, obwohl keine Naht gesetzt ist", werkzeug.Name)
		}
	}
}
