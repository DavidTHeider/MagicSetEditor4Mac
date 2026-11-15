package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ohne Datei zeichnet MSE sein Standardsymbol — ein Quadrat.**
//
//	SymbolP the_symbol;
//	if (filename.empty()) {
//	  the_symbol = default_symbol();
//	} else {
//	  the_symbol = package->readFile<SymbolP>(filename);
//	}
//
// `SymbolToImage::generate`, `src/gfx/generated_image.cpp`, Zeilen 453–462, und
//
//	SymbolShapeP default_symbol_part(double d) {
//	  part->points.push_back(make_intrusive<ControlPoint>(d + .2, d + .2));
//	  part->points.push_back(make_intrusive<ControlPoint>(d + .2, d + .8));
//	  part->points.push_back(make_intrusive<ControlPoint>(d + .8, d + .8));
//	  part->points.push_back(make_intrusive<ControlPoint>(d + .8, d + .2));
//	  part->name = _("Square");
//	}
//
// `src/data/symbol.cpp`, Zeilen 361–378, Fassung f9d9356 vom 2021-07-06.
//
// # Was das gekostet hat
//
// Ein Set ohne eigenes Symbol — und die vier Referenzkarten sind solche — hat
// `set.symbol` leer. Der Rasterer meldete „die Symbolzeichnung liegt nicht vor"
// und zeichnete nichts, während MSE an der Typzeile das Quadrat malt
// (#270).
func TestOhneDateiGiltDasStandardsymbol(t *testing.T) {
	zeichnung := msepaket.Standardsymbol()

	if len(zeichnung.Teile) != 1 {
		t.Fatalf("%d Teile", len(zeichnung.Teile))
	}

	teil := zeichnung.Teile[0]
	if teil.Art != msepaket.TeilForm || teil.Name != "Square" {
		t.Errorf("der Teil ist %q/%q", teil.Art, teil.Name)
	}

	// **Vier Ecken, gegen den Uhrzeigersinn** — genau die Reihenfolge von
	// `default_symbol_part`; eine andere ergäbe bei gerade-ungerader Füllung
	// dieselbe Fläche und wäre trotzdem eine andere Datei.
	erwartet := []msepaket.Vektor{
		{X: 0.2, Y: 0.2}, {X: 0.2, Y: 0.8}, {X: 0.8, Y: 0.8}, {X: 0.8, Y: 0.2},
	}

	if len(teil.Punkte) != len(erwartet) {
		t.Fatalf("%d Punkte", len(teil.Punkte))
	}

	for stelle, punkt := range erwartet {
		if teil.Punkte[stelle].Stelle != punkt {
			t.Errorf("Punkt %d ist %+v, erwartet %+v", stelle, teil.Punkte[stelle].Stelle, punkt)
		}

		// **Geraden und keine Kurven.** `ControlPoint(x, y)` ist MSEs
		// Konstruktor „for straight lines".
		if teil.Punkte[stelle].KurveDanach || teil.Punkte[stelle].KurveDavor {
			t.Errorf("Punkt %d ist eine Kurve", stelle)
		}
	}
}
