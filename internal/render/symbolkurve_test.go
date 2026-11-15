package render_test

import (
	"math"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Eine Strecke wird in gerade Stücke zerlegt, eine Gerade nicht.**
//
//	out.push_back(origin + p0.pos * m);
//	if (p0.segment_after == SEGMENT_CURVE) {
//	  BezierCurve curve(p0,p1);
//	  curve_subdivide(curve, p0.pos, p1.pos, 0, 1, origin, m, out, 5);
//	}
//
// `segment_subdivide`, `src/gfx/bezier.cpp`, Zeilen 84–95, Fassung f9d9356 vom
// 2021-07-06. **Der Endpunkt gehört nicht dazu** — er ist der Anfang des
// nächsten Stücks, und der Zug ist geschlossen
// (#323).
func TestEineGeradeBleibtEinPunkt(t *testing.T) {
	von := render.Symbolpunkt{Stelle: render.Punkt{X: 0, Y: 0}}
	nach := render.Symbolpunkt{Stelle: render.Punkt{X: 1, Y: 0}}

	punkte := render.Zerlege(von, nach, render.Punkt{}, render.Skaliert(100))
	if len(punkte) != 1 {
		t.Fatalf("%d Punkte, erwartet einen: %v", len(punkte), punkte)
	}

	if punkte[0].X != 0 || punkte[0].Y != 0 {
		t.Errorf("der Punkt liegt bei %+v, erwartet den Ursprung", punkte[0])
	}
}

// **Eine Kurve bekommt Zwischenpunkte, und ihr erster ist der Anfang.**
//
// Wie viele es werden, entscheidet die Schwelle aus `curve_subdivide`: der
// Winkel zwischen den beiden Hälften mal dem Quadrat ihres Abstands, gegen
// 0,0001 (Zeilen 63–82). Ein Viertelkreis von (0,0) nach (1,1) läuft damit auf
// eine zweistellige Zahl von Stücken hinaus — geprüft wird, dass sie zwischen
// den Enden liegen und der Reihe nach kommen.
func TestEineKurveBekommtZwischenpunkte(t *testing.T) {
	von := render.Symbolpunkt{
		Stelle:      render.Punkt{X: 0, Y: 0},
		GriffDanach: render.Punkt{X: 0.55, Y: 0},
		KurveDanach: true,
	}
	nach := render.Symbolpunkt{
		Stelle:     render.Punkt{X: 1, Y: 1},
		GriffDavor: render.Punkt{X: 0, Y: -0.55},
	}

	punkte := render.Zerlege(von, nach, render.Punkt{}, render.Skaliert(100))
	if len(punkte) < 8 {
		t.Fatalf("%d Punkte, erwartet mindestens acht", len(punkte))
	}

	if punkte[0].X != 0 || punkte[0].Y != 0 {
		t.Errorf("der erste Punkt liegt bei %+v, erwartet den Ursprung", punkte[0])
	}

	for stelle := 1; stelle < len(punkte); stelle++ {
		if punkte[stelle].X < punkte[stelle-1].X || punkte[stelle].Y < punkte[stelle-1].Y {
			t.Fatalf("Punkt %d (%+v) läuft zurück hinter %+v",
				stelle, punkte[stelle], punkte[stelle-1])
		}
	}

	// **Der Endpunkt gehört nicht dazu** — der letzte Zwischenpunkt liegt
	// davor, und ein Punkt auf (100,100) wäre der Anfang des nächsten Stücks.
	letzter := punkte[len(punkte)-1]
	if letzter.X == 100 && letzter.Y == 100 {
		t.Errorf("der letzte Punkt ist der Endpunkt: %+v", letzter)
	}

	// **Und jeder Punkt liegt auf einem Bildpunkt** (`to_int`).
	for _, punkt := range punkte {
		if punkt.X != math.Trunc(punkt.X) || punkt.Y != math.Trunc(punkt.Y) {
			t.Errorf("der Punkt %+v liegt zwischen zwei Bildpunkten", punkt)
		}
	}
}
