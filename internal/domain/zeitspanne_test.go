package domain_test

import (
	"math"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Anzahl mal eine Einheit läuft über, ohne es zu sagen** — J05
// (#2269), Befund 2 des nachgeholten Zweitblicks über F03: 5.124.096 Stunden
// sind in Go rund 25 Minuten, weil `time.Duration` bei 2⁶³ − 1 Nanosekunden
// endet und die Multiplikation umläuft.
func TestZeitspanne(t *testing.T) {
	for _, fall := range []struct {
		name    string
		anzahl  int64
		einheit time.Duration
		dauer   time.Duration
		passt   bool
	}{
		{"eine Stunde", 1, time.Hour, time.Hour, true},
		{"null ist null", 0, time.Second, 0, true},
		{"die längste Zahl ganzer Stunden", 2562047, time.Hour, 2562047 * time.Hour, true},
		{"eine Stunde mehr läuft über", 2562048, time.Hour, 0, false},
		{"Stunden aus dem Befund", 5124096, time.Hour, 0, false},
		{"Sekunden über die Grenze", 9223372037, time.Second, 0, false},
		{"die größte Anzahl", math.MaxInt64, time.Nanosecond, math.MaxInt64, true},
		{"negativ ist keine Spanne", -1, time.Second, 0, false},
		{"eine Einheit von null ist keine", 3, 0, 0, false},
	} {
		t.Run(fall.name, func(t *testing.T) {
			dauer, passt := domain.Zeitspanne(fall.anzahl, fall.einheit)
			if dauer != fall.dauer || passt != fall.passt {
				t.Errorf("Zeitspanne(%d, %v) = %v, %v — erwartet %v, %v",
					fall.anzahl, fall.einheit, dauer, passt, fall.dauer, fall.passt)
			}
		})
	}
}

// Höchstens ist die sättigende Lesart: eine Angabe, die nicht hineinpasst,
// heißt „länger als alles, was sich ausdrücken lässt", nicht „kurz".
func TestZeitspanneHoechstens(t *testing.T) {
	if got := domain.ZeitspanneHoechstens(9223372037, time.Second); got != math.MaxInt64 {
		t.Errorf("über der Grenze: %v, erwartet die längste Dauer", got)
	}

	if got := domain.ZeitspanneHoechstens(120, time.Second); got != 2*time.Minute {
		t.Errorf("im Bereich: %v", got)
	}

	if got := domain.ZeitspanneHoechstens(-5, time.Second); got != 0 {
		t.Errorf("negativ: %v, erwartet null", got)
	}
}
