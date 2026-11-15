package render

import "testing"

// **Kerning geht in alle vier Breiten ein** —
// #752.
//
// # Warum vier und nicht eine
//
// Es gibt vier Summenstellen über dieselbe Sache: der Umbruch
// ([vorsilbe.nimm]), der Vorschub beim Zeichnen ([vorschub]), die Zeilenbreite
// für die Ausrichtung ([breiteVon]) und die breiteste Zeile eines Blocks
// (`textbreite.go`). `satz.go` sagt, was daraus wird, wenn eine von ihnen
// etwas lernt und die anderen nicht.
//
// **Der erste Anlauf hat es vorgeführt**: Kerning ging nur in die gemessene
// Breite und nicht in den Vorschub. Die Zeile wurde schmaler gemessen und
// weiter ungekernt gesetzt — die Ausrichtung verschob sich, die Glyphen
// standen wie vorher, und der Abstand am Kanon **stieg** von 20,365 auf 20,377.
//
// **Der Test steht im Paket**, weil drei der vier Stellen dort leben und keine
// von ihnen eine Zusage an den Aufrufer ist.
func TestKerningGehtInAlleBreitenEin(t *testing.T) {
	t.Parallel()

	schrift := kernprobe{
		breiten: map[rune]float64{'T': 1, 'a': 1, ' ': 1, 'A': 1},
		paare:   map[[2]rune]float64{{'T', 'a'}: -0.25, {' ', 'A'}: -0.5},
	}
	satz := Satz{Groesse: 10}

	t.Run("die Zeilenbreite", func(t *testing.T) {
		t.Parallel()

		if got := breiteVon("Ta", satz, schrift); got != 17.5 {
			t.Errorf("breiteVon = %v, erwartet 17,5", got)
		}
	})

	t.Run("der Vorschub beim Zeichnen", func(t *testing.T) {
		t.Parallel()

		// **Das Paar gehört vor das Zeichen**: der Stift rückt erst um die
		// Korrektur, dann um den Vorschub.
		if got := vorschub('T', 'a', satz, schrift, 1, false); got != 7.5 {
			t.Errorf("vorschub = %v, erwartet 7,5", got)
		}

		// **Ohne vorheriges Zeichen gibt es kein Paar.**
		if got := vorschub(0, 'a', satz, schrift, 1, false); got != 10 {
			t.Errorf("vorschub am Anfang = %v, erwartet 10", got)
		}
	})

	t.Run("die breiteste Zeile", func(t *testing.T) {
		t.Parallel()

		// `textbreite` rechnet bei Größe 1 und nimmt sie am Ende.
		if got := Textbreite("Ta", 10, schrift); got != 17.5 {
			t.Errorf("Textbreite = %v, erwartet 17,5", got)
		}
	})

	t.Run("der Umbruch, auch über die Stücknaht", func(t *testing.T) {
		t.Parallel()

		var lauf vorsilbe

		breite := lauf.nimm(" ", satz, schrift) + lauf.nimm("A", satz, schrift)
		if breite != 15 {
			t.Errorf("über zwei Stücke = %v, erwartet 15", breite)
		}
	})
}

// kernprobe ist eine Schrift mit erklärten Breiten und Paaren.
type kernprobe struct {
	breiten map[rune]float64
	paare   map[[2]rune]float64
}

func (k kernprobe) Breite(zeichen rune) float64 { return k.breiten[zeichen] }
func (k kernprobe) Zeilenhoehe() float64        { return 1 }

func (k kernprobe) Kernung(vorher, jetzt rune) float64 {
	return k.paare[[2]rune{vorher, jetzt}]
}
