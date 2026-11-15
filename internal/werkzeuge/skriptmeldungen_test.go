// **Jede Meldung des Zerteilers steht in beiden Katalogen** —
// #1178.
//
// # Der Befund
//
// Seit #1164
// erreicht der Satz des Zerteilers das Fenster — **einsprachig**: `skriptFehler`
// trug ihn als Platzhalter, und er kam aus `internal/msescript`.
//
// # Was dieser Wächter hält
//
// Beide Richtungen, wie der Vertragstest über die Fehlerschlüssel:
//
//   - was Go setzt, steht in beiden Katalogen,
//   - was in einem Katalog steht, setzt Go.
//
// # Was er nicht hält
//
// **Ob der englische Satz dasselbe sagt.** Das sieht kein Wächter; er zählt
// Schlüssel. Was ihn hält, ist das Lesen — dieselbe Grenze wie bei jedem
// Katalogwächter dieses Verzeichnisses.
package werkzeuge_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// skriptkatalog findet die Einträge im `skript:`-Block eines Fehlerkatalogs.
var skriptkatalog = regexp.MustCompile(`(?m)^ {4}([a-zA-Z]+):`)

// skriptschluesselAus liest die Schlüssel aus einem Katalog.
func skriptschluesselAus(t *testing.T, datei string) map[string]bool {
	t.Helper()

	inhalt := lies(t, datei)

	anfang := strings.Index(inhalt, "  skript: {")
	if anfang < 0 {
		t.Fatalf("%s führt keinen Block »skript:«", datei)
	}

	rest := inhalt[anfang:]

	ende := strings.Index(rest, "\n  },")
	if ende < 0 {
		t.Fatalf("%s: der Block »skript:« ist nicht geschlossen", datei)
	}

	gefunden := map[string]bool{}
	for _, treffer := range skriptkatalog.FindAllStringSubmatch(rest[:ende], -1) {
		gefunden[treffer[1]] = true
	}

	return gefunden
}

func TestJedeSkriptmeldungStehtInBeidenKatalogen(t *testing.T) {
	t.Parallel()

	kataloge := map[string]map[string]bool{
		"de": skriptschluesselAus(t, "frontend/src/sprache/de/fehler.ts"),
		"en": skriptschluesselAus(t, "frontend/src/sprache/en/fehler.ts"),
	}

	// **Was Go setzt, steht in beiden Katalogen.**
	for _, schluessel := range msescript.Meldungsschluesselliste() {
		for sprache, katalog := range kataloge {
			if !katalog[string(schluessel)] {
				t.Errorf("der Katalog %q kennt die Meldung %q nicht", sprache, schluessel)
			}
		}
	}

	// **Und was im Katalog steht, setzt Go.** Ein Eintrag, den niemand
	// benutzt, ist ein Satz, den niemand liest — und er veraltet still.
	gesetzt := map[string]bool{}
	for _, schluessel := range msescript.Meldungsschluesselliste() {
		gesetzt[string(schluessel)] = true
	}

	for sprache, katalog := range kataloge {
		for schluessel := range katalog {
			if !gesetzt[schluessel] {
				t.Errorf("der Katalog %q führt die Meldung %q — Go setzt sie nicht",
					sprache, schluessel)
			}
		}
	}
}

// **Und die Zahl der Platzhalter stimmt zwischen Go und beiden Katalogen.**
//
// `nach {0} fehlt {1}` braucht zwei Werte; ein Katalogsatz mit nur einem
// zeigte den zweiten nie — und niemand merkte es, weil der Text lesbar bleibt.
func TestDieZahlDerPlatzhalterStimmt(t *testing.T) {
	t.Parallel()

	platzhalter := regexp.MustCompile(`\{(\d+)\}`)

	for _, datei := range []string{
		"frontend/src/sprache/de/fehler.ts",
		"frontend/src/sprache/en/fehler.ts",
	} {
		inhalt := lies(t, datei)
		anfang := strings.Index(inhalt, "  skript: {")
		rest := inhalt[anfang : anfang+strings.Index(inhalt[anfang:], "\n  },")]

		for _, zeile := range strings.Split(rest, "\n") {
			schluessel, satz, ist := strings.Cut(strings.TrimSpace(zeile), ": ")
			if !ist || !strings.HasPrefix(satz, "'") {
				continue
			}

			hoechste := -1
			for _, fund := range platzhalter.FindAllStringSubmatch(satz, -1) {
				if fund[1] > string(rune('0'+hoechste)) {
					hoechste = int(fund[1][0] - '0')
				}
			}

			if erwartet := msescript.PlatzhalterZahl(
				msescript.Meldungsschluessel(schluessel),
			); erwartet != hoechste+1 {
				t.Errorf("%s: %q hat %d Platzhalter, Go setzt %d",
					datei, schluessel, hoechste+1, erwartet)
			}
		}
	}
}
