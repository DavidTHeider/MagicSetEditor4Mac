package flaechen_test

// **Marken an jedem Fachobjekt über jede Fläche ohne Fenster** — UX-12
// (#1684).
//
// Geschrieben wird an allen dreien, gelesen wird an allen dreien — und alle
// sehen dasselbe: die Normalform, dieselbe Filterbedeutung, dieselbe Zählung.

import (
	"net/http"
	"strings"
	"testing"
)

func TestMarkenAnJedemObjektUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := setImProbebestand(t, umgebung)
	karteID := ersteKarte(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"))

	// **Drei Wege schreiben, jeder groß** — was ankommt, ist die Normalform.
	if gesetzt := mussKommandozeile(t, umgebung, "objektmarke", "set", setID, "Fertig", "--json"); !strings.Contains(gesetzt, `"fertig"`) {
		t.Fatalf("Kommandozeile: %s", gesetzt)
	}

	if text, istFehler := mcpRuf(t, umgebung, "schreiben.objektmarke",
		map[string]any{"art": "karte", "kennung": karteID, "marke": "Fertig"}); istFehler {
		t.Fatalf("MCP: %s", text)
	}

	an := starteDienst(t, umgebung)

	if status, rumpf := an.ruf(t, http.MethodPut, "/objektmarken/set/"+setID+"/Zum%20Tausch", ""); status != http.StatusOK ||
		!strings.Contains(rumpf, `"zum tausch"`) {
		t.Fatalf("Dienst: %d %s", status, rumpf)
	}

	t.Run("jede Fläche liest dasselbe", func(t *testing.T) {
		ausCLI := mussKommandozeile(t, umgebung, "--json", "objektmarken", "set")

		ausMCP, istFehler := mcpRuf(t, umgebung, "lesen.objektmarken", map[string]any{"art": "set"})
		if istFehler {
			t.Fatalf("MCP: %s", ausMCP)
		}

		status, ausREST := an.ruf(t, http.MethodGet, "/objektmarken/set", "")
		if status != http.StatusOK {
			t.Fatalf("Dienst: %d %s", status, ausREST)
		}

		gleicheAuskunft(t, "je Objekt — Zeile gegen Modell", ausCLI, ausMCP)
		gleicheAuskunft(t, "je Objekt — Modell gegen Dienst", ausMCP, ausREST)

		if !strings.Contains(ausCLI, `"fertig"`) || !strings.Contains(ausCLI, `"zum tausch"`) {
			t.Errorf("je Objekt: %s", ausCLI)
		}
	})

	t.Run("jede Fläche filtert mit derselben Bedeutung", func(t *testing.T) {
		und := mussKommandozeile(t, umgebung, "--json", "marken-filtern", "set", "fertig", "zum tausch")

		oder, istFehler := mcpRuf(t, umgebung, "lesen.markenfilter",
			map[string]any{"art": "set", "marken": []string{"fertig", "gibt es nicht"}, "verknuepfung": "oder"})
		if istFehler {
			t.Fatalf("MCP: %s", oder)
		}

		_, keines := an.ruf(t, http.MethodGet, "/objektmarken/set/filter?marke=fertig&marke=gibt%20es%20nicht", "")

		if !strings.Contains(und, `"`+setID+`"`) || !strings.Contains(oder, `"`+setID+`"`) || strings.Contains(keines, `"`+setID+`"`) {
			t.Errorf("und %s, oder %s, keines %s", und, oder, keines)
		}
	})

	t.Run("der Katalog zählt über alle Arten", func(t *testing.T) {
		ausCLI := mussKommandozeile(t, umgebung, "--json", "markenkatalog")

		ausMCP, istFehler := mcpRuf(t, umgebung, "lesen.markenkatalog", map[string]any{})
		if istFehler {
			t.Fatalf("MCP: %s", ausMCP)
		}

		_, ausREST := an.ruf(t, http.MethodGet, "/objektmarken", "")

		gleicheAuskunft(t, "Katalog — Zeile gegen Modell", ausCLI, ausMCP)
		gleicheAuskunft(t, "Katalog — Modell gegen Dienst", ausMCP, ausREST)

		// **„fertig" trägt das Set und die Karte** — zwei Tafeln, eine Zahl.
		if !strings.Contains(strings.Join(strings.Fields(ausCLI), ""), `{"marke":"fertig","anzahl":2}`) {
			t.Errorf("Katalog: %s", ausCLI)
		}
	})

	t.Run("Umbenennen auf einen vorhandenen Namen nur ausdrücklich", func(t *testing.T) {
		if _, _, code := kommandozeile(t, umgebung, "marke-umbenennen", "zum tausch", "fertig"); code == 0 {
			t.Error("die Kollision kam durch")
		}

		mussKommandozeile(t, umgebung, "marke-umbenennen", "zum tausch", "fertig", "--zusammenfuehren")

		status, rumpf := an.ruf(t, http.MethodDelete, "/marken/fertig", "")
		if status != http.StatusOK || !strings.Contains(rumpf, `"anzahl":2`) {
			t.Errorf("tilgen: %d %s", status, rumpf)
		}
	})

	t.Run("ein fremdes Objekt gibt es nicht", func(t *testing.T) {
		if _, _, code := kommandozeile(t, umgebung, "objektmarke", "deck", "4711", "x"); code == 0 {
			t.Error("ein Deck, das es nicht gibt, bekam eine Marke")
		}
	})

	// **Eine Kennung, die keine Zahl ist, ist ungültig und nicht unbekannt**
	// — an jeder Fläche dieselbe Fehlerart (Nachlauf über UX-12, Befund 6;
	// Vermerk 0135).
	t.Run("eine Kennung ohne Zahl ist ungültig", func(t *testing.T) {
		status, rumpf := an.ruf(t, http.MethodGet, "/objektmarken/deck/abc", "")
		if status != http.StatusBadRequest || !strings.Contains(rumpf, "validierung") {
			t.Errorf("Dienst: %d %s", status, rumpf)
		}

		if text, istFehler := mcpRuf(t, umgebung, "lesen.objektmarke",
			map[string]any{"art": "deck", "kennung": "abc"}); !istFehler || !strings.Contains(text, "ganze Zahl") {
			t.Errorf("MCP: %v %s", istFehler, text)
		}

		if _, fehler, code := kommandozeile(t, umgebung, "objektmarke", "deck", "abc"); code == 0 || !strings.Contains(fehler, "ganze Zahl") {
			t.Errorf("Zeile: %d %s", code, fehler)
		}
	})
}
