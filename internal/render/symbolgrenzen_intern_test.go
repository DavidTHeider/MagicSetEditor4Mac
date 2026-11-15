package render

import (
	"math"
	"testing"
)

// **`solve_quadratic` samt seinem Rückfall auf `solve_linear`** —
// #735.
//
//	UInt solve_linear(double a, double b, double* root) {
//	  if (a == 0) { if (b == 0) { root[0] = 0; return 1; } else return 0; }
//	  else { root[0] = -b / a; return 1; }
//	}
//	UInt solve_quadratic(double a, double b, double c, double* roots) {
//	  if (a == 0) return solve_linear(b, c, roots);
//	  double d = b*b - 4*a*c;
//	  if (d < 0) return 0;
//	  roots[0] = (-b - sqrt(d)) / (2*a);
//	  roots[1] = (-b + sqrt(d)) / (2*a);
//	  return 2;
//	}
//
// `src/gfx/polynomial.cpp`, Zeilen 15–38, Fassung f9d9356 vom 2021-07-06.
//
// **Der Test steht im Paket**, weil die Funktion keine Zusage an den Aufrufer
// ist, sondern eine gerechnete Zwischenstufe. Über [Seitenverhaeltnis] wären
// drei ihrer fünf Zweige nur mit Kurven erreichbar, deren Zahlen niemand mehr
// nachrechnet — und ein Test, den niemand nachrechnet, prüft die Umsetzung
// gegen sich selbst.
func TestDieNullstellenEinerQuadratischenGleichung(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		a, b, c  float64
		erwartet []float64
	}{
		{
			name: "zwei Lösungen", a: 1, b: 0, c: -1,
			erwartet: []float64{-1, 1},
		},
		{
			// **Die doppelte Nullstelle kommt zweimal.** MSE gibt sie so
			// zurück, und wer sie zu einer zusammenzöge, änderte die Zahl der
			// geprüften Stellen.
			name: "eine doppelte Lösung", a: 1, b: -2, c: 1,
			erwartet: []float64{1, 1},
		},
		{
			name: "keine reelle Lösung", a: 1, b: 0, c: 1,
			erwartet: nil,
		},
		{
			// a = 0 → `solve_linear(b, c)`
			name: "linear", a: 0, b: 2, c: -1,
			erwartet: []float64{0.5},
		},
		{
			// **`solve_linear(0, 0)` gibt die Null**, nicht „keine Lösung":
			// jede Zahl löst `0 = 0`, und MSE nimmt willkürlich die 0.
			name: "durchweg null", a: 0, b: 0, c: 0,
			erwartet: []float64{0},
		},
		{
			name: "waagerecht und nicht auf der Achse", a: 0, b: 0, c: 3,
			erwartet: nil,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			got := nullstellen(fall.a, fall.b, fall.c)
			if len(got) != len(fall.erwartet) {
				t.Fatalf("%v, erwartet %v", got, fall.erwartet)
			}

			for stelle, wert := range got {
				if math.Abs(wert-fall.erwartet[stelle]) > 1e-12 {
					t.Errorf("%d. Lösung = %v, erwartet %v",
						stelle+1, wert, fall.erwartet[stelle])
				}
			}
		})
	}
}
