//go:build unix

package msescript

import (
	"syscall"
	"testing"
	"time"
)

// rechenzeit gibt die Rechenzeit dieses Prozesses — Nutzer und System.
//
// **Rechenzeit und nicht Wanduhr** — ENG-03
// (#1728).
// Seit der Racelauf von `internal/app` in sechs Teilen läuft, ist die Maschine
// während der Teststufe voll, und eine Wanduhr misst dann die fremde Last mit:
// am 2026-09-24 brauchten 1600 Einheiten 0,50 s und 6400 Einheiten 3,77 s —
// das 7,5-Fache, und der Wächter über die Ordnung wurde rot, ohne dass der
// Zerteiler sich geändert hatte. **Die Rechenzeit zählt nur, was dieser
// Prozess tut**; wartet er auf einen Kern, zählt das nicht.
//
// **Was sie mitzählt**: den Race-Detektor, der mit der Arbeit wächst — und
// die Speicherbereinigung auf allen Kernen, die es nicht tut: ein langer Lauf
// löst Sammelzyklen aus, ein kurzer vielleicht keinen. Deshalb schalten die
// Messungen sie ab ([zerteilzeiten]). Und die Rechenzeit gilt nur, solange
// kein zweiter Test im selben Prozess rechnet: der Wächter ruft kein
// `t.Parallel`, und Go lässt die nebenläufigen erst nach allen anderen laufen.
func rechenzeit(t *testing.T) time.Duration {
	t.Helper()

	var verbrauch syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &verbrauch); err != nil {
		t.Fatalf("Rechenzeit lesen: %v", err)
	}

	return time.Duration(verbrauch.Utime.Nano() + verbrauch.Stime.Nano())
}
