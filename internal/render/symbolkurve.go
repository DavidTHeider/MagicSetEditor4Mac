package render

import "math"

// **Ein Symbol ist eine Vektorzeichnung, und der Rasterer bekommt sie als
// Plan.**
//
// Die Formen stehen hier noch einmal, obwohl `msepaket` sie schon liest: nach
// §2 kennt `internal/render` nichts aus dem Projekt. Was hier steht, ist die
// Gestalt, wie der Rasterer sie braucht — auf [0,1] normiert, ohne Namen und
// ohne Sperren (#323).

// Punkt ist eine Stelle oder eine Verschiebung in der Zeichnung.
type Punkt struct{ X, Y float64 }

// Symbolpunkt ist ein Kontrollpunkt eines Formzugs.
//
// **Die Griffe sind relativ zur Stelle** — `delta_before` und `delta_after`.
type Symbolpunkt struct {
	Stelle      Punkt
	GriffDavor  Punkt
	GriffDanach Punkt

	// KurveDanach sagt, dass die Strecke zum **nächsten** Punkt eine
	// Bézierkurve ist; KurveDavor gilt für die vom vorigen.
	KurveDanach bool
	KurveDavor  bool
}

// Zerlege macht aus einer Strecke die geraden Stücke, die sie ersetzen.
//
//	out.push_back(origin + p0.pos * m);
//	if (p0.segment_after == SEGMENT_CURVE) {
//	  BezierCurve curve(p0,p1);
//	  curve_subdivide(curve, p0.pos, p1.pos, 0, 1, origin, m, out, 5);
//	}
//
// `segment_subdivide`, `src/gfx/bezier.cpp`, Zeilen 84–95, Fassung f9d9356 vom
// 2021-07-06.
//
// **Der Endpunkt gehört nicht dazu.** Er ist der Anfang des nächsten Stücks,
// und der Zug ist geschlossen — wer ihn mitnähme, zeichnete jeden Punkt
// zweimal.
func Zerlege(von, nach Symbolpunkt, ursprung Punkt, m Matrix) []Punkt {
	punkte := []Punkt{versetzt(von.Stelle, ursprung, m)}
	if !von.KurveDanach {
		return punkte
	}

	kurve := kurveAus(von, nach)

	return unterteile(kurve, von.Stelle, nach.Stelle, 0, 1, ursprung, m, punkte, 5)
}

// Matrix ist MSEs `Matrix2D` — sie dreht, spiegelt und streckt.
//
//	inline Vector2D operator * (const Vector2D& a, const Matrix2D& m) {
//	  return Vector2D(dot(a,m.mx), dot(a,m.my));
//	}
//
// `src/util/vector2d.hpp`, Zeilen 148–164, Fassung f9d9356 vom 2021-07-06.
type Matrix struct{ A, B, C, D float64 }

// Skaliert ist die Matrix, die nur vergrößert — `Matrix2D(size,0, 0,size)`.
func Skaliert(faktor float64) Matrix { return Matrix{A: faktor, D: faktor} }

// angewandt ist `a * m` für einen Vektor.
func (m Matrix) angewandt(a Punkt) Punkt {
	return Punkt{X: a.X*m.A + a.Y*m.B, Y: a.X*m.C + a.Y*m.D}
}

// verkettet ist `a * m` für zwei Matrizen.
func (m Matrix) verkettet(aeussere Matrix) Matrix {
	x := aeussere.angewandt(Punkt{X: m.A, Y: m.B})
	y := aeussere.angewandt(Punkt{X: m.C, Y: m.D})

	return Matrix{A: x.X, B: x.Y, C: y.X, D: y.Y}
}

// bezier sind die Beiwerte von `(x,y) = at³ + bt² + ct + d`.
type bezier struct{ a, b, c, d Punkt }

// kurveAus rechnet die vier Griffe in Beiwerte um.
//
//	c = p0.delta_after * 3.0;
//	b = (p3.pos + p3.delta_before - p0.pos - p0.delta_after) * 3.0 - c;
//	a = (p3.pos - p0.pos) - c - b;
//	d = p0.pos;
//
// `BezierCurve::BezierCurve`, `src/gfx/bezier.cpp`, Zeilen 24–30.
//
// Fassung f9d9356 vom 2021-07-06.
func kurveAus(von, nach Symbolpunkt) bezier {
	c := mal(von.GriffDanach, 3)
	b := minus(mal(minus(plus(nach.Stelle, nach.GriffDavor),
		plus(von.Stelle, von.GriffDanach)), 3), c)
	a := minus(minus(minus(nach.Stelle, von.Stelle), c), b)

	return bezier{a: a, b: b, c: c, d: von.Stelle}
}

// stelleAn ist `pointAt` — `d + (c + (b + a * t) * t) * t`.
func (k bezier) stelleAn(t float64) Punkt {
	return plus(k.d, mal(plus(k.c, mal(plus(k.b, mal(k.a, t)), t)), t))
}

// unterteile ist `curve_subdivide`.
//
// **Die Schwelle entscheidet, nicht die Tiefe.** Unterteilt wird, wo der
// Winkel zwischen den beiden Hälften mal dem Quadrat ihres Abstands über
// 0,0001 liegt; die Tiefe von fünf begrenzt das nur nach unten (Zeilen 63–82).
// Eine flache Kurve bekommt so wenige Punkte, eine enge viele.
//
// **Beide Hälften laufen immer**, auch wenn dieser Mittelpunkt nicht
// mitgenommen wird — sonst fiele eine enge Stelle in einer sonst flachen
// Kurve weg.
func unterteile(
	kurve bezier, von, nach Punkt, t0, t1 float64,
	ursprung Punkt, m Matrix, punkte []Punkt, tiefe int,
) []Punkt {
	if tiefe <= 0 {
		return punkte
	}

	mitte := (t0 + t1) * 0.5
	mittelpunkt := kurve.stelleAn(mitte)

	d0 := minus(von, mittelpunkt)
	d1 := minus(mittelpunkt, nach)

	schwelle := math.Abs(math.Atan2(d0.X, d0.Y)-math.Atan2(d1.X, d1.Y)) *
		abstandsquadrat(von, nach)

	punkte = unterteile(kurve, von, mittelpunkt, t0, mitte, ursprung, m, punkte, tiefe-1)

	if schwelle >= 0.0001 {
		punkte = append(punkte, versetzt(mittelpunkt, ursprung, m))
	}

	return unterteile(kurve, mittelpunkt, nach, mitte, t1, ursprung, m, punkte, tiefe-1)
}

// versetzt rechnet eine Stelle der Zeichnung in Bildpunkte um — `origin + p * m`.
//
// **Und das Ergebnis ist ganzzahlig.** MSE sammelt die Stücke in `wxPoint`,
// und dessen Umrechnung rundet (`to_int`, `src/util/vector2d.hpp`, Zeilen
// 20–22). Die Kante eines Symbols liegt damit auf einem Bildpunkt und nicht
// dazwischen — wer hier mit Nachkommastellen weiterrechnete, bekäme eine
// geglättete Kante, wo MSE eine harte hat.
//
// Fassung f9d9356 vom 2021-07-06.
func versetzt(stelle, ursprung Punkt, m Matrix) Punkt {
	gedreht := m.angewandt(stelle)

	return Punkt{X: gerundet(ursprung.X + gedreht.X), Y: gerundet(ursprung.Y + gedreht.Y)}
}

// gerundet ist MSEs `to_int`.
//
// **Knapp unter der Hälfte**, und der Kommentar dort sagt warum: „to give a
// more consistent result when for instance something like `x/2` is used".
func gerundet(wert float64) float64 {
	if wert > 0 {
		return math.Trunc(wert + 0.4999995)
	}

	return math.Trunc(wert - 0.4999995)
}

func plus(a, b Punkt) Punkt  { return Punkt{X: a.X + b.X, Y: a.Y + b.Y} }
func minus(a, b Punkt) Punkt { return Punkt{X: a.X - b.X, Y: a.Y - b.Y} }

func mal(a Punkt, f float64) Punkt { return Punkt{X: a.X * f, Y: a.Y * f} }

func abstandsquadrat(a, b Punkt) float64 {
	d := minus(a, b)

	return d.X*d.X + d.Y*d.Y
}
