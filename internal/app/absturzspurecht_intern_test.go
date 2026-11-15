package app

// **Die Spur in den Gestalten, die Go wirklich schreibt** — Befund 5 des
// Nachlaufs über Phase 1 von 0.9.0 (F04, #2343). Die Dateien unter
// `testdaten/absturzspuren/` hat Go 1.27.2 geschrieben, je eine aus einem Fall
// von `erzeuger.go.txt`; in jedem freien Wert steht `hunter2` oder `4711`.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDieSpurFuerDasPaketInDenGestaltenVonGo(t *testing.T) {
	t.Parallel()

	for name, unveraendert := range map[string]bool{
		"mehrzeilig":        false, // eine Leerzeile im Wert, ein falscher Kopf, ein falsches Signal
		"laufzeitgestalt":   false, // ein Fehler, der mit „runtime error:" beginnt
		"verschachtelt":     false, // zwei Paniken, die zweite mit „fatal error:" im Wert
		"wiederaufgenommen": false,
		"beimschreiben":     false, // `String` wirft beim Schreiben des Werts
		"benannt":           false, // ein benannter Texttyp
		"zahl":              false,
		"index":             true,
		"nilmap":            true,
		"nilzeiger":         true,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			//nolint:gosec // ein fester Ort im Baum; der Name kommt aus diesem Test
			roh, err := os.ReadFile(filepath.Join("..", "..", "testdaten", "absturzspuren", name+".txt"))
			if err != nil {
				t.Fatal(err)
			}

			gezeigt := spurFuerDasPaket(string(roh))

			if unveraendert {
				if gezeigt != string(roh) {
					t.Errorf("ein Laufzeitfehler bleibt, wie er ist:\n%s", gezeigt)
				}

				return
			}

			for _, geheim := range []string{"hunter2", "4711"} {
				if strings.Contains(gezeigt, geheim) {
					t.Errorf("%q steht noch da:\n%s", geheim, gezeigt)
				}
			}

			// **Was bleibt**: gleich viele Zeilen, und ab dem ersten Kopf einer
			// Goroutine jede, wie sie war.
			vorher, nachher := strings.Split(string(roh), "\n"), strings.Split(gezeigt, "\n")
			if len(vorher) != len(nachher) {
				t.Fatalf("%d Zeilen statt %d", len(nachher), len(vorher))
			}

			ab := 0
			for ab < len(vorher) && !strings.HasPrefix(vorher[ab], "goroutine ") {
				ab++
			}

			if ab == len(vorher) {
				t.Fatal("die Spur hat keinen Kopf einer Goroutine")
			}

			for i := ab; i < len(vorher); i++ {
				if nachher[i] != vorher[i] {
					t.Errorf("Zeile %d ging verloren: %q wurde %q", i+1, vorher[i], nachher[i])
				}
			}
		})
	}
}
