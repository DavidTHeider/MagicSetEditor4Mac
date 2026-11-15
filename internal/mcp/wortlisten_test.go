package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// wortprobe ist eine Auswahl, die sich merkt, wonach gefragt wurde.
type wortprobe struct {
	bereiche []mcp.Wortbereich
	fehler   error

	karteID int64
	feld    string
	setID   int64
	bereich int
	weg     []int
}

func (w *wortprobe) Wortlisten(_ context.Context, karteID int64, feld string) (
	[]mcp.Wortbereich, error,
) {
	w.karteID, w.feld = karteID, feld

	return w.bereiche, w.fehler
}

func (w *wortprobe) Wortwahl(
	_ context.Context, setID, karteID int64, feld string, bereich int, weg []int,
) error {
	w.setID, w.karteID, w.feld, w.bereich, w.weg = setID, karteID, feld, bereich, weg

	return w.fehler
}

// rufeWortwerkzeug sucht ein Werkzeug und ruft es.
func rufeWortwerkzeug(
	t *testing.T, listen mcp.Wortlisten, name, argumente string,
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
		Wortlisten:  listen,
		Woerter:     nil,
		Pack:        nil,
		Anleger:     nil,
		Angaben:     neueAngabenprobe(),
		Stile:       neueStilprobe(),
		Bestand:     neueBestandsprobeWeit(),
	}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatalf("%s steht nicht in der Liste", name)

	return "", nil
}

// **Die Auswahl kommt als JSON zurück** —
// #1074.
func TestLesenWortlistenGibtDieBereiche(t *testing.T) {
	probe := &wortprobe{bereiche: []mcp.Wortbereich{{
		Liste: "type-en", Text: "Creature", Gewaehlt: 0,
		Eintraege: []mcp.Worteintrag{{Name: "Creature", Aktiv: true}},
	}}}

	antwort, err := rufeWortwerkzeug(t, probe, "lesen.wortlisten",
		`{"karteID": 7, "feld": "type"}`)
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	if probe.karteID != 7 || probe.feld != "type" {
		t.Errorf("gefragt wurde nach %d/%q", probe.karteID, probe.feld)
	}

	for _, erwartet := range []string{`"Liste": "type-en"`, `"Name": "Creature"`} {
		if !strings.Contains(antwort, erwartet) {
			t.Errorf("%s fehlt in\n%s", erwartet, antwort)
		}
	}
}

// **Eine leere Antwort ist `[]` und nicht `null`.**
//
// Gefunden hat das die Sitzung von Anfang bis Ende gegen einen leeren Bestand
// (`liste`); dieselbe Falle steht hier.
func TestLesenWortlistenGibtEineLeereListe(t *testing.T) {
	antwort, err := rufeWortwerkzeug(t, &wortprobe{}, "lesen.wortlisten",
		`{"karteID": 7, "feld": "type"}`)
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	if strings.Contains(antwort, "null") {
		t.Errorf("die Antwort ist %q", antwort)
	}
}

// **Eine Wahl reicht Bereich und Weg durch.**
func TestSchreibenWortwahlReichtAllesDurch(t *testing.T) {
	probe := &wortprobe{}

	antwort, err := rufeWortwerkzeug(t, probe, "schreiben.wortwahl",
		`{"setID": 1, "karteID": 7, "feld": "type", "bereich": 2, "weg": [4, 1]}`)
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	if probe.setID != 1 || probe.karteID != 7 || probe.feld != "type" ||
		probe.bereich != 2 || len(probe.weg) != 2 || probe.weg[1] != 1 {
		t.Errorf("angekommen ist %+v", probe)
	}

	if !strings.Contains(antwort, "type") {
		t.Errorf("die Antwort ist %q", antwort)
	}
}

// **Jede Kennung wird geprüft, und jeder Fehler kommt durch.**
func TestDieWortwerkzeugeWeisenAb(t *testing.T) {
	faelle := []struct {
		name      string
		werkzeug  string
		argumente string
		fehler    error
	}{
		{name: "kein JSON", werkzeug: "lesen.wortlisten", argumente: `{`},
		{
			name: "karteID ist null", werkzeug: "lesen.wortlisten",
			argumente: `{"karteID": 0, "feld": "type"}`,
		},
		{
			name: "die Auswahl scheitert", werkzeug: "lesen.wortlisten",
			argumente: `{"karteID": 7, "feld": "type"}`,
			fehler:    errors.New("die Platte ist voll"),
		},
		{name: "kein JSON beim Schreiben", werkzeug: "schreiben.wortwahl", argumente: `{`},
		{
			name: "setID ist null", werkzeug: "schreiben.wortwahl",
			argumente: `{"setID": 0, "karteID": 7, "feld": "type", "bereich": 0, "weg": [0]}`,
		},
		{
			name: "karteID ist null beim Schreiben", werkzeug: "schreiben.wortwahl",
			argumente: `{"setID": 1, "karteID": 0, "feld": "type", "bereich": 0, "weg": [0]}`,
		},
		{
			name: "die Wahl scheitert", werkzeug: "schreiben.wortwahl",
			argumente: `{"setID": 1, "karteID": 7, "feld": "type", "bereich": 0, "weg": [0]}`,
			fehler:    errors.New("der Weg führt ins Leere"),
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, err := rufeWortwerkzeug(t,
				&wortprobe{fehler: fall.fehler}, fall.werkzeug, fall.argumente)
			if err == nil {
				t.Error("der Fehler kam nicht durch")
			}
		})
	}
}
