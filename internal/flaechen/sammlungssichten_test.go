// **Die vier Fragen an eine geordnete Sammlung — an jeder Fläche** —
// Werkstück 5 von Phase 4.5
// (#1292,
// Vermerk [0057](../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md)).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// geordneteSammlung legt Besitz und Wunsch desselben Drucks an.
//
// **Derselbe Druck zweimal** — genau der Fall, für den die Eindeutigkeit
// `wunsch` aufnimmt (0057): wer eine gespielte Fassung hat und eine schöne
// will, hat beides.
func geordneteSammlung(t *testing.T) []string {
	t.Helper()

	umgebung := eigenerBestand(t)

	mussKommandozeile(t, umgebung, "exemplar", "probe-druck", "4",
		"nearMint", "en", "nonfoil", "", "", "", "", "", "", "", "Kiste 4", "besitz")
	mussKommandozeile(t, umgebung, "exemplar", "probe-druck", "2",
		"played", "en", "nonfoil", "", "", "", "", "", "", "", "", "wunsch")

	return umgebung
}

// **Besitz und Wunsch werden überall gleich getrennt.**
func TestDerBesitzWirdUeberallGleichGezaehlt(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		pruefeBesitz(t, mussKommandozeile(t, geordneteSammlung(t),
			"besitz", "probe-druck", "--json"))
	})

	t.Run("MCP", func(t *testing.T) {
		roh, istFehler := mcpRuf(t, geordneteSammlung(t), "lesen.sammlung.besitz",
			map[string]any{"printing": "probe-druck"})
		if istFehler {
			t.Fatalf("lesen.sammlung.besitz: %s", roh)
		}

		pruefeBesitz(t, roh)
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, geordneteSammlung(t))

		status, rumpf := an.ruf(t, http.MethodGet, "/sammlung/besitz/probe-druck", "")
		if status != http.StatusOK {
			t.Fatalf("GET /sammlung/besitz: %d (%s)", status, rumpf)
		}

		pruefeBesitz(t, rumpf)
	})
}

// pruefeBesitz hält die fachliche Zusage — an jeder Fläche dieselbe.
func pruefeBesitz(t *testing.T, roh string) {
	t.Helper()

	var stand struct {
		Stueck     int  `json:"stueck"`
		Gewuenscht int  `json:"gewuenscht"`
		Besitzt    bool `json:"besitzt"`
	}

	if err := json.Unmarshal([]byte(roh), &stand); err != nil {
		t.Fatalf("die Antwort lesen: %v (%s)", err, roh)
	}

	if stand.Stueck != 4 {
		t.Errorf("%d besessen, erwartet 4", stand.Stueck)
	}

	// **Der Wunsch zählt nicht zum Besitz** (0057, Punkt 2).
	if stand.Gewuenscht != 2 {
		t.Errorf("%d gewünscht, erwartet 2", stand.Gewuenscht)
	}

	if !stand.Besitzt {
		t.Error("vier Stück gelten als kein Besitz")
	}
}

// **Die Lagerorte, Dubletten und Wünsche gibt es an jeder Fläche.**
func TestDieOrdnungStehtAnJederFlaeche(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := geordneteSammlung(t)

		if orte := mussKommandozeile(t, umgebung, "lagerorte"); !strings.Contains(
			orte, "Kiste 4") {
			t.Errorf("die Lagerorte nennen »Kiste 4« nicht: %q", orte)
		}

		// **Behalten 1 heißt: drei sind zu viel.**
		if dubletten := mussKommandozeile(t, umgebung, "dubletten", "1"); !strings.Contains(
			dubletten, "3") {
			t.Errorf("die Dubletten nennen die drei Überzähligen nicht: %q", dubletten)
		}

		wuensche := mussKommandozeile(t, umgebung, "wuensche", "--json")
		if !strings.Contains(wuensche, "played") {
			t.Errorf("die Wunschliste nennt den Wunsch nicht: %q", wuensche)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := geordneteSammlung(t)

		for name, argumente := range map[string]map[string]any{
			"lesen.sammlung.lagerorte": {},
			"lesen.sammlung.dubletten": {"behalten": 1},
			"lesen.sammlung.wuensche":  {},
		} {
			if roh, istFehler := mcpRuf(t, umgebung, name, argumente); istFehler {
				t.Errorf("%s: %s", name, roh)
			}
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, geordneteSammlung(t))

		for _, weg := range []string{
			"/sammlung/lagerorte", "/sammlung/dubletten?behalten=1",
			"/sammlung/wuensche",
		} {
			if status, rumpf := an.ruf(t, http.MethodGet, weg, ""); status != http.StatusOK {
				t.Errorf("GET %s: %d (%s)", weg, status, rumpf)
			}
		}
	})
}
