package flaechen_test

// **Eine Datei ansehen über jede Fläche ohne Fenster** — UX-13
// (#1685).
//
// Dieselbe Datei, dieselbe Auskunft an Kommandozeile, Werkzeugserver und
// Dienst — und am Dienst der Pfadvertrag: ein Winkel des Zuhauses bleibt zu,
// auch zum Ansehen.

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEineDateiAnsehenUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	liste := filepath.Join(t.TempDir(), "liste.txt")
	if fehler := os.WriteFile(liste, []byte("4 Blitzschlag\n"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	ausCLI := mussKommandozeile(t, umgebung, "--json", "datei-ansehen", liste)

	ausMCP, istFehler := mcpRuf(t, umgebung, "lesen.datei.ansicht", map[string]any{"pfad": liste})
	if istFehler {
		t.Fatalf("lesen.datei.ansicht: %s", ausMCP)
	}

	an := starteDienst(t, umgebung)

	status, ausREST := an.ruf(t, http.MethodGet, "/dateien/ansicht?pfad="+url.QueryEscape(liste), "")
	if status != http.StatusOK {
		t.Fatalf("GET /dateien/ansicht: %d (%s)", status, ausREST)
	}

	gleicheAuskunft(t, "die Ansicht — Zeile gegen Modell", ausCLI, ausMCP)
	gleicheAuskunft(t, "die Ansicht — Modell gegen Dienst", ausMCP, ausREST)

	if !strings.Contains(ausCLI, "4 Blitzschlag") || !strings.Contains(ausCLI, `"art": "text"`) {
		t.Errorf("die Ansicht sagt das Falsche:\n%s", ausCLI)
	}

	// **Der Pfadvertrag gilt auch zum Ansehen** — `~/.ssh` bleibt zu.
	zuhause := ""
	for _, eintrag := range umgebung {
		if wert, ist := strings.CutPrefix(eintrag, "HOME="); ist {
			zuhause = wert
		}
	}

	geheim := filepath.Join(zuhause, ".ssh", "id_ed25519")
	if fehler := os.MkdirAll(filepath.Dir(geheim), 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.WriteFile(geheim, []byte("geheim"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	if status, antwort := an.ruf(t, http.MethodGet, "/dateien/ansicht?pfad="+url.QueryEscape(geheim), ""); status == http.StatusOK ||
		strings.Contains(antwort, "geheim\"") {
		t.Errorf("~/.ssh ließ sich ansehen: %d %s", status, antwort)
	}
}
