package service_test

// **Pakete als `.mse-installer` weitergeben** — FMT-02.AK-04
// (#1691).

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestPaketeGehenAlsInstallerHinaus(t *testing.T) {
	_, verwalter, _ := kopierplatz(t)
	ablage := t.TempDir()

	// **Ein Ordner als Ziel bekommt den Namen des ersten Pakets** —
	// `prefered_filename` (`Installer::addPackage`, `src/data/installer.cpp`,
	// Zeilen 177–186, Fassung f9d9356 vom 2021-07-06).
	datei, err := verwalter.Weitergeben(t.Context(), []string{"probe.mse-game", "probe-schlicht.mse-style"}, ablage)
	if err != nil {
		t.Fatalf("weitergeben: %v", err)
	}

	if datei != filepath.Join(ablage, "probe.mse-installer") {
		t.Errorf("geschrieben nach %q", datei)
	}

	installer, err := msepaket.OeffneInstaller(datei)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = installer.Schliessen() }() //nolint:errcheck // nur gelesen

	namen := make([]string, 0, len(installer.Pakete))
	for _, paket := range installer.Pakete {
		namen = append(namen, paket.Name)
	}

	if !slices.Equal(namen, []string{"probe.mse-game", "probe-schlicht.mse-style"}) {
		t.Errorf("im Installer: %v", namen)
	}
}

// **Ein genannter Dateiname gilt, und ein vorhandener wird ersetzt** — wie
// `saveAs` bei MSE2; gefragt hat vorher der Dialog.
func TestEinGenannterInstallerWirdErsetzt(t *testing.T) {
	_, verwalter, _ := kopierplatz(t)
	ziel := filepath.Join(t.TempDir(), "meins.mse-installer")

	if err := os.WriteFile(ziel, []byte("alt"), 0o600); err != nil {
		t.Fatal(err)
	}

	if datei, err := verwalter.Weitergeben(t.Context(), []string{"probe.mse-game"}, ziel); err != nil || datei != ziel {
		t.Fatalf("weitergeben: %q, %v", datei, err)
	}

	installer, err := msepaket.OeffneInstaller(ziel)
	if err != nil {
		t.Fatalf("der Installer ist keiner: %v", err)
	}

	_ = installer.Schliessen() //nolint:errcheck // nur gelesen
}

func TestEineWeitergabeOhneTaugendeAngabenWirdAbgewiesen(t *testing.T) {
	_, verwalter, _ := kopierplatz(t)
	ablage := t.TempDir()

	faelle := []struct {
		name   string
		pakete []string
		ziel   string
		feld   string
	}{
		{"ohne Paket", nil, ablage, "pakete"},
		{"ohne Ziel", []string{"probe.mse-game"}, " ", "ziel"},
		{"ein Ziel ohne Endung", []string{"probe.mse-game"}, filepath.Join(ablage, "meins.zip"), "ziel"},
		{"ein Name ohne Endung", []string{"probe"}, ablage, "pakete"},
		{"ein Set", []string{"probe.mse-set"}, ablage, "pakete"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, err := verwalter.Weitergeben(t.Context(), fall.pakete, fall.ziel)

			var feldfehler *domain.FieldError
			if !errors.As(err, &feldfehler) || feldfehler.Field != fall.feld {
				t.Errorf("Fehler %v, erwartet einen am Feld %q", err, fall.feld)
			}
		})
	}
}

func TestEinUnbekanntesPaketGehtNichtHinaus(t *testing.T) {
	_, verwalter, _ := kopierplatz(t)
	ablage := t.TempDir()

	if _, err := verwalter.Weitergeben(t.Context(), []string{"gibtsnicht.mse-game"}, ablage); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", err)
	}

	if eintraege, _ := os.ReadDir(ablage); len(eintraege) != 0 { //nolint:errcheck // leer ist leer
		t.Errorf("es blieb etwas liegen: %v", eintraege)
	}
}

func TestEinFehlenderOrdnerIstEinFehler(t *testing.T) {
	_, verwalter, _ := kopierplatz(t)
	ziel := filepath.Join(t.TempDir(), "fehlt", "meins.mse-installer")

	if _, err := verwalter.Weitergeben(t.Context(), []string{"probe.mse-game"}, ziel); err == nil {
		t.Error("in einen fehlenden Ordner wurde geschrieben")
	}
}

func TestEinAbbruchHaeltDieWeitergabe(t *testing.T) {
	_, verwalter, _ := kopierplatz(t)

	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if _, err := verwalter.Weitergeben(ctx, []string{"probe.mse-game"}, t.TempDir()); !errors.Is(err, context.Canceled) {
		t.Errorf("Fehler %v, erwartet den Abbruch", err)
	}
}

// **Ein Ziel, das sich nicht ansehen lässt**, ist ein Fehler mit Grund —
// und kein Versuch, hineinzuschreiben.
func TestEinVerschlossenesZielWirdGemeldet(t *testing.T) {
	_, verwalter, _ := kopierplatz(t)

	verschlossen := filepath.Join(t.TempDir(), "zu")
	if err := os.Mkdir(verschlossen, 0o000); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(verschlossen, 0o700) }) //nolint:errcheck,gosec // nur für das Aufräumen

	_, err := verwalter.Weitergeben(t.Context(), []string{"probe.mse-game"}, filepath.Join(verschlossen, "x.mse-installer"))
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet eine Abweisung", err)
	}
}
