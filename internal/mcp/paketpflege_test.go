// **Die Paketpflege über MCP** —
// #1291.
package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// pflegeprobe merkt sich, was gerufen wurde.
type pflegeprobe struct {
	installiert string
	uebernommen string
	entfernt    string
	fehler      error

	bildgesetzt   string
	vorlagelauf   string
	optiongesetzt string

	// ohneOrte stellt einen Rechner ohne MSE2 dar.
	ohneOrte bool

	// ohnePakete stellt eine Übernahme dar, die nichts gefunden hat.
	ohnePakete bool

	// angelegt hält fest, was angelegt wurde (#1350).
	angelegt string

	// kopiert hält fest, was kopiert wurde (FMT-02, #1691).
	kopiert, weitergegeben, eingelesen string

	// geplant, ersetzt und zurueck halten den Plan, das Ersetzen und das
	// Zurück fest (DOM-01, #1696).
	geplant, ersetzt, zurueck string
}

func (p *pflegeprobe) PaketInstallieren(_ context.Context, quelle string) (mcp.Paketzeile, error) {
	p.installiert = quelle

	if p.fehler != nil {
		return mcp.Paketzeile{}, p.fehler
	}

	return mcp.Paketzeile{Art: "mse-style", Name: "neu", Version: "2026-08-30"}, nil
}

// PaketAnlegen legt ein neues Paket an (#1350).
func (p *pflegeprobe) PaketAnlegen(_ context.Context, art, name, _, _ string) (
	mcp.Paketzeile, error,
) {
	p.angelegt = art + "/" + name

	if p.fehler != nil {
		return mcp.Paketzeile{}, p.fehler
	}

	return mcp.Paketzeile{Art: art, Name: name, Version: "2026-09-04"}, nil
}

// PaketKopieren legt eine eigene Kopie an (FMT-02, #1691).
func (p *pflegeprobe) PaketKopieren(_ context.Context, art, name, neu string) (mcp.Paketzeile, error) {
	p.kopiert = art + "/" + name + "→" + neu

	if p.fehler != nil {
		return mcp.Paketzeile{}, p.fehler
	}

	return mcp.Paketzeile{Art: art, Name: name + "-kopie"}, nil
}

func (p *pflegeprobe) PaketEntfernen(_ context.Context, art, name, pfad string) error {
	p.entfernt = art + "/" + name + " (" + pfad + ")"

	return p.fehler
}

func (p *pflegeprobe) PaketUebernehmen(_ context.Context, ort string) (mcp.Paketuebernahme, error) {
	p.uebernommen = ort

	uebernahme := mcp.Paketuebernahme{
		Uebernommen: []mcp.Paketzeile{{Art: "mse-game", Name: "magic"}},
		Uebergangen: []string{},
		Befunde:     []string{},
	}

	if p.ohnePakete {
		uebernahme.Uebernommen = nil
	}

	// **Auch im Fehlerfall mit Teilerfolg**: die Pakete liegen schon da, wenn
	// das Auffrischen des Katalogs scheitert (Nachlauf zu #1291).
	return uebernahme, p.fehler
}

// AusDatei setzt ein Bild von der Platte — dieselbe Probe bedient beides,
// weil beide Nähte in denselben Vorrichtungen gesetzt werden.
func (p *pflegeprobe) AusDatei(_ context.Context, _, _ int64, feld, pfad string) (
	mcp.Bildmasse, error,
) {
	p.bildgesetzt = feld + " <- " + pfad

	if p.fehler != nil {
		return mcp.Bildmasse{}, p.fehler
	}

	return mcp.Bildmasse{
		Feld: feld, QuelleBreite: 400, QuelleHoehe: 300,
		ZielBreite: 315, ZielHoehe: 230,
	}, nil
}

// MTGJSONLesen gibt einen **Teilerfolg** samt Fehler (#1291, Befund 1).
func (p *pflegeprobe) MTGJSONLesen(context.Context, string, string) (
	mcp.MTGJSONEinfahrt, error,
) {
	return mcp.MTGJSONEinfahrt{Angelegt: []int64{7}, Befunde: []string{}},
		errors.New("die Datei bricht ab")
}

func (p *pflegeprobe) MTGJSONSchreiben(context.Context, int64, string, string) (
	karten int, befunde []string, fehler error,
) {
	return 0, []string{}, nil
}

// Einfahren gibt einen **Teilerfolg** samt Fehler (#1291, Nachlauf, Befund 1):
// ein Archiv, das sich nach dem Einfahren nicht schließen lässt, hinterlässt
// ein Set im Bestand.
func (p *pflegeprobe) Einfahren(context.Context, string) (mcp.Einfahrt, error) {
	return mcp.Einfahrt{SetID: 7, Karten: 3, Befunde: []string{}},
		errors.New("das Archiv schließen")
}

func (p *pflegeprobe) Sichern(context.Context, int64, string) ([]string, error) {
	return []string{}, nil
}

func (p *pflegeprobe) Zeichnen(context.Context, int64, string, string) (
	breite, hoehe int, befunde []string, fehler error,
) {
	return 0, 0, []string{}, nil
}

func (p *pflegeprobe) Bogen(context.Context, int64, string) (
	seiten, karten int, befunde []string, fehler error,
) {
	return 0, 0, []string{}, nil
}

func (p *pflegeprobe) MTGJSONSpiele() []string { return []string{"magic"} }

// Die drei Züge der Exportvorlagen (#1291).
func (p *pflegeprobe) Vorlagen(context.Context, int64) ([]mcp.Vorlagenzeile, []string, error) {
	if p.fehler != nil {
		return nil, nil, p.fehler
	}

	return []mcp.Vorlagenzeile{{
		Paket: "probe-liste.mse-export-template", Name: "Liste",
		Dateityp: "*.txt|*.txt", Optionen: []mcp.Vorlagenoption{
			{Name: "trenner", Art: "text", Wert: ", "},
		},
	}}, []string{}, nil
}

func (p *pflegeprobe) Ausfuehren(
	_ context.Context, _ int64, paket, ziel string, _ map[string]string,
) ([]string, error) {
	p.vorlagelauf = paket + " -> " + ziel

	if p.fehler != nil {
		return nil, p.fehler
	}

	return []string{}, nil
}

func (p *pflegeprobe) OptionSetzen(_ context.Context, vorlage, feld, wert string) error {
	p.optiongesetzt = vorlage + "." + feld + "=" + wert

	return p.fehler
}

func (p *pflegeprobe) MSE2Orte() []string {
	if p.ohneOrte {
		return nil
	}

	return []string{"/Anwendungen/Magic Set Editor/data"}
}

// mitPflege sucht ein Werkzeug in der Liste mit Paketpflege.
func mitPflege(t *testing.T, probe *pflegeprobe, name string) mcp.Werkzeug {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Ausfuhr: probe, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{}, Angaben: neueAngabenprobe(),
		Stile: neueStilprobe(), Bestand: neueBestandsprobeWeit(),
		Pflege: probe, Kartenbilder: probe, Vorlagen: probe,
	}) {
		if werkzeug.Name == name {
			return werkzeug
		}
	}

	t.Fatalf("kein Werkzeug %q", name)

	return mcp.Werkzeug{}
}

// **Die vier stehen in der Liste und tun, was sie sagen.**
func TestDiePaketpflegeStehtUeberMCP(t *testing.T) {
	t.Run("Orte", func(t *testing.T) {
		antwort := rufeRoh(t, mitPflege(t, &pflegeprobe{}, "lesen.mse2orte"), `{}`)

		if !strings.Contains(antwort, "Magic Set Editor") {
			t.Errorf("die Antwort nennt keinen Ort: %s", antwort)
		}
	})

	t.Run("installieren", func(t *testing.T) {
		probe := &pflegeprobe{}

		antwort := rufeRoh(t, mitPflege(t, probe, "schreiben.paket.installieren"),
			`{"quelle":"/tmp/magic-m15.mse-style"}`)

		if probe.installiert != "/tmp/magic-m15.mse-style" {
			t.Errorf("eingebunden wurde %q", probe.installiert)
		}

		if !strings.Contains(antwort, `"name": "neu"`) {
			t.Errorf("die Antwort nennt das Paket nicht: %s", antwort)
		}
	})

	// **Anlegen steht neben Einbinden** (#1350): beides bringt ein Paket in
	// den Katalog, das eine von der Platte, das andere aus dem Nichts.
	t.Run("anlegen", func(t *testing.T) {
		probe := &pflegeprobe{}

		antwort := rufeRoh(t, mitPflege(t, probe, "schreiben.paket.anlegen"),
			`{"art":"mse-game","name":"eigenes","vollname":"Eigenes"}`)

		if probe.angelegt != "mse-game/eigenes" {
			t.Errorf("angelegt wurde %q", probe.angelegt)
		}

		if !strings.Contains(antwort, `"name": "eigenes"`) {
			t.Errorf("die Antwort nennt das Paket nicht: %s", antwort)
		}
	})

	// **Und die eigene Kopie** (FMT-02, #1691) — ohne Namen heißt sie wie
	// das Original mit `-kopie`.
	t.Run("kopieren", func(t *testing.T) {
		probe := &pflegeprobe{}

		antwort := rufeRoh(t, mitPflege(t, probe, "schreiben.paket.kopieren"),
			`{"art":"mse-style","name":"probe-schlicht"}`)

		if probe.kopiert != "mse-style/probe-schlicht→" {
			t.Errorf("kopiert wurde %q", probe.kopiert)
		}

		if !strings.Contains(antwort, `"name": "probe-schlicht-kopie"`) {
			t.Errorf("die Antwort nennt die Kopie nicht: %s", antwort)
		}
	})

	t.Run("übernehmen", func(t *testing.T) {
		probe := &pflegeprobe{}

		antwort := rufeRoh(t, mitPflege(t, probe, "schreiben.paket.uebernehmen"),
			`{"ort":"/Anwendungen/Magic Set Editor/data"}`)

		if probe.uebernommen != "/Anwendungen/Magic Set Editor/data" {
			t.Errorf("übernommen wurde aus %q", probe.uebernommen)
		}

		// **Die drei Listen sind nie `nil`** (§3).
		for _, feld := range []string{"uebernommen", "uebergangen", "befunde"} {
			if strings.Contains(antwort, `"`+feld+`": null`) {
				t.Errorf("%s ist null: %s", feld, antwort)
			}
		}
	})
}

// **Das Entfernen reicht den Pfad als Bestätigung durch.**
//
// **Geprüft wird er an der Wurzel** (`cmd/mse-mcp`), nicht hier: die
// Bestätigung ist eine Aussage über die Platte, und dieses Paket ist ein
// Blatt (0044). Was hier gilt, ist, dass der Wert **ankommt** — ginge er
// verloren, prüfte die Wurzel gegen eine leere Zeichenkette.
func TestDasEntfernenReichtDenPfadDurch(t *testing.T) {
	probe := &pflegeprobe{}

	rufeRoh(t, mitPflege(t, probe, "schreiben.paket.entfernen"),
		`{"art":"mse-style","name":"magic-m15","pfad":"/pakete/magic-m15.mse-style"}`)

	if probe.entfernt != "mse-style/magic-m15 (/pakete/magic-m15.mse-style)" {
		t.Errorf("entfernt wurde %q", probe.entfernt)
	}
}

// **Und ein krummer Rumpf kommt gar nicht erst durch.**
func TestDasEntfernenWeistEinenKrummenRumpfAb(t *testing.T) {
	probe := &pflegeprobe{}

	if _, err := mitPflege(t, probe, "schreiben.paket.entfernen").
		Ruf(t.Context(), json.RawMessage(`{`)); err == nil {
		t.Fatal("kein Fehler")
	}

	if probe.entfernt != "" {
		t.Errorf("es wurde trotzdem entfernt: %q", probe.entfernt)
	}
}

// **Und jedes weist ab, was es nicht lesen kann.**
func TestDiePaketpflegeWeistAb(t *testing.T) {
	faelle := map[string]struct {
		werkzeug  string
		argumente string
		probe     *pflegeprobe
	}{
		"installieren: krummer Rumpf": {"schreiben.paket.installieren", `{`, &pflegeprobe{}},
		"installieren: die Naht klagt": {
			"schreiben.paket.installieren", `{"quelle":"/tmp/x.mse-style"}`,
			&pflegeprobe{fehler: errors.New("kein Verzeichnis")},
		},
		"anlegen: krummer Rumpf": {"schreiben.paket.anlegen", `{`, &pflegeprobe{}},
		"anlegen: die Naht klagt": {
			"schreiben.paket.anlegen", `{"art":"mse-game","name":"x"}`,
			&pflegeprobe{fehler: errors.New("krummer Name")},
		},
		"kopieren: krummer Rumpf": {"schreiben.paket.kopieren", `{`, &pflegeprobe{}},
		"kopieren: die Naht klagt": {
			"schreiben.paket.kopieren", `{"art":"mse-game","name":"probe","neu":"x y"}`,
			&pflegeprobe{fehler: errors.New("krummer Name")},
		},
		"übernehmen: krummer Rumpf": {"schreiben.paket.uebernehmen", `{`, &pflegeprobe{}},
		"übernehmen: die Naht klagt": {
			"schreiben.paket.uebernehmen", `{"ort":"/tmp"}`,
			&pflegeprobe{fehler: errors.New("kein Verzeichnis")},
		},
		"entfernen: die Naht klagt": {
			"schreiben.paket.entfernen",
			`{"art":"mse-style","name":"magic-m15","pfad":"/pakete/magic-m15.mse-style"}`,
			&pflegeprobe{fehler: errors.New("nicht löschbar")},
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

// **Eine leere Ortsliste kommt als leere Liste an und nicht als `null`** (§3).
//
// Der Zweig, der `nil` zu `[]` macht: auf einem Rechner ohne MSE2 ist das der
// Normalfall.
func TestEineLeereOrtslisteKommtAlsListeAn(t *testing.T) {
	antwort := rufeRoh(t, mitPflege(t, &pflegeprobe{ohneOrte: true}, "lesen.mse2orte"), `{}`)

	if strings.Contains(antwort, "null") {
		t.Errorf("die Orte kamen als null an: %s", antwort)
	}
}

// **Ohne Naht gibt es die vier Werkzeuge nicht** — wie bei jeder anderen.
func TestOhnePflegenahtFehlenDieWerkzeuge(t *testing.T) {
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{}, Angaben: neueAngabenprobe(),
		Stile: neueStilprobe(), Bestand: neueBestandsprobeWeit(),
	}) {
		if strings.Contains(werkzeug.Name, "paket.installieren") ||
			strings.Contains(werkzeug.Name, "mse2orte") {
			t.Errorf("%q steht da, obwohl keine Naht gesetzt ist", werkzeug.Name)
		}
	}
}
