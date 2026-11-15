package service

// **Ein Farbfeld zeichnet eine Fläche** — der `ColorValueViewer` von MSE.
//
//	void ColorValueViewer::draw(RotatedDC& dc) {
//	  dc.SetPen(*wxTRANSPARENT_PEN);
//	  dc.SetBrush(value().value());
//	  … const AlphaMask& alpha_mask = getMask(dc);
//	  if (alpha_mask.isLoaded()) {
//	    dc.DrawImage(alpha_mask.colorImage(value().value()), RealPoint(0,0), style().combine);
//	  } else {
//	    bool clip = … ;
//	    dc.DrawRoundedRectangle(dc.getInternalRect(), style().radius);
//	  }
//	  drawFieldBorder(dc);
//	}
//
// `src/render/value/color.cpp`, Zeilen 18–66, Fassung f9d9356 vom 2021-07-06.
//
// # Was das gekostet hat
//
// **Der schwarze Kartenrand ist ein Farbfeld.** `magic-m15.mse-style` gibt
// `border color` die Maße der ganzen Karte, die Breiten 17/17/17/18, den Radius
// 18 und `mask: border_mask.png`. Bis zu diesem Werkstück zeichnete dieses
// Programm es nicht und meldete stattdessen, das Feld nenne „weder eine Schrift
// noch ein Bild" — 134-mal über die 190 Stile des Packs
// (#270).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// farbebene baut die Fläche eines Farbfeldes.
//
// **Die Maske wird zur Maske der Ebene, nicht zum Bild.** MSE malt bei einer
// geladenen Maske `alpha_mask.colorImage(value)` — ein Bild, das überall die
// Farbe des Feldes trägt und die Deckung der Maske. Eine gefärbte Fläche mit
// dieser Maske ist dasselbe Bild und geht durch denselben Weg wie jede andere
// Ebene.
// **Der Rahmen kommt von außen.** Er wird in [feldebene] einmal gerechnet —
// `left:` und `width:` sind Skripte, und sie zweimal auszuwerten kostete den
// halben Planbau noch einmal.
func farbebene(
	ctx context.Context, feld *msepaket.Feldstil, wert string, beschreibung farbfeld,
	paket string, rahmen render.Rechteck,
	umgebung *msescript.Umgebung, symbole Symbolquelle,
) ([]render.Ebene, []domain.Befund) {
	farbe := farbfeldfarbe(wert, beschreibung)

	ebene := render.Ebene{
		Art:     render.ArtFlaeche,
		Rahmen:  rahmen,
		Farbe:   farbe,
		Deckung: 1,
		Ecke:    masszahlOderNull(ctx, feld.Ecke, Eckenvorgabe, umgebung),
		Rand:    randbreiten(ctx, feld.Rand, umgebung),
	}

	maske, befunde := bildangabe(ctx, feld.Feld, "mask", feld.Maske, umgebung)
	ebene.Maske = alsZeichenbaum(maske, paket, symbole)

	return []render.Ebene{ebene}, befunde
}

// randbreiten rechnet die vier Breiten aus; ohne Angabe wird gefüllt.
//
//	ColorStyle::ColorStyle(const ColorFieldP& field)
//	  : … left_width(100000), right_width (100000)
//	  , top_width (100000), bottom_width(100000)
//
// `src/data/field/color.cpp`, Zeilen 58–63. **Die Vorgabe ist absichtlich
// riesig**: MSEs `clip` verlangt, dass *alle vier* kleiner als das Feld sind,
// und eine einzelne Angabe macht daraus deshalb noch keinen Rahmen. Wer hier
// mit null vorbelegte, ließe von jedem Feld mit nur einer Breite die Füllung
// verschwinden.
//
// Fassung f9d9356 vom 2021-07-06.
func randbreiten(
	ctx context.Context, angabe msepaket.Randangabe, umgebung *msescript.Umgebung,
) *render.Randbreiten {
	if angabe == (msepaket.Randangabe{}) {
		return nil
	}

	breite := func(roh string) float64 {
		if roh == "" {
			return Randvorgabe
		}

		return masszahlOderNull(ctx, roh, Randvorgabe, umgebung)
	}

	return &render.Randbreiten{
		Links:  breite(angabe.Links),
		Rechts: breite(angabe.Rechts),
		Oben:   breite(angabe.Oben),
		Unten:  breite(angabe.Unten),
	}
}
