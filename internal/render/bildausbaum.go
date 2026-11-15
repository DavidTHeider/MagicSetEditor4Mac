package render

import "image"

// **Ein Bildbaum als eigenes Bild** —
// #660.
//
// # Wofür
//
// `write_image_file` bekommt bei MSE **entweder** eine Karte **oder** einen
// Bildwert:
//
//	ScriptObject<CardP>* card = dynamic_cast<ScriptObject<CardP>*>(input.get());
//	if (card) {
//	  image = conform_image(export_bitmap(ei.set, card->getValue()).ConvertToImage(), options);
//	} else {
//	  image = input->toImage()->generateConform(options);
//	}
//
// `src/script/functions/export.cpp`, Zeilen 414–421, Fassung f9d9356 vom
// 2021-07-06. Die zweite Gestalt braucht `magic-spoiler` und
// `magic-symbol-extractor` für das Setsymbol.
//
// # Warum es hier steht
//
// Weil der Baum hier ausgeführt wird. [Rasterer.ausBaum] tut es seit jeher —
// für Ebenen, Masken und Symbolbilder; exportiert war es nicht, weil außerhalb
// niemand danach gefragt hat.

// BildAusBaum führt einen Bildbaum aus und bringt ihn auf die verlangte Größe.
//
//	if ((iw == options.width && ih == options.height) || (options.width == 0 && options.height == 0)) { … }
//	else if (options.height == 0) { int h = options.width * ih / iw; … }
//	else if (options.width == 0)  { int w = options.height * iw / ih; … }
//	else … resample(image, options.width, options.height);
//
// `conform_image`, `src/gfx/generated_image.cpp`, Zeilen 30–66, Fassung
// f9d9356 vom 2021-07-06.
//
// **Null heißt „so lassen".** Steht nur eine der beiden Zahlen da, folgt die
// andere dem Seitenverhältnis; stehen beide, wird gedehnt — `ASPECT_STRETCH`
// ist die Vorgabe von `Options` (`generated_image.hpp`, Zeile 30).
func (r *Rasterer) BildAusBaum(
	baum Bildbaum, breite, hoehe int,
) (bild *image.NRGBA, befund string) {
	// **Bild und Befund, nicht Bild oder Befund.** [Rasterer.ausBaum] gibt
	// beides: eine noch nicht gebaute Verknüpfung zeichnet das Ausgangsbild
	// und meldet es dazu (`bildausfuehrung.go`, Zeilen 207–215). Wer die
	// Meldung verwürfe, schriebe eine halb falsche Beigabe, ohne dass die
	// Ausfuhr etwas sagt — angestrichen vom Zweitblick durch Codex am
	// 2026-08-21.
	quelle, meldung := r.ausBaum(baum, image.Point{X: breite, Y: hoehe})
	if quelle == nil {
		return nil, meldung
	}

	zielbreite, zielhoehe := angepassteGroesse(quelle.Bounds(), breite, hoehe)

	ziel := image.NewNRGBA(image.Rect(0, 0, zielbreite, zielhoehe))
	feld := ziel.Bounds()

	for y := feld.Min.Y; y < feld.Max.Y; y++ {
		for x := feld.Min.X; x < feld.Max.X; x++ {
			r.punkt(ziel, x, y,
				// **Parität, und zwar mit Absicht**: der Symbolbaum gehört nicht
				// zum Kartenraster, für den die Güten gelten (Ω.4).
				abtasten(quelle, feld, x, y, GueteMittel), 1, nil)
		}
	}

	return ziel, meldung
}

// angepassteGroesse rechnet die Zielgröße nach `conform_image`.
func angepassteGroesse(
	masse image.Rectangle, breite, hoehe int,
) (zielbreite, zielhoehe int) {
	quellbreite, quellhoehe := masse.Dx(), masse.Dy()

	// **Ohne Quelle keine Rechnung.** Ein Bild ohne Fläche käme sonst mit einer
	// Division durch null; MSE fängt es über `image.Ok()` ab.
	if quellbreite <= 0 || quellhoehe <= 0 {
		return max(1, breite), max(1, hoehe)
	}

	switch {
	case breite <= 0 && hoehe <= 0:
		return quellbreite, quellhoehe

	case hoehe <= 0:
		return breite, max(1, breite*quellhoehe/quellbreite)

	case breite <= 0:
		return max(1, hoehe*quellbreite/quellhoehe), hoehe

	default:
		return breite, hoehe
	}
}
