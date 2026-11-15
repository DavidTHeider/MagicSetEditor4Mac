// **Was Platz belegt, an allen vier Flächen** — Werkstück 4 von Phase 4.13
// (#1352).
//
// # Warum an den gebauten Programmen
//
// **Weil die Zählung an Dateien hängt und nicht an einer Verbindung.** Ein
// Lauf gegen ein Doppel bewiese die Rechnung; hier wird bewiesen, dass die
// Programme dieselben Dateien sehen — und dass das Wegräumen wirklich
// wegräumt.
package flaechen_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **Der ganze Weg: zählen, wegräumen, nachzählen.**
func TestDerPlatzWirdAnJederFlaecheGezaehlt(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	// **Erst ein Bestand** — sonst zählte der Lauf ein leeres Verzeichnis.
	mussKommandozeile(t, umgebung, "sets")

	neben := filepath.Dir(bestandsortAus(t, umgebung))

	for _, name := range []string{
		"bestand-vor-0.1.0.db",
		"bestand-vor-0.2.0-4498ba0291e4.db",
		"bestand-vor-0.3.0-0e3c7fea30ac.db",
	} {
		if err := os.WriteFile(filepath.Join(neben, name),
			make([]byte, 4096), 0o600); err != nil {
			t.Fatalf("vorbereiten: %v", err)
		}
	}

	// **Die Kommandozeile zählt sie.**
	aus := mussKommandozeile(t, umgebung, "platz")
	if !strings.Contains(aus, "vorsicherungen") {
		t.Errorf("die Kommandozeile nennt die Vorsicherungen nicht:\n%s", aus)
	}

	// **MCP auch** — und mit denselben Zahlen.
	roh, istFehler := mcpRuf(t, umgebung, "lesen.platz", map[string]any{})
	if istFehler {
		t.Fatalf("MCP konnte nicht zählen: %s", roh)
	}

	if !strings.Contains(roh, "vorsicherungen") {
		t.Errorf("MCP nennt die Vorsicherungen nicht:\n%s", roh)
	}

	// **Und der Dienst.**
	an := starteDienst(t, umgebung)

	status, rumpf := an.ruf(t, http.MethodGet, "/platz", "")
	if status != http.StatusOK {
		t.Fatalf("GET /platz: %d (%s)", status, rumpf)
	}

	if !strings.Contains(rumpf, `"vorsicherungen"`) {
		t.Errorf("der Dienst nennt die Vorsicherungen nicht:\n%s", rumpf)
	}

	// **Wegräumen über die Kommandozeile — und die jüngste bleibt.**
	mussKommandozeile(t, umgebung, "platz", "freigeben", "vorsicherungen")

	uebrig, err := filepath.Glob(filepath.Join(neben, "bestand-vor-*.db"))
	if err != nil {
		t.Fatalf("nachsehen: %v", err)
	}

	if len(uebrig) != 1 {
		t.Errorf("es blieben %d Dateien: %v", len(uebrig), uebrig)
	}
}

// **Ein Posten, der nicht räumbar ist, wird an jeder Fläche abgewiesen.**
func TestEinNichtRaeumbarerPostenWirdUeberallAbgewiesen(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)
	mussKommandozeile(t, umgebung, "sets")

	aus, fehler, code := kommandozeile(t, umgebung, "platz", "freigeben", "bestand")
	if code == 0 {
		t.Errorf("die Kommandozeile räumte den Bestand weg: %s", aus+fehler)
	}

	roh, istFehler := mcpRuf(t, umgebung, "schreiben.platz.freigeben",
		map[string]any{"posten": "bestand"})

	if !istFehler {
		t.Errorf("MCP räumte den Bestand weg: %s", roh)
	}

	an := starteDienst(t, umgebung)

	status, rumpf := an.ruf(t, http.MethodDelete, "/platz/bestand", "")
	if status < 400 {
		t.Errorf("der Dienst räumte den Bestand weg: %d (%s)", status, rumpf)
	}
}

// bestandsortAus liest den Ort des Bestands aus der Umgebung.
//
// **Den letzten Eintrag und nicht den ersten** — so liest ihn `os/exec`, wenn
// ein Schlüssel zweimal dasteht, und `mitEigenerSaat` setzt ihn ein zweites
// Mal, damit Bestand und Saat an einem Ort liegen (QA-10.NAT-01).
func bestandsortAus(t *testing.T, umgebung []string) string {
	t.Helper()

	ort := ""

	for _, zeile := range umgebung {
		if nach, gefunden := strings.CutPrefix(zeile, "MSE_BESTAND="); gefunden {
			ort = nach
		}
	}

	if ort == "" {
		t.Fatal("die Umgebung nennt keinen Bestandsort")
	}

	return ort
}
