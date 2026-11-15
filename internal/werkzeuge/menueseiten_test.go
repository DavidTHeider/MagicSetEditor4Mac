package werkzeuge_test

// **Jeder Seitenschlüssel im Menü führt irgendwohin** — Werkstück 4 von
// Phase 3.7
// (#1168,
// Befund Z4-01).
//
// # Was dieser Wächter hält
//
// Die Menüleiste entsteht in Go und nennt Seiten mit ihrem Schlüssel
// (`seitenruf(handlungen.SeiteOeffnen, "ersetzen")`). Welche Schlüssel es
// gibt, sagt `frontend/src/navigation.ts`. **Zwei Listen, eine Wahrheit** —
// und ein Tippfehler wäre ein Menüeintrag, der ins Leere führt, ohne dass
// irgendetwas rot würde: das Fenster bekäme ein Ereignis mit einem Schlüssel,
// den sein `switch` nicht kennt, und täte nichts.
//
// # Was er nicht hält
//
// **Dass die Seite auch erreichbar ist.** Eine Setlasche ohne offenes Set
// führt nirgendwohin, und das ist richtig so; der Wächter prüft die Namen,
// nicht die Lage.

import (
	"os"
	"regexp"
	"testing"
)

// menueseiten findet die Schlüssel, die `menue.go` an `seitenruf` gibt.
var menueseiten = regexp.MustCompile(`seitenruf\(handlungen\.SeiteOeffnen,\s*"([a-z]+)"\)`)

// navigationsschluessel findet die Einträge der beiden Listen in
// `navigation.ts` — sie stehen dort als `'name',` in `as const`-Feldern.
var navigationsschluessel = regexp.MustCompile(`(?m)^\s*'([a-z]+)',`)

func TestJederSeitenschluesselImMenueFuehrtIrgendwohin(t *testing.T) {
	t.Parallel()

	genannt := treffer(t, "../app/menue.go", menueseiten)
	if len(genannt) == 0 {
		t.Fatal("die Menüleiste nennt keine Seite — der Wächter prüft nichts")
	}

	bekannt := map[string]bool{}
	for _, schluessel := range treffer(t, "../../frontend/src/navigation.ts",
		navigationsschluessel) {
		bekannt[schluessel] = true
	}

	if len(bekannt) < 10 {
		t.Fatalf("%d Seitenschlüssel gelesen — das Muster greift nicht mehr",
			len(bekannt))
	}

	for _, schluessel := range genannt {
		if !bekannt[schluessel] {
			t.Errorf("die Menüleiste führt auf %q — `navigation.ts` kennt es nicht",
				schluessel)
		}
	}
}

// treffer liest eine Datei und gibt die erste Gruppe jedes Treffers.
func treffer(t *testing.T, pfad string, muster *regexp.Regexp) []string {
	t.Helper()

	roh, err := os.ReadFile(pfad) //nolint:gosec // ein fester Pfad im Repositorium
	if err != nil {
		t.Fatalf("%s lesen: %v", pfad, err)
	}

	treffer := muster.FindAllStringSubmatch(string(roh), -1)
	gefunden := make([]string, 0, len(treffer))

	for _, einer := range treffer {
		gefunden = append(gefunden, einer[1])
	}

	return gefunden
}
