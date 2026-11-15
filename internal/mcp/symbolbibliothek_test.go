package mcp_test

// **Die Symbolbibliothek über MCP** — DOM-02
// (#1697).
//
// Hier steht, dass jedes Werkzeug seine Angaben **weiterreicht** und seinen
// Fehler nicht verschluckt; die fachliche Zusage prüft `internal/flaechen`.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// bibliotheksprobe ist ein Doppel für `mcp.Symbolbibliothek`.
type bibliotheksprobe struct {
	ergebnis any
	fehler   error

	gerufen    string
	id, setID  int64
	suche      string
	name       string
	inhalt     string
	bestaetigt bool
	teile      string
}

func (p *bibliotheksprobe) antwort(name string) (any, error) {
	p.gerufen = name

	return p.ergebnis, p.fehler
}

func (p *bibliotheksprobe) Symbolbibliothek(_ context.Context, suche string) (any, error) {
	p.suche = suche

	return p.antwort("Symbolbibliothek")
}

func (p *bibliotheksprobe) BibliothekssymbolText(_ context.Context, id int64) (any, error) {
	p.id = id

	return p.antwort("Text")
}

func (p *bibliotheksprobe) BibliothekssymbolAufnehmen(_ context.Context, setID int64, name string) (any, error) {
	p.setID, p.name = setID, name

	return p.antwort("Aufnehmen")
}

func (p *bibliotheksprobe) BibliothekssymbolKopieren(_ context.Context, id int64, name string) (any, error) {
	p.id, p.name = id, name

	return p.antwort("Kopieren")
}

func (p *bibliotheksprobe) BibliothekssymbolUmbenennen(_ context.Context, id int64, name string) (any, error) {
	p.id, p.name = id, name

	return p.antwort("Umbenennen")
}

func (p *bibliotheksprobe) BibliothekssymbolLoeschen(_ context.Context, id int64, bestaetigt bool) (any, error) {
	p.id, p.bestaetigt = id, bestaetigt

	return p.antwort("Loeschen")
}

func (p *bibliotheksprobe) BibliothekssymbolEinsetzen(_ context.Context, id, setID int64) (any, error) {
	p.id, p.setID = id, setID

	return p.antwort("Einsetzen")
}

func (p *bibliotheksprobe) BibliothekssymbolAusText(_ context.Context, inhalt, name string) (any, error) {
	p.inhalt, p.name = inhalt, name

	return p.antwort("AusText")
}

func (p *bibliotheksprobe) BibliothekssymbolTeile(_ context.Context, id int64) (any, error) {
	p.id = id

	return p.antwort("Teile")
}

func (p *bibliotheksprobe) BibliothekssymbolSichern(_ context.Context, id int64, teile json.RawMessage) (any, error) {
	p.id, p.teile = id, string(teile)

	return p.antwort("Sichern")
}

func (p *bibliotheksprobe) BibliothekssymbolZuruecknehmen(_ context.Context, id int64) (any, error) {
	p.id = id

	return p.antwort("Zuruecknehmen")
}

func (p *bibliotheksprobe) BibliothekssymbolNachziehen(_ context.Context, id int64) (any, error) {
	p.id = id

	return p.antwort("Nachziehen")
}

func rufeBibliothekswerkzeug(t *testing.T, probe *bibliotheksprobe, name, argumente string) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Symbolbibliothek: probe}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatalf("das Werkzeug %q gibt es nicht", name)

	return "", nil
}

// alleBibliothekswerkzeuge sind die zwölf Namen — einmal aufgeschrieben, samt
// dem, was ankommen muss.
var alleBibliothekswerkzeuge = []struct {
	werkzeug, argumente, erwartet string
	angekommen                    bibliotheksprobe
}{
	{"lesen.symbolbibliothek", `{"suche":"kro"}`, "Symbolbibliothek", bibliotheksprobe{suche: "kro"}},
	{"lesen.symbolbibliothek.symbol", `{"id":3}`, "Text", bibliotheksprobe{id: 3}},
	{
		"schreiben.symbolbibliothek.aufnehmen", `{"setID":7,"name":"Krone"}`, "Aufnehmen",
		bibliotheksprobe{setID: 7, name: "Krone"},
	},
	{
		"schreiben.symbolbibliothek.duplizieren", `{"id":3,"name":"Zacke"}`, "Kopieren",
		bibliotheksprobe{id: 3, name: "Zacke"},
	},
	{
		"schreiben.symbolbibliothek.umbenennen", `{"id":3,"name":"Reif"}`, "Umbenennen",
		bibliotheksprobe{id: 3, name: "Reif"},
	},
	{
		"schreiben.symbolbibliothek.loeschen", `{"id":3,"bestaetigt":true}`, "Loeschen",
		bibliotheksprobe{id: 3, bestaetigt: true},
	},
	{
		"schreiben.symbolbibliothek.einsetzen", `{"id":3,"setID":7}`, "Einsetzen",
		bibliotheksprobe{id: 3, setID: 7},
	},
	{
		"schreiben.symbolbibliothek.einlesen", `{"inhalt":"mse_version: 0.3.5","name":"Zinne"}`, "AusText",
		bibliotheksprobe{inhalt: "mse_version: 0.3.5", name: "Zinne"},
	},

	// **Die Bearbeitung** — DOM-02.AK-02.
	{"lesen.symbolbibliothek.teile", `{"id":3}`, "Teile", bibliotheksprobe{id: 3}},
	{
		"schreiben.symbolbibliothek.sichern", `{"id":3,"teile":[{"art":"Form"}]}`, "Sichern",
		bibliotheksprobe{id: 3, teile: `[{"art":"Form"}]`},
	},
	{"schreiben.symbolbibliothek.zuruecknehmen", `{"id":3}`, "Zuruecknehmen", bibliotheksprobe{id: 3}},
	{"schreiben.symbolbibliothek.nachziehen", `{"id":3}`, "Nachziehen", bibliotheksprobe{id: 3}},
}

func TestJedesBibliothekswerkzeugReichtSeineAngabenDurch(t *testing.T) {
	t.Parallel()

	for _, fall := range alleBibliothekswerkzeuge {
		t.Run(fall.werkzeug, func(t *testing.T) {
			t.Parallel()

			probe := &bibliotheksprobe{ergebnis: map[string]any{"id": "3"}}

			ausgabe, err := rufeBibliothekswerkzeug(t, probe, fall.werkzeug, fall.argumente)
			if err != nil || !strings.Contains(ausgabe, `"id": "3"`) {
				t.Fatalf("%q, %v", ausgabe, err)
			}

			// **Ohne das Ergebnis verglichen** — eine Tabelle lässt sich
			// nicht mit `!=` vergleichen, und sie ist nicht, was ankam.
			angekommen, erwartet := *probe, fall.angekommen
			angekommen.ergebnis, erwartet.gerufen = nil, fall.erwartet

			if angekommen != erwartet {
				t.Errorf("angekommen %+v, erwartet %+v", angekommen, erwartet)
			}
		})
	}
}

// **Und jedes verschluckt seinen Fehler nicht** — und weist unlesbare
// Argumente ab (§10: zum Ergebnis gehört auch das Nein).
func TestJedesBibliothekswerkzeugSagtNein(t *testing.T) {
	t.Parallel()

	for _, fall := range alleBibliothekswerkzeuge {
		t.Run(fall.werkzeug, func(t *testing.T) {
			t.Parallel()

			schweigt := &bibliotheksprobe{fehler: errors.New("die Naht schweigt")}
			if _, err := rufeBibliothekswerkzeug(t, schweigt, fall.werkzeug, fall.argumente); err == nil {
				t.Error("der Fehler ging verloren")
			}

			if _, err := rufeBibliothekswerkzeug(t, &bibliotheksprobe{}, fall.werkzeug, `{`); err == nil ||
				!strings.Contains(err.Error(), "JSON") {
				t.Errorf("unlesbar: %v", err)
			}
		})
	}
}
