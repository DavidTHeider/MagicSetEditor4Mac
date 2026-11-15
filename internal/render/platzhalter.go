package render

import (
	"image"
	"image/color"
	"image/draw"
)

// Platzhalterknoten ist die Knotenart des Schachbretts eines leeren
// Bildfelds — **keine Bildfunktion von MSE-Script**, sondern was
// `ImageValueViewer` zeichnet, wenn ein Bildfeld kein Bild hat
// (`src/render/value/image.cpp`, Zeilen 54–61, Fassung f9d9356 vom
// 2021-07-06). Den Knoten setzt `service`, kein Skript.
const Platzhalterknoten = "platzhalter"

// schachbrettfeld ist `checker_size` — zehn Bildpunkte des **Ziels**:
// `imagePlaceholder` zeichnet in ein Bild von `w × h` Bildpunkten mit dem
// Maßstab eins (ebd., Zeilen 102–110).
const schachbrettfeld = 10

// platzhalter ist `draw_checker`: weiß, darauf graue Felder (235), die erste
// Zeile beginnend bei null, die nächste um ein Feld versetzt
// (`src/gui/util.cpp`, Zeilen 114–132, Fassung f9d9356 vom 2021-07-06).
// **In der gefragten Größe**, damit ein Feld zehn Bildpunkte bleibt, bei
// jeder Vergrößerung — so wie bei MSE.
func (r *Rasterer) platzhalter() image.Image {
	breite := max(1, r.gefragteGroesse.X)
	hoehe := max(1, r.gefragteGroesse.Y)

	bild := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))
	draw.Draw(bild, bild.Bounds(), image.NewUniform(color.NRGBA{R: 255, G: 255, B: 255, A: 255}), image.Point{}, draw.Src)

	grau := image.NewUniform(color.NRGBA{R: 235, G: 235, B: 235, A: 255})
	versetzt := 0

	for y := 0; y < hoehe; y += schachbrettfeld {
		for x := versetzt * schachbrettfeld; x < breite; x += 2 * schachbrettfeld {
			draw.Draw(bild, image.Rect(x, y, x+schachbrettfeld, y+schachbrettfeld).Intersect(bild.Bounds()),
				grau, image.Point{}, draw.Src)
		}

		versetzt = 1 - versetzt
	}

	return bild
}
