// **Die Vorgangskennung in der Protokollzeile** — Werkstück O9 von Phase 5.2
// (#1433).
package app_test

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Handler setzt sie, nicht der Aufrufer.** Das ist die ganze Zusage:
// wer `…Context(ctx, …)` ruft, bekommt die Kennung ohne eine Zeile Code.
func TestEineZeileMitKontextTraegtDieVorgangskennung(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "mse.log")

	schreiber, schliessen, fehler := app.Protokoll(pfad)
	if fehler != nil {
		t.Fatalf("Protokoll anlegen: %v", fehler)
	}

	ctx := app.MitVorgang(t.Context(), domain.Vorgangskennung("0f1e2d3c"))
	schreiber.ErrorContext(ctx, "etwas ging schief", "anzahl", 1)

	// **Und eine ohne Kontext trägt keine** — sie gehört zu keinem Ruf.
	schreiber.Error("der Start ging schief", "anzahl", 2)

	if schliessfehler := schliessen(); schliessfehler != nil {
		t.Fatalf("schließen: %v", schliessfehler)
	}

	gelesen, lesefehler := os.ReadFile(pfad) // #nosec G304 -- Pfad aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	zeilen := strings.Split(strings.TrimSpace(string(gelesen)), "\n")
	if len(zeilen) != 2 {
		t.Fatalf("%d Zeilen, erwartet 2: %s", len(zeilen), gelesen)
	}

	if !strings.Contains(zeilen[0], "kennung=0f1e2d3c") {
		t.Errorf("die erste Zeile trägt keine Kennung: %s", zeilen[0])
	}

	if strings.Contains(zeilen[1], "kennung=") {
		t.Errorf("die zweite Zeile trägt eine Kennung, obwohl sie zu keinem "+
			"Ruf gehört: %s", zeilen[1])
	}
}

// **Und derselbe Fehler steht unter derselben Kennung an beiden Enden.** Das
// ist die „korrelierbare Diagnose" aus dem Issue: der Satz, den jemand sieht,
// und die Zeile, die den Grund nennt, tragen denselben Griff.
func TestDerVerdeckteGrundStehtUnterDerselbenKennung(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "mse.log")

	schreiber, schliessen, fehler := app.Protokoll(pfad)
	if fehler != nil {
		t.Fatalf("Protokoll anlegen: %v", fehler)
	}

	// **`slog.SetDefault` ist ein Prozesszustand** — deshalb läuft dieser
	// Test nicht parallel, und er gibt den vorigen Schreiber zurück.
	vorher := slog.Default()

	slog.SetDefault(schreiber)
	t.Cleanup(func() { slog.SetDefault(vorher) })

	ctx := app.MitNeuemVorgang(t.Context())

	befund := app.Uebersetze(ctx, domain.MitGrund(domain.ErrValidation,
		errors.New("no such table: cards"), "Es ist etwas schiefgegangen."))

	if schliessfehler := schliessen(); schliessfehler != nil {
		t.Fatalf("schließen: %v", schliessfehler)
	}

	gelesen, lesefehler := os.ReadFile(pfad) // #nosec G304 -- Pfad aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	switch {
	case befund.Vorgang == "":
		t.Fatal("die Meldung trägt keine Kennung")
	case !strings.Contains(string(gelesen), "kennung="+befund.Vorgang):
		t.Errorf("das Protokoll kennt die Kennung %q nicht: %s",
			befund.Vorgang, gelesen)
	case !strings.Contains(string(gelesen), "no such table"):
		t.Errorf("der verdeckte Grund fehlt: %s", gelesen)
	case strings.Contains(befund.Meldung, "no such table"):
		t.Errorf("der Grund erreichte die Oberfläche: %q", befund.Meldung)
	}
}
