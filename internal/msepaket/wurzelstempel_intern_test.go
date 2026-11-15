package msepaket

// **Die Änderungszeit allein trägt das Gemerkte nicht** — Befund 2 des
// Zweitblicks über Phase 0.7.0-I zu PERF-02
// (#2169).
//
// Zwei Wege, auf denen ein Ordner sich ändert und seine Änderungszeit nicht:
// ein Umbenennen, nach dem jemand die Zeit zurücksetzt (`os.Chtimes`, `touch
// -r`, ein Auspacker, der Zeiten bewahrt), und ein entzogenes Leserecht. In
// beiden gab die Wurzelsuche das alte Ergebnis, während eine frische Suche
// nichts mehr fand.

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestDasGemerkteFolgtJederAenderungAmOrdner(t *testing.T) {
	t.Parallel()

	for name, aendern := range map[string]func(t *testing.T, daten string){
		"umbenannt, die Zeit zurückgesetzt": func(t *testing.T, daten string) {
			t.Helper()

			vorher, fehler := os.Stat(daten)
			if fehler != nil {
				t.Fatal(fehler)
			}

			if fehler := os.Rename(filepath.Join(daten, "a.mse-style"), filepath.Join(daten, "a.normal")); fehler != nil {
				t.Fatal(fehler)
			}

			if fehler := os.Chtimes(daten, vorher.ModTime(), vorher.ModTime()); fehler != nil {
				t.Fatal(fehler)
			}
		},
		"das Leserecht entzogen": func(t *testing.T, daten string) {
			t.Helper()

			if fehler := os.Chmod(daten, 0o000); fehler != nil {
				t.Fatal(fehler)
			}

			// **Zurück auf 0o700**, sonst räumt `t.TempDir` den Ordner nicht:
			// ohne `x` lässt er sich nicht betreten.
			t.Cleanup(func() { _ = os.Chmod(daten, 0o700) }) //nolint:errcheck,gosec // aufräumen, was der Test verschlossen hat
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wurzel := t.TempDir()
			daten := filepath.Join(wurzel, "pack", "data")
			paketOrdner(t, filepath.Join(daten, "a.mse-style"))

			suche := NeueWurzelsuche()
			if gefunden := suche.Paketwurzeln(wurzel); !slices.Equal(gefunden, []string{daten}) {
				t.Fatalf("zuerst: %v", gefunden)
			}

			spaeter()
			aendern(t, daten)

			frisch := Paketwurzeln(wurzel)
			if gemerkt := suche.Paketwurzeln(wurzel); !slices.Equal(gemerkt, frisch) {
				t.Errorf("gemerkt %v, frisch %v", gemerkt, frisch)
			}
		})
	}
}
