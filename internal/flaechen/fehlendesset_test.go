// **Eine Setkennung, die auf nichts zeigt** —
// #1336.
//
// # Der Befund
//
// Fünf Wege gaben für ein Set, **das es nicht gibt**, dieselbe Antwort wie für
// ein leeres: Rückgabewert 0 und „nichts da". Wer ein Skript über die
// Setkennungen schreibt, merkte einen Tippfehler nie.
//
// **Die Fläche ohne Liste ist die, die es zeigt** — am Fenster kommt eine
// Setkennung immer aus einer Auswahl.
package flaechen_test

import (
	"strings"
	"testing"
)

// **Jeder der fünf Wege weist eine Kennung ab, die auf nichts zeigt.**
func TestJederSetwegWeistEineFehlendeKennungAb(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	// **Ein Set muss es geben**, sonst prüft der Lauf einen leeren Bestand und
	// nicht die Kennung.
	mussKommandozeile(t, umgebung, "set-neu", "Probeset", "probe", "probe-stil")

	for _, fall := range []struct {
		name      string
		argumente []string
	}{
		{name: "verweise", argumente: []string{"verweise", "999"}},
		{name: "vorbehalte", argumente: []string{"vorbehalte", "999"}},
		{name: "zuruecknehmen", argumente: []string{"zuruecknehmen", "999"}},
		{name: "wiederherstellen", argumente: []string{"wiederherstellen", "999"}},
		{name: "suche mit Set", argumente: []string{"suche", "Flamme", "--set=999"}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			aus, fehler, code := mitEingabeUndCode(t, umgebung, "", fall.argumente...)

			if code == 0 {
				t.Fatalf("Rückgabewert 0 für ein Set, das es nicht gibt: %s%s",
					aus, fehler)
			}

			// **Die Kennung steht in der Meldung** — sonst weiß niemand,
			// welche gemeint war.
			if !strings.Contains(fehler, "999") {
				t.Errorf("die Meldung nennt die Kennung nicht: %q", fehler)
			}
		})
	}
}

// **Und dieselben Wege antworten für ein Set, das es gibt.**
//
// Die Gegenprobe: ohne sie wäre eine Wache, die **jede** Kennung abweist,
// grün — und die fünf Wege wären kaputt statt streng.
func TestDieselbenWegeAntwortenFuerEinVorhandenesSet(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	mussKommandozeile(t, umgebung, "set-neu", "Probeset", "probe", "probe-stil")

	for _, argumente := range [][]string{
		{"verweise", "1"},
		{"vorbehalte", "1"},
		{"suche", "Flamme", "--set=1"},
	} {
		if _, fehler, code := mitEingabeUndCode(t, umgebung, "",
			argumente...); code != 0 {
			t.Errorf("%v: Rückgabewert %d (%s)", argumente, code, fehler)
		}
	}
}
