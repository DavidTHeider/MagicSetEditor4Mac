// Package dokumentsymbol zeichnet das Dokumentsymbol einer `.mse-set` —
// #891.
//
// # Warum es gezeichnet und nicht gemalt wird
//
// Ein Bild, das jemand von Hand anlegt, lässt sich nicht nachprüfen: wer es
// ändert, ändert eine Datei, und niemand sieht, **was** anders ist. Hier steht
// die Zeichnung als Code; der Wächter daneben hält die eingecheckte Datei
// dagegen und wird rot, sobald beide auseinanderlaufen.
//
// # Was es zeigt
//
// Ein Blatt mit umgeschlagener Ecke — die Gestalt, an der macOS ein Dokument
// erkennt — und darauf die Silhouette einer **Karte**: ein Rahmen, ein
// Bildfeld, zwei Textzeilen. Das ist der Gegenstand dieser Anwendung, und es
// ist von ihrem eigenen Symbol zu unterscheiden, wie es das Issue verlangt.
package dokumentsymbol

import (
	"image"
	"image/color"
)

// Kantenlaenge ist die Seitenlänge des erzeugten Bildes.
//
// **1024, wie das Anwendungssymbol.** Wails rechnet daraus die kleineren
// Stufen; ein größeres Bild brächte nichts, ein kleineres verlöre in der
// größten Stufe.
const Kantenlaenge = 1024

// Zeichnen gibt das Dokumentsymbol als Bild.
func Zeichnen() *image.NRGBA {
	bild := image.NewNRGBA(image.Rect(0, 0, Kantenlaenge, Kantenlaenge))

	blatt := blattmasse()
	ecke := blatt.Dx() * 28 / 100

	for y := range Kantenlaenge {
		for x := range Kantenlaenge {
			bild.SetNRGBA(x, y, farbeAn(x, y, blatt, ecke))
		}
	}

	return bild
}

// blattmasse gibt den Kasten des Blattes.
//
// **Hochkant und mit Rand.** macOS setzt Dokumentsymbole in ein Quadrat; ein
// Blatt, das es ausfüllte, sähe neben den Symbolen des Systems zu groß aus.
func blattmasse() image.Rectangle {
	breite := Kantenlaenge * 62 / 100
	hoehe := Kantenlaenge * 80 / 100

	links := (Kantenlaenge - breite) / 2
	oben := (Kantenlaenge - hoehe) / 2

	return image.Rect(links, oben, links+breite, oben+hoehe)
}

// farbeAn gibt die Farbe eines Bildpunktes.
func farbeAn(x, y int, blatt image.Rectangle, ecke int) color.NRGBA {
	if !imBlatt(x, y, blatt, ecke) {
		return color.NRGBA{}
	}

	// **Die umgeschlagene Ecke** — dieselbe Diagonale, nur die andere Seite.
	if imEck(x, y, blatt, ecke) {
		return farbeEcke
	}

	if farbe, aufDerKarte := kartenfarbe(x, y, blatt); aufDerKarte {
		return farbe
	}

	return farbeBlatt
}

// imBlatt sagt, ob ein Punkt auf dem Blatt liegt — die Ecke abgeschnitten.
func imBlatt(x, y int, blatt image.Rectangle, ecke int) bool {
	if x < blatt.Min.X || x >= blatt.Max.X || y < blatt.Min.Y || y >= blatt.Max.Y {
		return false
	}

	// Die Diagonale läuft von (Max.X-ecke, Min.Y) nach (Max.X, Min.Y+ecke).
	return (blatt.Max.X-x)+(y-blatt.Min.Y) > ecke
}

// imEck sagt, ob ein Punkt im umgeschlagenen Dreieck liegt.
func imEck(x, y int, blatt image.Rectangle, ecke int) bool {
	return x >= blatt.Max.X-ecke && y < blatt.Min.Y+ecke
}

// kartenfarbe gibt die Farbe der Kartensilhouette, wenn der Punkt darauf liegt.
//
// **Eine Karte im Seitenverhältnis 5:7** — dem einer Magic-Karte —, mittig auf
// dem Blatt und **unterhalb der umgeschlagenen Ecke**. Der erste Anlauf setzte
// sie über die ganze Blattbreite, und ihr Rahmen lief in die Ecke hinein.
//
// Rahmen, Bildfeld und zwei Zeilen: mehr trägt ein Symbol nicht, das auch mit
// sechzehn Bildpunkten noch erkennbar sein soll.
func kartenfarbe(x, y int, blatt image.Rectangle) (color.NRGBA, bool) {
	karte := kartenmasse(blatt)

	if !imKasten(x, y, karte) {
		return color.NRGBA{}, false
	}

	strich := max(karte.Dx()/24, 1)

	if amRand(x, y, karte, strich) {
		return farbeStrich, true
	}

	// Das Bildfeld sitzt im oberen Drittel, die Zeilen darunter.
	bildfeld := teilkasten(karte, strich, 10, 48)
	if imKasten(x, y, bildfeld) {
		return farbeStrich, true
	}

	for _, zeile := range [][2]int{{58, 68}, {74, 84}} {
		if imKasten(x, y, teilkasten(karte, strich, zeile[0], zeile[1])) {
			return farbeStrich, true
		}
	}

	return color.NRGBA{}, false
}

// kartenmasse gibt den Kasten der Karte auf dem Blatt.
//
// **Waagerecht mittig, senkrecht unter der Ecke.** Die umgeschlagene Ecke
// nimmt oben rechts 28 % der Blattbreite ein; die Karte beginnt darunter,
// damit ihr Rahmen nicht hineinläuft.
func kartenmasse(blatt image.Rectangle) image.Rectangle {
	breite := blatt.Dx() * 56 / 100
	hoehe := breite * 7 / 5

	links := blatt.Min.X + (blatt.Dx()-breite)/2
	obenGrenze := blatt.Min.Y + blatt.Dx()*28/100
	oben := obenGrenze + (blatt.Max.Y-obenGrenze-hoehe)/2

	return image.Rect(links, oben, links+breite, oben+hoehe)
}

// eingerueckt verkleinert einen Kasten um denselben Betrag auf allen Seiten.
func eingerueckt(kasten image.Rectangle, um int) image.Rectangle {
	return image.Rect(kasten.Min.X+um, kasten.Min.Y+um, kasten.Max.X-um, kasten.Max.Y-um)
}

// teilkasten gibt einen waagerechten Streifen der Karte, in Prozent ihrer Höhe.
func teilkasten(karte image.Rectangle, rand, von, bis int) image.Rectangle {
	innen := eingerueckt(karte, rand*2)

	return image.Rect(
		innen.Min.X,
		karte.Min.Y+karte.Dy()*von/100,
		innen.Max.X,
		karte.Min.Y+karte.Dy()*bis/100,
	)
}

// imKasten sagt, ob ein Punkt in einem Kasten liegt.
func imKasten(x, y int, kasten image.Rectangle) bool {
	return x >= kasten.Min.X && x < kasten.Max.X &&
		y >= kasten.Min.Y && y < kasten.Max.Y
}

// amRand sagt, ob ein Punkt auf dem Rahmen eines Kastens liegt.
func amRand(x, y int, kasten image.Rectangle, strich int) bool {
	return x < kasten.Min.X+strich || x >= kasten.Max.X-strich ||
		y < kasten.Min.Y+strich || y >= kasten.Max.Y-strich
}

// Die Farben. **Gemessen wird hier nichts** — ein Dokumentsymbol steht im
// Finder auf dessen Grund, nicht in dieser Anwendung, und `docs/KONTRASTE.md`
// führt die Farben der Oberfläche.
var (
	farbeBlatt  = color.NRGBA{R: 0xF8, G: 0xFA, B: 0xFC, A: 0xFF}
	farbeEcke   = color.NRGBA{R: 0xCB, G: 0xD5, B: 0xE1, A: 0xFF}
	farbeStrich = color.NRGBA{R: 0x1E, G: 0x29, B: 0x3B, A: 0xFF}
)
