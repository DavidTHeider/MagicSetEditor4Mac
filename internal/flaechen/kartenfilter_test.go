// **Der Schnellfilter über jede Fläche ohne Fenster** — Werkstück U5 von
// Phase 5.1 (#1416).
//
// Dieselbe Anfrage, dieselbe Antwort: `name:"Karte 2"` trifft genau eine der
// drei numerierten Karten (»Karte 1« steckt auch in »Karte 10« — Teilwort,
// wie in MSE2), `-1` lässt nur sie stehen — und alle drei Wege geben die
// Kennungen als Zeichenketten.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func filterkennungen(t *testing.T, roh string) []string {
	t.Helper()

	var antwort struct {
		KartenIDs []string `json:"kartenIDs"`
	}

	if err := json.Unmarshal([]byte(roh), &antwort); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	return antwort.KartenIDs
}

func TestDerSchnellfilterUeberJedeFlaecheOhneFenster(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := dreiNumerierteKarten(t, umgebung)

	t.Run("Kommandozeile", func(t *testing.T) {
		genau := filterkennungen(t, mussKommandozeile(t, umgebung,
			"kartenfilter", setID, `name:"Karte 2"`, "--json"))
		if len(genau) != 1 {
			t.Fatalf("name:\"Karte 2\" traf %v, erwartet genau eine", genau)
		}

		ohne := filterkennungen(t, mussKommandozeile(t, umgebung,
			"kartenfilter", setID, "-1", "--json"))
		if len(ohne) != 1 {
			t.Errorf("-1 ließ %v stehen, erwartet nur »Karte 2«", ohne)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		text, istFehler := mcpRuf(t, umgebung, "lesen.kartenfilter",
			map[string]any{"setID": zahl(t, setID), "anfrage": `name:"Karte 2"`})
		if istFehler {
			t.Fatalf("lesen.kartenfilter: %s", text)
		}

		if getroffen := filterkennungen(t, text); len(getroffen) != 1 {
			t.Errorf("MCP traf %v, erwartet genau eine", getroffen)
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodGet,
			"/sets/"+setID+"/kartenfilter?anfrage=name%3A%22Karte+2%22", "")
		if status != http.StatusOK {
			t.Fatalf("GET kartenfilter: %d (%s)", status, rumpf)
		}

		if getroffen := filterkennungen(t, rumpf); len(getroffen) != 1 {
			t.Errorf("REST traf %v, erwartet genau eine", getroffen)
		}

		if status, _ := an.ruf(t, http.MethodGet, "/sets/4711/kartenfilter?anfrage=x", ""); status != http.StatusNotFound {
			t.Errorf("ein fremdes Set antwortete mit %d, erwartet 404", status)
		}
	})
}
