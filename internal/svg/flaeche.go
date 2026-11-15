package svg

import (
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// flaeche schreibt eine gefüllte Fläche oder einen Rahmen.
func (s *schreiber) flaeche(aus *strings.Builder, ebene render.Ebene) {
	fuellung := alsFarbe(ebene.Farbe)
	if ebene.Verlauf != nil {
		fuellung = s.verlauf(ebene)
	}

	// **Ein Rahmen aus vier Bahnen und kein `stroke`.** MSEs vier Breiten
	// dürfen verschieden sein (`left_width`, `top_width`, `right_width`,
	// `bottom_width`), und eine Strichbreite kennt nur eine.
	if ebene.Rand.Rahmt(ebene.Rahmen.Breite, ebene.Rahmen.Hoehe) {
		s.bahnen(aus, ebene, fuellung)

		return
	}

	rechteck(aus, ebene.Feld, ebene.Rahmen, ebene.Ecke, fuellung)
}

// bahnen malt die vier Seiten eines Rahmens.
//
//	bool clip = style().left_width < style().width && …
//	if (clip) { … } else { dc.DrawRectangle(style_rect); }
//
// `ColorValueViewer::draw`, `src/render/value/color.cpp`, Zeilen 47–50,
// Fassung f9d9356 vom 2021-07-06. **Alle vier** müssen kleiner sein als das
// Feld; MSEs Vorgabe ist 100000, und dann wird gefüllt statt gerahmt — das
// entscheidet [render.Randbreiten.Rahmt], damit es an **einer** Stelle steht.
//
// # Und die runden Ecken schneiden auch den Rand
//
//	if (image.Point{X: x, Y: y}).In(innen) { return false }
//	return inRundemRechteck(rahmen, ebene.Ecke*vergroesserung, x, y)
//
// `inFlaechenform`, `internal/render/farbflaeche.go`: der Punkt muss nach dem
// Abzug des Innenrechtecks **weiterhin** im gerundeten Außenrechteck liegen —
// und MSE malt den Rand ebenso über ein gerundetes Außenrechteck.
//
// **Vier gewöhnliche Rechtecke haben diesen Beschnitt nicht.** Bei 100 × 100,
// Radius 10 und Randbreite 2 liegt (0,0) in der oberen Bahn und außerhalb des
// Bogens — √((0−10)² + (0−10)²) = 14,14 > 10; SVG malte dort Randfarbe, der
// Rasterer nicht. Angestrichen vom Nachlauf zum Zweitblick über Phase P
// (#860).
func (s *schreiber) bahnen(aus *strings.Builder, ebene render.Ebene, fuellung string) {
	r, rand := ebene.Rahmen, *ebene.Rand

	// **Der Beschnitt nur bei runden Ecken.** Ein `<clipPath>` aus einem
	// eckigen Rechteck schnitte ein Rechteck aus einem Rechteck — ein Knoten,
	// der nichts tut, und eine Zusage, die niemand prüfen kann.
	if ebene.Ecke > 0 {
		kennung := s.name("randform")

		s.vorlagen.WriteString(`<clipPath id="` + kennung + `">`)
		rechteck(&s.vorlagen, ebene.Feld, r, ebene.Ecke, fuellung)
		s.vorlagen.WriteString("</clipPath>")

		fmt.Fprintf(aus, `<g clip-path="url(#%s)">`, kennung)

		defer aus.WriteString("</g>")
	}

	// **Ohne runde Ecken an der einzelnen Bahn.** Der Rasterer prüft die
	// Eckenform an der **äußeren** Kante; eine Bahn mit `rx` rundete auch ihre
	// inneren Ecken, und das malte eine Form, die der Rasterer nicht malt.
	for _, bahn := range []render.Rechteck{
		{Links: r.Links, Oben: r.Oben, Breite: r.Breite, Hoehe: rand.Oben},
		{Links: r.Links, Oben: r.Oben + r.Hoehe - rand.Unten, Breite: r.Breite, Hoehe: rand.Unten},
		{Links: r.Links, Oben: r.Oben + rand.Oben, Breite: rand.Links, Hoehe: r.Hoehe - rand.Oben - rand.Unten},
		{
			Links: r.Links + r.Breite - rand.Rechts, Oben: r.Oben + rand.Oben,
			Breite: rand.Rechts, Hoehe: r.Hoehe - rand.Oben - rand.Unten,
		},
	} {
		rechteck(aus, ebene.Feld, bahn, 0, fuellung)
	}
}

// rechteck schreibt ein einzelnes `<rect>`.
func rechteck(aus *strings.Builder, feld string, r render.Rechteck, ecke float64, fuellung string) {
	fmt.Fprintf(aus, `<rect data-feld="%s" x="%s" y="%s" width="%s" height="%s"`,
		entschaerfe(feld), zahl(r.Links), zahl(r.Oben), zahl(r.Breite), zahl(r.Hoehe))

	// **`rx` **und** `ry`.** Ohne das zweite folgt es zwar dem ersten, aber
	// nur, solange kein Betrachter das anders auslegt; MSEs `radius` ist rund
	// und nicht elliptisch.
	if ecke > 0 {
		fmt.Fprintf(aus, ` rx="%s" ry="%s"`, zahl(ecke), zahl(ecke))
	}

	fmt.Fprintf(aus, ` fill="%s"/>`, fuellung)
}

// verlauf legt einen `<linearGradient>` an und gibt seine Fundstelle.
//
// **`gradientUnits="userSpaceOnUse"` und kein Einheitsquadrat**: der Winkel
// aus [render.Verlauf] ist ein Winkel in Vorlagenpunkten, und in einem auf
// 0…1 normierten Kasten wäre er ein anderer, sobald die Fläche nicht
// quadratisch ist.
func (s *schreiber) verlauf(ebene render.Ebene) string {
	kennung := s.name("verlauf")
	r := ebene.Rahmen

	// **Null Grad heißt von links nach rechts** ([render.Verlauf.Winkel]).
	x1, y1, x2, y2 := verlaufsachse(r, ebene.Verlauf.Winkel)

	fmt.Fprintf(&s.vorlagen,
		`<linearGradient id="%s" gradientUnits="userSpaceOnUse" `+
			`x1="%s" y1="%s" x2="%s" y2="%s">`,
		kennung, zahl(x1), zahl(y1), zahl(x2), zahl(y2))

	for _, stufe := range ebene.Verlauf.Stufen {
		fmt.Fprintf(&s.vorlagen, `<stop offset="%s" stop-color="%s"/>`,
			zahl(stufe.Stelle), alsFarbe(stufe.Farbe))
	}

	s.vorlagen.WriteString("</linearGradient>")

	return "url(#" + kennung + ")"
}
