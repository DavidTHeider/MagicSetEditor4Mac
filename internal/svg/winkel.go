package svg

import "math"

// gerade sagt, ob ein Winkel ein Vielfaches von 90 Grad ist.
//
// **Dieselbe Toleranz wie `render.geraderWinkel`.** Ein Winkel kommt aus einem
// Skript und darf `270.0000001` heißen; ein Vergleich auf Gleichheit machte
// daraus einen schrägen.
//
// **Und dieselbe Rechnung ein zweites Mal — mit Absicht.** Sie steht in
// `render` unexportiert, und sie exportieren hieße, eine innere Regel des
// Rasterers zur Schnittstelle zu machen, obwohl hier nur ein Befund davon
// abhängt. Was auseinanderlaufen könnte, ist die **Toleranz**; dass sie
// dieselbe ist, hält `internal/werkzeuge/svgebene_test.go`.
func gerade(grad float64) bool {
	umlauf := math.Mod(math.Mod(grad, 360)+360, 360)

	return math.Abs(umlauf-math.Round(umlauf/90)*90) <= 0.001
}
