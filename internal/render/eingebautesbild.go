package render

import (
	"image"
	"image/color"
)

// eingebauteBildkante ist die Seitenlänge der eingebauten Bilder.
//
// MSEs `bool_yes.png` und `bool_no.png` sind 16×16 (gemessen am 2026-09-03 an
// `resource/` der Fassung f9d9356 vom 2021-07-06). Die Zahl steht hier, weil
// eine Vorlage die Größe des Ergebnisses in ihre eigene Rechnung nimmt.
const eingebauteBildkante = 16

// eingebautesBild zeichnet eines der Bilder, die MSE aus seinen eigenen
// Ressourcen lädt.
//
// # Wie MSE es macht
//
//	Image BuiltInImage::generate(const Options& opt) const {
//	  Image img = load_resource_image(name);
//	  …
//	}
//
// `src/gfx/generated_image.cpp`, Zeile 436, Fassung f9d9356 vom 2021-07-06.
// Der Name benennt eine der 59 Dateien in `resource/`.
//
// # Was hier anders ist — Kennzeichnung: adapted (§15)
//
// **Gezeichnet statt mitgenommen.** MSEs Ressourcen sind fremdes Bildmaterial;
// §11 hält solches aus diesem Baum heraus, und ein grünes Häkchen und ein
// rotes Kreuz sind keine Gestalt, für die man eine Vorlage braucht. Was
// übernommen wird, ist die **Bedeutung** — `bool_yes` ist ein Ja, `bool_no`
// ein Nein —, und die trägt jede Zeichnung, die so aussieht.
//
// **Nur diese beiden.** Gezählt am 2026-09-03 über das volle Pack: von 59
// Ressourcennamen kommt kein anderer vor, und die übrigen 57 sind
// Bedienelemente von MSEs eigenem Fenster. Ein Rasterer, der fremde
// Fensterknöpfe nachzeichnet, hätte eine Zuständigkeit zu viel.
func (r *Rasterer) eingebautesBild(baum Bildbaum) (quelle image.Image, befund string) {
	name := baum.Texte["name"]

	ziel := image.NewNRGBA(image.Rect(0, 0, eingebauteBildkante, eingebauteBildkante))

	switch name {
	case "bool_yes":
		hakenZeichnen(ziel)
	case "bool_no":
		kreuzZeichnen(ziel)
	default:
		return ziel, "built_in_image kennt " + name + " nicht — " +
			"gezeichnet sind `bool_yes` und `bool_no`"
	}

	return ziel, ""
}

// hakenZeichnen malt ein grünes Häkchen: ein kurzer Zug nach unten rechts,
// ein langer nach oben rechts.
func hakenZeichnen(ziel *image.NRGBA) {
	gruen := color.NRGBA{G: 128, A: 255}

	for schritt := range 4 {
		strichpunkt(ziel, 3+schritt, 7+schritt, gruen)
	}

	for schritt := range 8 {
		strichpunkt(ziel, 6+schritt, 10-schritt, gruen)
	}
}

// kreuzZeichnen malt ein rotes Kreuz aus zwei Diagonalen.
func kreuzZeichnen(ziel *image.NRGBA) {
	rot := color.NRGBA{R: 192, A: 255}

	for schritt := range 11 {
		strichpunkt(ziel, 3+schritt, 3+schritt, rot)
		strichpunkt(ziel, 13-schritt, 3+schritt, rot)
	}
}

// strichpunkt setzt einen Punkt samt seinem rechten Nachbarn — das gibt der
// Linie die Breite, die sie bei 16 Punkten Kantenlänge sichtbar macht.
func strichpunkt(ziel *image.NRGBA, x, y int, farbe color.NRGBA) {
	ziel.SetNRGBA(x, y, farbe)

	if x+1 < eingebauteBildkante {
		ziel.SetNRGBA(x+1, y, farbe)
	}
}
