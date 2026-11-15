// **Die Symmetrie eines Symbols** —
// #890,
// Werkstück 3 von Phase X.
//
// Portiert aus `SymbolViewer::draw` (`src/render/symbol/viewer.cpp`, Zeilen
// 150–190) und `SymbolSymmetry::calculateBounds` (`src/data/symbol.cpp`,
// Zeilen 251–274), Fassung f9d9356 vom 2021-07-06.

package service

import (
	"math"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// SymmetriekopieN gibt die Zahl der Kopien, die wirklich gezeichnet werden.
//
//	int copies = s->kind == SYMMETRY_REFLECTION ? s->copies / 2 * 2 : s->copies;
//
// `src/render/symbol/viewer.cpp`, Zeile 155, Fassung f9d9356 vom 2021-07-06.
//
// **Eine Spiegelung braucht Paare** — Original und Bild. Eine ungerade Zahl
// ließe eines allein, und MSE rundet deshalb ab statt aufzurunden oder zu
// klagen.
func SymmetriekopieN(spiegel msepaket.Symbolteil) int {
	if spiegel.Kopien < 0 {
		return 0
	}

	if spiegel.Spiegelart == msepaket.SpiegelungAchse {
		return spiegel.Kopien / 2 * 2
	}

	return spiegel.Kopien
}

// SymmetrieAbbilden bildet einen Punkt auf seine `kopie` ab.
//
//	p' = ((p - d) * rot + d) * m + o
//
// `src/render/symbol/viewer.cpp`, Zeile 174, Fassung f9d9356 vom 2021-07-06.
// `d` ist die **Mitte** der Symmetrie: gedreht wird um sie, nicht um den
// Ursprung. `m` und `o` sind die Matrix und der Ursprung der umgebenden
// Zeichnung; hier steht der Punkt schon in seinem eigenen Raum, und beide
// entfallen.
//
// **Eine Kopie außerhalb bildet nicht ab.** Der Aufrufer bekommt den Punkt
// zurück, wie er kam — ein Fehler stünde an einer Stelle, die von Geometrie
// handelt.
func SymmetrieAbbilden(
	spiegel msepaket.Symbolteil, kopie int, punkt msepaket.Vektor,
) msepaket.Vektor {
	kopien := SymmetriekopieN(spiegel)
	if kopien <= 0 || kopie < 0 || kopie >= kopien {
		return punkt
	}

	relativ := minus(punkt, spiegel.Mitte)
	winkel := float64(kopie) * 2 * math.Pi / float64(kopien)

	var gedreht msepaket.Vektor

	if spiegel.Spiegelart == msepaket.SpiegelungAchse && kopie%2 == 1 {
		gedreht = gespiegelt(relativ, winkel+2*richtungswinkel(spiegel.Griff))
	} else {
		gedreht = verdreht(relativ, winkel)
	}

	return plus(gedreht, spiegel.Mitte)
}

// verdreht dreht gegen den Uhrzeigersinn.
//
//	Matrix2D rot(cos(a),-sin(a), sin(a),cos(a));
//
// Ebd., Zeile 176. MSE multipliziert **Zeilenvektor mal Matrix**
// (`operator *`, `src/util/vector2d.hpp`, Zeilen 158–160, Fassung f9d9356 vom
// 2021-07-06): `v * m = (dot(v, mx), dot(v, my))`. Wer sie als Spaltenvektoren
// läse, drehte in die falsche Richtung.
func verdreht(v msepaket.Vektor, winkel float64) msepaket.Vektor {
	sin, cos := math.Sincos(winkel)

	return msepaket.Vektor{
		X: v.X*cos - v.Y*sin,
		Y: v.X*sin + v.Y*cos,
	}
}

// gespiegelt spiegelt an der Geraden durch den Ursprung im halben Winkel.
//
//	Matrix2D rot(cos(a+b),sin(a+b), sin(a+b),-cos(a+b));
//
// Ebd., Zeile 182. Die Matrix spiegelt an der Geraden im Winkel `(a+b)/2` —
// deshalb steht in `b` das **Doppelte** des Griffwinkels (`Radians b = 2 *
// s->handle.angle()`, Zeile 152).
func gespiegelt(v msepaket.Vektor, doppelterWinkel float64) msepaket.Vektor {
	sin, cos := math.Sincos(doppelterWinkel)

	return msepaket.Vektor{
		X: v.X*cos + v.Y*sin,
		Y: v.X*sin - v.Y*cos,
	}
}

// richtungswinkel gibt den Winkel eines Vektors.
//
// **Der Nullvektor zeigt nach rechts.** `Vector2D::angle` gibt bei MSE
// `atan2(y, x)`, und `atan2(0, 0)` ist dort null — dieselbe Antwort, nur
// ausgeschrieben.
func richtungswinkel(v msepaket.Vektor) float64 {
	return math.Atan2(v.Y, v.X)
}
