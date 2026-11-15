package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Fehlerwege der Beigaben** —
// #657.
//
// Sie stehen hier und nicht am Weg durch die Fassade, weil sie von dort nur
// über eine Platte zu erreichen wären, die gerade voll ist oder unter der
// jemand ein Verzeichnis wegzieht. Ein ungeprüfter Fehlerzweig ist eine
// Behauptung darüber, was im Ernstfall geschieht (§3).

// setmitschrift ist ein Speicherer, der scheitern kann.
// setmitschrift steht für einen Setschreiber, der glückt oder scheitert.
//
// **Er legt das Verzeichnis wirklich an**, seit `LegeSet` daraus ein Archiv
// packt (Werkstück C von Phase 4.9). Vorher tat er gar nichts — eine
// Vorrichtung, die großzügiger ist als der Betrieb (§3): ein geglücktes
// Schreiben hinterlässt ein Verzeichnis mit einer Hauptdatei, und wer das
// nicht nachstellt, prüft einen Zustand, den es nicht gibt.
type setmitschrift struct{ fehler error }

func (s setmitschrift) Schreibe(
	_ context.Context, _ int64, ziel string,
) (*service.Speicherung, error) {
	if s.fehler != nil {
		return nil, s.fehler
	}

	if err := os.MkdirAll(ziel, 0o750); err != nil {
		return nil, err
	}

	if err := os.WriteFile(filepath.Join(ziel, "set"),
		[]byte("mse version: 2.0.0\n"), 0o600); err != nil {
		return nil, err
	}

	return &service.Speicherung{}, nil
}

func TestDieBeigabenMeldenIhreFehler(t *testing.T) {
	t.Parallel()

	t.Run("ein Vorlagenpaket, das es nicht gibt", func(t *testing.T) {
		t.Parallel()

		beigabe := beigaben{vorlagenpaket: filepath.Join(t.TempDir(), "weg.mse-export-template")}

		if _, err := beigabe.AusVorlage("x.txt"); err == nil {
			t.Error("das fehlende Paket blieb ohne Fehler")
		}
	})

	t.Run("ein Ziel, das es nicht gibt", func(t *testing.T) {
		t.Parallel()

		beigabe := beigaben{ziel: filepath.Join(t.TempDir(), "gibtsnicht")}

		if _, err := beigabe.Lege("x.txt", []byte("x")); err == nil {
			t.Error("das fehlende Verzeichnis blieb ohne Fehler")
		}
	})

	t.Run("ein Setname, der ein Pfad ist", func(t *testing.T) {
		t.Parallel()

		beigabe := beigaben{ziel: t.TempDir(), setschreibt: setmitschrift{}}

		if _, err := beigabe.LegeSet(probekontext(t), "../weg.mse-set"); err == nil {
			t.Error("der Pfad ging durch")
		}
	})

	t.Run("ein Set, das sich nicht schreiben lässt", func(t *testing.T) {
		t.Parallel()

		beigabe := beigaben{
			ziel:        t.TempDir(),
			setschreibt: setmitschrift{fehler: errors.New("voll")},
		}

		if _, err := beigabe.LegeSet(probekontext(t), "kopie.mse-set"); err == nil {
			t.Error("der Schreibfehler blieb liegen")
		}
	})

	t.Run("und ein Set, das sich schreiben lässt", func(t *testing.T) {
		t.Parallel()

		beigabe := beigaben{ziel: t.TempDir(), setschreibt: setmitschrift{}}

		name, err := beigabe.LegeSet(probekontext(t), "kopie.mse-set")
		if err != nil || name != "kopie.mse-set" {
			t.Errorf("Name %q, Fehler %v", name, err)
		}
	})
}
