package render

import (
	"image"
	"image/color"
	"strings"
)

// maskenbild ist ein geladenes Graustufenbild oder nichts.
//
// Nichts heißt „keine Maske", und `anteil` gibt dann 1 zurück. Das spart im
// Aufrufer eine Fallunterscheidung an einer Stelle, die für jeden Bildpunkt
// einmal durchlaufen wird.
type maskenbild struct {
	bild image.Image

	// guete ist die Güte des Laufs, aus dem diese Maske stammt.
	//
	// **Am Bild und nicht am Aufruf**, weil [maskenbild.anteil] je Bildpunkt
	// gerufen wird und den Parameter sonst millionenfach durchreichte.
	guete Guete
}

// anteil sagt, wie viel von der Ebene an dieser Stelle durchkommt.
//
// **Der rote Kanal ist die Deckung**, nicht der Alphakanal und nicht die
// Helligkeit. MSE sagt es in einem Kommentar und drei Zeilen:
//
//	// Copy red chanel to alpha
//	Byte* from = img.GetData(), *to = alpha;
//	for (size_t i = 0 ; i < n ; ++i) {
//	  to[i] = from[3*i];
//	}
//
// `src/gfx/mask_image.cpp`, Zeilen 41–44, Fassung f9d9356 vom 2021-07-06.
//
// Ein Alphakanal wäre die andere naheliegende Lesart und ergäbe bei einer
// undurchsichtigen Maske überall 1 — also eine Maske, die nichts tut.
//
// **Auch nicht die übliche Gewichtung nach Helligkeitsempfinden**
// (`0,2126 R + 0,7152 G + 0,0722 B`). Bei einer echten Graustufenmaske ist das
// dasselbe; bei einer farbigen nicht — rot ergäbe 54 statt 255, grün 182
// statt 0. Und `maskenrot` nimmt im selben Paket Rot; zwei Lesarten für
// dieselbe Sache wären eine zu viel
// (#471).
func (m *maskenbild) anteil(rahmen image.Rectangle, x, y int) float64 {
	if m == nil || m.bild == nil {
		return 1
	}

	farbe := abtasten(m.bild, rahmen, x, y, m.guete)

	return begrenzt01(float64(farbe.R) / 255)
}

// maske lädt die Maske einer Ebene.
func (r *Rasterer) maske(
	ebene Ebene, vergroesserung float64, guete Guete,
) (bild *maskenbild, maengel []string) {
	if ebene.Maske.Leer() {
		return nil, nil
	}

	geladen, befunde := r.ausBaumMaengel(ebene.Maske, gefragteGroesse(ebene, vergroesserung))
	if geladen == nil {
		return nil, []string{"die Maske " + ebene.Maske.Beschreibung() + " ist nicht lesbar: " +
			strings.Join(befunde, "; ")}
	}

	// **Auch hier reist der Befund mit.** Eine nicht gebaute Verknüpfung
	// liefert ihr Ausgangsbild samt Warnung; der erste Anlauf verwarf sie,
	// sobald ein Bild herauskam — für die Kinder von `masked_blend` war das
	// behoben, **für die Maske nicht**. Der gleichartige Nebenkanal blieb
	// stehen, angestrichen vom Nachlauf des Zweitblicks. **Je Befund eine
	// Meldung** (QA-03, #1656).
	for _, befund := range befunde {
		maengel = append(maengel, "die Maske "+ebene.Maske.Beschreibung()+": "+befund)
	}

	return &maskenbild{bild: geladen, guete: guete}, maengel
}

// bild zeichnet ein Bild in seinen Rahmen.
func (r *Rasterer) bild(
	ziel *image.NRGBA, ebene Ebene, maske *maskenbild, vergroesserung float64,
	guete Guete,
) []string {
	// **Gezeichnet wird auch, wenn ein Befund mitkommt.** Eine Verknüpfung, die
	// wir nicht können, liefert ihr Ausgangsbild — sichtbar falsch statt
	// unsichtbar fehlend. Der Befund reist trotzdem mit (§8).
	quelle, maengel := r.ausBaumMaengel(ebene.Bild, gefragteGroesse(ebene, vergroesserung))
	if quelle == nil {
		return maengel
	}

	// **Die Mischart der Ebene** ([Ebene.Mischart]). Eine unbekannte wird
	// benannt und normal gezeichnet: sichtbar falsch statt unsichtbar fehlend
	// (§8).
	var mischart func(a, b int) int

	if ebene.Mischart != "" && ebene.Mischart != "normal" {
		rechnung, bekannt := mischarten[ebene.Mischart]
		if !bekannt {
			maengel = append(maengel, "die Mischart "+ebene.Mischart+" ist hier nicht bekannt")
		}

		mischart = rechnung
	}

	rahmen := eingepassterRahmen(skaliert(ebene.Rahmen, vergroesserung), quelle, ebene.Passung)

	// **Die Schnittmenge statt einer Prüfung je Bildpunkt** (#672): was außerhalb
	// des Ziels läge, wird gar nicht erst durchlaufen. Dieselben Punkte, ein
	// Vergleich weniger — 196.125 Mal bei einem Kartenrahmen von 375 × 523.
	sichtbar := rahmen.Intersect(ziel.Bounds())

	// **Einmal abgetastet, dann nur noch gelesen** (#672): derselbe Rahmen in
	// derselben Größe ist für jede Karte des Stils dasselbe Bild.
	abgetastet := r.abgetasteteEbene(quelle, rahmen, guete, ebene.Bild.IstBlatt())

	for y := sichtbar.Min.Y; y < sichtbar.Max.Y; y++ {
		for x := sichtbar.Min.X; x < sichtbar.Max.X; x++ {
			r.punkt(ziel, x, y,
				farbeAusEbene(abgetastet, x-rahmen.Min.X, y-rahmen.Min.Y),
				ebene.Deckung*maske.anteil(rahmen, x, y), mischart)
		}
	}

	return maengel
}

// punktfarbe liest einen Bildpunkt und hält sich am Rand fest.
//
// Festhalten statt Wiederholen: am Rand eines Kartenrahmens sähe eine
// Wiederholung wie ein schmaler Streifen der Gegenseite aus.
func punktfarbe(quelle image.Image, innen image.Rectangle, x, y int) Farbe {
	if x < innen.Min.X {
		x = innen.Min.X
	}

	if x >= innen.Max.X {
		x = innen.Max.X - 1
	}

	if y < innen.Min.Y {
		y = innen.Min.Y
	}

	if y >= innen.Max.Y {
		y = innen.Max.Y - 1
	}

	rot, gruen, blau, alpha := vorgewichteteWerte(quelle, x, y)
	if alpha == 0 {
		return Farbe{}
	}

	// `RGBA()` gibt vorgewichtete Werte in 0…65535. Hier wird zurückgerechnet,
	// weil dieses Paket ohne Vorgewicht arbeitet (`mischen.go`).
	zurueck := func(kanal uint32) uint8 {
		return alsAnteil(float64(kanal) / float64(alpha) * 255)
	}

	return Farbe{
		R: zurueck(rot), G: zurueck(gruen), B: zurueck(blau),
		A: alsAnteil(float64(alpha) / 257),
	}
}

// vorgewichteteWerte gibt dasselbe wie `quelle.At(x, y).RGBA()` — **ohne den
// Umweg über die Schnittstelle** (#672).
//
// # Warum es diese Abkürzung gibt
//
// `At` gibt ein `color.Color` als Schnittstellenwert; `RGBA()` darauf ist ein
// zweiter Sprung. Beides geschieht **viermal je Zielbildpunkt** — die
// bilineare Abtastung liest vier Nachbarn.
//
// Gemessen über 120 Vorschauen einer M15-Karte: `At(…).RGBA()` allein war
// **2,35 s von 18,57 s** — 12,7 % der gesamten Rechenzeit.
//
// # Warum das Ergebnis dasselbe ist
//
// Die beiden Zweige rechnen **Zeichen für Zeichen** das, was `color.NRGBA` und
// `color.RGBA` in ihrem `RGBA()` rechnen (`image/color/color.go`):
//
//	// NRGBA
//	r = uint32(c.R); r |= r << 8; r *= uint32(c.A); r /= 0xff
//	// RGBA — schon vorgewichtet
//	r = uint32(c.R); r |= r << 8
//
// Kein Runden, kein Näherungswert, keine andere Reihenfolge. Was hier
// herauskommt, ist bitgleich — und der Goldbildvergleich (§3) ist der Beleg
// dafür, nicht diese Zeile.
//
// **Alles andere geht den alten Weg.** Ein Bild, dessen Art hier nicht steht,
// wird nicht falsch gelesen, sondern langsam.
func vorgewichteteWerte(quelle image.Image, x, y int) (rot, gruen, blau, alpha uint32) {
	switch bild := quelle.(type) {
	case *image.NRGBA:
		stelle := bild.PixOffset(x, y)
		punkt := bild.Pix[stelle : stelle+4 : stelle+4]

		roh := uint32(punkt[3])
		if roh == 0 {
			return 0, 0, 0, 0
		}

		return vorgewichtet8(punkt[0], roh), vorgewichtet8(punkt[1], roh),
			vorgewichtet8(punkt[2], roh), breit8(punkt[3])

	case *image.RGBA:
		stelle := bild.PixOffset(x, y)
		punkt := bild.Pix[stelle : stelle+4 : stelle+4]

		return breit8(punkt[0]), breit8(punkt[1]), breit8(punkt[2]), breit8(punkt[3])

	// **Und Bilder mit Farbtafel**, denn davon lebt das Pack: die Rahmen von
	// `magic-modules.mse-include` sind PNG mit Tafel. `Paletted.At` schlägt
	// den Eintrag nach und gibt ihn als Schnittstellenwert zurück; hier wird
	// er gleich gelesen.
	case *image.Paletted:
		// **Ohne Prüfung auf einen Wert jenseits der Tafel.** `Paletted.At`
		// greift selbst ungeprüft zu und bricht dort ab
		// (`image/image.go`, `(*Paletted).At`); ein Zweig hier gäbe diesem
		// Bild ein Verhalten, das es ohne den Schnellweg nicht hätte — und er
		// wäre nicht zu prüfen, weil der Vergleichswert nicht zu holen ist.
		switch farbe := bild.Palette[bild.Pix[bild.PixOffset(x, y)]].(type) {
		case color.NRGBA:
			roh := uint32(farbe.A)
			if roh == 0 {
				return 0, 0, 0, 0
			}

			return vorgewichtet8(farbe.R, roh), vorgewichtet8(farbe.G, roh),
				vorgewichtet8(farbe.B, roh), breit8(farbe.A)

		case color.RGBA:
			return breit8(farbe.R), breit8(farbe.G), breit8(farbe.B), breit8(farbe.A)
		}
	}

	return quelle.At(x, y).RGBA()
}

// breit8 macht aus 0…255 den Wert 0…65535, den `RGBA()` gibt.
func breit8(kanal uint8) uint32 {
	wert := uint32(kanal)

	return wert | wert<<8
}

// vorgewichtet8 gewichtet einen Kanal mit dem **rohen** Alpha.
//
//	r = uint32(c.R); r |= r << 8; r *= uint32(c.A); r /= 0xff
//
// `image/color/color.go`. **Das Alpha geht roh ein, nicht verbreitert** — der
// erste Anlauf verbreiterte es und war damit um den Faktor 257 daneben; jeder
// Kanal lief in die Klemmung auf 255. Gefunden von
// `randfestigkeit_intern_test.go`, sofort.
func vorgewichtet8(kanal uint8, rohesAlpha uint32) uint32 {
	return breit8(kanal) * rohesAlpha / 0xff
}

// eingepassterRahmen verkleinert den Rahmen auf das Seitenverhältnis des Bildes.
//
//	} else if (options.preserve_aspect == ASPECT_FIT) {
//	  if (iw * options.height > ih * options.width) { // too much height requested
//	    w = options.width;  h = options.width * ih / iw;
//	  } else {
//	    w = options.height * iw / ih; h = options.height;
//	  }
//	}
//	…
//	RealRect rect(align_in_rect(style.alignment, size, dc.getInternalRect()), size);
//
// `conform_image`, `src/gfx/generated_image.cpp`, Zeilen 49–59, und
// `draw_choice_viewer`, `src/render/value/choice.cpp`, Zeile 83 — Fassung
// f9d9356 vom 2021-07-06.
//
// **Ohne Passung bleibt der Rahmen, wie er ist.** Das ist `ASPECT_STRETCH`, und
// ein Bildfeld — das Kunstfenster — wird bei MSE genau so gezeichnet.
func eingepassterRahmen(rahmen image.Rectangle, quelle image.Image, passung *Bildpassung) image.Rectangle {
	if passung == nil {
		return rahmen
	}

	innen := quelle.Bounds()
	if innen.Empty() || rahmen.Empty() {
		return rahmen
	}

	breite, hoehe := rahmen.Dx(), rahmen.Dy()

	if innen.Dx()*hoehe > innen.Dy()*breite {
		hoehe = breite * innen.Dy() / innen.Dx()
	} else {
		breite = hoehe * innen.Dx() / innen.Dy()
	}

	links := rahmen.Min.X + int(waagerechterVersatz(
		passung.Ausrichtung, float64(rahmen.Dx()), float64(breite)))
	oben := rahmen.Min.Y + int(senkrechterVersatz(
		passung.Ausrichtung, float64(rahmen.Dy()), float64(hoehe)))

	return image.Rect(links, oben, links+breite, oben+hoehe)
}

// mit hängt an die Mängel einer Ebene, was nicht leer ist.
//
// **Eine Liste und keine Fuge**: jeder Mangel wird sein eigener Befund mit
// seiner eigenen Art. Zwei mit »; « gefügte Meldungen ordnete `ArtAus` nach
// dem ersten Treffer (QA-03, #1656).
func mit(maengel []string, weitere ...string) []string {
	for _, meldung := range weitere {
		if meldung != "" {
			maengel = append(maengel, meldung)
		}
	}

	return maengel
}

// gefragteGroesse ist die Größe, in der eine Ebene ihr Bild braucht — MSEs
// `Options.width`/`height` (#509).
func gefragteGroesse(ebene Ebene, vergroesserung float64) image.Point {
	rahmen := skaliert(ebene.Rahmen, vergroesserung)

	return image.Point{X: rahmen.Dx(), Y: rahmen.Dy()}
}
