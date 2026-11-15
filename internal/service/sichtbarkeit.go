package service

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`visible:` entscheidet, ob eine Ebene gezeichnet wird.**
//
//	bool ValueViewer::isVisible() const {
//	  return getStyle()->visible
//	    && bounding_box.width > 0 && bounding_box.height > 0
//	    && fabs(bounding_box.x) < 100000 && fabs(bounding_box.y) < 100000;
//	}
//
// `src/render/value/viewer.cpp`, Zeilen 36–42, Fassung f9d9356 vom
// 2021-07-06; gerufen in `CardViewer::draw`, Zeile 62.
//
// **Was das kostet, wenn es fehlt:** `magic-m15` schreibt am Kasten für Stärke
// und Widerstandskraft `visible: { card.pt != "" }`. Ohne die Angabe steht auf
// **jeder Verzauberung** ein leerer weißer Kasten unten rechts — auf
// *Old-Money* zu sehen
// (#270).

// sichtbar rechnet `visible:` aus.
//
// **Der Nullwert ist sichtbar**, so wie MSEs `visible(true)`
// (`src/data/field.cpp`, Zeile 106): eine Vorlage ohne die Angabe zeichnet.
//
// **Und was sich nicht ausrechnen lässt, wird gezeichnet.** Eine Ebene
// wegzulassen, weil ihre Bedingung scheiterte, versteckte den Mangel — sichtbar
// falsch schlägt unsichtbar fehlend (§8). Gemeldet wird das nicht: die
// Bedingung steht an **jeder** Ebene, und ein Befund je Ebene machte die Liste
// unlesbar. Wer wissen will, warum eine Ebene fehlt, findet es an ihrem Bild.
//
// Fassung f9d9356 vom 2021-07-06.
func sichtbar(ctx context.Context, angabe string, umgebung *msescript.Umgebung) bool {
	roh := strings.TrimSpace(angabe)
	if roh == "" {
		return true
	}

	wert, err := msescript.AuswertenQuelle(ctx, skriptrumpfOder(roh), umgebung)
	if err != nil {
		return true
	}

	trifft, brauchbar := wert.AlsWahrheit()

	return !brauchbar || trifft
}
