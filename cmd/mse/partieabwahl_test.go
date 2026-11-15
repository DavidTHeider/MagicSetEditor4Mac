package main

// **Die Form von `--ohne=`** — P1a
// (#2335).
// Ob ein Bereich gilt, prüft der Kern; dass dieselbe Abwahl an jeder Fläche
// dieselbe Partie gibt, hält `internal/flaechen/partieabwahl_test.go`.

import (
	"slices"
	"testing"
)

func TestTrenneAbwahl(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		rein               []string
		uebrig, abgewaehlt []string
	}{
		"ohne Schalter": {
			rein:   []string{"magic", "a:A:1", "b:B:2"},
			uebrig: []string{"magic", "a:A:1", "b:B:2"},
		},
		"ein Bereich": {
			rein:       []string{"magic", "--ohne=zugstruktur", "a:A:1"},
			uebrig:     []string{"magic", "a:A:1"},
			abgewaehlt: []string{"zugstruktur"},
		},
		"zwei Schalter und leere Teile": {
			rein:       []string{"--ohne=combat,, stapel ", "magic", "--ohne=", "--ohne=marken"},
			uebrig:     []string{"magic"},
			abgewaehlt: []string{"combat", "stapel", "marken"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			uebrig, abgewaehlt := trenneAbwahl(fall.rein)

			if !slices.Equal(uebrig, fall.uebrig) ||
				!slices.Equal(abgewaehlt, fall.abgewaehlt) {
				t.Errorf("übrig %q, abgewählt %q", uebrig, abgewaehlt)
			}
		})
	}
}
