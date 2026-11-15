package service_test

// **Ein Installer zählt seine Pakete und lässt sich anhalten** —
// FMT-03.AK-02 und AK-04
// (#1692).

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func zweiPaketeImInstaller(t *testing.T) string {
	t.Helper()

	return installerdatei(t, map[string]string{
		"installer":                   "package:\n\tname: probe-a.mse-include\npackage:\n\tname: probe-b.mse-include\n",
		"probe-a.mse-include/include": "mse version: 2.0.0\n",
		"probe-b.mse-include/include": "mse version: 2.0.0\n",
	})
}

// paketzaehler merkt sich, was gemeldet wurde, und bricht bei `bei` ab.
func paketzaehler(t *testing.T, bei int) (ctx context.Context, erledigt *[]int) {
	t.Helper()

	ctx, abbrechen := context.WithCancel(t.Context())
	t.Cleanup(abbrechen)

	erledigt = &[]int{}

	return service.MitFortschritt(ctx, func(stand domain.Fortschritt) {
		if stand.Gesamt != 2 || stand.Einheit != domain.EinheitPaket {
			t.Errorf("Stand: %+v", stand)
		}

		*erledigt = append(*erledigt, stand.Erledigt)

		if stand.Erledigt == bei {
			abbrechen()
		}
	}), erledigt
}

func TestEinInstallerZaehltSeinePakete(t *testing.T) {
	_, verwalter, _ := kopierplatz(t)
	ctx, erledigt := paketzaehler(t, -1)

	if _, err := verwalter.Einlesen(ctx, zweiPaketeImInstaller(t)); err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(*erledigt, []int{0, 1, 2}) {
		t.Errorf("erledigt: %v", *erledigt)
	}
}

// **Ein Abbruch zwischen zwei Paketen lässt das erste ganz und vermerkt** —
// ein Paket im Verzeichnis ohne Herkunft wäre ein halbes (AK-04).
func TestEinAbgebrochenerInstallerVermerktWasDaliegt(t *testing.T) {
	ort, verwalter, _ := kopierplatz(t)
	ctx, _ := paketzaehler(t, 1)

	ergebnis, err := verwalter.Einlesen(ctx, zweiPaketeImInstaller(t))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Fehler: %v", err)
	}

	if ergebnis == nil || len(ergebnis.Uebernommen) != 1 ||
		ergebnis.Uebernommen[0].Herkunft == "" {
		t.Fatalf("Ergebnis: %+v", ergebnis)
	}

	if _, statfehler := os.Stat(filepath.Join(ort, "probe-b.mse-include")); !errors.Is(statfehler, os.ErrNotExist) {
		t.Errorf("das zweite Paket liegt da: %v", statfehler)
	}
}

func TestDieWeitergabeZaehltIhrePakete(t *testing.T) {
	_, verwalter, _ := kopierplatz(t)
	ctx, erledigt := paketzaehler(t, -1)

	if _, err := verwalter.Weitergeben(ctx,
		[]string{"probe.mse-game", "probe-schlicht.mse-style"}, t.TempDir()); err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(*erledigt, []int{0, 1, 2}) {
		t.Errorf("erledigt: %v", *erledigt)
	}
}

// **Eine abgebrochene Weitergabe schreibt keinen Installer** — er entsteht
// erst, wenn alle Pakete gelesen sind.
func TestEineAbgebrocheneWeitergabeSchreibtNichts(t *testing.T) {
	_, verwalter, _ := kopierplatz(t)
	ctx, _ := paketzaehler(t, 2)
	ablage := t.TempDir()

	if _, err := verwalter.Weitergeben(ctx,
		[]string{"probe.mse-game", "probe-schlicht.mse-style"}, ablage); !errors.Is(err, context.Canceled) {
		t.Fatalf("Fehler: %v", err)
	}

	if eintraege, _ := os.ReadDir(ablage); len(eintraege) != 0 { //nolint:errcheck // leer heißt: nichts geschrieben
		t.Errorf("liegt da: %v", eintraege)
	}
}
