package mcp_test

// **Der zweite Faktor über MCP** — Werkstück 7 von Phase 4.12
// (#1351).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Die Faktoren eines Nutzers kommen an.**
func TestDieFaktorenKommenAn(t *testing.T) {
	t.Parallel()

	probe := &nutzerprobe{
		faktorzeilen: []mcp.Faktorzeile{{Art: "totp", Bezeichnung: "iPhone"}},
	}

	antwort := rufeRoh(t, mitNutzern(t, probe, "lesen.faktoren"),
		`{"nutzerID":"1"}`)

	if !strings.Contains(antwort, "iPhone") {
		t.Errorf("die Antwort ist %q", antwort)
	}
}

// **Eine leere Liste ist eine Antwort, `null` ist keine** (§3).
func TestEineLeereFaktorlisteIstEineListe(t *testing.T) {
	t.Parallel()

	antwort := rufeRoh(t, mitNutzern(t, &nutzerprobe{}, "lesen.faktoren"),
		`{"nutzerID":"1"}`)

	if strings.Contains(antwort, "null") {
		t.Errorf("die Liste ist null: %s", antwort)
	}
}

// **Ein Faktor lässt sich herausnehmen.**
func TestEinFaktorLaesstSichUeberMCPHerausnehmen(t *testing.T) {
	t.Parallel()

	probe := &nutzerprobe{}

	antwort := rufeRoh(t, mitNutzern(t, probe, "schreiben.faktor.loeschen"),
		`{"nutzerID":"1","art":"totp"}`)

	if probe.geloescht != "totp" {
		t.Errorf("gelöscht wurde %q", probe.geloescht)
	}

	if !strings.Contains(antwort, "totp") {
		t.Errorf("die Antwort ist %q", antwort)
	}
}

// **Und jeder Weg reicht seinen Fehler durch** (§8).
func TestJederFaktorwegUeberMCPReichtSeinenFehlerDurch(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct{ werkzeug, argumente string }{
		"lesen":   {"lesen.faktoren", `{"nutzerID":"1"}`},
		"löschen": {"schreiben.faktor.loeschen", `{"nutzerID":"1","art":"totp"}`},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			probe := &nutzerprobe{fehler: errors.New("kaputt")}

			if !rufeMitFehler(t, mitNutzern(t, probe, fall.werkzeug), fall.argumente) {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Und ein krummes Argument ebenso.**
func TestEinKrummesArgumentAmFaktorWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range []string{"lesen.faktoren", "schreiben.faktor.loeschen"} {
		t.Run(werkzeug, func(t *testing.T) {
			t.Parallel()

			if !rufeMitFehler(t, mitNutzern(t, &nutzerprobe{}, werkzeug), `{`) {
				t.Error("das krumme Argument kam durch")
			}
		})
	}
}
