package flaechen_test

// **Weggeräumt ist weggeräumt — an jeder Fläche** — §7, Regel 3
// (#1970).
//
// # Was hier gemessen wird
//
// Ein Verzeichnis unter den Bildern, das niemand lesen darf: das Wegräumen
// der Vorsicherungen liest es nicht, die Zählung danach schon. **Die
// Kommandozeile** endet mit 0 und sagt es, **der Dienst** antwortet 200 mit
// `nachladefehler`. **MCP** gibt zwei Teile: die Liste, wie sie wäre, und
// den Satz dahinter — ihr Ergebnis ist eine Liste, und ein Feld darin gäbe
// es nur um den Preis einer anderen Gestalt (0033).

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// zweiVorsicherungen legt eine ältere und eine jüngere neben den Bestand.
func zweiVorsicherungen(t *testing.T, neben string) {
	t.Helper()

	for stelle, name := range []string{"bestand-vor-0.1.0.db", "bestand-vor-0.2.0.db"} {
		pfad := filepath.Join(neben, name)
		if err := os.WriteFile(pfad, make([]byte, 4096), 0o600); err != nil {
			t.Fatalf("vorbereiten: %v", err)
		}

		alt := time.Now().Add(time.Duration(stelle-2) * time.Hour)
		if err := os.Chtimes(pfad, alt, alt); err != nil {
			t.Fatalf("vorbereiten: %v", err)
		}
	}
}

// eineGeblieben prüft, dass die ältere weg ist.
func eineGeblieben(t *testing.T, flaeche, neben string) {
	t.Helper()

	if uebrig, _ := filepath.Glob(filepath.Join(neben, "bestand-vor-*.db")); len(uebrig) != 1 { //nolint:errcheck // ein fester Ausdruck
		t.Errorf("%s: es blieben %v", flaeche, uebrig)
	}
}

func TestWeggeraeumtIstWeggeraeumtAnJederFlaeche(t *testing.T) {
	umgebung := eigenerBestand(t)
	mussKommandozeile(t, umgebung, "sets")

	an := starteDienst(t, umgebung)
	neben := filepath.Dir(bestandsortAus(t, umgebung))

	gesperrt := filepath.Join(neben, "beiwerk", "gesperrt")
	if err := os.MkdirAll(gesperrt, 0o700); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	if err := os.Chmod(gesperrt, 0o000); err != nil {
		t.Fatalf("sperren: %v", err)
	}

	t.Cleanup(func() { _ = os.Chmod(gesperrt, 0o700) }) //nolint:errcheck,gosec // zurück für t.TempDir

	zweiVorsicherungen(t, neben)

	if aus, fehlerstrom, code := kommandozeile(t, umgebung,
		"platz", "freigeben", "vorsicherungen"); code != 0 ||
		!strings.Contains(fehlerstrom, "geschrieben") {
		t.Fatalf("Kommandozeile: %d, %q, %q", code, aus, fehlerstrom)
	}

	eineGeblieben(t, "Kommandozeile", neben)
	zweiVorsicherungen(t, neben)

	status, rumpf := an.ruf(t, http.MethodDelete, "/platz/vorsicherungen", "")
	if dienst := nachgeladenAus(t, rumpf); status != http.StatusOK || dienst.Nachladefehler == nil {
		t.Fatalf("Dienst: %d, %s", status, rumpf)
	}

	eineGeblieben(t, "Dienst", neben)
	zweiVorsicherungen(t, neben)

	if teile, istFehler := mcpTeile(t, umgebung, "schreiben.platz.freigeben",
		map[string]any{"posten": "vorsicherungen"}); istFehler || len(teile) != 2 ||
		!strings.HasPrefix(teile[0], "[") || !strings.Contains(teile[1], "geschrieben") {
		t.Fatalf("MCP: %v, %q", istFehler, teile)
	}

	eineGeblieben(t, "MCP", neben)
}
