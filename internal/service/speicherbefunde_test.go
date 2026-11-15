package service_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Was beim Beiwerk schiefgeht, hält das Speichern nicht auf — es wird
// benannt.**
//
// # Warum das eine eigene Datei ist
//
// Die Zusage in `speichern_test.go` lautet: die Bilder kommen mit. Hier steht
// die andere Hälfte, und sie ist die, die man vergisst: **was tut das
// Speichern, wenn eines davon nicht geht?** Die Hauptdatei steht in der
// Datenbank und kommt heraus; alles daneben liegt auf einer Platte, die voll
// sein kann, deren Rechte fremd sind und in der schon etwas anderes liegt.
//
// Ein Speichern, das daran scheitert, verlöre die Änderungen; eines, das
// schweigt, verlöre die Bilder unbemerkt (§8).
func TestBeiwerkbefunde(t *testing.T) {
	faelle := []struct {
		name string

		// stellen richtet Herkunft und Ziel so ein, dass **eine** Stelle
		// scheitert, und gibt beide zurück.
		stellen func(t *testing.T) (herkunft, ziel string)
	}{
		{
			name: "eine unlesbare Datei",
			stellen: func(t *testing.T) (string, string) {
				t.Helper()

				herkunft := herkunftspaket(t)
				unlesbar(t, filepath.Join(herkunft, "image1"))

				return herkunft, filepath.Join(t.TempDir(), "woanders.mse-set")
			},
		},
		{
			name: "ein unlesbares Verzeichnis",
			stellen: func(t *testing.T) (string, string) {
				t.Helper()

				herkunft := herkunftspaket(t)
				verzeichnis := filepath.Join(herkunft, "bilder")

				if err := os.MkdirAll(verzeichnis, 0o750); err != nil {
					t.Fatalf("anlegen: %v", err)
				}

				unlesbar(t, verzeichnis)

				return herkunft, filepath.Join(t.TempDir(), "woanders.mse-set")
			},
		},
		{
			name: "ein Ziel, in dem eine Datei im Weg steht",
			stellen: func(t *testing.T) (string, string) {
				t.Helper()

				// Im Herkunftspaket ist `bilder` ein Verzeichnis, im Ziel liegt
				// eine gleichnamige **Datei**: das Verzeichnis lässt sich dort
				// nicht anlegen.
				herkunft := herkunftspaket(t)
				lege(t, filepath.Join(herkunft, "bilder", "eins.png"), "\x89PNG")

				ziel := filepath.Join(t.TempDir(), "woanders.mse-set")
				lege(t, filepath.Join(ziel, "bilder"), "keine Verzeichnis")

				return herkunft, ziel
			},
		},
		{
			name: "ein Ziel, in dem ein Verzeichnis im Weg steht",
			stellen: func(t *testing.T) (string, string) {
				t.Helper()

				// Umgekehrt: `image1` ist im Ziel ein Verzeichnis, und in ein
				// Verzeichnis lässt sich keine Datei schreiben.
				ziel := filepath.Join(t.TempDir(), "woanders.mse-set")

				if err := os.MkdirAll(filepath.Join(ziel, "image1"), 0o750); err != nil {
					t.Fatalf("anlegen: %v", err)
				}

				return herkunftspaket(t), ziel
			},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			herkunft, ziel := fall.stellen(t)

			ergebnis, err := service.NeuerSetschreiber(mitHerkunft(t, herkunft), &ortsschreiber{}).
				Schreibe(t.Context(), 7, ziel)
			if err != nil {
				t.Fatalf("schreiben: %v", err)
			}

			if len(ergebnis.Befunde) == 0 {
				t.Error("kein Befund — der Verlust bliebe unsichtbar")
			}

			// **Und die Hauptdatei ist trotzdem da.** Das ist der Grund, warum
			// es ein Befund ist und kein Fehler.
			if _, err := os.Stat(filepath.Join(ziel, "set")); err != nil {
				t.Errorf("die Hauptdatei fehlt: %v", err)
			}
		})
	}
}

// unlesbar nimmt einem Eintrag jedes Recht — und gibt es zum Aufräumen zurück.
//
// **Ohne das Zurückgeben scheiterte `t.TempDir`** beim Wegräumen, und der Test
// wäre rot, ohne dass am Prüfling etwas falsch wäre.
func unlesbar(t *testing.T, pfad string) {
	t.Helper()

	if err := os.Chmod(pfad, 0o000); err != nil {
		t.Fatalf("die Rechte nehmen: %v", err)
	}

	t.Cleanup(func() {
		// 0700 und nicht 0600: bei einem **Verzeichnis** ist das x-Recht das,
		// woran `t.TempDir` sonst scheitert.
		if err := os.Chmod(pfad, 0o700); err != nil { // #nosec G302 -- siehe darüber
			t.Errorf("die Rechte zurückgeben: %v", err)
		}
	})
}

// lege schreibt eine Datei samt ihrer Verzeichnisse.
func lege(t *testing.T, pfad, inhalt string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(pfad), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(pfad, []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}
}
