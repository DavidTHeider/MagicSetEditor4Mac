package render

// zeilenmaske ist, was MSE aus der Maske eines Textfelds für den Umbruch
// liest: je Bildzeile den ersten und den letzten weißen Punkt
// (#2299).
//
//	for (int x = 0 ; x < size.x ; ++x) {
//	  if (alpha[y * size.x + x] >= 128) { // white enough
//	    rights[y] = x;
//	    if (x < lefts[y]) lefts[y] = x;
//	  }
//	}
//
// `AlphaMask::loadRowSizes`, `src/gfx/mask_image.cpp`, Zeilen 145–160;
// `alpha` ist der Rotkanal des Bilds (`AlphaMask::load`, Zeilen 29–46) —
// **roh**, ohne Vormultiplikation ([nrgbaAn]): wxWidgets gibt die Farbwerte
// heraus, wie sie in der Datei stehen, auch unter einem durchsichtigen
// Punkt. Eine Zeile ohne weißen Punkt behält `lefts = size.x` und
// `rights = 0`. Fassung f9d9356 vom 2021-07-06.
//
// **In der Größe des Felds** ([zeilenmaskeAus]); [zeilenmaske.raender] rechnet
// trotzdem um, wie MSE, denn das Feld misst in Bruchteilen und die Maske in
// ganzen Punkten.
type zeilenmaske struct {
	links, rechts []int
	breite, hoehe int
}

// zeilenmaskeAus liest die Maske unter pfad und bringt sie auf die Größe des
// Felds, breite mal hoehe in ganzen Punkten. Gibt es keine oder ist sie nicht
// zu lesen, ist das Ergebnis nil, und jede Zeile hat die ganze Breite — den
// Befund meldet das Zeichnen, das dieselbe Maske lädt ([Rasterer.maske]).
//
// **Erst umgerechnet, dann gemessen**: `ScriptableImage::generate` gibt das
// Bild über `conform_image` in der Größe zurück, nach der gefragt wird
// (`src/script/image.cpp`, Zeile 36), und gefragt wird nach der des Felds
// (`ValueViewer::getMask`, `src/render/value/viewer.cpp`, Zeilen 94–100).
// Beides Fassung f9d9356 vom 2021-07-06. Eine ganz weiße Zeile endet deshalb
// einen Punkt vor dem Rand, nicht `Breite/Maskenbreite` davor.
func zeilenmaskeAus(bilder Bildquelle, pfad string, breite, hoehe float64) *zeilenmaske {
	// **Ein Feld ohne Maß hat keine Maske**: die Messstufe setzt ein Feld,
	// dessen Breite sich nicht ausrechnen ließ, mit `math.MaxFloat64`, damit es
	// nie umbricht. Darauf eine Maske umzurechnen hieße, ein Bild dieser Größe
	// anzulegen.
	if pfad == "" || bilder == nil || !(breite >= 1 && breite <= groessteMaske) ||
		!(hoehe >= 1 && hoehe <= groessteMaske) {
		return nil
	}

	return gelesenesMaskenbild(bilder, pfad, int(breite), int(hoehe))
}

// groessteMaske ist das größte Feld, auf das eine Maske gebracht wird — weit
// über jeder Karte; MSE kennt die Frage nicht, weil es nie ohne Feldmaß misst.
const groessteMaske = 1 << 14

// gelesenesMaskenbild lädt die Maske und misst ihre Zeilen.
func gelesenesMaskenbild(bilder Bildquelle, pfad string, breite, hoehe int) *zeilenmaske {
	bild, err := bilder.Bild(pfad)
	if err != nil || bild == nil || bild.Bounds().Empty() {
		return nil
	}

	kanaele := kanaeleAus(bild).umgerechnet(breite, hoehe)
	maske := &zeilenmaske{
		links:  make([]int, hoehe),
		rechts: make([]int, hoehe),
		breite: breite,
		hoehe:  hoehe,
	}

	for y := range hoehe {
		maske.links[y], maske.rechts[y] = breite, 0

		for x := range breite {
			if kanaele.rot[y*breite+x] >= 128 {
				maske.rechts[y] = x
				maske.links[y] = min(maske.links[y], x)
			}
		}
	}

	return maske
}

// raender gibt den linken und den rechten Rand der Maske auf der Höhe y eines
// Felds von breite mal hoehe — y vom oberen Rand des Felds an.
//
//	if (!lefts || y < 0 || y >= resize.height) {
//	  // no mask, or outside it
//	  return 0;
//	}
//	return lefts[(int)(y * size.y / resize.height)] * resize.width / size.x;
//
// `AlphaMask::rowLeft`, `src/gfx/mask_image.cpp`, Zeilen 162–178, Fassung
// f9d9356 vom 2021-07-06, und `rowRight` ebenso mit `resize.width` außerhalb.
func (m *zeilenmaske) raender(y, breite, hoehe float64) (links, rechts float64) {
	if m == nil || y < 0 || y >= hoehe {
		return 0, breite
	}

	// **Nicht über die letzte Zeile hinaus** — `y < hoehe`, aber das Produkt
	// kann in Gleitkomma auf `size.y` aufrunden.
	zeile := min(int(y*float64(m.hoehe)/hoehe), m.hoehe-1)

	return float64(m.links[zeile]) * breite / float64(m.breite),
		float64(m.rechts[zeile]) * breite / float64(m.breite)
}

// zeilenraum sagt, wo eine Zeile beginnt und wie weit sie reicht — bei MSE
// `lineLeft` und `lineRight`:
//
//	return style.mask.getFromCache().rowLeft(y, dc.getInternalSize()) + style.padding_left;
//	return style.mask.getFromCache().rowRight(y, dc.getInternalSize()) - style.padding_right;
//
// `src/render/text/viewer.cpp`, Zeilen 736–741, Fassung f9d9356 vom
// 2021-07-06. **Gefragt wird mit der Oberkante der Zeile beim Umbruch**
// (`lineRight(dc, style, line.top)`, ebd. Zeile 616), also vor Dehnung und
// Ausrichtung.
type zeilenraum struct {
	maske   *zeilenmaske
	feld    Rechteck
	polster Polster
}

// bei gibt für eine Zeile, deren Oberkante oben unter dem inneren Kasten
// liegt, den Versatz ihres Anfangs und ihre Breite. Ohne Maske ist das der
// innere Kasten selbst.
func (r zeilenraum) bei(oben float64) (versatz, breite float64) {
	links, rechts := r.maske.raender(oben+r.polster.Oben, r.feld.Breite, r.feld.Hoehe)

	return links, max(rechts-r.polster.Rechts-links-r.polster.Links, 0)
}

// abgerueckt rückt eine Zeile so weit hinunter, bis an ihrer Stelle etwas
// passt — je einen Punkt, höchstens bis unter das Feld:
//
//	while (line.top < dc.getHeight() && line_size.width + 1 >= dc.getWidth() - style.padding_right - line.margin_right) {
//	  // nothing fits on this line, move down one pixel
//	  line.top += 1;
//	  line_size.width = line.margin_left + lineLeft(dc, style, line.top);
//	}
//
// `TextViewer::fitLineWidth`, `src/render/text/viewer.cpp`, Zeilen 546–554,
// Fassung f9d9356 vom 2021-07-06. **Mit und ohne Maske**: auch ein Feld, das schmaler ist als sein Polster
// und seine Ränder, schiebt die Zeile hinunter.
func (r zeilenraum) abgerueckt(oben float64, marge Margen) float64 {
	for oben+r.polster.Oben < r.feld.Hoehe {
		links, _ := r.maske.raender(oben+r.polster.Oben, r.feld.Breite, r.feld.Hoehe)
		if marge.Links+links+r.polster.Links+1 < r.feld.Breite-r.polster.Rechts-marge.Rechts {
			break
		}

		oben++
	}

	return oben
}
