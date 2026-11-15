//go:build darwin

package app

// **Die Rechnung hinter dem freien Platz** — Werkstück O6 von Phase 5.2
// (#1430).
//
// **Intern und nicht über die Naht**: `Statfs` meldet auf dieser Maschine,
// was diese Maschine hat, und keine Größe der Welt bringt es dazu, den
// Überlauf zu zeigen. Die Schranke ist genau deshalb eine eigene Funktion.

import (
	"math"
	"testing"
)

func TestFreieBytesRechnetUndDeckelt(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name     string
		bloecke  uint64
		groesse  uint64
		erwartet int64
	}{
		{"der Normalfall", 1_000_000, 4096, 4_096_000_000},
		{"nichts frei", 0, 4096, 0},
		{"ein Volumen ohne Blockgröße zählt nichts", 1_000_000, 0, 0},
		{"genau an der Grenze", uint64(math.MaxInt64) / 4096, 4096, (math.MaxInt64 / 4096) * 4096},
		{"darüber wird gedeckelt statt zu kippen", math.MaxUint64, 4096, math.MaxInt64},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if gerechnet := freieBytes(fall.bloecke, fall.groesse); gerechnet != fall.erwartet {
				t.Errorf("%d × %d ergab %d, erwartet war %d",
					fall.bloecke, fall.groesse, gerechnet, fall.erwartet)
			}
		})
	}
}

// **Und die Gegenprobe zum Deckel**: ohne ihn wäre die Zahl **negativ** — und
// genau das ist der Schaden, gegen den er steht.
func TestOhneDeckelWaereDieZahlNegativ(t *testing.T) {
	t.Parallel()

	// **Über Variablen und nicht als Ausdruck**: als Konstante rechnet Go
	// exakt und weist den Überlauf beim Übersetzen ab — die Gegenprobe käme
	// gar nicht bis zum Lauf.
	bloecke, groesse := uint64(math.MaxUint64), uint64(4096)

	if roh := int64(bloecke * groesse); roh >= 0 { // #nosec G115 -- die Gegenprobe zeigt den Überlauf
		t.Errorf("der Überlauf blieb aus: %d — dann misst der Deckel nichts", roh)
	}
}
