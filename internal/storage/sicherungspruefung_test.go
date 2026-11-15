// **Was eine Sicherung erlebt, bevor sie übernommen wird** (#940).
//
// Die Prüfung ist der Punkt: eine Datei, die erst beim Öffnen auffällt, hat
// den alten Stand schon ersetzt. Deshalb wird sie **vorher** gelesen — als
// eigene Verbindung, ohne Migration, ohne Schreiben.
package storage_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// sicherungsdatei legt eine migrierte Datenbank als Datei an und gibt ihren Pfad.
func sicherungsdatei(t *testing.T) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "sicherung.db")

	db, err := storage.Oeffnen(probekontext(t), "file:"+pfad)
	if err != nil {
		t.Fatalf("die Sicherung ließ sich nicht anlegen: %v", err)
	}

	if err := storage.Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("die Migration schlug fehl: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("die Sicherung ließ sich nicht schließen: %v", err)
	}

	return pfad
}

func TestPruefeSicherungNimmtEineGueltigeDatei(t *testing.T) {
	t.Parallel()

	befund, err := storage.PruefeSicherung(probekontext(t), sicherungsdatei(t))
	if err != nil {
		t.Fatalf("eine frisch migrierte Datenbank sollte durchgehen: %v", err)
	}

	if befund.AngelegtAm == "" {
		t.Error("der Befund nennt nicht, wann der Bestand angelegt wurde")
	}

	if befund.Schemastand == "" {
		t.Error("der Befund nennt nicht, bis zu welcher Migration die Datei reicht")
	}
}

func TestPruefeSicherungWeistEineNeuereFassungBenanntAb(t *testing.T) {
	t.Parallel()

	pfad := sicherungsdatei(t)

	// **Eine Migration, die es hier nicht gibt** — genau das, was eine
	// Sicherung aus einer neueren Fassung mitbrächte.
	zukunftEintragen(t, pfad)

	_, err := storage.PruefeSicherung(probekontext(t), pfad)
	if !errors.Is(err, storage.ErrMigrationen) {
		t.Fatalf("erwartet war ErrMigrationen, bekommen: %v", err)
	}

	// **Benannt heißt benannt.** Eine Abweisung, die nicht sagt, *was* sie
	// nicht kennt, ist für die Empfangenden dasselbe wie ein Absturz.
	if !strings.Contains(err.Error(), "9999_aus_der_zukunft.sql") {
		t.Errorf("die Meldung nennt die unbekannte Migration nicht: %v", err)
	}
}

func TestPruefeSicherungWeistFremdeDateienAb(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(t *testing.T) string{
		"eine Datei, die es nicht gibt": func(t *testing.T) string {
			t.Helper()

			return filepath.Join(t.TempDir(), "fehlt.db")
		},
		"eine Datei, die keine Datenbank ist": func(t *testing.T) string {
			t.Helper()

			pfad := filepath.Join(t.TempDir(), "kein.db")
			if err := os.WriteFile(pfad, []byte("nur Text"), 0o600); err != nil {
				t.Fatalf("schreiben: %v", err)
			}

			return pfad
		},
		"eine Datenbank ohne Bestand": func(t *testing.T) string {
			t.Helper()

			pfad := filepath.Join(t.TempDir(), "leer.db")

			db, err := storage.Oeffnen(probekontext(t), "file:"+pfad)
			if err != nil {
				t.Fatalf("öffnen: %v", err)
			}

			if err := db.Close(); err != nil {
				t.Fatalf("schließen: %v", err)
			}

			return pfad
		},
	}

	for name, bauen := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := storage.PruefeSicherung(probekontext(t), bauen(t)); err == nil {
				t.Fatal("erwartet war eine Abweisung, bekommen: nil")
			}
		})
	}
}

// zukunftEintragen verbucht in einer Datei eine Migration aus einer neueren
// Fassung.
//
// **Eingetragen und nicht nachgestellt.** Eine Attrappe, die „aus der Zukunft"
// behauptet, prüfte den Aufrufer; hier steht die Zeile, die eine echte neuere
// Fassung hinterließe.
func zukunftEintragen(t *testing.T, pfad string) {
	t.Helper()

	db, err := storage.FuerLaeufe("file:" + pfad)
	if err != nil {
		t.Fatalf("die Sicherung ließ sich nicht öffnen: %v", err)
	}

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO migrationen (name, angewendet, pruefsumme)
		 VALUES ('9999_aus_der_zukunft.sql', '2026-08-23T00:00:00.000Z', 'x')`); fehler != nil {
		t.Fatalf("die künftige Migration ließ sich nicht eintragen: %v", fehler)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}
}
