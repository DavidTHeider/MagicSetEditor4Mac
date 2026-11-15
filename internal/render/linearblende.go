package render

import "image"

// blendeLinear führt `linear_blend(image1:, image2:, x1:, y1:, x2:, y2:)` aus.
//
// # Die Regel, nachgelesen in MSE2
//
// `src/gfx/blend_image.cpp:18` (Fassung f9d9356 vom 2021-07-06):
//
//	int mult = x * xm + y * ym + d;      // 0 bei (x1,y1), fixed bei (x2,y2)
//	data1[0] = data1[0] + mult * (data2[0] - data1[0]) / fixed;
//
// Bei (x1,y1) gilt `image1`, bei (x2,y2) `image2`, dazwischen linear und
// außerhalb geklemmt. Das Pack baut damit den Hybridrahmen
// (`magic-blends.mse-include/new-blends:70`):
//
//	card_hybrid_2 := {linear_blend(
//		image1: template(colors[0]),
//		image2: template(colors[1]),
//		x1: 0.4, y1: 0
//		x2: 0.6, y2: 0)}
//
// **Die Koordinaten sind Anteile, keine Bildpunkte.** MSE rechnet sie über
// `x1 * width` in Bildpunkte um; 0,4 und 0,6 sind vierzig und sechzig Prozent
// der Breite.
//
// # Drei Einzelheiten, die erst der Quelltext gezeigt hat
//
//   - **Der Punktindex, nicht seine Mitte.** MSE rechnet mit `x`, nicht mit
//     `x + 0,5`. Der erste Anlauf nahm die Mitte, weil sie beim Abtasten
//     richtig ist — beim Blenden ist sie es nicht.
//   - **Nur RGB.** `data1[0..2]`; das Alpha bleibt, wie es ist.
//   - **Gleiche Größe, sonst ein Fehler** — hier ein Befund statt eines
//     Abbruchs (§8).
//
// **Beide Achsen, keine angenommen.** Zeile 91 desselben Moduls bindet
// `linear_blend` um und setzt `y1: x1, y2: x2` — eine senkrechte Version
// derselben Mischung. Wer hier die Waagerechte voraussetzte, bekäme dort einen
// Verlauf ohne Verlauf.
//
// # Warum das die schlechteste Zahl des Bestands betrifft
//
// Eine Hybridkarte bekommt ohne diese Verknüpfung **gar keinen Rahmen** und
// liegt bei 92,1 Abstand gegen MSE2 — 99,9 Prozent abweichende Bildpunkte
// (#306).
func (r *Rasterer) blendeLinear(baum Bildbaum, tiefe int) (gemischt image.Image, befund string) {
	bilder, fehlt := r.kinderbilder(baum, tiefe, "image1", "image2")
	if fehlt != "" {
		return nil, fehlt
	}

	eins, zwei := bilder[0], bilder[1]

	if abweichend := gleicheGroesse(eins, zwei); abweichend != "" {
		return nil, "linear_blend: " + abweichend
	}

	x1, y1 := baum.Zahlen["x1"], baum.Zahlen["y1"]
	x2, y2 := baum.Zahlen["x2"], baum.Zahlen["y2"]

	rahmen := eins.Bounds()
	breite, hoehe := float64(rahmen.Dx()), float64(rahmen.Dy())

	// **Gewichtet mit den Kantenlängen, nicht in reinen Anteilen.**
	//
	// MSE rechnet `a = fixed / (w² dx² + h² dy²)` und dann
	// `xm = (x2-x1) * w * a`, `ym = (y2-y1) * h * a`. Der erste Anlauf teilte
	// durch `dx² + dy²` und behandelte beide Achsen gleich — bei einem
	// **nicht quadratischen** Bild ist das etwas anderes: auf 8×4 mit
	// `(0,0) → (1,1)` und Bildpunkt (4,0) gibt MSE rund 102, der erste Anlauf
	// 64. Angestrichen vom Nachlauf des Zweitblicks.
	spanneX, spanneY := (x2-x1)*breite, (y2-y1)*hoehe

	laenge := spanneX*spanneX + spanneY*spanneY
	if laenge == 0 {
		// **Kein geratener Verlauf.** Fallen die beiden Punkte zusammen, ist
		// die Richtung nicht bestimmt; eine zu erfinden hieße, eine Vorlage zu
		// deuten, die nichts gesagt hat (§8).
		return nil, "linear_blend hat keine Richtung: (x1,y1) und (x2,y2) fallen zusammen"
	}

	ziel := image.NewNRGBA(image.Rect(0, 0, rahmen.Dx(), rahmen.Dy()))

	for y := range rahmen.Dy() {
		for x := range rahmen.Dx() {
			// **Der Punktindex, nicht seine Mitte** — so rechnet MSE, und in
			// Bildpunkten statt in Anteilen.
			anteil := begrenzt01(
				((float64(x)-x1*breite)*spanneX + (float64(y)-y1*hoehe)*spanneY) / laenge)

			a := geradeFarbe(eins, eins.Bounds().Min.X+x, eins.Bounds().Min.Y+y)
			b := geradeFarbe(zwei, zwei.Bounds().Min.X+x, zwei.Bounds().Min.Y+y)

			stelle := ziel.PixOffset(x, y)
			ziel.Pix[stelle] = zwischenKanal(a.R, b.R, anteil)
			ziel.Pix[stelle+1] = zwischenKanal(a.G, b.G, anteil)
			ziel.Pix[stelle+2] = zwischenKanal(a.B, b.B, anteil)

			// **Das Alpha des ersten Bildes bleibt.** MSE mischt drei Kanäle.
			ziel.Pix[stelle+3] = a.A
		}
	}

	return ziel, ""
}

// zwischenKanal ist die Zeile aus `blend_image.cpp:52`, in Go.
//
//	data1[0] = data1[0] + mult * (data2[0] - data1[0]) / fixed
func zwischenKanal(eins, zwei uint8, anteil float64) uint8 {
	return alsAnteil(float64(eins) + anteil*(float64(zwei)-float64(eins)))
}
