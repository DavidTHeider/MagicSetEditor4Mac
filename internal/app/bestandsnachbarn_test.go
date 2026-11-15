// **Was dem Bestand gehört, folgt ihm — was dem Rechner gehört, nicht** —
// QA-10.NAT-01 (#1741).
//
// # Warum es diesen Test gibt
//
// **Bis hierher folgten dem Schalter `MSE_BESTAND` Datenbank und Protokoll —
// die eingesetzten Bilder nicht.** Ein Bestand mit eigenem Ort legte seine
// Bilder in den Vorgabeort, neben einen anderen Bestand: eine Sichtprüfung
// schrieb sie in das echte Profil dieses Rechners, und zwei Instanzen teilten
// sich, was jeder für sich gehört.
//
// # Und warum Pakete nicht mitgehen
//
// **Weil sie dem Rechner gehören** (`docs/AUTORISIERUNG.md`), mit ihren
// Schriften und dem Zwischenspeicher von Scryfall. Ein Skript, das mit
// `MSE_BESTAND` an einem Wegwerfbestand übt, zeichnet mit den bezogenen Paketen
// dieses Rechners — genau dafür gibt es den Schalter (`bestandsort.go`). Wer
// auch sie trennen will, trennt das Zuhause; so baut das Labor jede Instanz.
package app_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestWasDemBestandGehoertFolgtIhm(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	ort := t.TempDir()
	t.Setenv(app.Bestandsumgebung, filepath.Join(ort, "bestand.db"))

	for name, frage := range map[string]func() (string, error){
		"die Bilder":     app.Bilderverzeichnis,
		"das Protokoll":  app.Protokollpfad,
		"der Ort selbst": app.BestandsortFuerFlaechen,
	} {
		pfad, fehler := frage()
		if fehler != nil {
			t.Fatalf("%s: %v", name, fehler)
		}

		if filepath.Dir(pfad) != ort {
			t.Errorf("%s liegen unter %q und nicht neben dem Bestand in %q", name, pfad, ort)
		}
	}
}

func TestWasDemRechnerGehoertBleibtBeimZuhause(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	t.Setenv(app.Bestandsumgebung, filepath.Join(t.TempDir(), "bestand.db"))

	for name, frage := range map[string]func() (string, error){
		"die Pakete":               app.Paketverzeichnis,
		"der Scryfall-Speicher":    app.Scryfallverzeichnis,
		"die Schriften der Pakete": func() (string, error) { return app.Schriftwurzel(), nil },
		"die Drossel":              func() (string, error) { return app.ScryfallAbspracheort(), nil },
	} {
		pfad, fehler := frage()
		if fehler != nil {
			t.Fatalf("%s: %v", name, fehler)
		}

		if !strings.HasPrefix(pfad, zuhause) {
			t.Errorf("%s liegen unter %q und nicht beim Rechner unter %q", name, pfad, zuhause)
		}
	}

	// **Und die Schriften kommen aus denselben Paketen wie die Stile** —
	// sonst zeichnete ein Stil von hier mit Schriften von dort.
	pakete, fehler := app.Paketverzeichnis()
	if fehler != nil {
		t.Fatal(fehler)
	}

	if schriften := app.Schriftwurzel(); schriften != pakete {
		t.Errorf("die Schriften kommen aus %q, die Pakete liegen in %q", schriften, pakete)
	}
}

// **Die Diagnose nennt den Bestand, der wirklich offen ist** — und nicht den
// Vorgabeort, an dem keiner liegt.
func TestDieDiagnoseNenntDenEigenenBestand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	eigener := filepath.Join(t.TempDir(), "bestand.db")
	t.Setenv(app.Bestandsumgebung, eigener)

	if text := app.NeueApp("").Diagnose(t.Context()).Text; !strings.Contains(text, eigener) {
		t.Errorf("die Diagnose nennt %q nicht:\n%s", eigener, text)
	}
}

// **Ohne Zuhause gibt es keine Paketschriften** — und keinen Absturz: die
// Schriftsuche fragt, sie richtet nichts ein.
func TestOhneZuhauseGibtEsKeineSchriftwurzel(t *testing.T) {
	t.Setenv(app.Bestandsumgebung, "")
	t.Setenv("HOME", "")

	if wurzel := app.Schriftwurzel(); wurzel != "" {
		t.Errorf("die Schriftwurzel ist %q, erwartet leer", wurzel)
	}
}
