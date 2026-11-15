package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Die beiden Ersetzungswerkzeuge** — Werkstück 2 von Phase 3.7
// (#1168).

// ersetzprobe ist ein Doppel der Naht.
type ersetzprobe struct {
	fehler  error
	getan   bool
	gefragt domain.Ersetzungsfrage
}

func (e *ersetzprobe) ErsetzungVorschau(
	_ context.Context, _ int64, frage domain.Ersetzungsfrage,
) (domain.Ersetzungsvorschau, error) {
	e.gefragt = frage

	return domain.Ersetzungsvorschau{
		Stellen: []domain.Ersetzungsstelle{{
			KarteID: 7, Nummer: 2, Feld: "rule text",
			Vorher: "Fliegend", Nachher: "Rennend", Treffer: 1,
		}},
		Karten: 1, Treffer: 1,
	}, e.fehler
}

func (e *ersetzprobe) ErsetzungAusfuehren(
	ctx context.Context, setID int64, frage domain.Ersetzungsfrage,
) (domain.Ersetzungsvorschau, error) {
	e.getan = true

	return e.ErsetzungVorschau(ctx, setID, frage)
}

// rufeErsetzwerkzeug ruft eines der beiden über die **echte** Werkzeugliste.
func rufeErsetzwerkzeug(
	t *testing.T, ersetzung mcp.Fremdersetzung, name, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser:     probenbestand(),
		Schreiber: probenbestand(),
		Ausfuhr:   ruhendeAusfuhr{},
		Auskunft:  stilleAuskunft{},
		Ersetzung: ersetzung,
		Angaben:   neueAngabenprobe(),
		Stile:     neueStilprobe(),
		Bestand:   neueBestandsprobeWeit(),
	}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatalf("es gibt kein Werkzeug %q", name)

	return "", nil
}

// **Das lesende zeigt und schreibt nicht, das schreibende tut es.**
func TestDieErsetzwerkzeugeGehenIhreZweiWege(t *testing.T) {
	t.Parallel()

	t.Run("lesen.ersetzen zeigt nur", func(t *testing.T) {
		t.Parallel()

		probe := &ersetzprobe{}

		text, fehler := rufeErsetzwerkzeug(t, probe, "lesen.ersetzen",
			`{"setID":1,"text":"Flieg","ersatz":"Renn","feld":"rule text",
			  "grossKleinBeachten":true,"ganzesWort":true}`)
		if fehler != nil {
			t.Fatalf("Ruf: %v", fehler)
		}

		if probe.getan {
			t.Error("das lesende Werkzeug hat ausgeführt")
		}

		for _, teil := range []string{"Rennend", `"treffer": 1`, `"nummer": 3`} {
			if !strings.Contains(text, teil) {
				t.Errorf("die Antwort führt %q nicht: %s", teil, text)
			}
		}

		if probe.gefragt.Feld != "rule text" || !probe.gefragt.GrossKleinBeachten ||
			!probe.gefragt.GanzesWort {
			t.Errorf("gefragt wurde %+v", probe.gefragt)
		}
	})

	t.Run("schreiben.ersetzen tut es", func(t *testing.T) {
		t.Parallel()

		probe := &ersetzprobe{}

		if _, fehler := rufeErsetzwerkzeug(t, probe, "schreiben.ersetzen",
			`{"setID":1,"text":"Flieg","ersatz":"Renn"}`); fehler != nil {
			t.Fatalf("Ruf: %v", fehler)
		}

		if !probe.getan {
			t.Error("ausgeführt wurde nicht")
		}
	})
}

// **Jede Wache greift, und zwar an beiden Werkzeugen.**
func TestDieErsetzwerkzeugeWeisenUnbrauchbareArgumenteAb(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, werkzeug, argumente string }{
		{"lesen: kein JSON", "lesen.ersetzen", `{`},
		{"lesen: Set ist null", "lesen.ersetzen", `{"setID":0,"text":"a"}`},
		{"lesen: kein Suchtext", "lesen.ersetzen", `{"setID":1,"text":"  "}`},
		{"schreiben: kein JSON", "schreiben.ersetzen", `{`},
		{"schreiben: Set ist null", "schreiben.ersetzen", `{"setID":0,"text":"a"}`},
		{"schreiben: kein Suchtext", "schreiben.ersetzen", `{"setID":1}`},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := rufeErsetzwerkzeug(
				t, &ersetzprobe{}, fall.werkzeug, fall.argumente); fehler == nil {
				t.Error("das Argument kam durch")
			}
		})
	}
}

// **Ein Fehler der Naht kommt an beiden durch.**
func TestDieErsetzwerkzeugeReichenIhrenFehlerDurch(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"lesen.ersetzen", "schreiben.ersetzen"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			probe := &ersetzprobe{fehler: domain.ErrNotFound}

			_, fehler := rufeErsetzwerkzeug(t, probe, name,
				`{"setID":1,"text":"a","ersatz":"b"}`)
			if !errors.Is(fehler, domain.ErrNotFound) {
				t.Errorf("Fehler: %v", fehler)
			}
		})
	}
}

// **Ohne Naht steht keines der beiden in der Liste.**
func TestOhneErsetzungsnahtGibtEsKeineErsetzwerkzeuge(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser:     probenbestand(),
		Schreiber: probenbestand(),
		Ausfuhr:   ruhendeAusfuhr{},
		Auskunft:  stilleAuskunft{},
		Angaben:   neueAngabenprobe(),
		Stile:     neueStilprobe(),
		Bestand:   neueBestandsprobeWeit(),
	}) {
		if strings.Contains(werkzeug.Name, "ersetzen") {
			t.Errorf("ohne Naht steht %q in der Liste", werkzeug.Name)
		}
	}
}
