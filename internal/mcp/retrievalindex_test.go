package mcp_test

// **Der lokale Suchindex über MCP** — Werkstück A15 von Phase 5.3
// (#1514).
//
// **Zwei Werkzeuge, beide lesend** — und das dritte Nein: der Aufbau steht
// nicht hier, weil er Minuten läuft und von hier aus nicht abbrechbar ist.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// indexprobe hält, was die Werkzeuge bekommen sollen.
type indexprobe struct {
	stand  mcp.Indexstand
	suche  mcp.Indexsuche
	fehler error

	// gesehen hält, womit gesucht wurde.
	frage  string
	grenze int
	filter mcp.Indexfilter
}

func (p *indexprobe) Indexstand(context.Context) (mcp.Indexstand, error) {
	return p.stand, p.fehler
}

func (p *indexprobe) Indexsuche(
	_ context.Context, frage string, grenze int, filter mcp.Indexfilter,
) (mcp.Indexsuche, error) {
	p.frage, p.grenze, p.filter = frage, grenze, filter

	return p.suche, p.fehler
}

func TestOhneNahtGibtEsKeinIndexwerkzeug(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if strings.HasPrefix(werkzeug.Name, "lesen.index") {
			t.Fatalf("%s stand ohne Naht da", werkzeug.Name)
		}
	}
}

// **Der Stand nennt Modell, Laufzeit und die Stände je Quelle.**
func TestDasIndexstandwerkzeugNenntDieStaende(t *testing.T) {
	t.Parallel()

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{Index: &indexprobe{
		stand: mcp.Indexstand{
			Aktiv: true, Modell: "all-minilm", Laufzeit: "ollama 0.33.3",
			Staende: []mcp.Korpusstand{{
				Quelle: "oracle", Revision: "2026-09-01", Datum: "2026-09-01",
			}},
			Suchbar: true,
			Aufbau:  mcp.Indexaufbaulage{Geschrieben: 4711},
		},
	}}, "lesen.indexstand")

	antwort, fehler := werkzeug.Ruf(t.Context(), []byte(`{}`))
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	for _, teil := range []string{
		`"modell": "all-minilm"`, `"quelle": "oracle"`,
		`"revision": "2026-09-01"`, `"suchbar": true`, `"geschrieben": 4711`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("Antwort %q nennt %q nicht", antwort, teil)
		}
	}

	if werkzeug.Art != mcp.Lesend || !werkzeug.Idempotent {
		t.Errorf("die Marken sind %q und %v", werkzeug.Art, werkzeug.Idempotent)
	}
}

// **Die Suche reicht Frage, Grenze und Filter durch.**
//
// **Und die Farben bleiben zwei** — sie sind nicht dasselbe, und ein Werkzeug,
// das sie zusammenlegte, machte jede Kommandeur-Frage falsch.
func TestDasIndexsuchwerkzeugReichtSeineAngabenDurch(t *testing.T) {
	t.Parallel()

	probe := &indexprobe{suche: mcp.Indexsuche{
		Modus: "hybrid",
		Treffer: []mcp.Indextreffer{{
			Rang: 1, Kennung: "scryfall-oracle:o1:0", Quelle: "oracle",
			Titel: "Lightning Bolt", Quellrevision: "2026-09-01",
			Abgerufen: "2026-09-11T08:00:00Z",
		}},
	}}

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{Index: probe},
		"lesen.indexsuche")

	antwort, fehler := werkzeug.Ruf(t.Context(), []byte(
		`{"frage":"bolt","grenze":3,"filter":{"farben":["R"],`+
			`"farbidentitaet":["R","G"],"quellen":["oracle"],`+
			`"format":"modern","formatstand":"heute"}}`))
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if probe.frage != "bolt" || probe.grenze != 3 {
		t.Errorf("die Naht sah %q und %d", probe.frage, probe.grenze)
	}

	if len(probe.filter.Farben) != 1 || len(probe.filter.Farbidentitaet) != 2 ||
		probe.filter.Formatstand != "heute" {
		t.Errorf("der Filter ist %+v", probe.filter)
	}

	// **Die Herkunft steht am Treffer** — ohne sie ist er eine Behauptung.
	for _, teil := range []string{
		`"modus": "hybrid"`, `"quellrevision": "2026-09-01"`,
		`"abgerufen": "2026-09-11T08:00:00Z"`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("Antwort %q nennt %q nicht", antwort, teil)
		}
	}
}

// **Ohne Grenze sind es zehn.**
//
// **Und nicht null.** Ein Modell, das die Grenze weglässt, meint „die üblichen"
// — eine Suche, die deshalb nichts fände, sähe aus wie ein leerer Index.
func TestDasIndexsuchwerkzeugNimmtZehnOhneGrenze(t *testing.T) {
	t.Parallel()

	probe := &indexprobe{}

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{Index: probe},
		"lesen.indexsuche")

	for name, eingabe := range map[string]string{
		"gar keine": `{"frage":"bolt"}`,
		"null":      `{"frage":"bolt","grenze":0}`,
		"negativ":   `{"frage":"bolt","grenze":-3}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, fehler := werkzeug.Ruf(t.Context(), []byte(eingabe)); fehler != nil {
				t.Fatalf("Ruf: %v", fehler)
			}

			if probe.grenze != 10 {
				t.Errorf("die Grenze ist %d", probe.grenze)
			}
		})
	}
}

// **Beide Werkzeuge melden ihren Fehler — und eine krumme Eingabe.**
func TestDieIndexwerkzeugeMeldenIhreFehler(t *testing.T) {
	t.Parallel()

	naehte := mcp.Naehte{Index: &indexprobe{
		fehler: errors.New("die Datenbank ist zu"),
	}}

	for _, name := range []string{"lesen.indexstand", "lesen.indexsuche"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			werkzeug := gesundheitswerkzeug(t, naehte, name)

			if _, fehler := werkzeug.Ruf(t.Context(),
				[]byte(`{"frage":"x"}`)); fehler == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}

	// **Die krumme Eingabe gibt es nur bei der Suche** — der Stand nimmt
	// keine Angaben entgegen, und ein `{` ginge dort durch.
	werkzeug := gesundheitswerkzeug(t, naehte, "lesen.indexsuche")

	if _, fehler := werkzeug.Ruf(t.Context(), []byte(`{"frage":`)); fehler == nil {
		t.Error("die krumme Eingabe kam durch")
	}
}

// **Über MCP wird kein Index gebaut** — mit Absicht.
func TestUeberMCPWirdKeinIndexAufgebaut(t *testing.T) {
	t.Parallel()

	gefunden := false

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Index: &indexprobe{}}) {
		if strings.Contains(werkzeug.Name, "index") {
			gefunden = true
		}

		if strings.HasPrefix(werkzeug.Name, "schreiben."+"index") {
			t.Errorf("%s baut über MCP einen Index", werkzeug.Name)
		}
	}

	// **Die Gegenprobe zum Nein**: ohne die lesenden Werkzeuge sagte der
	// Lauf nichts über das Fehlen des schreibenden (§10).
	if !gefunden {
		t.Error("es steht gar kein Indexwerkzeug da")
	}
}
