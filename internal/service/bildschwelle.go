// **Ein Bild nachzeichnen: die Schwelle** —
// #1340,
// Werkstück F von Phase 4.11.
//
// Portiert aus `greyscale` und `threshold`
// (`src/data/format/image_to_symbol.cpp`, Zeilen 36–92, Fassung `f9d9356` vom
// 2021-07-06).
//
// # Warum die Schwelle nicht bei 128 liegt
//
// Sie wird **gemessen**, nicht gesetzt: MSE baut ein Histogramm und sucht die
// Helligkeit, an der die Hälfte der Bildpunkte liegt. Eine feste Schwelle
// machte jedes zu dunkle oder zu helle Bild zu einer leeren Fläche — und
// „scannen Sie heller" ist keine Antwort, die ein Programm geben darf.

package service

import (
	"image"
)

// schwellenfeld ist ein Bild, das nur noch **voll** und **leer** kennt.
//
// **Und ein Rand aus lauter Leere**: MSEs `ImageData::operator()` gibt für
// jede Stelle außerhalb `EMPTY` (Zeilen 116–127). Das ist der Grund, warum der
// Rundgang an einer Form hält, die den Bildrand berührt.
type schwellenfeld struct {
	breite, hoehe int

	// zellen ist zeilenweise. `true` heißt voll.
	zellen []bool

	// besucht merkt, wo ein Rundgang schon einmal angesetzt hat — MSEs
	// `MARKED`. Ohne das fände die Suche dieselbe Form endlos wieder.
	besucht []bool
}

func (f *schwellenfeld) voll(x, y int) bool {
	if x < 0 || x >= f.breite || y < 0 || y >= f.hoehe {
		return false
	}

	return f.zellen[x+y*f.breite]
}

func (f *schwellenfeld) merken(x, y int) {
	if x < 0 || x >= f.breite || y < 0 || y >= f.hoehe {
		return
	}

	f.besucht[x+y*f.breite] = true
}

func (f *schwellenfeld) gemerkt(x, y int) bool {
	if x < 0 || x >= f.breite || y < 0 || y >= f.hoehe {
		return false
	}

	return f.besucht[x+y*f.breite]
}

// schwelle macht aus einem Bild ein Feld aus voll und leer.
func schwelle(gemalt image.Image) *schwellenfeld {
	kasten := gemalt.Bounds()
	breite, hoehe := kasten.Dx(), kasten.Dy()

	// **Grau ist der Mittelwert der drei Kanäle**, nicht die wahrgenommene
	// Helligkeit: `(r+g+b)/3` (`greyscale`, Zeile 40). Eine gewichtete Formel
	// wäre die bessere Umrechnung und gäbe hier eine **andere** Schwelle — und
	// damit eine andere Silhouette als MSE.
	grau := make([]uint8, breite*hoehe)

	for y := range hoehe {
		for x := range breite {
			r, g, b, _ := gemalt.At(kasten.Min.X+x, kasten.Min.Y+y).RGBA()

			// **`RGBA()` gibt 16 Bit je Kanal**; `>>8` bringt jeden auf 0…255,
			// und ihr Mittel liegt wieder darin. Die Klemmung steht trotzdem
			// da: sie kostet nichts und macht die Zusage sichtbar, statt sie
			// einem Leser der Umrechnung zu überlassen.
			gemittelt := (r>>8 + g>>8 + b>>8) / 3
			grau[x+y*breite] = uint8(min(gemittelt, 255)) //nolint:gosec // durch min geklemmt
		}
	}

	feld := &schwellenfeld{
		breite:  breite,
		hoehe:   hoehe,
		zellen:  make([]bool, breite*hoehe),
		besucht: make([]bool, breite*hoehe),
	}

	grenze := gemesseneGrenze(grau)

	for i, ton := range grau {
		feld.zellen[i] = ton >= grenze
	}

	if randIstVoll(feld) {
		for i := range feld.zellen {
			feld.zellen[i] = !feld.zellen[i]
		}
	}

	return feld
}

// gemesseneGrenze sucht die Helligkeit, an der die Hälfte der Punkte liegt.
//
// **Die halbe Klasse zählt mit** — `below + hist[i]/2 > threshold_pos`
// (Zeile 63). Ohne diese Hälfte spränge die Grenze bei einem Bild mit einer
// sehr großen Klasse um eine ganze Stufe daneben.
func gemesseneGrenze(grau []uint8) uint8 {
	var haeufigkeit [256]int
	for _, ton := range grau {
		haeufigkeit[ton]++
	}

	mitte := len(grau) / 2
	darunter := 0

	for i := range 255 {
		if darunter+haeufigkeit[i]/2 > mitte {
			return uint8(i) //nolint:gosec // i < 255
		}

		darunter += haeufigkeit[i]

		if darunter >= mitte {
			return uint8(i + 1) //nolint:gosec // i < 255
		}
	}

	return 255
}

// randIstVoll sagt, ob mehr als die Hälfte des Randes voll ist.
//
// **Die Zählung ist MSEs**, samt ihrer Eigenheit: die vier Ecken werden
// **zweimal** gezählt, einmal in der Zeilen- und einmal in der Spaltenrunde
// (Zeilen 78–83). Verglichen wird gegen `w + h`, also gegen die Hälfte der
// gezählten Stellen. Wer die Ecken einmal zählte, bekäme bei einem Bild mit
// vollen Ecken eine andere Antwort als MSE.
func randIstVoll(feld *schwellenfeld) bool {
	gezaehlt := 0

	for x := range feld.breite {
		if feld.voll(x, 0) {
			gezaehlt++
		}

		if feld.voll(x, feld.hoehe-1) {
			gezaehlt++
		}
	}

	for y := range feld.hoehe {
		if feld.voll(0, y) {
			gezaehlt++
		}

		if feld.voll(feld.breite-1, y) {
			gezaehlt++
		}
	}

	return gezaehlt > feld.breite+feld.hoehe
}
