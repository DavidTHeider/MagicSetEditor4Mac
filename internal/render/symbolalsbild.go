package render

import "image"

// **Ein einzelnes Symbol als eigenes Bild** —
// #658.
//
// # Wofür
//
// `symbols_to_html` legt für jedes Manasymbol eine PNG-Datei neben die Ausgabe:
//
//	Image img = symbol_font.getImage(size, sym);
//	img.SaveFile(fn.GetFullPath());
//
// `src/script/functions/export.cpp`, Zeilen 193–199, Fassung f9d9356 vom
// 2021-07-06.
//
// # Warum es hier steht, und nicht in `symbolbild.go`
//
// Weil es Rasterung ist — und weil es etwas **anderes** rastert als die Datei
// nebenan: dort steht die Vektorzeichnung eines `symbol_variation`, hier das
// Bild eines Zeichens aus einer `.mse-symbol-font`. Die beiden teilen den
// Namen „Symbol" und sonst nichts. Bisher gab es das Symbol nur **in** einer Karte —
// [Rasterer.zeichneSymbole] setzt es an eine Stelle in einer Zeile. Eine
// zweite Rechnung in `service` liefe auseinander, sobald eine von beiden etwas
// lernt (§10); deshalb teilen sich beide Wege [Symbolmasse.kasten],
// [Rasterer.ausBaum] und [Rasterer.symboltext].

// SymbolAlsBild rastert ein Symbol in ein Bild seiner eigenen Größe.
//
//	Image resampled_image((int)(actual_size.GetWidth()  * size / img_size),
//	                      (int)(actual_size.GetHeight() * size / img_size), false);
//	if (!resampled_image.Ok()) return Image(1,1);
//
// `SymbolInFont::getImage`, `src/data/symbol_font.cpp`, Zeilen 116–129,
// Fassung f9d9356 vom 2021-07-06 — dieselbe Formel, die [Symbolmasse.kasten]
// rechnet. **Ohne den Abstand**:
// `image font size` bestimmt das Bild, der Abstand nur den Platz daneben.
//
// **Ein fehlendes Bild gibt einen Befund und kein leeres Bild.** Anders als
// beim Zeichnen einer Karte, wo eine Lücke bleibt und die übrigen Symbole
// stehen (§8): hier entstünde eine **Datei**, die Vorlage verwiese darauf, und
// im Browser stünde ein leerer Rahmen.
func (r *Rasterer) SymbolAlsBild(
	symbol Symbol, masse Symbolmasse,
) (bild *image.NRGBA, befund string) {
	kasten := masse.kasten(&symbol, r.quelle)

	breite := max(1, int(kasten.Bildbreite))
	hoehe := max(1, int(kasten.Bildhoehe))

	// **Bild und Befund, nicht Bild oder Befund** — derselbe Fall wie in
	// [Rasterer.BildAusBaum], und der Zweitblick hat beide genannt.
	quelle, meldung := r.ausBaum(symbol.Bild, image.Point{X: breite, Y: hoehe})
	if quelle == nil {
		return nil, meldung
	}

	ziel := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))
	feld := ziel.Bounds()

	for y := feld.Min.Y; y < feld.Max.Y; y++ {
		for x := feld.Min.X; x < feld.Max.X; x++ {
			r.punkt(ziel, x, y,
				// Parität wie in [Rasterer.BildAusBaum] — außerhalb des
				// Kartenrasters gibt es keine Güte.
				abtasten(quelle, feld, x, y, GueteMittel), 1, nil)
		}
	}

	// **Und die Ziffer darauf.** Derselbe Weg wie auf der Karte; was er nicht
	// findet — eine fehlende Schrift —, kommt als Befund zurück.
	//
	// **Ohne Maske und ohne Stauchung**: beide gehören zur Ebene, in der ein
	// Symbol auf einer Karte steht. Eine eigene Datei hat keine.
	if fehlend := r.symboltext(ziel, Ebene{Deckung: 1}, nil, feld,
		&symbol, kasten, masse.Groesse, 0, 0, 1, gestauchtVon(0, 1)); len(fehlend) > 0 {
		return nil, fehlend[0]
	}

	return ziel, meldung
}
