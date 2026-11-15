package render

import (
	"image"
	"image/color"
)

// geradeFarbe liest einen Bildpunkt **ungewichtet**, auch wo er durchsichtig ist.
//
// # Warum das für die Blenden nötig ist
//
// MSE hält RGB und Alpha getrennt (`wxImage`): `mask_blend` und `linear_blend`
// laufen über `GetData()` und sehen **nur RGB**, ungewichtet, unabhängig vom
// Alpha. Ein vollständig durchsichtiger Bildpunkt trägt dort trotzdem eine
// Farbe, und die geht in die Mischung ein.
//
// [punktfarbe] gibt an solchen Stellen `Farbe{}` — schwarz. Für das Übereinander
// legen ist das richtig, denn ein unsichtbarer Punkt soll nichts beitragen; für
// das **Mischen** ist es falsch und erzeugt Farbsäume an den Kanten, wo eine
// Vorlage durchsichtige Ränder hat.
//
// Angestrichen vom Zweitblick über die Phasen B bis D: „für transparente Kinder
// werden entgewichtete RGB- und Alphakanäle unabhängig interpoliert … das ist
// bei Bildkanten materiell".
//
// # Was sie nicht kann
//
// **Ein vorgewichtetes Bild gibt seine Farbe bei Alpha null nicht mehr her.**
// `image.RGBA` speichert vorgewichtet; dort ist die Farbe verloren, und es
// bleibt bei schwarz. Die Bilder der Pakete sind PNG und JPEG, und `image/png`
// liefert für Bilder mit Alpha ein [image.NRGBA] — der Fall, auf den es
// ankommt, ist also gedeckt.
func geradeFarbe(bild image.Image, x, y int) Farbe {
	rahmen := bild.Bounds()

	x = max(rahmen.Min.X, min(x, rahmen.Max.X-1))
	y = max(rahmen.Min.Y, min(y, rahmen.Max.Y-1))

	// **Über [color.NRGBAModel] und nicht über einen Typvergleich.**
	//
	// Der erste Anlauf prüfte auf `*image.NRGBA` und fiel für alles andere
	// zurück. `image.Decode` liefert aber auch `*image.Paletted` — das Pack
	// führt palettierte PNG mit Alpha, etwa `magic-mana-maps/…/mana_wubrg.png`
	// — und `*image.NRGBA64`. Beide bewahren ihre gerade Farbe, und beide
	// fielen durch. Angestrichen vom Nachlauf des Zweitblicks: „nur das
	// Beispiel geschlossen".
	//
	// Das Farbmodell fragt nach der **Eigenschaft** statt nach dem Typ: jede
	// `color.Color` lässt sich nach NRGBA umrechnen, und wo die Farbe noch da
	// ist, kommt sie heraus.
	// **Nach der Art der Farbe, nicht nach der des Bildes.**
	//
	// Der erste Anlauf prüfte auf `*image.NRGBA` und fiel für alles andere
	// zurück — `*image.Paletted` und `*image.NRGBA64` bewahren ihre gerade
	// Farbe aber ebenso. Ein palettiertes PNG mit Alpha liegt im Pack
	// (`magic-mana-maps/…/mana_wubrg.png`), und `image.Decode` gibt es als
	// `*image.Paletted` zurück; seine Palette führt `color.NRGBA`.
	switch farbe := bild.At(x, y).(type) {
	case color.NRGBA:
		return Farbe{R: farbe.R, G: farbe.G, B: farbe.B, A: farbe.A}

	case color.NRGBA64:
		return Farbe{
			R: uint8(farbe.R >> 8), G: uint8(farbe.G >> 8), //nolint:gosec // Verschiebung um acht
			B: uint8(farbe.B >> 8), A: uint8(farbe.A >> 8), //nolint:gosec // dito
		}
	}

	// **Vorgewichtet: die Farbe ist bei Alpha null verloren**, und [punktfarbe]
	// sagt das mit Schwarz. Das ist keine Lücke, sondern die Grenze — sie hat
	// ihren eigenen Wächter.
	return punktfarbe(bild, rahmen, x, y)
}
