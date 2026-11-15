package render

import "image"

// ErzeugtesMass gibt die Größe des Bilds, das ein Baum für die gefragte Größe
// erzeugt — so, wie eine Wahl sie als `content_width` und `content_height`
// meldet (#2303).
//
// MSE2 erzeugt das Stilbild mit den Optionen der Wahl und übernimmt die Größe,
// die `generateCached` zurückgibt (`prepare_choice_viewer`,
// `src/render/value/choice.cpp`, Zeilen 35–58): das Bild nach
// `conform_image` (`src/gfx/generated_image.cpp`, Zeilen 30–66). **Ein leerer
// Baum ist `BlankImage`** — so groß wie gefragt, mindestens ein Punkt (Zeilen
// 83–89); `nil` und die leere Zeichenkette werden dort zu ihm
// (`src/script/value.cpp`, Zeilen 369–375 und 450–452). Alles Fassung f9d9356
// vom 2021-07-06.
//
// einpassen ist `ASPECT_FIT`; ohne gilt `ASPECT_STRETCH`, die Vorgabe der
// Optionen. Ohne Maß — kein Bild, das sich erzeugen lässt — ist da falsch:
// dann bleibt bei MSE2 das vorige Maß stehen.
func (r *Rasterer) ErzeugtesMass(baum Bildbaum, gefragt image.Point, einpassen bool) (mass image.Point, da bool) {
	if baum.Leer() {
		return image.Point{X: max(1, gefragt.X), Y: max(1, gefragt.Y)}, true
	}

	quelle, _ := r.ausBaum(baum, gefragt)
	if quelle == nil {
		return image.Point{}, false
	}

	masse := quelle.Bounds()

	if einpassen && gefragt.X > 0 && gefragt.Y > 0 && masse.Dx() > 0 && masse.Dy() > 0 &&
		(masse.Dx() != gefragt.X || masse.Dy() != gefragt.Y) {
		// **Zu viel Höhe gefragt**, oder zu viel Breite — `conform_image`,
		// Zeilen 49–59, in ganzen Zahlen.
		if masse.Dx()*gefragt.Y > masse.Dy()*gefragt.X {
			return image.Point{X: gefragt.X, Y: gefragt.X * masse.Dy() / masse.Dx()}, true
		}

		return image.Point{X: gefragt.Y * masse.Dx() / masse.Dy(), Y: gefragt.Y}, true
	}

	breite, hoehe := angepassteGroesse(masse, gefragt.X, gefragt.Y)

	return image.Point{X: breite, Y: hoehe}, true
}
