package render

import "math"

// **Wie flach oder wie hoch eine Symbolzeichnung ist** —
// #735.
//
// # Wofür
//
// `render_symbol` klemmt das Bild auf dieses Verhältnis, statt das Symbol zu
// dehnen: wer ein 200 × 100 großes Setsymbol verlangt und ein rundes bekommt,
// bekommt es rund und mittig, nicht in die Breite gezogen
// (`src/render/symbol/viewer.cpp`, Zeilen 18–25, Fassung f9d9356 vom
// 2021-07-06).
//
// # Warum die Grenzen exakt sein müssen
//
// Die Grenze einer Bézierkurve liegt **innerhalb** ihrer Griffe. Wer die
// Kontrollpunkte nähme, bekäme ein zu großes Rechteck und damit ein anderes
// Verhältnis — MSE löst dafür die Ableitung
// (`bezier_bounds`, `src/gfx/bezier.cpp`, Zeilen 105–130).

// bereich sind die Grenzen einer Zeichnung — MSEs `Bounds`.
//
// **Der Nullwert ist leer, nicht Null**: `Bounds()` beginnt bei ±unendlich,
// damit das erste `update` beide Ecken setzt (`src/data/symbol.hpp`, Zeile 113,
// Fassung f9d9356 vom 2021-07-06).
type bereich struct{ minX, minY, maxX, maxY float64 }

// leererBereich ist `Bounds()`.
func leererBereich() bereich {
	return bereich{
		minX: math.Inf(1), minY: math.Inf(1),
		maxX: math.Inf(-1), maxY: math.Inf(-1),
	}
}

// nimm ist `Bounds::update(const Vector2D&)`.
func (b bereich) nimm(p Punkt) bereich {
	return bereich{
		minX: math.Min(b.minX, p.X), minY: math.Min(b.minY, p.Y),
		maxX: math.Max(b.maxX, p.X), maxY: math.Max(b.maxY, p.Y),
	}
}

// vereint ist `Bounds::update(const Bounds&)`.
func (b bereich) vereint(a bereich) bereich {
	return bereich{
		minX: math.Min(b.minX, a.minX), minY: math.Min(b.minY, a.minY),
		maxX: math.Max(b.maxX, a.maxX), maxY: math.Max(b.maxY, a.maxY),
	}
}

// Seitenverhaeltnis ist Breite durch Höhe der Zeichnung — `Symbol::aspectRatio`.
//
//	double margin_x = min(0.4999, max(0., min(bounds.min.x, 1-bounds.max.x)));
//	double margin_y = min(0.4999, max(0., min(bounds.min.y, 1-bounds.max.y)));
//	double delta = 2 * (margin_y - margin_x);
//	if (delta > 0) return 1 / (1 - delta); else return 1 + delta;
//
// `src/data/symbol.cpp`, Zeilen 342–358, Fassung f9d9356 vom 2021-07-06.
//
// **Gemessen wird der Rand, nicht das Symbol.** Je Richtung zählt der kleinere
// der beiden Abstände zum Bildrand; ist der senkrechte größer, liegt das Symbol
// flach. Die Klemme auf 0,4999 fängt den Strich in der Mitte ab — ohne sie
// teilte die Rechnung durch null.
//
// **Ohne Teile kommt 1 heraus.** Die Grenzen bleiben bei ±unendlich, beide
// Ränder fallen auf die Klemme, und ihre Differenz ist null.
func Seitenverhaeltnis(teile []Symbolteil) float64 {
	grenzen := grenzenVon(teile, Punkt{}, Matrix{A: 1, D: 1})

	randX := rand(grenzen.minX, grenzen.maxX)
	randY := rand(grenzen.minY, grenzen.maxY)

	unterschied := 2 * (randY - randX)
	if unterschied > 0 {
		return 1 / (1 - unterschied)
	}

	return 1 + unterschied
}

// rand ist der geklemmte Abstand zum Bildrand in einer Richtung.
func rand(kleinste, groesste float64) float64 {
	return math.Min(0.4999, math.Max(0, math.Min(kleinste, 1-groesste)))
}

// grenzenVon ist `calculateBounds` über eine Liste von Teilen.
func grenzenVon(teile []Symbolteil, ursprung Punkt, m Matrix) bereich {
	grenzen := leererBereich()

	for _, teil := range teile {
		grenzen = grenzen.vereint(grenzenTeil(teil, ursprung, m))
	}

	return grenzen
}

// grenzenTeil ist `SymbolPart::calculateBounds` für einen Teil.
func grenzenTeil(teil Symbolteil, ursprung Punkt, m Matrix) bereich {
	switch teil.Art {
	case TeilForm:
		return grenzenForm(teil.Punkte, ursprung, m)

	case TeilSymmetrie:
		return grenzenSymmetrie(teil, ursprung, m)

	default:
		return grenzenVon(teil.Teile, ursprung, m)
	}
}

// grenzenForm ist `SymbolShape::calculateBounds`.
//
//	for (int i = 0 ; i < (int)points.size() ; ++i)
//	  bounds.update(segment_bounds(origin, m, *getPoint(i), *getPoint(i + 1)));
//
// `src/data/symbol.cpp`, Zeilen 214–220, Fassung f9d9356 vom 2021-07-06.
// **Der Zug ist geschlossen** — der letzte Punkt hat den ersten als Nachfolger.
func grenzenForm(punkte []Symbolpunkt, ursprung Punkt, m Matrix) bereich {
	grenzen := leererBereich()

	for stelle := range punkte {
		von := punkte[stelle]
		nach := punkte[(stelle+1)%len(punkte)]

		grenzen = grenzen.vereint(grenzenStrecke(von, nach, ursprung, m))
	}

	return grenzen
}

// grenzenStrecke ist `segment_bounds` — Gerade oder Kurve.
func grenzenStrecke(von, nach Symbolpunkt, ursprung Punkt, m Matrix) bereich {
	anfang := gelegt(von.Stelle, ursprung, m)
	ende := gelegt(nach.Stelle, ursprung, m)

	grenzen := leererBereich().nimm(anfang).nimm(ende)

	if !von.KurveDanach {
		return grenzen
	}

	// **Die Kurve wird im Zielraum gebildet**, nicht im Ausgangsraum: MSE
	// verschiebt die vier Kontrollpunkte erst und rechnet dann die Beiwerte
	// (`bezier_bounds`, Zeilen 107–115).
	kurve := kurveAus(
		Symbolpunkt{Stelle: anfang, GriffDanach: m.angewandt(von.GriffDanach)},
		Symbolpunkt{Stelle: ende, GriffDavor: m.angewandt(nach.GriffDavor)},
	)

	for _, t := range append(
		nullstellen(3*kurve.a.X, 2*kurve.b.X, kurve.c.X),
		nullstellen(3*kurve.a.Y, 2*kurve.b.Y, kurve.c.Y)...,
	) {
		if t >= 0 && t <= 1 {
			grenzen = grenzen.nimm(kurve.stelleAn(t))
		}
	}

	return grenzen
}

// gelegt ist `origin + p * m` — **ohne Rundung**.
//
// **Nicht [versetzt]**: der rundet auf ganze Bildpunkte, weil er Ecken für
// `wxPoint` liefert. Hier wird im Einheitsquadrat gerechnet, und dort machte
// eine Rundung aus jeder Zeichnung ein Quadrat — genau das ist beim ersten
// Anlauf herausgekommen: jedes Verhältnis war 1.
func gelegt(stelle, ursprung Punkt, m Matrix) Punkt {
	return plus(ursprung, m.angewandt(stelle))
}

// nullstellen ist `solve_quadratic`, samt seinem Rückfall auf `solve_linear`.
//
//	if (a == 0) return solve_linear(b, c, roots);
//	double d = b*b - 4*a*c;
//	if (d < 0) return 0;
//
// `src/gfx/polynomial.cpp`, Zeilen 15–38, Fassung f9d9356 vom 2021-07-06.
//
// **`solve_linear(0, 0)` gibt die Null zurück** und nicht „keine Lösung" —
// jede Zahl löst `0 = 0`, und MSE nimmt willkürlich die 0. Das ist hier kein
// Detail ohne Wirkung: bei einer Kurve ohne Krümmung in einer Richtung landet
// damit der Anfangspunkt noch einmal in den Grenzen, und der steht ohnehin drin.
func nullstellen(a, b, c float64) []float64 {
	if a == 0 {
		if b == 0 {
			if c == 0 {
				return []float64{0}
			}

			return nil
		}

		return []float64{-c / b}
	}

	unter := b*b - 4*a*c
	if unter < 0 {
		return nil
	}

	wurzel := math.Sqrt(unter)

	return []float64{(-b - wurzel) / (2 * a), (-b + wurzel) / (2 * a)}
}

// grenzenSymmetrie ist `SymbolSymmetry::calculateBounds`.
//
// **Dieselbe Rechnung wie beim Zeichnen** (`symbolraster.go`, Zeilen 249–288),
// und das ist Absicht: liefen die beiden auseinander, klemmte `render_symbol`
// auf ein Verhältnis, das die Zeichnung nicht hat.
func grenzenSymmetrie(teil Symbolteil, ursprung Punkt, m Matrix) bereich {
	kopien := teil.Kopien
	if teil.Spiegelart == SpiegelungAchse {
		kopien = kopien / 2 * 2
	}

	grenzen := leererBereich()

	if kopien < 1 {
		return grenzen
	}

	spiegelwinkel := 2 * math.Atan2(teil.Griff.Y, teil.Griff.X)

	for i := range kopien {
		winkel := float64(i) * 2 * math.Pi / float64(kopien)

		drehung := Matrix{
			A: math.Cos(winkel), B: -math.Sin(winkel),
			C: math.Sin(winkel), D: math.Cos(winkel),
		}

		if teil.Spiegelart == SpiegelungAchse && i%2 != 0 {
			gesamt := winkel + spiegelwinkel
			drehung = Matrix{
				A: math.Cos(gesamt), B: math.Sin(gesamt),
				C: math.Sin(gesamt), D: -math.Cos(gesamt),
			}
		}

		grenzen = grenzen.vereint(grenzenVon(teil.Teile,
			plus(ursprung, m.angewandt(minus(teil.Mitte, drehung.angewandt(teil.Mitte)))),
			drehung.verkettet(m)))
	}

	return grenzen
}
