package svg

import (
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// huelle ist das `<g>`, in dem eine Ebene steht — oder gar keines.
//
// **Kein `<g>`, wenn nichts daran hängt.** Deckung, Drehung und Maske sind der
// Ausnahmefall; eine Gruppe je Ebene machte aus einer Karte mit vierzig Ebenen
// vierzig überflüssige Knoten.
type huelle struct{ auf, zu string }

// huelle baut die Gruppe für Deckung, Drehung und Maske.
func (s *schreiber) huelle(nummer int, ebene render.Ebene) huelle {
	var teile []string

	if ebene.Deckung < 1 {
		teile = append(teile, `opacity="`+zahl(ebene.Deckung)+`"`)
	}

	if ebene.Winkel != 0 {
		teile = append(teile, s.drehung(ebene))
	}

	if kennung := s.maske(nummer, ebene); kennung != "" {
		teile = append(teile, fmt.Sprintf(`mask="url(#%s)"`, kennung))
	}

	if len(teile) == 0 {
		return huelle{}
	}

	return huelle{auf: "<g " + strings.Join(teile, " ") + ">", zu: "</g>"}
}

// drehung rechnet MSEs Winkel in eine SVG-Umformung.
//
// # Woher die Formel kommt
//
//	return RealPoint(c * x + s * y + origin.x,
//	                -s * x + c * y + origin.y);
//
// `Rotation::tr`, `src/util/rotation.cpp`, Zeilen 37–42, Fassung f9d9356 vom
// 2021-07-06, mit `origin` gleich der linken oberen Ecke des Feldkastens
// (`ValueViewer::getRotation`, `src/render/value/viewer.cpp`, Zeilen 44–46).
//
// Eine SVG-Umformung `matrix(a,b,c,d,e,f)` bildet `(x,y)` auf
// `(a·x + c·y + e, b·x + d·y + f)` ab. Der Vergleich gibt
// `a = cos, b = −sin, c = sin, d = cos, e = links, f = oben` — und das ist
// genau `translate(links, oben) rotate(−Winkel)`.
//
// # Und hier ist SVG genauer als der Rasterer
//
// `render.ebeneGedreht` kann **nur gerade** Winkel und meldet einen schrägen
// als Befund; SVG dreht jeden exakt. Das ist keine stille Besserung, sondern
// die Rückkehr zu MSE, das ebenfalls jeden Winkel dreht — **gemeldet** wird
// es trotzdem, denn für dieselbe Karte sähen PNG und SVG dann verschieden aus
// (§15, Punkt 7). Im vollen Pack stehen schräge Winkel viermal gegen 142
// gerade (#486).
func (s *schreiber) drehung(ebene render.Ebene) string {
	return fmt.Sprintf(`transform="translate(%s,%s) rotate(%s)"`,
		zahl(ebene.Rahmen.Links), zahl(ebene.Rahmen.Oben), zahl(-ebene.Winkel))
}
