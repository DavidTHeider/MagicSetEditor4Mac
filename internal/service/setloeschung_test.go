// **Ein Set wegnehmen, mit seinen Bildern** —
// #1218.

package service_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// loeschdoppel merkt sich, womit es gerufen wurde.
type loeschdoppel struct {
	setID   int64
	name    string
	gerufen int
	fehler  error
}

func (l *loeschdoppel) SetLoeschen(_ context.Context, setID int64, name string) error {
	l.setID, l.name, l.gerufen = setID, name, l.gerufen+1

	return l.fehler
}

// bilderMit legt ein Bildverzeichnis für ein Set an und gibt Ablage und Ort.
func bilderMit(t *testing.T, setID int64) (ablage *service.Bildablage, ort string) {
	t.Helper()

	ablage = service.NeueBildablage(t.TempDir())
	ort = ablage.Verzeichnis(setID)

	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("das Bildverzeichnis anlegen: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ort, "image1"), []byte("PNG"), 0o600); err != nil {
		t.Fatalf("ein Bild ablegen: %v", err)
	}

	return ablage, ort
}

func TestSetloeschungNimmtDieBilderMit(t *testing.T) {
	doppel := &loeschdoppel{}
	ablage, ort := bilderMit(t, 7)

	befunde, err := service.NeueSetloeschung(doppel).MitBildablage(ablage).
		Loeschen(t.Context(), 7, "Mein Set")
	if err != nil {
		t.Fatalf("löschen: %v", err)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde, wo keine sein sollten: %v", befunde)
	}
	if doppel.setID != 7 || doppel.name != "Mein Set" {
		t.Errorf("der Speicher bekam (%d, %q)", doppel.setID, doppel.name)
	}
	if _, statfehler := os.Stat(ort); !errors.Is(statfehler, os.ErrNotExist) {
		t.Errorf("das Bildverzeichnis steht noch: %v", statfehler)
	}
}

// TestSetloeschungOhneBilderablage ist der Normalfall der Nähte ohne Ablage.
func TestSetloeschungOhneBilderablage(t *testing.T) {
	doppel := &loeschdoppel{}

	if _, err := service.NeueSetloeschung(doppel).Loeschen(t.Context(), 7, "Mein Set"); err != nil {
		t.Fatalf("löschen: %v", err)
	}
	if doppel.gerufen != 1 {
		t.Errorf("der Speicher wurde %d mal gerufen", doppel.gerufen)
	}
}

// TestSetloeschungLaesstDieBilderStehenWennNichtsGeloescht wurde ist §7 Regel 3
// rückwärts: was nicht geschrieben wurde, wird auch nicht nachgeräumt.
func TestSetloeschungLaesstDieBilderStehenWennNichtsGeloescht(t *testing.T) {
	doppel := &loeschdoppel{fehler: domain.NewFieldError(
		"name", domain.SchluesselKonflikt, "der Name stimmt nicht")}
	ablage, ort := bilderMit(t, 7)

	_, err := service.NeueSetloeschung(doppel).MitBildablage(ablage).
		Loeschen(t.Context(), 7, "Falsch")
	if err == nil {
		t.Fatal("der Fehler des Speichers kam nicht durch")
	}

	if _, statfehler := os.Stat(ort); statfehler != nil {
		t.Errorf("die Bilder sind weg, obwohl das Set steht: %v", statfehler)
	}
}

// TestSetloeschungMeldetDieBilderAlsBefund ist §7 Regel 3.
//
// **Geschrieben ist geschrieben.** Das Set ist weg; dass die Bilder daneben
// nicht wegzuräumen waren, ist ein Befund und keine Rücknahme. Wer hier einen
// Fehler zurückgäbe, ließe die Oberfläche ein Set anzeigen, das es nicht mehr
// gibt.
func TestSetloeschungMeldetDieBilderAlsBefund(t *testing.T) {
	doppel := &loeschdoppel{}
	ablage, ort := bilderMit(t, 7)

	// **Ein Verzeichnis, das sich nicht räumen lässt**: ohne Schreibrecht am
	// Elternverzeichnis kann `RemoveAll` den Eintrag nicht lösen.
	eltern := filepath.Dir(ort)

	// #nosec G302 -- ein Verzeichnis ohne Schreibrecht ist genau der Fall
	if err := os.Chmod(eltern, 0o500); err != nil {
		t.Fatalf("die Rechte setzen: %v", err)
	}

	t.Cleanup(func() {
		// #nosec G302 -- zurück auf schreibbar, damit `t.TempDir` aufräumen kann
		if err := os.Chmod(eltern, 0o700); err != nil {
			t.Errorf("die Rechte zurücksetzen: %v", err)
		}
	})

	befunde, err := service.NeueSetloeschung(doppel).MitBildablage(ablage).
		Loeschen(t.Context(), 7, "Mein Set")
	if err != nil {
		t.Fatalf("das Löschen ist am Aufräumen gescheitert: %v", err)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet war einer: %v", len(befunde), befunde)
	}
	if doppel.gerufen != 1 {
		t.Errorf("der Speicher wurde %d mal gerufen", doppel.gerufen)
	}
}
