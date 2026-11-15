package service

// **Ein Bild nachzeichnen: der Rundgang** —
// #1340.
//
// Portiert aus `find_symbol_shape_start` und `read_symbol_shape`
// (`src/data/format/image_to_symbol.cpp`, Zeilen 129–204, Fassung `f9d9356`
// vom 2021-07-06).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// TestDerRundgangUmschliesstEinQuadrat ist der einfachste Fall.
func TestDerRundgangUmschliesstEinQuadrat(t *testing.T) {
	t.Parallel()

	feld := schwelle(bild(
		"......",
		"......",
		"..##..",
		"..##..",
		"......",
		"......",
	))

	form, gefunden := umrissLesen(feld)
	if !gefunden {
		t.Fatal("keine Form gefunden")
	}

	// **Ein geschlossener Zug**: der Rundgang endet, wo er begann, und der
	// letzte Punkt ist der Anfangspunkt — MSEs `while (x != xs || y != ys)`.
	if len(form.Punkte) < 4 {
		t.Fatalf("%d Punkte, ein Quadrat braucht mindestens vier", len(form.Punkte))
	}

	// **Der Umriss liegt im Einheitsquadrat**, weil MSE durch Breite und Höhe
	// teilt (Zeilen 186–189).
	for i, punkt := range form.Punkte {
		if punkt.Stelle.X < 0 || punkt.Stelle.X > 1 || punkt.Stelle.Y < 0 || punkt.Stelle.Y > 1 {
			t.Errorf("Punkt %d liegt außerhalb: %+v", i, punkt.Stelle)
		}
	}

	// **Eine äußere Form wird vereint** — `SYMBOL_COMBINE_MERGE` (Zeile 201).
	if form.Kombination != msepaket.KombinationVereinen {
		t.Errorf("Kombination %q, erwartet %q", form.Kombination, msepaket.KombinationVereinen)
	}
}

// **Zwei getrennte Flecken geben zwei Formen.**
//
// `image_to_symbol` liest so lange Formen, wie es welche findet (Zeilen
// 215–219) — und die Markierung sorgt dafür, dass es nicht immer dieselbe
// findet.
func TestZweiFleckenGebenZweiFormen(t *testing.T) {
	t.Parallel()

	feld := schwelle(bild(
		"........",
		".##..##.",
		".##..##.",
		"........",
	))

	gefundene := 0

	for {
		if _, gefunden := umrissLesen(feld); !gefunden {
			break
		}

		gefundene++

		if gefundene > 5 {
			t.Fatal("der Rundgang findet endlos neue Formen — die Markierung greift nicht")
		}
	}

	if gefundene != 2 {
		t.Errorf("%d Formen, erwartet 2", gefundene)
	}
}

// **Ein leeres Bild gibt keine Form.**
func TestEinLeeresBildGibtKeineForm(t *testing.T) {
	t.Parallel()

	feld := &schwellenfeld{
		breite:  4,
		hoehe:   4,
		zellen:  make([]bool, 16),
		besucht: make([]bool, 16),
	}

	if _, gefunden := umrissLesen(feld); gefunden {
		t.Error("eine Form in einem leeren Feld")
	}
}
