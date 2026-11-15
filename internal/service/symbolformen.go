// **Die Grundformen des Symboleditors** —
// #890,
// Werkstück 3 von Phase X.
//
// Portiert aus `SymbolBasicShapeEditor::makeCenteredShape`
// (`src/gui/symbol/basic_shape_editor.cpp`, Zeilen 188–297, Fassung f9d9356
// vom 2021-07-06).
//
// # Warum das nachgelesen und nicht hergeleitet ist
//
// Die Zahlen sehen aus, als könnte man sie ausrechnen, und man kann es — aber
// nicht nebenbei. Der Kreis benutzt eine Konstante, die MSE selbst „magic"
// nennt; das Vieleck eine Herleitung über drei Gleichungen und den Sinussatz;
// der Stern eine Näherung, die MSE ausdrücklich als **nur für n ≤ 6 exakt**
// bezeichnet. Wer das nachrechnet statt nachzulesen, bekommt eine Form, die
// fast stimmt — und „fast" sieht man einem Symbol an.

package service

import (
	"math"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Formart ist eine der vier Grundformen.
type Formart string

const (
	FormKreis    Formart = "Kreis"
	FormRechteck Formart = "Rechteck"
	FormVieleck  Formart = "Vieleck"
	FormStern    Formart = "Stern"
)

// kreisgriff ist die Länge der Bézierhandgriffe eines Kreisviertels.
//
//	const double kr = 0.5522847498f; // = 4/3 * (sqrt(2) - 1)
//
// Ebd., Zeile 200. Vier Bögen nähern einen Kreis genau dann gut, wenn die
// Griffe diese Länge haben.
const kreisgriff = 0.5522847498

// Grundform baut eine Form um eine Mitte.
//
// `halb` ist der halbe Durchmesser in beiden Richtungen — bei MSE `r`. `ecken`
// gilt nur für Vieleck und Stern; `eng` heißt, dass die Form gleichseitig ist
// (Kreis statt Ellipse, Quadrat statt Rechteck) und ändert allein den **Namen**
// — die Gleichseitigkeit stellt der Aufrufer über `halb` her, so wie MSEs
// Editor es über die Maus tut.
func Grundform(art Formart, mitte, halb msepaket.Vektor, ecken int, eng bool) msepaket.Symbolteil {
	form := msepaket.Symbolteil{Art: msepaket.TeilForm, Kombination: msepaket.KombinationUeberdecken}

	switch art {
	case FormKreis:
		form.Name, form.Punkte = kreisname(eng), kreispunkte(mitte, halb)
	case FormRechteck:
		form.Name, form.Punkte = rechteckname(eng), rechteckpunkte(mitte, halb)
	case FormStern:
		form.Name, form.Punkte = "Stern", sternpunkte(mitte, halb, ecken)
	case FormVieleck:
		form.Name, form.Punkte = vieleckname(ecken), vieleckpunkte(mitte, halb, ecken)
	}

	return form
}

// kreispunkte legt vier Punkte auf die Achsen und gibt ihnen ihre Bögen.
//
// **Reihenfolge und Vorzeichen sind die von MSE** (ebd., Zeilen 201–205):
// rechts, oben, links, unten. Die Sperre ist `size` — die beiden Griffe eines
// Kreispunktes zeigen entgegengesetzt und sind gleich lang; wer sie freigäbe,
// verlöre die Rundung beim ersten Ziehen.
func kreispunkte(mitte, halb msepaket.Vektor) []msepaket.Kontrollpunkt {
	kx, ky := kreisgriff*halb.X, kreisgriff*halb.Y

	return []msepaket.Kontrollpunkt{
		bogen(mitte.X+halb.X, mitte.Y, 0, ky, 0, -ky),
		bogen(mitte.X, mitte.Y-halb.Y, kx, 0, -kx, 0),
		bogen(mitte.X-halb.X, mitte.Y, 0, -ky, 0, ky),
		bogen(mitte.X, mitte.Y+halb.Y, -kx, 0, kx, 0),
	}
}

// rechteckpunkte legt die vier Ecken.
func rechteckpunkte(mitte, halb msepaket.Vektor) []msepaket.Kontrollpunkt {
	return []msepaket.Kontrollpunkt{
		gerade(mitte.X-halb.X, mitte.Y-halb.Y),
		gerade(mitte.X+halb.X, mitte.Y-halb.Y),
		gerade(mitte.X+halb.X, mitte.Y+halb.Y),
		gerade(mitte.X-halb.X, mitte.Y+halb.Y),
	}
}

// vieleckpunkte legt `ecken` Punkte auf einen Kreis.
//
// # Die Herleitung, wie MSE sie aufschreibt
//
//	double alpha = 2 * M_PI / n;      // Innenwinkel /_aOb
//	double beta  = alpha * ((n+1)/4); // Winkel /_aOc
//	double lb = 1 / cos(beta/2);
//	double ra = lb / sin(beta) * cos(beta/2);
//
// Ebd., Zeilen 238–252. `ra` ist der Umkreisradius, gemessen so, dass die
// **Breite** der Form 2 ist — deshalb steht darunter `y = c.y + (ra - 1) * r.y`
// und nicht einfach `c.y`: die Form wird an ihrer Breite ausgerichtet, nicht
// an ihrem Umkreis.
func vieleckpunkte(mitte, halb msepaket.Vektor, ecken int) []msepaket.Kontrollpunkt {
	if ecken < 3 {
		return nil
	}

	alpha, ra := vieleckmasse(ecken)
	y := mitte.Y + (ra-1)*halb.Y

	punkte := make([]msepaket.Kontrollpunkt, 0, ecken)
	for stelle := range ecken {
		theta := alpha * float64(stelle)
		punkte = append(punkte, gerade(
			mitte.X+ra*halb.X*math.Sin(theta),
			y-ra*halb.Y*math.Cos(theta)))
	}

	return punkte
}

// sternpunkte legt außen und innen abwechselnd.
//
//	double lc = ra * sin(alpha) / cos(alpha/2);
//	double delta = alpha * 2;
//	double ld = ra * sin(delta) / cos(delta/2);
//	double rb = (ld - 2 * lc * (lc/ld)) * ra / lb;
//
// Ebd., Zeilen 271–279. **MSE nennt das selbst eine Näherung**: „This is only
// correct for n<=6, but gives acceptable results for higher n". Der Satz wird
// mitportiert, nicht stillschweigend verbessert (§15, Regel 4).
func sternpunkte(mitte, halb msepaket.Vektor, ecken int) []msepaket.Kontrollpunkt {
	if ecken < 3 {
		return nil
	}

	alpha, ra := vieleckmasse(ecken)
	lb := 1 / math.Cos(alpha*float64((ecken+1)/4)/2)
	y := mitte.Y + (ra-1)*halb.Y

	lc := ra * math.Sin(alpha) / math.Cos(alpha/2)
	delta := alpha * 2
	ld := ra * math.Sin(delta) / math.Cos(delta/2)
	rb := (ld - 2*lc*(lc/ld)) * ra / lb

	punkte := make([]msepaket.Kontrollpunkt, 0, 2*ecken)

	for stelle := range ecken {
		theta := alpha * float64(stelle)
		punkte = append(punkte, gerade(
			mitte.X+ra*halb.X*math.Sin(theta),
			y-ra*halb.Y*math.Cos(theta)))

		theta = alpha * (float64(stelle) + 0.5)
		punkte = append(punkte, gerade(
			mitte.X+rb*halb.X*math.Sin(theta),
			y-rb*halb.Y*math.Cos(theta)))
	}

	return punkte
}

// vieleckmasse gibt Innenwinkel und Umkreisradius.
//
// **Die Ganzzahldivision `(n+1)/4` ist Absicht** und steht so im Original: sie
// zählt die Seiten zwischen dem obersten Punkt und dem, der die Seite berührt.
func vieleckmasse(ecken int) (alpha, ra float64) {
	alpha = 2 * math.Pi / float64(ecken)
	beta := alpha * float64((ecken+1)/4)
	lb := 1 / math.Cos(beta/2)

	return alpha, lb / math.Sin(beta) * math.Cos(beta/2)
}

// bogen baut einen Kurvenpunkt mit beiden Griffen.
func bogen(x, y, xDavor, yDavor, xDanach, yDanach float64) msepaket.Kontrollpunkt {
	return msepaket.Kontrollpunkt{
		Stelle:      msepaket.Vektor{X: x, Y: y},
		GriffDavor:  msepaket.Vektor{X: xDavor, Y: yDavor},
		GriffDanach: msepaket.Vektor{X: xDanach, Y: yDanach},
		KurveDavor:  true,
		KurveDanach: true,
		Sperre:      "size",
	}
}

// gerade baut einen Eckpunkt ohne Bögen.
func gerade(x, y float64) msepaket.Kontrollpunkt {
	return msepaket.Kontrollpunkt{
		Stelle: msepaket.Vektor{X: x, Y: y},
		Sperre: "free",
	}
}

// kreisname, rechteckname und vieleckname geben der Form ihren Namen.
//
// **Ein Dreieck heißt nicht „Vieleck"** — MSE benennt drei bis sechs Ecken
// einzeln (ebd., Zeilen 222–231). Die Namen sind **deutsch**, weil sie in die
// Teilliste geschrieben werden und dort ein Oberflächentext sind; MSE schreibt
// sie über seinen eigenen Übersetzungskatalog (`_TYPE_`), nicht fest.
func kreisname(eng bool) string {
	if eng {
		return "Kreis"
	}

	return "Ellipse"
}

func rechteckname(eng bool) string {
	if eng {
		return "Quadrat"
	}

	return "Rechteck"
}

func vieleckname(ecken int) string {
	switch ecken {
	case 3:
		return "Dreieck"
	case 4:
		return "Raute"
	case 5:
		return "Fünfeck"
	case 6:
		return "Sechseck"
	default:
		return "Vieleck"
	}
}
