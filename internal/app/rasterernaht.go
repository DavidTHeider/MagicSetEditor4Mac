// Die Naht zum Rasterer.
//
// **Eine eigene Datei**, weil `anwendung.go` an der Zeilengrenze aus §2 stand
// — und weil der Schnitt der Frage folgt: dort **was die Anwendung hält**,
// hier **was sie zum Zeichnen braucht**.

package app

import (
	"context"
	"image"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Rasterer zeichnet einen Plan — die Naht zu `internal/render`.
type Rasterer interface {
	Zeichne(
		ctx context.Context, plan *render.Plan, vergroesserung float64,
		guete render.Guete,
	) (
		*image.NRGBA, []render.Befund, error,
	)

	// SymbolAlsBild rastert ein einzelnes Symbol — für `symbols_to_html`
	// (#658).
	//
	// **Hier und nicht in einer zweiten Schnittstelle daneben**: es ist
	// derselbe Rasterer, mit derselben Bildquelle und derselben Schriftsuche.
	// Zwei Nähte auf dasselbe Ding ließen zu, dass eine Attrappe die eine
	// belegt und die andere nicht.
	SymbolAlsBild(symbol render.Symbol, masse render.Symbolmasse) (
		bild *image.NRGBA, befund string,
	)

	// BildAusBaum löst einen Bildbaum auf — für die SVG-Ausgabe
	// (#852).
	//
	// **Aus demselben Grund hier wie [Rasterer.SymbolAlsBild]**: derselbe
	// Rasterer, dieselbe Bildquelle. Damit gehen Umfärbung, Mischbild und
	// Symbolzeichnung im SVG durch die Maschine, die auch das PNG malt.
	BildAusBaum(baum render.Bildbaum, breite, hoehe int) (
		bild *image.NRGBA, befund string,
	)
}
