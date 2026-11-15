package schriften_test

// **Schriften in Unterordnern der Schriftorte** —
// #2411.
//
// macOS aktiviert eine Schrift in einem Unterordner von `~/Library/Fonts`
// (gemessen am 2026-10-15 in einem Gast, `schriftordner-27.0.1.md`). Wer den
// Ordner `Magic - Fonts` des Packs als Ordner dorthin zieht, sieht die
// Schriften in jedem Programm — und diese Anwendung zeichnete die Karte ohne
// Text, weil sie die Orte nur flach las.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
)

// legeIn legt die Systemschrift unter `name` in einen Unterordner von `ort`.
func legeIn(t *testing.T, ort, unter, name string) string {
	t.Helper()

	ordner := filepath.Join(ort, unter)
	if err := os.MkdirAll(ordner, 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(ordner, name), systemschrift(t), 0o600); err != nil {
		t.Fatal(err)
	}

	return ordner
}

func TestEineSchriftImUnterordnerWirdGefunden(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	ordner := legeIn(t, ort, "Magic - Fonts", "MPlantin.ttf")
	suche := schriften.Neue([]string{ort})

	if _, err := suche.Suche("MPlantin", false); err != nil {
		t.Fatalf("über den Dateinamen: %v", err)
	}

	if fundort, err := suche.Fundort("MPlantin", false); err != nil || filepath.Dir(fundort) != ordner {
		t.Errorf("gefunden in %q (%v), erwartet %q", fundort, err, ordner)
	}
}

// **Auch über den Namen, den die Schrift über sich selbst führt** — so findet
// `magic-m15` seine `Beleren Bold` in `beleren-bold_P1.01.ttf`.
func TestEineSchriftImUnterordnerWirdUeberIhrenEigenenNamenGefunden(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	ordner := filepath.Join(ort, "Magic - Fonts")
	if err := os.MkdirAll(ordner, 0o700); err != nil {
		t.Fatal(err)
	}

	legeBenannt(t, ordner, "probe-unten_P1.ttf", map[uint16]string{1: "Probe Unten", 2: "Regular", 4: "Probe Unten"})

	if _, err := schriften.Neue([]string{ort}).Suche("Probe Unten", false); err != nil {
		t.Errorf("über den eigenen Namen: %v", err)
	}
}

// **Was oben liegt, geht vor** — auch vor einem Unterordner, dessen Name
// vorher sortiert: der Ort selbst ist, was jemand dorthin gelegt hat.
func TestDieSchriftImOrtGehtDemUnterordnerVor(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	legeIn(t, ort, "A Fonts", "MPlantin.ttf")
	legeIn(t, ort, ".", "MPlantin.ttf")

	if fundort, err := schriften.Neue([]string{ort}).Fundort("MPlantin", false); err != nil || filepath.Dir(fundort) != ort {
		t.Errorf("gefunden in %q (%v), erwartet der Ort selbst", fundort, err)
	}
}

// **Mit Grenze**: vier Ebenen unter dem Ort, nicht tiefer — ein Ort wird bei
// jedem ersten Fehlschlag gelesen.
func TestTieferAlsVierEbenenWirdNichtGesucht(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	legeIn(t, ort, filepath.Join("a", "b", "c", "d"), "Vier.ttf")
	legeIn(t, ort, filepath.Join("a", "b", "c", "d", "e"), "Fuenf.ttf")
	suche := schriften.Neue([]string{ort})

	if _, err := suche.Suche("Vier", false); err != nil {
		t.Errorf("vier Ebenen tief: %v", err)
	}

	if _, err := suche.Suche("Fuenf", false); err == nil {
		t.Error("fünf Ebenen tief gefunden")
	}
}

// **Ein Verweis auf eine Schriftdatei wird gefunden**, wie in der flachen
// Suche davor — und einem Verweis auf einen Ordner wird nicht gefolgt: ein
// Ring aus Verweisen ginge sonst nie zu Ende.
func TestVerweiseAufDateienJaAufOrdnerNein(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	anderswo := legeIn(t, t.TempDir(), "echt", "Verwiesen.ttf")

	if err := os.Symlink(filepath.Join(anderswo, "Verwiesen.ttf"), filepath.Join(ort, "Verwiesen.ttf")); err != nil {
		t.Fatal(err)
	}

	legeIn(t, anderswo, "drin", "Hinter.ttf")

	if err := os.Symlink(anderswo, filepath.Join(ort, "Ordnerverweis")); err != nil {
		t.Fatal(err)
	}

	suche := schriften.Neue([]string{ort})

	if _, err := suche.Suche("Verwiesen", false); err != nil {
		t.Errorf("über einen Verweis auf die Datei: %v", err)
	}

	if _, err := suche.Suche("Hinter", false); err == nil {
		t.Error("hinter einem Verweis auf einen Ordner gefunden")
	}
}
