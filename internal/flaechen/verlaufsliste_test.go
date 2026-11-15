package flaechen_test

import (
	"net/http"
	"strings"
	"testing"
)

// **Der Verlauf über jede Fläche** — Werkstück U6 von Phase 5.1 (#1417):
// nach einer Massenänderung über zwei Karten nennt jede Fläche denselben
// Schritt — mit Umfang 2 und dem Vorgang, den `zuruecknehmen` nimmt.
func TestDerVerlaufNenntDenSchrittUeberJedeFlaeche(t *testing.T) {
	for name, verlauf := range map[string]func(*testing.T, []string, string) string{
		"Kommandozeile": func(t *testing.T, umgebung []string, setID string) string {
			t.Helper()

			return mussKommandozeile(t, umgebung, "verlauf", setID, "--json")
		},
		"MCP": func(t *testing.T, umgebung []string, setID string) string {
			t.Helper()

			text, istFehler := mcpRuf(t, umgebung, "lesen.verlauf", map[string]any{
				"setID": zahl(t, setID),
			})
			if istFehler {
				t.Fatalf("lesen.verlauf: %s", text)
			}

			return text
		},
		"REST": func(t *testing.T, umgebung []string, setID string) string {
			t.Helper()

			an := starteDienst(t, umgebung)
			status, rumpf := an.ruf(t, http.MethodGet, "/sets/"+setID+"/verlauf", "")
			if status != http.StatusOK {
				t.Fatalf("GET: %d (%s)", status, rumpf)
			}

			return rumpf
		},
	} {
		t.Run(name, func(t *testing.T) {
			umgebung, setID, karteIDs := setMitKarten(t)
			zwei := karteIDs[:2]

			aus := mussKommandozeile(t, umgebung, append(
				[]string{"felder-setzen", setID, "artist", "Im Verlauf", "--json"}, zwei...)...)
			if !strings.Contains(aus, `"vorgang"`) {
				t.Fatalf("kein Vorgang: %s", aus)
			}

			text := verlauf(t, umgebung, setID)

			if !strings.Contains(text, `"umfang": 2`) && !strings.Contains(text, `"umfang":2`) {
				t.Errorf("der Schritt über zwei Karten fehlt: %s", text)
			}

			if !strings.Contains(text, `"nachher": "Im Verlauf"`) && !strings.Contains(text, `"nachher":"Im Verlauf"`) {
				t.Errorf("der Nachher-Wert fehlt: %s", text)
			}
		})
	}
}
