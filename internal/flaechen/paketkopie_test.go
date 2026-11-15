// **Eine eigene Kopie — an jeder Fläche** — FMT-02.AK-04
// (#1691).
//
// Die Saat ist mitgeliefert; wer sie anpassen will, kopiert sie. Dieselbe
// Kopie, derselbe Name und dieselbe Abweisung an jeder Fläche (§16).
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"
)

// kopiesatz ist der Satz, mit dem der Dienst einen ungültigen Namen abweist —
// an jeder Fläche derselbe.
const kopiesatz = "Ein Paketname besteht aus Buchstaben, Ziffern, Bindestrichen und Unterstrichen"

func TestEinPaketWirdAnJederFlaecheKopiert(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)

		if aus, fehler, code := kommandozeile(t, umgebung,
			"paket-kopieren", "mse-style", "probe-schlicht"); code != 0 || !strings.Contains(aus, "probe-schlicht-kopie") {
			t.Fatalf("kopieren: %d (%s) %s", code, fehler, aus)
		}

		if _, fehler, code := kommandozeile(t, umgebung,
			"paket-kopieren", "mse-style", "probe-schlicht", "mein stil"); code == 0 || !strings.Contains(fehler, kopiesatz) {
			t.Errorf("ein ungültiger Name: %d (%s)", code, fehler)
		}

		imKatalog(t, umgebung, "die Kommandozeile", "probe-schlicht-kopie")
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)

		if text, istFehler := mcpRuf(t, umgebung, "schreiben.paket.kopieren",
			map[string]any{"art": "mse-style", "name": "probe-schlicht"}); istFehler ||
			!strings.Contains(text, "probe-schlicht-kopie") {
			t.Fatalf("kopieren: %s", text)
		}

		if text, istFehler := mcpRuf(t, umgebung, "schreiben.paket.kopieren",
			map[string]any{"art": "mse-style", "name": "probe-schlicht", "neu": "mein stil"}); !istFehler ||
			!strings.Contains(text, kopiesatz) {
			t.Errorf("ein ungültiger Name: %s", text)
		}

		imKatalog(t, umgebung, "MCP", "probe-schlicht-kopie")
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		an := starteDienst(t, umgebung)

		if status, rumpf := an.ruf(t, http.MethodPost, "/pakete/mse-style/probe-schlicht/kopie",
			`{}`); status != http.StatusCreated || !strings.Contains(rumpf, "probe-schlicht-kopie") {
			t.Fatalf("POST: %d (%s)", status, rumpf)
		}

		if status, rumpf := an.ruf(t, http.MethodPost, "/pakete/mse-style/probe-schlicht/kopie",
			`{"neu":"mein stil"}`); status != http.StatusBadRequest || !strings.Contains(rumpf, kopiesatz) {
			t.Errorf("ein ungültiger Name: %d (%s)", status, rumpf)
		}

		imKatalog(t, umgebung, "REST", "probe-schlicht-kopie")
	})
}
