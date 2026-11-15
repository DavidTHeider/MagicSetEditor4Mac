package mcp_test

// **Wo ein Handel steht, über MCP** —
// #1637.
//
// # Was hier gemessen wird
//
//   - **Dass das Werkzeug die Kennung durchreicht und den Stand ausgibt.**
//   - **Dass es ohne Naht gar nicht dasteht** — eines in der Liste, das beim
//     Ruf abstürzt, wäre schlimmer als keines (eigener Befund aus `T8`).
//   - **Und dass es kein zweites gibt.** Die Gegenprobe zur Zusage im Kopf
//     von `handelsstand.go`: kein Werkzeug dieser Fläche bewegt einen
//     Handel. Ohne sie wäre „ein Modell handelt nicht" ein Satz in einem
//     Kommentar.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// handelsprobe merkt sich die Kennung und gibt einen Stand.
type handelsprobe struct {
	vorgang    string
	gegenueber string
	filter     mcp.Handelsfilter
	fehler     error
}

func (h *handelsprobe) Handelsuebersicht(_ context.Context, filter mcp.Handelsfilter) (
	[]mcp.Handelsuebersichtszeile, error,
) {
	h.filter = filter

	if h.fehler != nil {
		return nil, h.fehler
	}

	return []mcp.Handelsuebersichtszeile{{
		Vorgang: mcp.Handelsstand{Vorgang: "h-8", Zustand: "angeboten"},
		Lage:    "dran", Naechster: "zustimmen",
		Gegenueber: []mcp.Handelspartner{{Kennung: "bdu", Name: "Turfer"}},
	}}, nil
}

func (h *handelsprobe) Handelsvorgaenge(_ context.Context, gegenueber string) (
	[]mcp.Handelsstand, error,
) {
	h.gegenueber = gegenueber

	if h.fehler != nil {
		return nil, h.fehler
	}

	return []mcp.Handelsstand{{
		Vorgang: "h-7", Revision: 2, Zustand: "angeboten", Art: "tausch",
		Parteien: []string{"bich", gegenueber},
	}}, nil
}

func (h *handelsprobe) Handelszusagen(_ context.Context, vorgang string) (
	[]mcp.Handelszusage, error,
) {
	h.vorgang = vorgang

	if h.fehler != nil {
		return nil, h.fehler
	}

	return []mcp.Handelszusage{{ExemplarID: 7, Menge: 2, Revision: 1}}, nil
}

func (h *handelsprobe) Handelsstand(_ context.Context, vorgang string) (
	mcp.Handelsstand, error,
) {
	h.vorgang = vorgang

	if h.fehler != nil {
		return mcp.Handelsstand{}, h.fehler
	}

	return mcp.Handelsstand{
		Vorgang: vorgang, Revision: 2, Abdruck: "ab12", Art: "tausch",
		Initiator: "bich", Parteien: []string{"bich", "bdu"},
		Zustand: "angeboten",
		Positionen: []mcp.Handelszeile{{
			Von: "bich", Printing: "a1b2", Menge: 1,
			Guete: "nearMint", Sprache: "de", Finish: "nonfoil",
		}},
		Zustimmungen: []string{"bich"},
	}, nil
}

// rufeHandel sucht das Werkzeug und ruft es.
func rufeHandel(t *testing.T, naht mcp.Handelsauskunft, argumente string) (
	string, error,
) {
	t.Helper()

	return rufeHandelswerkzeug(t, naht, "lesen.handel", argumente)
}

// rufeHandelswerkzeug ruft ein benanntes Werkzeug dieser Naht.
func rufeHandelswerkzeug(
	t *testing.T, naht mcp.Handelsauskunft, name, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Handel: naht}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), []byte(argumente))
		}
	}

	t.Fatalf("das Werkzeug %s steht nicht in der Liste", name)

	return "", nil
}

func TestDasHandelswerkzeugGibtDenStand(t *testing.T) {
	t.Parallel()

	probe := &handelsprobe{}

	antwort, fehler := rufeHandel(t, probe, `{"vorgang":"v-7"}`)
	if fehler != nil {
		t.Fatalf("lesen.handel: %v", fehler)
	}

	if probe.vorgang != "v-7" {
		t.Errorf("durchgereicht wurde %q", probe.vorgang)
	}

	for _, wort := range []string{"v-7", "angeboten", "ab12", "tausch"} {
		if !strings.Contains(antwort, wort) {
			t.Errorf("in der Antwort fehlt %q: %s", wort, antwort)
		}
	}
}

// **Ohne Naht steht es nicht in der Liste.**
func TestOhneHandelsnahtGibtEsKeinHandelswerkzeug(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if werkzeug.Name == "lesen.handel" {
			t.Fatal("lesen.handel steht ohne Naht in der Liste")
		}
	}
}

// **Und es gibt kein Werkzeug, das einen Handel bewegt.**
//
// **Die Gegenprobe zur Zusage** — ohne sie wäre „ein Modell handelt nicht"
// ein Satz in einem Kommentar (§10).
func TestKeinWerkzeugBewegtEinenHandel(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(probenaehte(&bestandsprobe{})) {
		if !strings.Contains(werkzeug.Name, "handel") {
			continue
		}

		if werkzeug.Art != mcp.Lesend {
			t.Errorf("%q ist nicht lesend", werkzeug.Name)
		}
	}
}

// **Ein krummes Argument und eine klagende Naht enden vor der Antwort.**
//
// **Beide Wege, weil beide Absagen sind** (§10: zum Ergebnis gehört das
// Nein): ein Rumpf, der kein JSON ist, und eine Naht, die klagt.
func TestDasHandelswerkzeugReichtSeineAbsagenDurch(t *testing.T) {
	t.Parallel()

	if _, fehler := rufeHandel(t, &handelsprobe{}, `{`); fehler == nil {
		t.Error("ein krummer Rumpf kam durch")
	}

	if _, fehler := rufeHandel(t,
		&handelsprobe{fehler: errors.New("klemmt")},
		`{"vorgang":"v-7"}`); fehler == nil {
		t.Error("die klagende Naht kam durch")
	}
}

// **Die Zusagen kommen über MCP — lesend.**
//
// **Welches Exemplar weggeht, wählt ein Mensch** (#1643); das Werkzeug sagt
// nur, was gewählt wurde.
func TestDieZusagenKommenUeberMCP(t *testing.T) {
	t.Parallel()

	probe := &handelsprobe{}

	roh, fehler := rufeHandelswerkzeug(
		t, probe, "lesen.handelszusagen", `{"vorgang":"v-1"}`)
	if fehler != nil {
		t.Fatalf("der Ruf: %v", fehler)
	}

	if probe.vorgang != "v-1" {
		t.Errorf("gefragt wurde nach %q", probe.vorgang)
	}

	if !strings.Contains(roh, `"exemplarId": 7`) {
		t.Errorf("die Antwort ist %s", roh)
	}
}

// **Und beide Absagen kommen durch.**
//
// **Ein krummer Rumpf und eine klagende Naht** — dieselbe Gestalt wie beim
// Werkzeug daneben (§10: zum Ergebnis gehört das Nein).
func TestEinFehlerDerZusagenKommtUeberMCPDurch(t *testing.T) {
	t.Parallel()

	if _, fehler := rufeHandelswerkzeug(t, &handelsprobe{},
		"lesen.handelszusagen", `{`); fehler == nil {
		t.Error("ein krummer Rumpf kam durch")
	}

	if _, fehler := rufeHandelswerkzeug(t,
		&handelsprobe{fehler: errors.New("der Bestand klemmt")},
		"lesen.handelszusagen", `{"vorgang":"v-1"}`); fehler == nil {
		t.Error("der Fehler kam nicht durch")
	}
}
