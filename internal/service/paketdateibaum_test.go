// **Der Dateibaum der Werkstatt** — FMT-02.AK-01
// (#1691):
// Dateien anlegen, von der Platte einfügen, umbenennen und entfernen, und die
// Liste nennt auch das Beiwerk, das sich nicht als Text bearbeiten lässt.
package service_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieWerkstattNenntAuchDasBeiwerk(t *testing.T) {
	t.Parallel()

	dateien, beiwerk, err := service.NeuePaketwerkstatt(werkwurzel(t)).DateienUndBeiwerk("probe.mse-style")
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(dateien, []string{"skript.txt", "style"}) || !slices.Equal(beiwerk, []string{"bilder/rahmen.png"}) {
		t.Errorf("Dateien %v, Beiwerk %v", dateien, beiwerk)
	}
}

func TestEineVorlageWirdInDerWerkstattAngelegt(t *testing.T) {
	t.Parallel()

	werkstatt := service.NeuePaketwerkstatt(werkwurzel(t))

	stand, err := werkstatt.Anlegen(t.Context(), "probe.mse-style", "skripte/neu", "# neu\n")
	if err != nil {
		t.Fatal(err)
	}

	inhalt, gelesen, err := werkstatt.Lesen("probe.mse-style", "skripte/neu")
	if err != nil || inhalt != "# neu\n" || gelesen != stand {
		t.Errorf("gelesen %q (%v), Stand %q gegen %q", inhalt, err, gelesen, stand)
	}
}

func TestEineDateiVonDerPlatteWirdEingefuegt(t *testing.T) {
	t.Parallel()

	wurzel := werkwurzel(t)
	quelle := filepath.Join(t.TempDir(), "maske.png")

	if err := os.WriteFile(quelle, []byte("\x89PNG\r\n\x1a\nmaske"), 0o600); err != nil {
		t.Fatal(err)
	}

	werkstatt := service.NeuePaketwerkstatt(wurzel)
	if err := werkstatt.Einfuegen(t.Context(), "probe.mse-style", quelle, "bilder/maske.png"); err != nil {
		t.Fatal(err)
	}

	da, err := os.ReadFile(filepath.Join(wurzel, "probe.mse-style", "bilder", "maske.png")) //nolint:gosec // Testpfad
	if err != nil || string(da) != "\x89PNG\r\n\x1a\nmaske" {
		t.Errorf("eingefügt %q", da)
	}
}

func TestEineDateiWirdInDerWerkstattUmbenanntUndEntfernt(t *testing.T) {
	t.Parallel()

	werkstatt := service.NeuePaketwerkstatt(werkwurzel(t))
	_, stand, err := werkstatt.Lesen("probe.mse-style", "skript.txt")
	if err != nil {
		t.Fatal(err)
	}

	if err = werkstatt.Umbenennen(t.Context(), "probe.mse-style", "skript.txt", "skripte/eins.txt", stand); err != nil {
		t.Fatal(err)
	}

	if err = werkstatt.Entfernen(t.Context(), "probe.mse-style", "bilder/rahmen.png", ""); err != nil {
		t.Fatal(err)
	}

	dateien, beiwerk, err := werkstatt.DateienUndBeiwerk("probe.mse-style")
	if err != nil || !slices.Equal(dateien, []string{"skripte/eins.txt", "style"}) || len(beiwerk) != 0 {
		t.Errorf("Dateien %v, Beiwerk %v", dateien, beiwerk)
	}
}

func TestDerDateibaumWeistAb(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		lauf func(werkstatt *service.Paketwerkstatt, t *testing.T) error
		will error
	}{
		"eine leere Vorlage als Bild": {
			func(w *service.Paketwerkstatt, t *testing.T) error {
				_, err := w.Anlegen(t.Context(), "probe.mse-style", "rahmen.png", "")

				return err
			}, domain.ErrValidation,
		},
		"ein Paket außerhalb": {
			func(w *service.Paketwerkstatt, t *testing.T) error {
				_, err := w.Anlegen(t.Context(), "../fremd.mse-style", "neu", "")

				return err
			}, domain.ErrValidation,
		},
		"eine fehlende Quelle": {
			func(w *service.Paketwerkstatt, t *testing.T) error {
				return w.Einfuegen(t.Context(), "probe.mse-style", filepath.Join(t.TempDir(), "fehlt.png"), "a.png")
			}, domain.ErrValidation,
		},
		"Einfügen außerhalb": {
			func(w *service.Paketwerkstatt, t *testing.T) error {
				return w.Einfuegen(t.Context(), "../x.mse-style", "egal", "a.png")
			}, domain.ErrValidation,
		},
		"Umbenennen außerhalb": {
			func(w *service.Paketwerkstatt, t *testing.T) error {
				return w.Umbenennen(t.Context(), "../x.mse-style", "a", "b", "")
			}, domain.ErrValidation,
		},
		"Entfernen außerhalb": {
			func(w *service.Paketwerkstatt, t *testing.T) error {
				return w.Entfernen(t.Context(), "../x.mse-style", "a", "")
			}, domain.ErrValidation,
		},
		"Beiwerk außerhalb": {
			func(w *service.Paketwerkstatt, _ *testing.T) error {
				_, _, err := w.DateienUndBeiwerk("../x.mse-style")

				return err
			}, domain.ErrValidation,
		},
		// **Eine Vorlage nur mit Stand** — ohne ihn wäre nicht zu erkennen,
		// ob inzwischen jemand anderes geschrieben hat (§7, Regel 2).
		"eine Vorlage ohne Stand umbenennen": {
			func(w *service.Paketwerkstatt, t *testing.T) error {
				return w.Umbenennen(t.Context(), "probe.mse-style", "skript.txt", "b.txt", "")
			}, domain.ErrValidation,
		},
		"eine Vorlage ohne Stand entfernen": {
			func(w *service.Paketwerkstatt, t *testing.T) error {
				return w.Entfernen(t.Context(), "probe.mse-style", "skript.txt", "")
			}, domain.ErrValidation,
		},
		"eine Vorlage, die schon dasteht": {
			func(w *service.Paketwerkstatt, t *testing.T) error {
				_, err := w.Anlegen(t.Context(), "probe.mse-style", "skript.txt", "")

				return err
			}, domain.ErrConflict,
		},
	} {
		if err := fall.lauf(service.NeuePaketwerkstatt(werkwurzel(t)), t); !errors.Is(err, fall.will) {
			t.Errorf("%s: %v, erwartet %v", name, err, fall.will)
		}
	}
}

// **Ohne Sperre kein Handgriff** — lässt sich die Bezugssperre neben dem Baum
// nicht anlegen, bleibt das Paket, wie es ist.
func TestOhneBezugssperreGehtKeinHandgriff(t *testing.T) {
	t.Parallel()

	wurzel := werkwurzel(t)
	if err := os.Chmod(wurzel, 0o500); err != nil { //nolint:gosec // eng mit Absicht: die Sperrdatei soll sich nicht anlegen lassen
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(wurzel, 0o750) }) //nolint:errcheck,gosec // Aufräumen

	err := service.NeuePaketwerkstatt(wurzel).Entfernen(t.Context(), "probe.mse-style", "bilder/rahmen.png", "")
	if err == nil {
		t.Fatal("entfernt ohne Sperre")
	}

	if _, statfehler := os.Stat(filepath.Join(wurzel, "probe.mse-style", "bilder", "rahmen.png")); statfehler != nil {
		t.Errorf("die Datei ist fort: %v", statfehler)
	}
}
