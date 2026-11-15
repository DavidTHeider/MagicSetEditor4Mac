// **Die Fehlerwege der Dateioperationen, an der echten Platte** — FMT-02
// (#1691).
// Was sich nur über ersetzte Griffe erreichen lässt, steht in
// `dateibaumgriffe_intern_test.go`.
package msepaket_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// eng macht einen Pfad unbeschreibbar, bis der Test endet.
func eng(t *testing.T, pfad string, rechte os.FileMode) {
	t.Helper()

	if err := os.Chmod(pfad, rechte); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(pfad, 0o750) }) //nolint:errcheck,gosec // Aufräumen
}

func TestDieDateioperationenScheiternAnDerPlatte(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]func(t *testing.T, ort string) error{
		"ein Ordnerteil ist eine Datei": func(t *testing.T, ort string) error {
			return msepaket.DateiAnlegen(t.Context(), ort, "style/neu", nil)
		},
		"der neue Ordner lässt sich nicht anlegen": func(t *testing.T, ort string) error {
			eng(t, ort, 0o500)

			return msepaket.DateiAnlegen(t.Context(), ort, "neu/datei", nil)
		},
		"der Zielordner beim Umbenennen lässt sich nicht anlegen": func(t *testing.T, ort string) error {
			eng(t, ort, 0o500)

			return msepaket.DateiUmbenennen(t.Context(), ort, "unterordner/skript", "neu/skript", "")
		},
		"der Rückweg lässt sich nicht anlegen": func(t *testing.T, ort string) error {
			if err := os.WriteFile(ort+msepaket.Sicherungsendung, nil, 0o600); err != nil {
				t.Fatal(err)
			}

			return msepaket.DateiEntfernen(t.Context(), ort, "unterordner/skript", "")
		},
		"die Datei lässt sich nicht lesen": func(t *testing.T, ort string) error {
			eng(t, filepath.Join(ort, "unterordner", "skript"), 0o000)

			return msepaket.DateiEntfernen(t.Context(), ort, "unterordner/skript", summe([]byte("1 + 1\n")))
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := fall(t, probeverzeichnis(t)); err == nil {
				t.Fatal("kein Fehler")
			}
		})
	}
}

// **Wer die Sperre nicht bekommt, wird abgewiesen** — wie beim Schreiben
// (`TestEineGehalteneSperreWeistDenZweitenAb`), und aus demselben Grund
// nicht parallel: die Frist ist eine Paketvariable.
func TestEineGehalteneSperreWeistDieDateioperationenAb(t *testing.T) {
	ort := probeverzeichnis(t)

	griff, err := dateisperre.Oeffnen(ort + msepaket.Schreibsperrendung)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = griff.Schliessen() }) //nolint:errcheck // Aufräumen

	if fehler := griff.Allein(); fehler != nil {
		t.Fatal(fehler)
	}

	frist, takt := msepaket.Schreibfrist, msepaket.Schreibtakt
	msepaket.Schreibfrist, msepaket.Schreibtakt = 40*time.Millisecond, 10*time.Millisecond

	t.Cleanup(func() { msepaket.Schreibfrist, msepaket.Schreibtakt = frist, takt })

	if fehler := msepaket.DateiAnlegen(t.Context(), ort, "neu", nil); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet ErrConflict", fehler)
	}
}
