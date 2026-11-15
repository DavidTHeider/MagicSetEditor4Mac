package service_test

// **Der ganze Weg zum Ordner, nicht nur sein letzter Teil** — Nachlauf über
// Phase 5 von 0.9.0, Befund 3 (SEC02, #2355). `MkdirAll` folgt Verweisen im
// Elternpfad, und `Lstat` sah nur die letzte Komponente: bei `eltern →
// fremd` entstand der Ordner im fremden Ziel, und erst danach wurde geprüft.
// Wer einen Ordner auf dem Weg verändern kann, kann den Ordner tauschen —
// deshalb zählt, **wer** dort schreiben darf, und das **vor** dem Anlegen.

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// offenerOrdner legt einen Ordner an, in dem jeder schreiben darf.
func offenerOrdner(t *testing.T, pfad string, sticky bool) {
	t.Helper()

	if err := os.Mkdir(pfad, 0o700); err != nil {
		t.Fatal(err)
	}

	rechte := os.FileMode(0o777)
	if sticky {
		rechte |= os.ModeSticky
	}

	if err := os.Chmod(pfad, rechte); err != nil { //nolint:gosec // G302: der Fall, gegen den geprüft wird
		t.Fatal(err)
	}
}

func TestEinPrivaterOrdnerPrueftDenGanzenWeg(t *testing.T) {
	t.Parallel()

	t.Run("lehnt einen Vorfahren ab, in dem andere schreiben dürfen — und legt nichts an", func(t *testing.T) {
		t.Parallel()

		offen := filepath.Join(t.TempDir(), "offen")
		offenerOrdner(t, offen, false)

		ordner := filepath.Join(offen, "schnittstellen")
		if fehler := service.PrivaterOrdner(ordner); !errors.Is(fehler, domain.ErrConflict) {
			t.Errorf("Fehler %v, erwartet ErrConflict", fehler)
		}

		if _, fehler := os.Lstat(ordner); !errors.Is(fehler, os.ErrNotExist) {
			t.Errorf("angelegt wurde trotzdem: %v", fehler)
		}
	})

	t.Run("lehnt einen Verweis im Elternpfad auf einen offenen Ordner ab, bevor er etwas anlegt", func(t *testing.T) {
		t.Parallel()

		wurzel := t.TempDir()
		fremd := filepath.Join(wurzel, "fremd")
		offenerOrdner(t, fremd, false)

		eltern := filepath.Join(wurzel, "eltern")
		if err := os.Symlink(fremd, eltern); err != nil {
			t.Fatal(err)
		}

		if fehler := service.PrivaterOrdner(filepath.Join(eltern, "schnittstellen")); !errors.Is(fehler, domain.ErrConflict) {
			t.Errorf("Fehler %v, erwartet ErrConflict", fehler)
		}

		if _, fehler := os.Lstat(filepath.Join(fremd, "schnittstellen")); !errors.Is(fehler, os.ErrNotExist) {
			t.Errorf("im fremden Ziel angelegt: %v", fehler)
		}
	})

	t.Run("nimmt einen eigenen Verweis im Elternpfad", func(t *testing.T) {
		t.Parallel()

		wurzel := t.TempDir()
		eigen := filepath.Join(wurzel, "eigen")

		if err := os.Mkdir(eigen, 0o700); err != nil {
			t.Fatal(err)
		}

		eltern := filepath.Join(wurzel, "eltern")
		if err := os.Symlink(eigen, eltern); err != nil {
			t.Fatal(err)
		}

		if fehler := service.PrivaterOrdner(filepath.Join(eltern, "schnittstellen")); fehler != nil {
			t.Errorf("ein eigener Verweis wurde abgewiesen: %v", fehler)
		}

		if angaben, fehler := os.Lstat(filepath.Join(eigen, "schnittstellen")); fehler != nil || angaben.Mode().Perm() != 0o700 {
			t.Errorf("im eigenen Ziel: %v, %v", angaben, fehler)
		}
	})

	t.Run("nimmt einen Vorfahren mit Sticky-Bit wie /tmp", func(t *testing.T) {
		t.Parallel()

		klebend := filepath.Join(t.TempDir(), "klebend")
		offenerOrdner(t, klebend, true)

		if fehler := service.PrivaterOrdner(filepath.Join(klebend, "schnittstellen")); fehler != nil {
			t.Errorf("ein Ordner wie /tmp wurde abgewiesen: %v", fehler)
		}
	})

	t.Run("legt mehrere fehlende Ebenen an und prüft den Vorfahren, den es gibt", func(t *testing.T) {
		t.Parallel()

		ordner := filepath.Join(t.TempDir(), "a", "b", "schnittstellen")
		if fehler := service.PrivaterOrdner(ordner); fehler != nil {
			t.Fatalf("drei Ebenen: %v", fehler)
		}

		if angaben, fehler := os.Lstat(ordner); fehler != nil || angaben.Mode().Perm() != 0o700 {
			t.Errorf("angelegt: %v, %v", angaben, fehler)
		}
	})

	t.Run("nimmt einen relativen Pfad", func(t *testing.T) {
		t.Parallel()

		ordner, fehler := filepath.Rel(mussArbeitsordner(t), filepath.Join(t.TempDir(), "relativ"))
		if fehler != nil {
			t.Fatal(fehler)
		}

		if fehler := service.PrivaterOrdner(ordner); fehler != nil {
			t.Errorf("relativ: %v", fehler)
		}
	})
}

func TestEinPrivaterOrdnerSagtWarumNicht(t *testing.T) {
	t.Parallel()

	t.Run("ein Vorfahr, den niemand betreten darf", func(t *testing.T) {
		t.Parallel()

		zu := filepath.Join(t.TempDir(), "zu")
		if err := os.Mkdir(zu, 0o700); err != nil {
			t.Fatal(err)
		}

		if err := os.Chmod(zu, 0); err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() { _ = os.Chmod(zu, 0o700) }) //nolint:errcheck,gosec // damit TempDir aufräumen kann

		if fehler := service.PrivaterOrdner(filepath.Join(zu, "a", "b")); fehler == nil {
			t.Error("hinter einem gesperrten Ordner angelegt")
		}
	})

	t.Run("ein Verweis ins Leere", func(t *testing.T) {
		t.Parallel()

		eltern := filepath.Join(t.TempDir(), "eltern")
		if err := os.Symlink(filepath.Join(filepath.Dir(eltern), "fort"), eltern); err != nil {
			t.Fatal(err)
		}

		if fehler := service.PrivaterOrdner(filepath.Join(eltern, "schnittstellen")); fehler == nil {
			t.Error("durch einen Verweis ins Leere angelegt")
		}
	})

	// **Ein Ordner, der einem anderen gehört**, wird nicht verändert — der
	// Wurzelordner ist der eine, den es auf jedem Rechner gibt.
	t.Run("ein Ordner eines anderen Besitzers", func(t *testing.T) {
		t.Parallel()

		if fehler := service.PrivaterOrdner("/"); !errors.Is(fehler, domain.ErrConflict) {
			t.Errorf("Fehler %v, erwartet ErrConflict", fehler)
		}
	})
}

func mussArbeitsordner(t *testing.T) string {
	t.Helper()

	ordner, fehler := os.Getwd()
	if fehler != nil {
		t.Fatal(fehler)
	}

	return ordner
}
