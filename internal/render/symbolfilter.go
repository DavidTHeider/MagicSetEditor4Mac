package render

import (
	"image"
	"image/color"
	"math"
)

// **Erst der Filter macht aus der Maske ein Bild.**
//
// Der Rasterer gibt drei Bereiche her — außen, Rand, innen —, und die Variante
// eines Setsymbols sagt, welche Farbe jeder bekommt. **Außen ist
// durchsichtig**: ein Symbol steht auf einer Karte und nicht auf einem Kasten
// (`src/render/symbol/filter.cpp`, Zeilen 90–94, Fassung f9d9356 vom
// 2021-07-06, #323).

// Symbolfilter gibt einer Stelle der Maske ihre Farbe.
//
// **Die Stelle kommt als Anteil**, nicht in Bildpunkten: `filter.color((double)x
// / width, (double)y / height, point)`. Ein Verlauf ist damit von der Größe des
// Bildes unabhängig.
type Symbolfilter interface {
	Farbe(x, y float64, bereich Symbolbereich) color.NRGBA
}

// FaerbeSymbol macht aus einer Maske ein Bild.
func FaerbeSymbol(maske Symbolmaske, filter Symbolfilter) *image.NRGBA {
	bild := image.NewNRGBA(image.Rect(0, 0, maske.Breite, maske.Hoehe))

	for y := range maske.Hoehe {
		for x := range maske.Breite {
			bild.SetNRGBA(x, y, filter.Farbe(
				float64(x)/float64(maske.Breite),
				float64(y)/float64(maske.Hoehe),
				maske.Bei(x, y)))
		}
	}

	return bild
}

// Vollfilter ist `SolidFillSymbolFilter` — je eine Farbe für innen und Rand.
type Vollfilter struct{ Fuellfarbe, Randfarbe color.NRGBA }

// Farbe erfüllt [Symbolfilter].
func (f Vollfilter) Farbe(_, _ float64, bereich Symbolbereich) color.NRGBA {
	switch bereich {
	case SymbolInnen:
		return f.Fuellfarbe

	case SymbolRand:
		return f.Randfarbe

	default:
		return color.NRGBA{}
	}
}

// Verlaufsfilter ist `LinearGradientSymbolFilter`.
//
//	double t = fabs((x-cx)*(ex-cx) + (y-cy)*(ey-cy)) / len;
//	return min(1., max(0., t));
//
// Zeilen 152–161. `len` ist der quadrierte Abstand von der Mitte zum Ende;
// ist er null, gilt eins — sonst teilte MSE durch null.
type Verlaufsfilter struct {
	Fuellfarbe1, Fuellfarbe2 color.NRGBA
	Randfarbe1, Randfarbe2   color.NRGBA

	MitteX, MitteY float64
	EndeX, EndeY   float64
}

// Farbe erfüllt [Symbolfilter].
func (f Verlaufsfilter) Farbe(x, y float64, bereich Symbolbereich) color.NRGBA {
	anteil := f.anteil(x, y)

	switch bereich {
	case SymbolInnen:
		return gemischt(f.Fuellfarbe1, f.Fuellfarbe2, anteil)

	case SymbolRand:
		return gemischt(f.Randfarbe1, f.Randfarbe2, anteil)

	default:
		return color.NRGBA{}
	}
}

func (f Verlaufsfilter) anteil(x, y float64) float64 {
	laenge := quadrat(f.EndeX-f.MitteX) + quadrat(f.EndeY-f.MitteY)
	if laenge == 0 {
		laenge = 1
	}

	anteil := math.Abs((x-f.MitteX)*(f.EndeX-f.MitteX)+
		(y-f.MitteY)*(f.EndeY-f.MitteY)) / laenge

	return math.Min(1, math.Max(0, anteil))
}

// Kreisfilter ist `RadialGradientSymbolFilter`.
//
//	return sqrt((sqr(x - 0.5) + sqr(y - 0.5)) * 2);
//
// Zeilen 184–186. **Ohne Abschneiden** — in der Ecke wird der Anteil eins, und
// weiter kommt die Fläche nicht.
type Kreisfilter struct {
	Fuellfarbe1, Fuellfarbe2 color.NRGBA
	Randfarbe1, Randfarbe2   color.NRGBA
}

// Farbe erfüllt [Symbolfilter].
func (f Kreisfilter) Farbe(x, y float64, bereich Symbolbereich) color.NRGBA {
	anteil := math.Sqrt((quadrat(x-0.5) + quadrat(y-0.5)) * 2)

	switch bereich {
	case SymbolInnen:
		return gemischt(f.Fuellfarbe1, f.Fuellfarbe2, anteil)

	case SymbolRand:
		return gemischt(f.Randfarbe1, f.Randfarbe2, anteil)

	default:
		return color.NRGBA{}
	}
}

// gemischt ist MSEs `lerp` über alle vier Kanäle.
func gemischt(von, nach color.NRGBA, anteil float64) color.NRGBA {
	return color.NRGBA{
		R: kanal(von.R, nach.R, anteil),
		G: kanal(von.G, nach.G, anteil),
		B: kanal(von.B, nach.B, anteil),
		A: kanal(von.A, nach.A, anteil),
	}
}

func kanal(von, nach uint8, anteil float64) uint8 {
	return uint8(math.Round(float64(von) + (float64(nach)-float64(von))*anteil))
}

func quadrat(wert float64) float64 { return wert * wert }
