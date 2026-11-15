package render

import (
	"image"
	"image/color"
)

// **Wie ein eingesetztes Bild auf die Größe seines Feldes kommt** —
// #246.
//
// MSE legt beim Einsetzen **das zugeschnittene** Bild ab, nicht das rohe:
//
//	ImageSliceWindow s(…, image, style().getSize(), mask);
//	if (s.ShowModal() == wxID_OK) {
//	  LocalFileName new_image_file = getLocalPackage().newFileName(field().name, "");
//	  img.SaveFile(getLocalPackage().nameOut(new_image_file), wxBITMAP_TYPE_PNG);
//	}
//
// `ImageValueEditor::sliceImage`, `src/gui/value/image.cpp`, Zeilen 42–51,
// Fassung f9d9356 vom 2021-07-06.
//
// **Hier steht die Vorgabewahl, nicht das Fenster.** MSE öffnet einen Dialog,
// in dem sich der Ausschnitt schieben und zoomen lässt; was er beim Öffnen
// zeigt, ist das Folgende. Der Dialog kommt später
// ([0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md)).

// Zuschnittflaeche gibt den Ausschnitt, den MSE beim Öffnen des Fensters zeigt.
//
//	ImageSlice::ImageSlice(const Image& source, const wxSize& target_size)
//	  : selection(0, 0, source.GetWidth(), source.GetHeight())
//	  , allow_outside(false), aspect_fixed(true)
//
//	int diff = selection.width * target_size.GetHeight()
//	         - selection.height * target_size.GetWidth();
//	if ((diff > 0 && prefer != PREFER_WIDTH) || prefer == PREFER_HEIGHT) {
//	  selection.width  -= int(diff / target_size.GetHeight());
//	} else {
//	  selection.height -= int(-diff / target_size.GetWidth());
//	}
//
// `ImageSlice::constrain`, `src/gui/image_slice_window.cpp`, Zeilen 20–52,
// gerufen im Konstruktor des Fensters (Zeile 84). Fassung f9d9356 vom
// 2021-07-06.
//
// **Oben links verankert**, denn `selection.x` und `selection.y` bleiben null.
// Das ist keine Schönheitsentscheidung, sondern MSEs — mittig wäre bei einem
// Porträt die bessere Wahl und wäre trotzdem eine andere Karte.
//
// **Sie ist [Zuschnittwahl.Begrenze] über der ganzen Quelle** und keine
// zweite Rechnung daneben
// (#576).
// Der Dialog ruft sie nach jeder Bewegung, und zwei Fassungen einer
// portierten Regel geben zwei Antworten, sobald jemand eine davon nachzieht.
func Zuschnittflaeche(quelle image.Rectangle, breite, hoehe int) image.Rectangle {
	return Zuschnittwahl{
		Quelle:  quelle,
		Ziel:    image.Pt(breite, hoehe),
		Auswahl: quelle,
	}.Begrenze(BevorzugtKeines)
}

// Zuschnitt schneidet ein Bild zu und rechnet es auf die Größe des Feldes.
//
//	Image target(target_size.GetWidth(), target_size.GetHeight(), false);
//	resample_and_clip(source, target, selection);
//
// `ImageSlice::getSlice`, Zeilen 54–66, Fassung f9d9356 vom 2021-07-06.
//
// **Über [abtasten] und nicht über einen eigenen Filter.** Es ist derselbe, mit
// dem der Rasterer jedes Bild auf sein Feld bringt
// (#532); ein
// zweiter daneben hieße, dass ein eingesetztes Bild anders aussieht als
// dasselbe Bild aus einem Paket.
func Zuschnitt(quelle image.Image, breite, hoehe int) *image.NRGBA {
	return ZuschnittAus(quelle, Zuschnittflaeche(quelle.Bounds(), breite, hoehe), breite, hoehe)
}

// ZuschnittAus schneidet ein Bild auf eine **genannte** Fläche zu.
//
// **Für den Zuschnittdialog** (#576): dort wählt die Auswahl, wer schneidet,
// und nicht die Vorgabewahl. [Zuschnitt] ist derselbe Aufruf mit ihr — damit
// bekommt, wer nichts ändert, genau das bisherige Ergebnis.
//
// **Die Fläche wird hier nicht mehr begrenzt.** Das tut [Zuschnittwahl.Begrenze],
// und zwar dort, wo sie entsteht; eine zweite Begrenzung hier verdeckte eine
// fehlende dort.
func ZuschnittAus(quelle image.Image, flaeche image.Rectangle, breite, hoehe int) *image.NRGBA {
	if flaeche.Empty() || breite <= 0 || hoehe <= 0 {
		return image.NewNRGBA(image.Rectangle{})
	}

	ziel := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))

	for y := range hoehe {
		for x := range breite {
			// Der Zuschnitt bedient den Bildeditor, nicht das Kartenraster.
			farbe := abtasten(ausschnitt{quelle, flaeche},
				image.Rect(0, 0, breite, hoehe), x, y, GueteMittel)
			ziel.SetNRGBA(x, y, color.NRGBA{R: farbe.R, G: farbe.G, B: farbe.B, A: farbe.A})
		}
	}

	return ziel
}

// ausschnitt zeigt einen Teil eines Bildes so, als wäre er das ganze.
//
// **Damit [abtasten] unverändert bleibt.** Es rechnet Zielpunkte auf die
// **Grenzen** der Quelle um; ein zweiter Parameter für „nur dieser Teil" wäre
// eine Zeile im Rasterer, die dort niemand braucht.
type ausschnitt struct {
	bild    image.Image
	flaeche image.Rectangle
}

func (a ausschnitt) Bounds() image.Rectangle { return a.flaeche }
func (a ausschnitt) ColorModel() color.Model { return a.bild.ColorModel() }
func (a ausschnitt) At(x, y int) color.Color { return a.bild.At(x, y) }
