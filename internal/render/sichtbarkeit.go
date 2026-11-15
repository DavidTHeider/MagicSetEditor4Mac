package render

import "math"

// Sichtbar sagt, ob eine Ebene überhaupt betreten wird.
//
//	bool ValueViewer::isVisible() const {
//	  return getStyle()->visible
//	    && bounding_box.width > 0
//	    && bounding_box.height > 0
//	    && fabs(bounding_box.x) < 100000
//	    && fabs(bounding_box.y) < 100000;
//	}
//
// `src/render/value/viewer.cpp`, Zeilen 36–42, Fassung f9d9356 vom 2021-07-06.
// MSE fragt das **vor** dem Zeichnen (`src/render/card/viewer.cpp`, Zeile 62)
// — und damit vor jedem Bild und jeder Maske, die die Ebene bräuchte.
//
// # Warum das mehr ist als eine Ersparnis
//
// **Eine unsichtbare Ebene darf nichts melden.** So verbirgt eine Vorlage ein
// Feld: `magic-m15-mainframe-dfc.mse-style` gibt `art mirror` die Höhe 0,
// wenn die Karte keinen Stempel trägt — und lässt daneben eine `mask:`
// stehen, die auf `image_masks/mythic_stamp_none.png` zeigt. Diese Datei gibt
// es nicht und soll es nicht geben; MSE fragt nie nach ihr, und dieses
// Programm meldete sie
// (#442).
//
// **`visible` steht nicht hier**, sondern schon im Plan: der Planbauer legt
// eine unsichtbare Ebene gar nicht erst an.
//
// # Eine Prüfung für jede Art
//
// Sie gilt für **jede** Art — für den Text etwa schreibt
// `magic-m15-saga.mse-style` `width: {if is_saga() then 0 else 153}` —, und
// sie muss kommen, bevor eine Maske geladen wird.
// **Ausgeführt (`Sichtbar`), weil `internal/app` dieselbe Frage stellt.** MSE
// stellt sie auch zweimal — beim Zeichnen und beim Bestimmen der bedienbaren
// Felder (`src/gui/control/card_editor.cpp`, Zeilen 155–157) — und beide Male
// mit **derselben** Antwort. Zwei Fassungen liefen auseinander, sobald eine
// von beiden nachgezogen wird.
//
// **Über dem Rechteck und nicht über der Ebene**: MSE fragt es am
// `bounding_box`, und die Fläche eines Feldes ohne Inhalt hat keine Ebene
// ([Feldflaeche]).
func Sichtbar(rahmen Rechteck) bool {
	return rahmen.Breite > 0 && rahmen.Hoehe > 0 &&
		math.Abs(rahmen.Links) < 100000 && math.Abs(rahmen.Oben) < 100000
}
