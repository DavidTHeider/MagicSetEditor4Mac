package service_test

// **Ersetzen und zurück** — DOM-01.AK-02 und AK-03
// (#1696):
// eine neue Fassung ersetzt die bisherige, die als **vorige** aufgehoben
// wird; zurück tauscht beide. Ein lokal geändertes Paket wird nur mit
// Bestätigung ersetzt — und auch dann geht die Änderung nicht verloren.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// ersatzplatz bindet Fassung 1 eines Stils ein und legt Fassung 2 bereit.
func ersatzplatz(t *testing.T) (verwalter *service.Paketverwalter, katalog *ablage, verwaltet, neu string) {
	t.Helper()

	alt, neuort := t.TempDir(), t.TempDir()
	paket(t, alt, "fremd.mse-style", "mse version: 2.0.0\ngame: probe\nversion: 2026-01-01\n")
	paket(t, neuort, "fremd.mse-style", "mse version: 2.0.0\ngame: probe\nversion: 2026-02-02\n")

	verwaltet = filepath.Join(t.TempDir(), "pakete")
	if err := os.MkdirAll(verwaltet, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	katalog = &ablage{}
	verwalter = service.NeuerPaketverwalter(katalog, verwaltet, verwaltet)

	if _, err := verwalter.Installieren(t.Context(), filepath.Join(alt, "fremd.mse-style")); err != nil {
		t.Fatalf("installieren: %v", err)
	}

	return verwalter, katalog, verwaltet, filepath.Join(neuort, "fremd.mse-style")
}

// fassung liest die Fassung des Stils aus dem Katalog.
func fassungImKatalog(t *testing.T, katalog *ablage) string {
	t.Helper()

	for _, eines := range katalog.pakete {
		if eines.Name == "fremd" {
			return eines.Version
		}
	}

	t.Fatal("der Stil fehlt im Katalog")

	return ""
}

func TestEinPaketLaesstSichErsetzenUndZurueckholen(t *testing.T) {
	verwalter, katalog, _, neu := ersatzplatz(t)

	ersetzt, err := verwalter.Ersetzen(t.Context(), neu, false)
	if err != nil || ersetzt.Version != "2026-02-02" || fassungImKatalog(t, katalog) != "2026-02-02" {
		t.Fatalf("ersetzen: %+v, %v", ersetzt, err)
	}

	// **Der Plan kennt jetzt eine vorige Fassung.**
	plan, err := verwalter.PlanZurueck(t.Context(), "mse-style", "fremd")
	if err != nil || !plan.Moeglich || plan.Von != "2026-02-02" || plan.Zu != "2026-01-01" {
		t.Errorf("der Plan zurück: %+v, %v", plan, err)
	}

	zurueck, err := verwalter.Zurueck(t.Context(), "mse-style", "fremd")
	if err != nil || zurueck.Version != "2026-01-01" || fassungImKatalog(t, katalog) != "2026-01-01" {
		t.Fatalf("zurück: %+v, %v", zurueck, err)
	}

	// **Zurück ist selbst umkehrbar** — die ersetzte Fassung ist jetzt die
	// vorige, und die Herkunft der alten kam mit.
	if nochmal, err := verwalter.Zurueck(t.Context(), "mse-style", "fremd"); err != nil || nochmal.Version != "2026-02-02" {
		t.Errorf("noch einmal zurück: %+v, %v", nochmal, err)
	}

	for _, eines := range katalog.pakete {
		if eines.Name == "fremd" && eines.Herkunft != neu {
			t.Errorf("die Herkunft ist %q", eines.Herkunft)
		}
	}
}

// **Ein geändertes Paket wird nicht still überschrieben** (AK-03) — mit
// Bestätigung schon, und die Änderung liegt danach als vorige Fassung da.
func TestEinGeaendertesPaketBrauchtEineBestaetigung(t *testing.T) {
	verwalter, _, verwaltet, neu := ersatzplatz(t)

	datei := filepath.Join(verwaltet, "fremd.mse-style", "style")
	if err := os.WriteFile(datei, []byte("mse version: 2.0.0\ngame: probe\nversion: 2026-01-01\n# eigene Zeile, lang genug\n"), 0o600); err != nil {
		t.Fatalf("bearbeiten: %v", err)
	}

	var feldfehler *domain.FieldError
	if _, err := verwalter.Ersetzen(t.Context(), neu, false); !errors.Is(err, domain.ErrConflict) ||
		!errors.As(err, &feldfehler) || feldfehler.Key != domain.SchluesselPaketGeaendert {
		t.Fatalf("ohne Bestätigung: %v", err)
	}

	if _, err := verwalter.Ersetzen(t.Context(), neu, true); err != nil {
		t.Fatalf("mit Bestätigung: %v", err)
	}

	vorige, err := os.ReadFile(filepath.Join(verwaltet+"-vorige", "fremd.mse-style", "style")) //nolint:gosec // ein Pfad dieses Laufs
	if err != nil || !strings.Contains(string(vorige), "eigene Zeile") {
		t.Errorf("die Änderung ging verloren: %q (%v)", vorige, err)
	}
}

func TestErsetzenUndZurueckSagenWasNichtGeht(t *testing.T) {
	verwalter, _, _, neu := ersatzplatz(t)

	if _, err := verwalter.Zurueck(t.Context(), "mse-style", "fremd"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("zurück ohne vorige Fassung: %v", err)
	}

	if _, err := verwalter.Zurueck(t.Context(), "mse-style", "gibtsnicht"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("zurück ohne Paket: %v", err)
	}

	quelle := t.TempDir()
	paket(t, quelle, "anderes.mse-style", "mse version: 2.0.0\ngame: probe\n")

	if _, err := verwalter.Ersetzen(t.Context(), filepath.Join(quelle, "anderes.mse-style"), false); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ersetzen ohne Vorgänger: %v", err)
	}

	if _, err := verwalter.Ersetzen(t.Context(), filepath.Join(quelle, "gibtsnicht.mse-style"), false); err == nil {
		t.Error("ersetzen aus dem Nichts")
	}

	// **Außerhalb des verwalteten Verzeichnisses wird nichts ersetzt** — und
	// nichts zurückgeholt.
	fremd := service.NeuerPaketverwalter(&ablage{pakete: []domain.Paket{{
		Art: "mse-style", Name: "fremd", Pfad: "/buendel/saat/fremd.mse-style",
	}}}, t.TempDir())

	if _, err := fremd.Ersetzen(t.Context(), neu, true); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("ersetzen außerhalb: %v", err)
	}

	if _, err := fremd.Zurueck(t.Context(), "mse-style", "fremd"); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("zurück außerhalb: %v", err)
	}

	if _, err := service.NeuerPaketverwalter(&ablage{}, "").Ersetzen(t.Context(), neu, false); err == nil {
		t.Error("ersetzen ohne Verwaltungsort")
	}
}
