package app

// **Ein Bild als Symbol nachzeichnen — an der Naht** —
// #1340.

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// kreisdatei schreibt ein PNG mit einem gefüllten Kreis und gibt seinen Pfad.
func kreisdatei(t *testing.T) string {
	t.Helper()

	const mass = 80

	gemalt := image.NewRGBA(image.Rect(0, 0, mass, mass))
	mitte := float64(mass) / 2

	for y := range mass {
		for x := range mass {
			ton := uint8(255)
			if math.Hypot(float64(x)-mitte, float64(y)-mitte) <= float64(mass)*0.35 {
				ton = 0
			}

			gemalt.Set(x, y, color.RGBA{R: ton, G: ton, B: ton, A: 255})
		}
	}

	pfad := filepath.Join(t.TempDir(), "kreis.png")

	datei, err := os.Create(pfad) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		t.Fatalf("die Datei anlegen: %v", err)
	}

	if err := png.Encode(datei, gemalt); err != nil {
		t.Fatalf("das PNG schreiben: %v", err)
	}

	if err := datei.Close(); err != nil {
		t.Fatalf("die Datei schließen: %v", err)
	}

	return pfad
}

// TestEinPNGWirdEineZeichnung ist die Abnahme an der Naht.
func TestEinPNGWirdEineZeichnung(t *testing.T) {
	t.Parallel()

	antwort := (&App{}).SymbolNachzeichnen(t.Context(), kreisdatei(t))
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if len(antwort.Teile) != 1 {
		t.Fatalf("%d Teile, erwartet 1", len(antwort.Teile))
	}

	if len(antwort.Teile[0].Punkte) < 4 {
		t.Errorf("%d Punkte — das ist kein Kreis", len(antwort.Teile[0].Punkte))
	}
}

// **Eine Datei, die es nicht gibt, ist ein fachliches Nein — und nennt sich.**
func TestEineFehlendeBilddateiWirdBenannt(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "gibtesnicht.png")

	antwort := (&App{}).SymbolNachzeichnen(t.Context(), pfad)
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler für eine fehlende Datei")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "gibtesnicht.png") {
		t.Errorf("die Meldung nennt den Pfad nicht: %q", antwort.Fehler.Meldung)
	}
}

// **Eine Datei, die kein Bild ist, ebenso.**
func TestEineDateiOhneBildWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "kein-bild.png")
	if err := os.WriteFile(pfad, []byte("das ist Text"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if antwort := (&App{}).SymbolNachzeichnen(t.Context(), pfad); antwort.Fehler == nil {
		t.Error("kein Fehler für eine Datei, die kein Bild ist")
	}
}

// **Ohne Pfad gibt es nichts nachzuzeichnen.**
func TestOhnePfadWirdNichtNachgezeichnet(t *testing.T) {
	t.Parallel()

	if antwort := (&App{}).SymbolNachzeichnen(t.Context(), ""); antwort.Fehler == nil {
		t.Error("kein Fehler für einen leeren Pfad")
	}
}

// **Und die leere Liste ist `[]` und nicht `null`** — dieselbe Zusage wie
// überall (§3, `leerelistenlaufzeit_test.go`).
func TestEinFehlerGibtEineLeereTeileliste(t *testing.T) {
	t.Parallel()

	antwort := (&App{}).SymbolNachzeichnen(t.Context(), "")
	if antwort.Teile == nil {
		t.Error("Teile ist null statt []")
	}
}
