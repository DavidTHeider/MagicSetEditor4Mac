package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// schluesselwortprobe merkt sich, wonach gefragt wurde.
type schluesselwortprobe struct {
	liste  mcp.Schluesselwortliste
	fehler error

	setID, wortID int64
	feld, wert    string
}

func (p *schluesselwortprobe) Schluesselwoerter(_ context.Context, setID int64) (
	mcp.Schluesselwortliste, error,
) {
	p.setID = setID

	return p.liste, p.fehler
}

func (p *schluesselwortprobe) SchluesselwortAnlegen(
	_ context.Context, setID int64,
) (int64, error) {
	p.setID = setID

	return 9, p.fehler
}

func (p *schluesselwortprobe) SchluesselwortSetzen(
	_ context.Context, setID, wortID int64, feld, wert string,
) error {
	p.setID, p.wortID, p.feld, p.wert = setID, wortID, feld, wert

	return p.fehler
}

func (p *schluesselwortprobe) SchluesselwortLoeschen(
	_ context.Context, setID, wortID int64,
) error {
	p.setID, p.wortID = setID, wortID

	return p.fehler
}

// rufeSchluesselwortwerkzeug sucht ein Werkzeug und ruft es.
func rufeSchluesselwortwerkzeug(
	t *testing.T, woerter mcp.Schluesselwoerter, name, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser:       probenbestand(),
		Schreiber:   probenbestand(),
		Ausfuhr:     ruhendeAusfuhr{},
		Abgleich:    nil,
		Bilder:      nil,
		Auskunft:    stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Anlage:      nil,
		Saetze:      nil,
		Konsole:     nil,
		Wortlisten:  nil,
		Woerter:     woerter,
		Pack:        nil,
		Anleger:     nil,
		Angaben:     neueAngabenprobe(),
		Stile:       neueStilprobe(),
		Bestand:     neueBestandsprobeWeit(),
		Sammlung:    neueSammlungsprobe(),
		Preise:      neuePreisprobe(),
	}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatalf("%s steht nicht in der Liste", name)

	return "", nil
}

// **Die Liste kommt als JSON zurück, samt fest und gültig** —
// #1062.
func TestLesenSchluesselwoerterGibtDieListe(t *testing.T) {
	probe := &schluesselwortprobe{liste: mcp.Schluesselwortliste{
		Woerter: []mcp.Schluesselwort{
			{ID: 7, Wort: "Beben", Muster: "Beben", Gueltig: true},
			{Wort: "Fliegend", Fest: true, Gueltig: true},
		},
		Modi: []string{"core"},
	}}

	antwort, err := rufeSchluesselwortwerkzeug(t, probe,
		"lesen.schluesselwoerter", `{"setID": 3}`)
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	if probe.setID != 3 {
		t.Errorf("gefragt wurde nach Set %d", probe.setID)
	}

	for _, erwartet := range []string{`"Wort": "Beben"`, `"Fest": true`, `"Modi"`} {
		if !strings.Contains(antwort, erwartet) {
			t.Errorf("%s fehlt in\n%s", erwartet, antwort)
		}
	}
}

// **Eine leere Antwort ist `[]` und nicht `null`.**
func TestLesenSchluesselwoerterGibtLeereListen(t *testing.T) {
	antwort, err := rufeSchluesselwortwerkzeug(t, &schluesselwortprobe{},
		"lesen.schluesselwoerter", `{"setID": 3}`)
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	if strings.Contains(antwort, "null") {
		t.Errorf("die Antwort ist %q", antwort)
	}
}

// **Die drei schreibenden reichen alles durch.**
func TestDieSchreibendenSchluesselwortwerkzeuge(t *testing.T) {
	probe := &schluesselwortprobe{}

	antwort, err := rufeSchluesselwortwerkzeug(t, probe,
		"schreiben.schluesselwort.anlegen", `{"setID": 3}`)
	if err != nil || !strings.Contains(antwort, "9") {
		t.Fatalf("anlegen: %q / %v", antwort, err)
	}

	if _, err := rufeSchluesselwortwerkzeug(t, probe,
		"schreiben.schluesselwort.setzen",
		`{"setID": 3, "wortID": 7, "feld": "muster", "wert": "Beben"}`); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if probe.setID != 3 || probe.wortID != 7 ||
		probe.feld != "muster" || probe.wert != "Beben" {
		t.Errorf("angekommen ist %+v", probe)
	}

	if _, err := rufeSchluesselwortwerkzeug(t, probe,
		"schreiben.schluesselwort.loeschen", `{"setID": 3, "wortID": 7}`); err != nil {
		t.Errorf("loeschen: %v", err)
	}
}

// **Jede Kennung wird geprüft, und jeder Fehler kommt durch.**
func TestDieSchluesselwortwerkzeugeWeisenAb(t *testing.T) {
	faelle := []struct {
		name      string
		werkzeug  string
		argumente string
		fehler    error
	}{
		{name: "kein JSON", werkzeug: "lesen.schluesselwoerter", argumente: `{`},
		{
			name: "setID ist null", werkzeug: "lesen.schluesselwoerter",
			argumente: `{"setID": 0}`,
		},
		{
			name: "das Lesen scheitert", werkzeug: "lesen.schluesselwoerter",
			argumente: `{"setID": 3}`, fehler: errors.New("die Platte ist voll"),
		},
		{name: "kein JSON beim Anlegen", werkzeug: "schreiben.schluesselwort.anlegen", argumente: `{`},
		{
			name: "setID ist null beim Anlegen", werkzeug: "schreiben.schluesselwort.anlegen",
			argumente: `{"setID": 0}`,
		},
		{
			name: "das Anlegen scheitert", werkzeug: "schreiben.schluesselwort.anlegen",
			argumente: `{"setID": 3}`, fehler: errors.New("voll"),
		},
		{name: "kein JSON beim Setzen", werkzeug: "schreiben.schluesselwort.setzen", argumente: `{`},
		{
			name: "wortID ist null", werkzeug: "schreiben.schluesselwort.setzen",
			argumente: `{"setID": 3, "wortID": 0, "feld": "muster", "wert": "B"}`,
		},
		{
			name: "das Setzen scheitert", werkzeug: "schreiben.schluesselwort.setzen",
			argumente: `{"setID": 3, "wortID": 7, "feld": "muster", "wert": "B"}`,
			fehler:    errors.New("voll"),
		},
		{name: "kein JSON beim Löschen", werkzeug: "schreiben.schluesselwort.loeschen", argumente: `{`},
		{
			name: "setID ist null beim Löschen", werkzeug: "schreiben.schluesselwort.loeschen",
			argumente: `{"setID": 0, "wortID": 7}`,
		},
		{
			name: "das Löschen scheitert", werkzeug: "schreiben.schluesselwort.loeschen",
			argumente: `{"setID": 3, "wortID": 7}`, fehler: errors.New("voll"),
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, err := rufeSchluesselwortwerkzeug(t,
				&schluesselwortprobe{fehler: fall.fehler}, fall.werkzeug, fall.argumente)
			if err == nil {
				t.Error("der Fehler kam nicht durch")
			}
		})
	}
}
