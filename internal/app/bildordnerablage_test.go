package app_test

// **Der Ordnerdialog der Bildausgabe geht im zuletzt gewählten Ordner auf** —
// AP-10 (#2367),
// **ported** aus MSE2: `wxFileSelector` beginnt bei `settings.default_export_dir`
// (`src/gui/images_export_window.cpp`, Zeile 63, Fassung f9d9356 vom 2021-07-06).
//
// **Ohne gemerkten Ordner beim Ordner fürs Hinausgeben**, nicht bei dem fürs
// Hereinholen: Bilder gehen hinaus. Bis AP-10 galt der zweite, weil der Dialog
// kein Sicherndialog ist — eine benannte Änderung.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// bildordnerStart gibt den Ordner, in dem der Dialog der Bildausgabe aufging.
func bildordnerStart(t *testing.T, anwendung *app.App) string {
	t.Helper()

	gesehen := mitgeschrieben(anwendung)
	app.NeueBindungen(anwendung).Bildordner(t.Context())

	if len(*gesehen) != 1 {
		t.Fatalf("%d Dialoge statt eines", len(*gesehen))
	}

	return (*gesehen)[0].Startverzeichnis
}

func TestDerBildordnerdialogBeginntBeimGemerktenOrdner(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	gemerkt, ausfuhr := t.TempDir(), t.TempDir()

	setzeOrdner(t, anwendung, domain.VorgabeAusfuhrpfad, ausfuhr)
	setzeOrdner(t, anwendung, domain.EinstellungBildordner, gemerkt)

	if start := bildordnerStart(t, anwendung); start != gemerkt {
		t.Errorf("der Dialog begann bei %q statt %q", start, gemerkt)
	}
}

func TestOhneGemerktenOrdnerGiltDerFuersHinausgeben(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	ausfuhr, herein := t.TempDir(), t.TempDir()

	setzeOrdner(t, anwendung, domain.VorgabeHochladepfad, herein)
	setzeOrdner(t, anwendung, domain.VorgabeAusfuhrpfad, ausfuhr)

	if start := bildordnerStart(t, anwendung); start != ausfuhr {
		t.Errorf("der Dialog begann bei %q statt %q", start, ausfuhr)
	}
}

// **Und nie beim Ordner fürs Hereinholen** — das war der Weg vor AP-10.
func TestDerBildordnerdialogNimmtNichtDenFuersHereinholen(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	setzeOrdner(t, anwendung, domain.VorgabeHochladepfad, t.TempDir())

	if start := bildordnerStart(t, anwendung); start != "" {
		t.Errorf("der Dialog begann bei %q", start)
	}
}

// **Ein verschwundener Ordner hält die Wahl nicht an** — sie ist ein Öffnen
// und kein Sichern: wer ein Laufwerk abgezogen hat, wählt einen anderen.
func TestEinVerschwundenerBildordnerLaesstDieWahlZu(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	ordner := filepath.Join(t.TempDir(), "platte")

	if fehler := os.Mkdir(ordner, 0o750); fehler != nil {
		t.Fatal(fehler)
	}

	setzeOrdner(t, anwendung, domain.EinstellungBildordner, ordner)

	if fehler := os.Remove(ordner); fehler != nil {
		t.Fatal(fehler)
	}

	if start := bildordnerStart(t, anwendung); start != "" {
		t.Errorf("der Dialog begann bei %q", start)
	}
}
