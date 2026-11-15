package msescript_test

// **Kalt und warm, gemessen** — OPS-08.AK-04
// (#1725).

import (
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func BenchmarkZerteileKaltUndWarm(b *testing.B) {
	quelle := strings.Repeat(`x := if a == "b" then 1 + 2 * 3 else to_upper("abc")`+"\n", 20)

	b.Run("kalt", func(b *testing.B) {
		stelle := 0
		for b.Loop() {
			stelle++
			// **Jede Quelle neu**, sonst träfe der Speicher.
			if _, fehler := msescript.Zerteile(quelle + "# " + strconv.Itoa(stelle)); fehler != nil {
				b.Fatal(fehler)
			}
		}
	})

	b.Run("warm", func(b *testing.B) {
		for b.Loop() {
			if _, fehler := msescript.Zerteile(quelle); fehler != nil {
				b.Fatal(fehler)
			}
		}
	})
}
