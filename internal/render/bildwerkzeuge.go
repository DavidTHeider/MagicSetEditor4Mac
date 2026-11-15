package render

import (
	"image"
	"image/color"
)

// **Vier Verknüpfungen, die aus einem Bild ein anderes machen.**
//
// `saturate`, `invert_image`, `crop` und die beiden Spiegelungen — die
// einfachste Sorte, weil sie kein zweites Bild brauchen
// (#270).

// saettige führt `saturate` aus.
//
//	int factor = int(256 * amount);
//	if (factor == 0) return;
//	else if (factor == 256) { … r+r > g+b ? 255 : 0 … }
//	else if (factor > 0) { int div = 768 - 3*factor;
//	                       pix[0] = col((768*r - factor*(r+g+b)) / div); … }
//	else { int f1 = -factor, f2 = 768 - 3*f1;
//	       pix[0] = (f2*r + f1*(r+g+b)) / 768; … }
//
// `src/gfx/image_effects.cpp`, Zeilen 15–61, Fassung f9d9356 vom 2021-07-06.
//
// **Drei Zweige und eine Division.** Bei genau eins teilte MSE durch null;
// dort gilt „größer als die beiden anderen zusammen" — der Kommentar sagt
// warum: „if we take infty to be 255, then it is a >avg test".
func (r *Rasterer) saettige(baum Bildbaum, tiefe int) (quelle image.Image, befund string) {
	unten, tiefer := r.ausBaumTief(baum.Kinder[0].Baum, tiefe+1)
	if unten == nil {
		return nil, tiefer
	}

	menge, da := baum.Zahlen["amount"]
	if !da {
		return unten, "saturate braucht ein Argument `amount:`"
	}

	faktor := int(256 * menge)
	if faktor == 0 {
		return unten, ""
	}

	return jederPunkt(unten, func(farbe color.NRGBA) color.NRGBA {
		return gesaettigt(farbe, faktor)
	}), ""
}

// gesaettigt rechnet einen Bildpunkt um.
func gesaettigt(farbe color.NRGBA, faktor int) color.NRGBA {
	rot, gruen, blau := int(farbe.R), int(farbe.G), int(farbe.B)

	if faktor == 256 {
		return color.NRGBA{
			R: schwelle(rot+rot > gruen+blau), G: schwelle(gruen+gruen > blau+rot),
			B: schwelle(blau+blau > rot+gruen), A: farbe.A,
		}
	}

	summe := rot + gruen + blau

	if faktor > 0 {
		teiler := 768 - 3*faktor

		return color.NRGBA{
			R: geklemmt((768*rot - faktor*summe) / teiler),
			G: geklemmt((768*gruen - faktor*summe) / teiler),
			B: geklemmt((768*blau - faktor*summe) / teiler),
			A: farbe.A,
		}
	}

	erster := -faktor
	zweiter := 768 - 3*erster

	return color.NRGBA{
		R: geklemmt((zweiter*rot + erster*summe) / 768),
		G: geklemmt((zweiter*gruen + erster*summe) / 768),
		B: geklemmt((zweiter*blau + erster*summe) / 768),
		A: farbe.A,
	}
}

// schwelle macht aus der Entscheidung einen Kanalwert.
func schwelle(trifft bool) uint8 {
	if trifft {
		return 255
	}

	return 0
}

// geklemmt ist MSEs `col` — es begrenzt auf 0…255.
func geklemmt(wert int) uint8 {
	return uint8(max(0, min(255, wert))) //nolint:gosec // durch min/max begrenzt
}

// kehreUm führt `invert_image` aus.
//
//	for (int i = 0 ; i < n ; ++i) data[i] = 255 - data[i];
//
// `src/gfx/image_effects.cpp`, Zeilen 66–72. `n` ist **dreimal** die Zahl der
// Bildpunkte: die Durchsichtigkeit bleibt.
//
// Fassung f9d9356 vom 2021-07-06.
func (r *Rasterer) kehreUm(baum Bildbaum, tiefe int) (quelle image.Image, befund string) {
	unten, tiefer := r.ausBaumTief(baum.Kinder[0].Baum, tiefe+1)
	if unten == nil {
		return nil, tiefer
	}

	return jederPunkt(unten, func(farbe color.NRGBA) color.NRGBA {
		return color.NRGBA{R: 255 - farbe.R, G: 255 - farbe.G, B: 255 - farbe.B, A: farbe.A}
	}), ""
}

// schneide führt `crop` aus.
//
//	return image->generate(opt).Size(wxSize((int)width, (int)height),
//	                                 wxPoint(-(int)offset_x, -(int)offset_y));
//
// `CropImage::generate`, `src/gfx/generated_image.cpp`, Zeilen 309–311.
//
// **Der Versatz zählt negativ**: `offset_x: 2` schiebt das Bild um zwei Punkte
// nach links, zeigt also ab Spalte zwei. Was über den Rand ragt, bleibt leer —
// `wxImage::Size` füllt mit der Hintergrundfarbe, und ohne eine ist das
// durchsichtig.
//
// Fassung f9d9356 vom 2021-07-06.
func (r *Rasterer) schneide(baum Bildbaum, tiefe int) (quelle image.Image, befund string) {
	unten, tiefer := r.ausBaumTief(baum.Kinder[0].Baum, tiefe+1)
	if unten == nil {
		return nil, tiefer
	}

	breite, breiteDa := baum.Zahlen["width"]
	hoehe, hoeheDa := baum.Zahlen["height"]

	if !breiteDa || !hoeheDa || breite <= 0 || hoehe <= 0 {
		return unten, "crop braucht `width:` und `height:` über null"
	}

	versatzX, versatzY := int(baum.Zahlen["offset_x"]), int(baum.Zahlen["offset_y"])

	rahmen := unten.Bounds()
	ziel := image.NewNRGBA(image.Rect(0, 0, int(breite), int(hoehe)))

	for y := range int(hoehe) {
		for x := range int(breite) {
			quellX, quellY := rahmen.Min.X+x+versatzX, rahmen.Min.Y+y+versatzY
			if !(image.Point{X: quellX, Y: quellY}).In(rahmen) {
				continue
			}

			ziel.SetNRGBA(x, y, nrgbaAn(unten, quellX, quellY))
		}
	}

	return ziel, ""
}

// spiegele führt `flip_horizontal` und `flip_vertical` aus.
//
//	Image FlipImageHorizontal::generate(const Options& opt) const {
//	  return image->generate(opt).Mirror(true);
//	}
//
// `src/gfx/generated_image.cpp`, Zeilen 254–270.
//
// Fassung f9d9356 vom 2021-07-06.
// spiegeleWaagerecht und spiegeleSenkrecht sind die zwei Namen, unter denen
// [gebauteVerknuepfungen] die Spiegelung führt.
//
// **Zwei Methoden statt eines Kennzeichens in der Liste**: die Liste trägt je
// Knotenart **eine** Ausführung, und ein zusätzliches Feld nur für diesen
// einen Fall wäre eine Sonderregel, die jeder Leser mitdenken müsste.
func (r *Rasterer) spiegeleWaagerecht(
	baum Bildbaum, tiefe int,
) (quelle image.Image, befund string) {
	return r.spiegele(baum, tiefe, true)
}

func (r *Rasterer) spiegeleSenkrecht(
	baum Bildbaum, tiefe int,
) (quelle image.Image, befund string) {
	return r.spiegele(baum, tiefe, false)
}

func (r *Rasterer) spiegele(
	baum Bildbaum, tiefe int, waagerecht bool,
) (quelle image.Image, befund string) {
	unten, tiefer := r.ausBaumTief(baum.Kinder[0].Baum, tiefe+1)
	if unten == nil {
		return nil, tiefer
	}

	rahmen := unten.Bounds()
	ziel := image.NewNRGBA(image.Rect(0, 0, rahmen.Dx(), rahmen.Dy()))

	for y := range rahmen.Dy() {
		for x := range rahmen.Dx() {
			quellX, quellY := x, y
			if waagerecht {
				quellX = rahmen.Dx() - 1 - x
			} else {
				quellY = rahmen.Dy() - 1 - y
			}

			ziel.SetNRGBA(x, y, nrgbaAn(unten, rahmen.Min.X+quellX, rahmen.Min.Y+quellY))
		}
	}

	return ziel, ""
}

// jederPunkt legt eine Umrechnung über jeden Bildpunkt.
func jederPunkt(quelle image.Image, tue func(color.NRGBA) color.NRGBA) image.Image {
	rahmen := quelle.Bounds()
	ziel := image.NewNRGBA(image.Rect(0, 0, rahmen.Dx(), rahmen.Dy()))

	for y := range rahmen.Dy() {
		for x := range rahmen.Dx() {
			ziel.SetNRGBA(x, y, tue(nrgbaAn(quelle, rahmen.Min.X+x, rahmen.Min.Y+y)))
		}
	}

	return ziel
}

// vergroessere führt `enlarge` aus — ein durchsichtiger Rand rundherum.
//
//	Image EnlargeImage::generate(const Options& opt) const {
//	  …
//	  int w  = img.GetWidth(),  h = img.GetHeight();
//	  int dw = int(w * border_size), dh = int(h * border_size);
//	  int w2 = w + dw + dw,     h2 = h + dh + dh;
//	  Image larger(w2,h2);
//	  larger.InitAlpha();
//	  memset(larger.GetAlpha(),0,w2*h2); // blank
//	  … // das Bild in die Mitte kopieren
//	}
//
// `src/gfx/generated_image.cpp`, Zeilen 271–300, Fassung f9d9356 vom
// 2021-07-06.
//
// **Der Rand ist ein Anteil, keine Punktzahl**: `border_size` von 0,1 legt an
// jede Seite ein Zehntel der Bildbreite beziehungsweise -höhe. Das Bild wächst
// also um das Doppelte davon.
//
// **Abgeschnitten, nicht gerundet** — `int(w * border_size)`. Bei 0,004 auf
// 100 Punkten kommt null heraus, und das Bild bleibt, wie es war.
//
// # Was hier abweicht und benannt ist
//
// MSE fragt das Kind mit **verkleinerten** Maßen an (`sub_opt`), damit das
// Ergebnis am Ende die gewünschte Größe hat. Dieses Programm kennt keine
// Wunschgröße im Bildbaum — der Zeichenplan gibt sie erst an der Ebene. Was
// hier herauskommt, ist deshalb um den Rand **größer** statt gleich groß; die
// Ebene passt es ein (#472).
func (r *Rasterer) vergroessere(baum Bildbaum, tiefe int) (quelle image.Image, befund string) {
	unten, tiefer := r.ausBaumTief(baum.Kinder[0].Baum, tiefe+1)
	if unten == nil {
		return nil, tiefer
	}

	anteil, da := baum.Zahlen["border_size"]
	if !da {
		return unten, "enlarge braucht `border_size:`"
	}

	rahmen := unten.Bounds()
	randX, randY := int(float64(rahmen.Dx())*anteil), int(float64(rahmen.Dy())*anteil)

	if randX <= 0 && randY <= 0 {
		return unten, ""
	}

	ziel := image.NewNRGBA(image.Rect(0, 0,
		rahmen.Dx()+2*randX, rahmen.Dy()+2*randY))

	for y := range rahmen.Dy() {
		for x := range rahmen.Dx() {
			ziel.SetNRGBA(randX+x, randY+y,
				nrgbaAn(unten, rahmen.Min.X+x, rahmen.Min.Y+y))
		}
	}

	return ziel, ""
}

// leeresBild führt `BlankImage` aus — durchsichtig, in der gefragten Größe.
//
//	Image BlankImage::generate(const Options& opt) const {
//	  int w = max(1, opt.width >= 0  ? opt.width  : opt.height);
//	  int h = max(1, opt.height >= 0 ? opt.height : opt.width);
//	  Image img(w, h);
//	  img.InitAlpha();
//	  memset(img.GetAlpha(), 0, w * h);
//	  return img;
//	}
//
// `src/gfx/generated_image.cpp`, Zeilen 83–91, Fassung f9d9356 vom
// 2021-07-06.
//
// **Es entsteht aus der leeren Zeichenkette und aus `nil`**, nicht aus einem
// eigenen Skriptnamen: `ScriptString::toImage` und `ScriptNil::toImage` geben
// beide eines (`src/script/value.cpp`, Zeilen 369–375 und 450). Der Kanonlauf
// meldete dafür „linear_blend bekam für image1 kein Bild, sondern
// Zeichenkette" — die Zeichenkette war leer
// (#509).
//
// **Die Größe ist die gefragte** — die der Ebene, die diesen Baum ausführt.
// MSE gibt sie über die `Options` an jeden Erzeuger weiter; hier reicht
// [Rasterer.ausBaum] sie durch. Ohne sie käme ein Punkt heraus, und
// `linear_blend` verglichen 1×1 mit 375×523.
func (r *Rasterer) leeresBild() image.Image {
	breite := max(1, r.gefragteGroesse.X)
	hoehe := max(1, r.gefragteGroesse.Y)

	return image.NewNRGBA(image.Rect(0, 0, breite, hoehe))
}
