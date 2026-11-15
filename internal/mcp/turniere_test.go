package mcp_test

// **Turniere über MCP** — Werkstück 5 von Phase 4.7
// (#1294).
//
// **Die fachliche Zusage prüft `internal/flaechen`** — dort läuft derselbe
// Ablauf über alle drei Flächen. Hier steht, dass jedes Werkzeug seine
// Angaben **weiterreicht** und seinen Fehler nicht verschluckt.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// turnierprobe ist ein Doppel für `mcp.Turniere`.
type turnierprobe struct {
	ergebnis any
	fehler   error

	gerufen      string
	turnierID    int64
	paarungID    int64
	teilnehmerID int64
	deckID       int64
	spiel        string
	name         string
	notiz        string
	einzelspiel  bool
	ausgestiegen bool
	stand        map[string]any

	// formatID und mitTeilnehmenden kamen mit DOM-04 (#1699), inhalt mit
	// dem Einlesen (DOM-04.AK-03).
	formatID         int64
	mitTeilnehmenden bool
	inhalt           string
}

func (p *turnierprobe) antwort(name string) (any, error) {
	p.gerufen = name

	return p.ergebnis, p.fehler
}

func (p *turnierprobe) Turniere(_ context.Context, spiel string) (any, error) {
	p.spiel = spiel

	return p.antwort("Turniere")
}

func (p *turnierprobe) Turnier(_ context.Context, turnierID int64) (any, error) {
	p.turnierID = turnierID

	return p.antwort("Turnier")
}

func (p *turnierprobe) TurnierRangliste(_ context.Context, turnierID int64) (
	any, error,
) {
	p.turnierID = turnierID

	return p.antwort("TurnierRangliste")
}

func (p *turnierprobe) TurnierPaarungsvorschlag(_ context.Context, turnierID int64) (
	any, error,
) {
	p.turnierID = turnierID

	return p.antwort("TurnierPaarungsvorschlag")
}

func (p *turnierprobe) TurnierAnlegen(
	_ context.Context, name, spiel, notiz string, einzelspiel bool, formatID int64,
) (any, error) {
	p.name, p.spiel, p.notiz, p.einzelspiel = name, spiel, notiz, einzelspiel
	p.formatID = formatID

	return p.antwort("TurnierAnlegen")
}

func (p *turnierprobe) TurnierAbschliessen(_ context.Context, turnierID int64) error {
	p.turnierID = turnierID
	p.gerufen = "TurnierAbschliessen"

	return p.fehler
}

func (p *turnierprobe) TurnierWiedereroeffnen(_ context.Context, turnierID int64) error {
	p.turnierID = turnierID
	p.gerufen = "TurnierWiedereroeffnen"

	return p.fehler
}

func (p *turnierprobe) TurnierProtokoll(_ context.Context, turnierID int64) (any, error) {
	p.turnierID = turnierID

	return p.antwort("TurnierProtokoll")
}

func (p *turnierprobe) TurnierKorrekturfolgen(
	_ context.Context, turnierID, paarungID int64, stand map[string]any,
) (any, error) {
	p.turnierID, p.paarungID, p.stand = turnierID, paarungID, stand

	return p.antwort("TurnierKorrekturfolgen")
}

func (p *turnierprobe) TurnierKopieren(
	_ context.Context, turnierID int64, name string, mitTeilnehmenden bool,
) (any, error) {
	p.turnierID, p.name, p.mitTeilnehmenden = turnierID, name, mitTeilnehmenden

	return p.antwort("TurnierKopieren")
}

func (p *turnierprobe) TurnierLoeschen(_ context.Context, turnierID int64) error {
	p.turnierID = turnierID
	p.gerufen = "TurnierLoeschen"

	return p.fehler
}

func (p *turnierprobe) TurnierAnmelden(
	_ context.Context, turnierID int64, name string, deckID int64,
) (any, error) {
	p.turnierID, p.name, p.deckID = turnierID, name, deckID

	return p.antwort("TurnierAnmelden")
}

func (p *turnierprobe) TurnierAussteigen(
	_ context.Context, turnierID, teilnehmerID int64, ausgestiegen bool,
) error {
	p.turnierID, p.teilnehmerID, p.ausgestiegen = turnierID, teilnehmerID, ausgestiegen
	p.gerufen = "TurnierAussteigen"

	return p.fehler
}

func (p *turnierprobe) TurnierRundeLegen(_ context.Context, turnierID int64) (
	any, error,
) {
	p.turnierID = turnierID

	return p.antwort("TurnierRundeLegen")
}

func (p *turnierprobe) TurnierErgebnisSetzen(
	_ context.Context, turnierID, paarungID int64, stand map[string]any,
) error {
	p.turnierID, p.paarungID, p.stand = turnierID, paarungID, stand
	p.gerufen = "TurnierErgebnisSetzen"

	return p.fehler
}

// rufeTurnierwerkzeug sucht ein Werkzeug und ruft es.
func rufeTurnierwerkzeug(
	t *testing.T, probe *turnierprobe, name, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Turniere: probe}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatalf("das Werkzeug %q gibt es nicht", name)

	return "", nil
}

// alleTurnierwerkzeuge sind die achtzehn Namen — einmal aufgeschrieben.
var alleTurnierwerkzeuge = []struct {
	werkzeug  string
	argumente string
	erwartet  string
	pruefe    func(*testing.T, *turnierprobe)
}{
	{"lesen.turniere", `{"spiel":"magic"}`, "Turniere", func(t *testing.T, p *turnierprobe) {
		if p.spiel != "magic" {
			t.Errorf("Spiel %q", p.spiel)
		}
	}},
	{"lesen.turnier", `{"turnierID":7}`, "Turnier", func(t *testing.T, p *turnierprobe) {
		if p.turnierID != 7 {
			t.Errorf("Turnier %d", p.turnierID)
		}
	}},
	{"lesen.turnier.rangliste", `{"turnierID":7}`, "TurnierRangliste", nil},
	{"lesen.turnier.protokoll", `{"turnierID":7}`, "TurnierProtokoll", nil},
	{"lesen.turnier.ausgabe", `{"turnierID":7}`, "TurnierAusgabe", nil},
	{"lesen.turnier.bericht", `{"turnierID":7}`, "TurnierBericht", nil},
	{"schreiben.turnier.einlesen", `{"inhalt":"{}"}`, "TurnierEinlesen", func(t *testing.T, p *turnierprobe) {
		if p.inhalt != "{}" {
			t.Errorf("Inhalt %q", p.inhalt)
		}
	}},
	{"schreiben.turnier.abschliessen", `{"turnierID":7}`, "TurnierAbschliessen", nil},
	{"schreiben.turnier.wiedereroeffnen", `{"turnierID":7}`, "TurnierWiedereroeffnen", nil},
	{
		"lesen.turnier.folgen", `{"turnierID":7,"paarungID":12,"spieleB":2}`,
		"TurnierKorrekturfolgen", func(t *testing.T, p *turnierprobe) {
			if p.paarungID != 12 || p.stand["spieleB"] != 2 {
				t.Errorf("%+v", p)
			}
		},
	},
	{
		"schreiben.turnier.vorschlag", `{"turnierID":7}`,
		"TurnierPaarungsvorschlag", nil,
	},
	{
		"schreiben.turnier.anlegen",
		`{"name":"Küchentisch","spiel":"magic","notiz":"N","einzelspiel":true,"formatID":4}`,
		"TurnierAnlegen", func(t *testing.T, p *turnierprobe) {
			if p.name != "Küchentisch" || p.notiz != "N" || !p.einzelspiel || p.formatID != 4 {
				t.Errorf("%+v", p)
			}
		},
	},
	{
		"schreiben.turnier.kopieren",
		`{"turnierID":7,"name":"Küchentisch II","mitTeilnehmenden":true}`,
		"TurnierKopieren", func(t *testing.T, p *turnierprobe) {
			if p.turnierID != 7 || p.name != "Küchentisch II" || !p.mitTeilnehmenden {
				t.Errorf("%+v", p)
			}
		},
	},
	{"schreiben.turnier.loeschen", `{"turnierID":7}`, "TurnierLoeschen", nil},
	{
		"schreiben.turnier.anmelden", `{"turnierID":7,"name":"Ada","deckID":5}`,
		"TurnierAnmelden", func(t *testing.T, p *turnierprobe) {
			if p.name != "Ada" || p.deckID != 5 {
				t.Errorf("%+v", p)
			}
		},
	},
	{
		"schreiben.turnier.aussteigen",
		`{"turnierID":7,"teilnehmerID":3,"ausgestiegen":true}`,
		"TurnierAussteigen", func(t *testing.T, p *turnierprobe) {
			if p.teilnehmerID != 3 || !p.ausgestiegen {
				t.Errorf("%+v", p)
			}
		},
	},
	{"schreiben.turnier.runde", `{"turnierID":7}`, "TurnierRundeLegen", nil},
	{
		"schreiben.turnier.ergebnis",
		`{"turnierID":7,"paarungID":12,"spieleA":2,"spieleB":1,"unentschieden":0}`,
		"TurnierErgebnisSetzen", func(t *testing.T, p *turnierprobe) {
			if p.paarungID != 12 || p.stand["spieleA"] != 2 {
				t.Errorf("%+v", p)
			}
		},
	},
}

// **Jedes Werkzeug reicht seine Angaben durch.**
func TestJedesTurnierwerkzeugReichtSeineAngabenDurch(t *testing.T) {
	t.Parallel()

	for _, fall := range alleTurnierwerkzeuge {
		t.Run(fall.werkzeug, func(t *testing.T) {
			t.Parallel()

			probe := &turnierprobe{ergebnis: map[string]any{"id": "1"}}

			if _, err := rufeTurnierwerkzeug(
				t, probe, fall.werkzeug, fall.argumente); err != nil {
				t.Fatalf("%v", err)
			}

			if probe.gerufen != fall.erwartet {
				t.Errorf("gerufen wurde %q, erwartet %q", probe.gerufen, fall.erwartet)
			}

			if fall.pruefe != nil {
				fall.pruefe(t, probe)
			}
		})
	}
}

// **Und jedes verschluckt seinen Fehler nicht.**
//
// **Zum Ergebnis gehört auch das Nein** (§10): ein Werkzeug, von dem nur der
// Erfolgsweg belegt ist, hat die Hälfte ausgelassen.
func TestJedesTurnierwerkzeugReichtSeinenFehlerDurch(t *testing.T) {
	t.Parallel()

	for _, fall := range alleTurnierwerkzeuge {
		t.Run(fall.werkzeug, func(t *testing.T) {
			t.Parallel()

			probe := &turnierprobe{fehler: errors.New("die Naht schweigt")}

			if _, err := rufeTurnierwerkzeug(
				t, probe, fall.werkzeug, fall.argumente); err == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Und jedes weist unlesbare Argumente ab.**
func TestJedesTurnierwerkzeugWeistUnlesbareArgumenteAb(t *testing.T) {
	t.Parallel()

	for _, fall := range alleTurnierwerkzeuge {
		t.Run(fall.werkzeug, func(t *testing.T) {
			t.Parallel()

			probe := &turnierprobe{}

			_, err := rufeTurnierwerkzeug(t, probe, fall.werkzeug, `{`)
			if err == nil {
				t.Fatal("angenommen, erwartet abgewiesen")
			}

			if !strings.Contains(err.Error(), "JSON") {
				t.Errorf("die Meldung sagt nicht, was falsch ist: %v", err)
			}
		})
	}
}

// **`zuruecknehmen` schickt gar keinen Stand.**
//
// Der Unterschied ist der ganze Punkt: ein Stand von 0–0–0 ist ein anderer
// Satz als „doch nicht gespielt".
func TestZuruecknehmenSchicktKeinenStand(t *testing.T) {
	t.Parallel()

	probe := &turnierprobe{}

	if _, err := rufeTurnierwerkzeug(t, probe, "schreiben.turnier.ergebnis",
		`{"turnierID":7,"paarungID":12,"zuruecknehmen":true}`); err != nil {
		t.Fatalf("%v", err)
	}

	if probe.stand != nil {
		t.Errorf("geschickt wurde %+v, erwartet nichts", probe.stand)
	}
}

// **Ohne Naht gibt es keine Turnierwerkzeuge.**
//
// Ein MCP-Server ohne Turniere soll sie nicht anbieten — ein Werkzeug, das
// beim Ruf abstürzt, wäre schlimmer als eines, das fehlt.
func TestOhneTurniernahtGibtEsKeineTurnierwerkzeuge(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if strings.Contains(werkzeug.Name, "turnier") {
			t.Errorf("das Werkzeug %q steht ohne Naht da", werkzeug.Name)
		}
	}
}
