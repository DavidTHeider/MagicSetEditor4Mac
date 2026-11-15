package service

// **Ein Bild nachzeichnen: die Schwelle** —
// #1340,
// Werkstück F von Phase 4.11.
//
// Portiert aus `greyscale` und `threshold`
// (`src/data/format/image_to_symbol.cpp`, Zeilen 36–92, Fassung `f9d9356` vom
// 2021-07-06).

import (
	"image"
	"image/color"
	"testing"
)

// bild baut ein Bild aus einer Zeichnung: `#` ist dunkel, `.` ist hell.
func bild(zeilen ...string) image.Image {
	hoehe := len(zeilen)
	breite := len(zeilen[0])
	gemalt := image.NewRGBA(image.Rect(0, 0, breite, hoehe))

	for y, zeile := range zeilen {
		for x, zeichen := range zeile {
			ton := uint8(255)
			if zeichen == '#' {
				ton = 0
			}

			gemalt.Set(x, y, color.RGBA{R: ton, G: ton, B: ton, A: 255})
		}
	}

	return gemalt
}

// TestDieSchwelleTrenntVordergrundVomRand prüft die Grundzusage.
func TestDieSchwelleTrenntVordergrundVomRand(t *testing.T) {
	t.Parallel()

	feld := schwelle(bild(
		"....",
		".##.",
		".##.",
		"....",
	))

	// **Der Rand ist leer, die Mitte voll.** MSE dreht das Bild um, wenn mehr
	// als die Hälfte des Randes voll wäre (`threshold`, Zeilen 76–91) — hier
	// ist der Rand hell, und hell ist nach der Schwelle zunächst `voll`.
	for _, fall := range []struct {
		x, y int
		voll bool
	}{
		{0, 0, false},
		{1, 1, true},
		{2, 2, true},
		{3, 3, false},
	} {
		if gilt := feld.voll(fall.x, fall.y); gilt != fall.voll {
			t.Errorf("(%d,%d): voll=%v, erwartet %v", fall.x, fall.y, gilt, fall.voll)
		}
	}
}

// **Außerhalb ist leer** — MSEs `ImageData::operator()` gibt für jede Stelle
// außerhalb `EMPTY` zurück (Zeilen 116–127).
//
// **Das ist keine Bequemlichkeit, sondern der Grund, warum der Rundgang
// überhaupt hält**: eine Form am Bildrand hätte sonst keine Kante, an der sie
// umkehrt.
func TestAusserhalbIstLeer(t *testing.T) {
	t.Parallel()

	feld := schwelle(bild("##", "##"))

	for _, stelle := range [][2]int{{-1, 0}, {0, -1}, {2, 0}, {0, 2}} {
		if feld.voll(stelle[0], stelle[1]) {
			t.Errorf("(%d,%d) ist voll, außerhalb muss leer sein", stelle[0], stelle[1])
		}
	}
}

// **Ein Bild mit dunklem Rand wird umgedreht.**
//
// `threshold`, Zeilen 76–91: gezählt wird der Rand; sind mehr als die Hälfte
// seiner Stellen voll, kippt das ganze Bild. Ohne das wäre eine weiße Form auf
// schwarzem Grund ein Loch.
func TestEinDunklerRandDrehtDasBildUm(t *testing.T) {
	t.Parallel()

	feld := schwelle(bild(
		"####",
		"#..#",
		"#..#",
		"####",
	))

	if feld.voll(0, 0) {
		t.Error("die Ecke ist voll — das Bild wurde nicht umgedreht")
	}

	if !feld.voll(1, 1) {
		t.Error("die Mitte ist leer — das Bild wurde nicht umgedreht")
	}
}
