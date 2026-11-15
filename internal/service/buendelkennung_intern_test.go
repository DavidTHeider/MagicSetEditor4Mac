package service

// **Die eigenen Kennungen stehen zweimal da** — im Bau und hier, für den
// Nachweis, dass ein altes Bündel das eigene war (Nachlauf über Phase 6.12,
// N3). Eine Zusage in zwei Dateien läuft auseinander (§12); dieser Wächter
// liest beide Baustellen und verlangt jede ihrer Kennungen hier.

import (
	"os"
	"regexp"
	"testing"
)

func TestJedeKennungDesBausGiltAlsEigene(t *testing.T) {
	t.Parallel()

	plist, fehler := os.ReadFile("../../build/darwin/Info.plist")
	if fehler != nil {
		t.Fatal(fehler)
	}

	bauskript, fehler := os.ReadFile("../../scripts/bauen.sh")
	if fehler != nil {
		t.Fatal(fehler)
	}

	gefunden := map[string][][]byte{
		"build/darwin/Info.plist": buendelkennung.FindSubmatch(plist),
		"scripts/bauen.sh": regexp.MustCompile(`CFBundleIdentifier -string "([^"]+)"`).
			FindSubmatch(bauskript),
	}

	for ort, treffer := range gefunden {
		if treffer == nil {
			t.Errorf("%s nennt keine Kennung — der Wächter liest an der falschen Stelle", ort)

			continue
		}

		if !eigeneKennungen[string(treffer[1])] {
			t.Errorf("%s nennt %q, und die gilt hier nicht als eigene", ort, treffer[1])
		}
	}
}
