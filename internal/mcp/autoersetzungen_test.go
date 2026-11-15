package mcp_test

// **Die Autoersetzungen über MCP** — Werkstück E von Phase 4.11
// (#1341).

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// regelprobe ist ein Doppel für `mcp.Autoersetzungen`.
type regelprobe struct {
	liste  any
	fehler error

	setID      int64
	muster     string
	ersatz     string
	ganzesWort bool
	aktiv      bool
	an         *bool
	zurueck    bool
	geloescht  string
}

func (r *regelprobe) Autoersetzungen(_ context.Context, setID int64) (any, error) {
	r.setID = setID

	return r.liste, r.fehler
}

func (r *regelprobe) AutoersetzungSetzen(
	_ context.Context, setID int64, muster, ersatz string, ganzesWort, aktiv bool,
) error {
	r.setID, r.muster, r.ersatz = setID, muster, ersatz
	r.ganzesWort, r.aktiv = ganzesWort, aktiv

	return r.fehler
}

func (r *regelprobe) AutoersetzungLoeschen(
	_ context.Context, setID int64, muster string,
) error {
	r.setID, r.geloescht = setID, muster

	return r.fehler
}

func (r *regelprobe) AutoersetzungenZuruecksetzen(_ context.Context, setID int64) error {
	r.setID, r.zurueck = setID, true

	return r.fehler
}

func (r *regelprobe) AutoersetzungSchalten(
	_ context.Context, setID int64, an bool,
) error {
	r.setID, r.an = setID, &an

	return r.fehler
}

// rufeRegelwerkzeug sucht ein Werkzeug und ruft es.
func rufeRegelwerkzeug(
	t *testing.T, probe *regelprobe, name, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Autoersetzungen: probe}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatalf("das Werkzeug %q gibt es nicht", name)

	return "", nil
}

// **Die Liste kommt durch.**
func TestDieAutoersetzungenKommenUeberMCP(t *testing.T) {
	t.Parallel()

	probe := &regelprobe{liste: map[string]any{"an": true}}

	roh, fehler := rufeRegelwerkzeug(t, probe, "lesen.autoersetzungen", `{"setID":7}`)
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if probe.setID != 7 || !strings.Contains(roh, "an") {
		t.Errorf("gefragt wurde nach %d, zurück kam %s", probe.setID, roh)
	}
}

// **Setzen trägt alle fünf Angaben.**
func TestDasSetzenTraegtAlleAngabenUeberMCP(t *testing.T) {
	t.Parallel()

	probe := &regelprobe{}

	if _, fehler := rufeRegelwerkzeug(t, probe, "schreiben.autoersetzung",
		`{"setID":7,"muster":"(W)","ersatz":"weiss","ganzesWort":true,"aktiv":true}`,
	); fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if probe.muster != "(W)" || probe.ersatz != "weiss" ||
		!probe.ganzesWort || !probe.aktiv {
		t.Errorf("gesetzt wurde %+v", probe)
	}
}

// **Löschen, Zurücksetzen und Schalten kommen an.**
func TestDieUebrigenAutoersetzungswerkzeugeKommenAn(t *testing.T) {
	t.Parallel()

	probe := &regelprobe{}

	if _, fehler := rufeRegelwerkzeug(t, probe,
		"schreiben.autoersetzung.loeschen", `{"setID":7,"muster":"(W)"}`,
	); fehler != nil || probe.geloescht != "(W)" {
		t.Errorf("Löschen: %v / %q", fehler, probe.geloescht)
	}

	if _, fehler := rufeRegelwerkzeug(t, probe,
		"schreiben.autoersetzungen.zuruecksetzen", `{"setID":7}`,
	); fehler != nil || !probe.zurueck {
		t.Errorf("Zurücksetzen: %v / %v", fehler, probe.zurueck)
	}

	if _, fehler := rufeRegelwerkzeug(t, probe,
		"schreiben.autoersetzung.schalter", `{"setID":7,"an":false}`,
	); fehler != nil || probe.an == nil || *probe.an {
		t.Errorf("Schalten: %v / %v", fehler, probe.an)
	}
}

// **Jedes Werkzeug prüft seine Kennung.**
func TestJedesAutoersetzungswerkzeugPrueftSeineKennung(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"lesen.autoersetzungen":                   `{"setID":0}`,
		"schreiben.autoersetzung":                 `{"setID":0,"muster":"x","ersatz":"y"}`,
		"schreiben.autoersetzung.loeschen":        `{"setID":0,"muster":"x"}`,
		"schreiben.autoersetzungen.zuruecksetzen": `{"setID":0}`,
		"schreiben.autoersetzung.schalter":        `{"setID":0,"an":true}`,
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := rufeRegelwerkzeug(
				t, &regelprobe{}, name, argumente); fehler == nil {
				t.Error("die Null ging als Kennung durch")
			}
		})
	}
}

// **Und kaputte Argumente.**
func TestKaputteArgumenteAnDenAutoersetzungswerkzeugen(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"lesen.autoersetzungen",
		"schreiben.autoersetzung",
		"schreiben.autoersetzung.loeschen",
		"schreiben.autoersetzungen.zuruecksetzen",
		"schreiben.autoersetzung.schalter",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := rufeRegelwerkzeug(
				t, &regelprobe{}, name, `{`); fehler == nil {
				t.Error("die kaputten Argumente gingen durch")
			}
		})
	}
}

// **Ein Fehler der Naht kommt durch — an jedem Weg.**
func TestEinFehlerDerRegelnahtKommtUeberMCP(t *testing.T) {
	t.Parallel()

	eigener := errors.New("kein Bestand")

	faelle := map[string]string{
		"lesen.autoersetzungen":                   `{"setID":7}`,
		"schreiben.autoersetzung":                 `{"setID":7,"muster":"x","ersatz":"y"}`,
		"schreiben.autoersetzung.loeschen":        `{"setID":7,"muster":"x"}`,
		"schreiben.autoersetzungen.zuruecksetzen": `{"setID":7}`,
		"schreiben.autoersetzung.schalter":        `{"setID":7,"an":true}`,
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := rufeRegelwerkzeug(t, &regelprobe{fehler: eigener},
				name, argumente); !errors.Is(fehler, eigener) {
				t.Errorf("der Fehler lautet %v", fehler)
			}
		})
	}
}

// **Ohne Naht gibt es die Werkzeuge nicht.**
func TestOhneNahtGibtEsKeineAutoersetzungswerkzeuge(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if strings.Contains(werkzeug.Name, "autoersetzung") {
			t.Errorf("ohne Naht steht %q im Menü", werkzeug.Name)
		}
	}
}
