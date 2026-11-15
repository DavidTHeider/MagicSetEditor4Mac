// **Weitergeben und Einlesen ohne Fenster** — FMT-02.AK-04
// (#1691):
// der Weg, den Kommandozeile, REST und MCP nehmen — ohne `service`-Typen.
package app_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestDiePaketnahtGibtWeiterUndLiestEin(t *testing.T) {
	weg, quelle := pflegeweg(t, "probe-weiter.mse-style")
	ctx := probekontext(t)

	if _, err := weg.Installieren(ctx, quelle); err != nil {
		t.Fatalf("installieren: %v", err)
	}

	datei, err := weg.Weitergeben(ctx, []string{"probe-weiter.mse-style"}, t.TempDir())
	if err != nil || filepath.Base(datei) != "probe-weiter.mse-installer" {
		t.Fatalf("weitergeben: %q, %v", datei, err)
	}

	// **Es liegt schon da** — übergangen, und keine Liste ist `nil`.
	nochmal, err := weg.Einlesen(ctx, datei)
	if err != nil || !slices.Equal(nochmal.Uebergangen, []string{"probe-weiter.mse-style"}) ||
		nochmal.Uebernommen == nil || nochmal.Befunde == nil {
		t.Fatalf("einlesen: %+v, %v", nochmal, err)
	}

	// **Auf einem zweiten Rechner** kommt es an.
	zweiter, _ := pflegeweg(t, "unbenutzt.mse-style")

	angekommen, err := zweiter.Einlesen(ctx, datei)
	if err != nil || len(angekommen.Uebernommen) != 1 || angekommen.Uebernommen[0].Name != "probe-weiter" {
		t.Errorf("angekommen: %+v, %v", angekommen, err)
	}
}

func TestDiePaketnahtNenntEinenUnlesbarenInstaller(t *testing.T) {
	weg, _ := pflegeweg(t, "unbenutzt.mse-style")

	kein := filepath.Join(t.TempDir(), "kein.mse-installer")
	if err := os.WriteFile(kein, []byte("kein Archiv"), 0o600); err != nil {
		t.Fatal(err)
	}

	if ergebnis, err := weg.Einlesen(probekontext(t), kein); err == nil || ergebnis != nil {
		t.Errorf("Ergebnis %+v, Fehler %v", ergebnis, err)
	}
}

// **Das Ziel, das MCP und REST vorher reservieren** — dieselbe Regel wie
// beim Schreiben: ein Ordner bekommt den Namen des ersten Pakets.
func TestDasInstallerzielNenntDieDatei(t *testing.T) {
	ordner := t.TempDir()

	if pfad, err := app.Installerziel([]string{"probe.mse-game"}, ordner); err != nil ||
		pfad != filepath.Join(ordner, "probe.mse-installer") {
		t.Errorf("Ziel %q, %v", pfad, err)
	}

	if _, err := app.Installerziel(nil, ordner); err == nil {
		t.Error("ein Ziel ohne Paket")
	}
}
