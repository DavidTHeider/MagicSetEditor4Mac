package werkzeuge_test

// **Die Grenzen einer Spaltenbreite stehen zweimal da, und gleich** — AP-12
// (#2369).
//
// `domain.BreiteMindestens` und `domain.BreiteHoechstens` prüfen, was in
// `arbeitsplatz.navigatorbreite` und `…inspektorbreite` stehen darf; der
// Trenner im Fenster (`frontend/src/bausteine/Arbeitsspalten.tsx`) zieht
// innerhalb derselben Grenzen. Eine weitere Grenze im Fenster ließe den
// Trenner auf eine Breite gehen, die der Bestand beim Schreiben abweist —
// die Spalte spränge dann nach dem Loslassen zurück.
//
// # Was er nicht hält
//
// **Die obere Grenze aus dem Platz im Fenster** (480 Punkte für das
// Zentrum). Die rechnet das Fenster selbst; hier stehen zwei Zahlenpaare
// gegeneinander.

import (
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

var spaltengrenze = regexp.MustCompile(`(?m)^export const breite(Mindestens|Hoechstens) = (\d+)$`)

func TestDieGrenzenDerSpaltenbreiteStehenInGoUndImFensterGleich(t *testing.T) {
	t.Parallel()

	quelle, fehler := os.ReadFile("../../frontend/src/bausteine/Arbeitsspalten.tsx")
	if fehler != nil {
		t.Fatal(fehler)
	}

	soll := map[string]int{
		"Mindestens": domain.BreiteMindestens,
		"Hoechstens": domain.BreiteHoechstens,
	}
	gefunden := map[string]bool{}

	for _, treffer := range spaltengrenze.FindAllSubmatch(quelle, -1) {
		name := string(treffer[1])

		zahl, umwandlung := strconv.Atoi(string(treffer[2]))
		if umwandlung != nil {
			t.Fatal(umwandlung)
		}

		gefunden[name] = true

		if zahl != soll[name] {
			t.Errorf("das Fenster hat breite%s = %d, der Bestand %d", name, zahl, soll[name])
		}
	}

	for name := range soll {
		if !gefunden[name] {
			t.Errorf("breite%s steht nicht mehr in Arbeitsspalten.tsx — der Leser greift daneben", name)
		}
	}
}
