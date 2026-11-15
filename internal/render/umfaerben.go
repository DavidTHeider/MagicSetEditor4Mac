package render

import (
	"image"
	"image/color"
)

// **`recolor_image` färbt ein Bild um — die häufigste Verknüpfung im Pack.**
//
// # Die Regel, nachgelesen in MSE2
//
//	int lo = min(x.r,min(x.g,x.b));
//	int nr = x.r - lo; int ng = x.g - lo; int nb = x.b - lo; int nw = lo;
//	int total = max(255, nr+ng+nb+nw);
//	return RGB((nr*cr.r + ng*cg.r + nb*cb.r + nw*cw.r) / total, …);
//
// `recolor`, `src/gfx/image_effects.cpp`, Zeilen 76–93, Fassung f9d9356 vom
// 2021-07-06.
//
// **Ein Bildpunkt wird als Mischung gelesen**, nicht als Farbe: der Abstand
// jedes Kanals zum dunkelsten ist der Anteil dieser Grundfarbe, der Rest ist
// Weiß. Ein Rahmen wird in MSE deshalb einmal in Rot, Grün, Blau und Weiß
// gezeichnet und danach in beliebige vier Farben übersetzt.
//
// **Geteilt wird durch die Summe, nicht durch 255.** MSEs Kommentar sagt
// warum: „We should have that nr+ng+bw+nw < 255, otherwise the input is not a
// mixture. Just to be sure, divide by the sum instead of 255."
//
// **Ohne Alpha** — `GetData()` sind drei Kanäle. Die Durchsichtigkeit gehört
// zur Form eines Rahmens, nicht zu seiner Farbe
// (#270).
func (r *Rasterer) faerbeUm(baum Bildbaum, tiefe int) (quelle image.Image, befund string) {
	unten, tiefer := r.ausBaumTief(baum.Kinder[0].Baum, tiefe+1)
	if unten == nil {
		return nil, tiefer
	}

	rot, gruen, blau, weiss, fehlt := umfaerbefarben(baum)
	if fehlt != "" {
		// **Das Ausgangsbild bleibt.** Sichtbar falsch schlägt unsichtbar
		// fehlend — und der Befund geht trotzdem mit.
		return unten, "recolor_image braucht " + fehlt
	}

	rahmen := unten.Bounds()
	ziel := image.NewNRGBA(image.Rect(0, 0, rahmen.Dx(), rahmen.Dy()))

	for y := range rahmen.Dy() {
		for x := range rahmen.Dx() {
			ziel.SetNRGBA(x, y,
				umgefaerbt(nrgbaAn(unten, rahmen.Min.X+x, rahmen.Min.Y+y),
					rot, gruen, blau, weiss))
		}
	}

	return ziel, ""
}

// umfaerbefarben liest die Farbargumente — beide Schreibweisen.
//
//	SCRIPT_OPTIONAL_PARAM(Color, red) {
//	  SCRIPT_PARAM(Color, green); SCRIPT_PARAM(Color, blue);
//	  SCRIPT_PARAM_DEFAULT(Color, white, Color(255,255,255));
//	  return make_intrusive<RecolorImage2>(input,red,green,blue,white);
//	} else {
//	  SCRIPT_PARAM(Color, color);
//	  return make_intrusive<RecolorImage>(input,color);
//	}
//
// `src/script/functions/image.cpp`, Zeilen 88–99.
//
// **`red` entscheidet, welche Form gemeint ist.** Steht es da, sind `green` und
// `blue` Pflicht und `white` hat eine Vorgabe; sonst gilt die einfarbige Form,
// und dann ist `color` Pflicht.
//
// Fassung f9d9356 vom 2021-07-06.
func umfaerbefarben(baum Bildbaum) (rot, gruen, blau, weiss Farbe, fehlt string) {
	weiss = Farbe{R: 255, G: 255, B: 255, A: 255}

	if rot, da := baum.Farben["red"]; da {
		gruen, gruenDa := baum.Farben["green"]
		blau, blauDa := baum.Farben["blue"]

		if !gruenDa || !blauDa {
			return Farbe{}, Farbe{}, Farbe{}, Farbe{}, "neben `red:` auch `green:` und `blue:`"
		}

		if eigenes, da := baum.Farben["white"]; da {
			weiss = eigenes
		}

		return rot, gruen, blau, weiss, ""
	}

	einzelne, da := baum.Farben["color"]
	if !da {
		return Farbe{}, Farbe{}, Farbe{}, Farbe{}, "ein Argument `color:` oder `red:`"
	}

	// **Grün wird die ähnliche, Blau die entgegengesetzte Farbe.**
	//
	//	bool dark = to_grayscale(cr) < 100;
	//	recolor(img, cr, dark ? black : white, dark ? white : black, white);
	//
	// Zeilen 107–112. So bleibt ein Symbol lesbar, gleich in welcher Farbe es
	// steht: das, was auf dem Vorlagenbild grün war, geht mit der Farbe mit,
	// das Blaue setzt sich davon ab.
	schwarz := Farbe{A: 255}
	hell := Farbe{R: 255, G: 255, B: 255, A: 255}

	if graustufe(einzelne) < 100 {
		return einzelne, schwarz, hell, weiss, ""
	}

	return einzelne, hell, schwarz, weiss, ""
}

// graustufe ist MSEs `to_grayscale`.
//
//	return (Byte)((6969 * x.r + 23434 * x.g + 2365 * x.b) / 32768); // from libpng
//
// Zeilen 103–105. **Die Gewichte kommen aus libpng**, nicht aus Rec. 601 — wer
// hier 0,299/0,587/0,114 einsetzte, bekäme bei Grenzfällen die andere
// Entscheidung.
func graustufe(farbe Farbe) int {
	return (6969*int(farbe.R) + 23434*int(farbe.G) + 2365*int(farbe.B)) / 32768
}

// umgefaerbt rechnet einen Bildpunkt um.
func umgefaerbt(punkt color.NRGBA, rot, gruen, blau, weiss Farbe) color.NRGBA {
	tiefste := min(int(punkt.R), int(punkt.G), int(punkt.B))

	anteilRot := int(punkt.R) - tiefste
	anteilGruen := int(punkt.G) - tiefste
	anteilBlau := int(punkt.B) - tiefste
	anteilWeiss := tiefste

	gesamt := max(255, anteilRot+anteilGruen+anteilBlau+anteilWeiss)

	// **Der Wert kann nicht über 255 gehen**, und das ist der Sinn des
	// `max(255, …)`: die Summe der Anteile ist höchstens `gesamt`, also ist
	// jeder Kanal höchstens der größte der vier Farbwerte. `min` steht hier
	// trotzdem, weil eine Zusicherung im Kommentar keine Zusicherung ist.
	kanal := func(r, g, b, w uint8) uint8 {
		gemischt := (anteilRot*int(r) + anteilGruen*int(g) +
			anteilBlau*int(b) + anteilWeiss*int(w)) / gesamt

		// #nosec G115 -- durch min auf 0…255 begrenzt
		return uint8(min(gemischt, 255)) //nolint:gosec // dito
	}

	return color.NRGBA{
		R: kanal(rot.R, gruen.R, blau.R, weiss.R),
		G: kanal(rot.G, gruen.G, blau.G, weiss.G),
		B: kanal(rot.B, gruen.B, blau.B, weiss.B),
		A: punkt.A,
	}
}

// nrgbaAn liest einen Bildpunkt ohne Vormultiplikation.
//
// **Weil `color.RGBA` vormultipliziert ist** und MSE mit den rohen Kanälen
// rechnet: ein halbdurchsichtiges Rot ist dort (255,0,0) mit Alpha 128, nicht
// (128,0,0).
// **Ohne Prüfung der Umwandlung**: `color.NRGBAModel.Convert` gibt eine
// `color.NRGBA` und nichts anderes; ein Wächter dagegen wäre nie betreten
// worden (§3).
func nrgbaAn(bild image.Image, x, y int) color.NRGBA {
	farbe, _ := color.NRGBAModel.Convert(bild.At(x, y)).(color.NRGBA) //nolint:errcheck // NRGBAModel gibt eine NRGBA

	return farbe
}
