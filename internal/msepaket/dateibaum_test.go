// **Dateien im Paket anlegen, umbenennen und entfernen** — FMT-02.AK-01
// (#1691).
//
// MSE2 hat keinen Paketeditor und damit keine dieser Operationen (gesucht in
// `src/util/io/package.cpp` und `src/gui/`, Fassung f9d9356 vom 2021-07-06);
// die Regeln hier sind die des Schreibwegs: dieselbe Sperre, der Stand wird
// geprüft, und was verschwindet, liegt danach in der Sicherung.
package msepaket_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func lies(t *testing.T, pfad string) string {
	t.Helper()

	inhalt, err := os.ReadFile(pfad) //nolint:gosec // Pfad aus dem Test
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	return string(inhalt)
}

func TestEineDateiWirdAngelegt(t *testing.T) {
	t.Parallel()

	ort := probeverzeichnis(t)

	if err := msepaket.DateiAnlegen(t.Context(), ort, "neu/symbole", []byte("# leer\n")); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if got := lies(t, filepath.Join(ort, "neu", "symbole")); got != "# leer\n" {
		t.Errorf("Inhalt %q", got)
	}

	// **Keine Nachbardatei bleibt liegen.**
	eintraege, err := os.ReadDir(filepath.Join(ort, "neu"))
	if err != nil || len(eintraege) != 1 {
		t.Errorf("%d Einträge in neu/, erwartet einer", len(eintraege))
	}
}

func TestEineVorhandeneDateiWirdNichtUeberschrieben(t *testing.T) {
	t.Parallel()

	ort := probeverzeichnis(t)

	err := msepaket.DateiAnlegen(t.Context(), ort, "unterordner/skript", []byte("anders"))
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("Fehler %v, erwartet ErrConflict", err)
	}

	if got := lies(t, filepath.Join(ort, "unterordner", "skript")); got != "1 + 1\n" {
		t.Errorf("überschrieben: %q", got)
	}
}

func TestEineDateiWirdUmbenannt(t *testing.T) {
	t.Parallel()

	ort := probeverzeichnis(t)

	err := msepaket.DateiUmbenennen(t.Context(), ort, "unterordner/skript", "skripte/eins", summe([]byte("1 + 1\n")))
	if err != nil {
		t.Fatalf("umbenennen: %v", err)
	}

	if got := lies(t, filepath.Join(ort, "skripte", "eins")); got != "1 + 1\n" {
		t.Errorf("Inhalt %q", got)
	}

	if _, err := os.Stat(filepath.Join(ort, "unterordner", "skript")); !os.IsNotExist(err) {
		t.Errorf("die alte Datei steht noch: %v", err)
	}
}

func TestEineDateiWirdEntferntUndLiegtInDerSicherung(t *testing.T) {
	t.Parallel()

	ort := probeverzeichnis(t)

	if err := msepaket.DateiEntfernen(t.Context(), ort, "unterordner/skript", summe([]byte("1 + 1\n"))); err != nil {
		t.Fatalf("entfernen: %v", err)
	}

	if _, err := os.Stat(filepath.Join(ort, "unterordner", "skript")); !os.IsNotExist(err) {
		t.Errorf("die Datei steht noch: %v", err)
	}

	if got := lies(t, filepath.Join(ort+msepaket.Sicherungsendung, "unterordner", "skript")); got != "1 + 1\n" {
		t.Errorf("Sicherung %q", got)
	}
}

// **Ohne Stand geht es bei Beiwerk** — ein Bild liest die Werkstatt nicht,
// also hat sie keinen Stand dafür; die Sicherung macht es trotzdem umkehrbar.
func TestBeiwerkGehtOhneStand(t *testing.T) {
	t.Parallel()

	ort := probeverzeichnis(t)
	if err := os.WriteFile(filepath.Join(ort, "rahmen.png"), []byte{0x89, 'P', 'N', 'G'}, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := msepaket.DateiUmbenennen(t.Context(), ort, "rahmen.png", "bilder/rahmen.png", ""); err != nil {
		t.Fatalf("umbenennen: %v", err)
	}

	if err := msepaket.DateiEntfernen(t.Context(), ort, "bilder/rahmen.png", ""); err != nil {
		t.Fatalf("entfernen: %v", err)
	}

	if _, err := os.Stat(filepath.Join(ort+msepaket.Sicherungsendung, "bilder", "rahmen.png")); err != nil {
		t.Errorf("keine Sicherung: %v", err)
	}
}

func TestDieDateioperationenWeisenAb(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		lauf func(ort string) error
		will error
	}{
		"ein veralteter Stand beim Umbenennen": {
			func(ort string) error {
				return msepaket.DateiUmbenennen(t.Context(), ort, "unterordner/skript", "b", summe([]byte("alt")))
			}, domain.ErrConflict,
		},
		"ein veralteter Stand beim Entfernen": {
			func(ort string) error {
				return msepaket.DateiEntfernen(t.Context(), ort, "unterordner/skript", summe([]byte("alt")))
			}, domain.ErrConflict,
		},
		"ein belegter neuer Name": {
			func(ort string) error {
				return msepaket.DateiUmbenennen(t.Context(), ort, "unterordner/skript", "style", summe([]byte("1 + 1\n")))
			}, domain.ErrConflict,
		},
		"eine fehlende Datei beim Umbenennen": {
			func(ort string) error { return msepaket.DateiUmbenennen(t.Context(), ort, "fehlt", "b", "") },
			domain.ErrNotFound,
		},
		"eine fehlende Datei beim Entfernen": {
			func(ort string) error { return msepaket.DateiEntfernen(t.Context(), ort, "fehlt", "") },
			domain.ErrNotFound,
		},
		"die Hauptdatei entfernen": {
			func(ort string) error {
				return msepaket.DateiEntfernen(t.Context(), ort, "style", summe([]byte("mse version: 2.0.0\ngame: probe\n")))
			}, domain.ErrValidation,
		},
		"die Hauptdatei umbenennen": {
			func(ort string) error {
				return msepaket.DateiUmbenennen(t.Context(), ort, "style", "stil", summe([]byte("mse version: 2.0.0\ngame: probe\n")))
			}, domain.ErrValidation,
		},
		"ein Ausbruch beim Anlegen": {
			func(ort string) error { return msepaket.DateiAnlegen(t.Context(), ort, "../draussen", nil) },
			msepaket.ErrAusbruch,
		},
		"ein Ausbruch beim neuen Namen": {
			func(ort string) error {
				return msepaket.DateiUmbenennen(t.Context(), ort, "unterordner/skript", "../x", summe([]byte("1 + 1\n")))
			}, msepaket.ErrAusbruch,
		},
		"ein Ausbruch beim Entfernen": {
			func(ort string) error { return msepaket.DateiEntfernen(t.Context(), ort, "../x", "") },
			msepaket.ErrAusbruch,
		},
		"ein Ordner ist keine Datei": {
			func(ort string) error { return msepaket.DateiEntfernen(t.Context(), ort, "unterordner", "") },
			domain.ErrValidation,
		},
	} {
		ort := probeverzeichnis(t)

		if err := fall.lauf(ort); !errors.Is(err, fall.will) {
			t.Errorf("%s: %v, erwartet %v", name, err, fall.will)
		}

		// **Und nichts ist geschehen.**
		if got := lies(t, filepath.Join(ort, "unterordner", "skript")); got != "1 + 1\n" {
			t.Errorf("%s: skript ist %q", name, got)
		}
	}
}

// **Ein gepacktes Paket wird nicht umgebaut** — es wäre ein neues Archiv je
// Handgriff. Wer daran arbeiten will, legt eine eigene Kopie an; sie ist ein
// Verzeichnis.
func TestEinArchivpaketWeistDieDateioperationenAb(t *testing.T) {
	t.Parallel()

	ort := probearchiv(t)

	for name, err := range map[string]error{
		"anlegen":    msepaket.DateiAnlegen(t.Context(), ort, "neu", nil),
		"umbenennen": msepaket.DateiUmbenennen(t.Context(), ort, "a", "b", ""),
		"entfernen":  msepaket.DateiEntfernen(t.Context(), ort, "a", ""),
		"ohne Paket": msepaket.DateiAnlegen(t.Context(), filepath.Join(t.TempDir(), "fehlt.mse-style"), "neu", nil),
	} {
		if name == "ohne Paket" {
			if !errors.Is(err, domain.ErrNotFound) {
				t.Errorf("%s: %v, erwartet ErrNotFound", name, err)
			}

			continue
		}

		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("%s: %v, erwartet ErrValidation", name, err)
		}
	}
}
