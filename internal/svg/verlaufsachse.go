package svg

import (
	"math"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// verlaufsachse rechnet Winkel und Rahmen in die beiden Endpunkte um.
//
// # Warum die Achse durch die Mitte geht
//
// Ein Verlauf ist eine Richtung und eine Strecke. Die Richtung kommt aus dem
// Winkel; die Strecke muss so lang sein, dass die Stufen 0 und 1 die
// **Randpunkte** des Rechtecks in dieser Richtung treffen — sonst liefe der
// Verlauf schon vor der Kante aus.
//
// Die Länge ist deshalb die Ausdehnung des Rechtecks in Richtung der Achse:
//
//	|Breite · cos| + |Höhe · sin|
//
// Das ist dieselbe Rechnung, die `linearblende.go` im Rasterer über den
// Skalarprodukt-Bereich macht — dort punktweise, hier einmal für die beiden
// Enden.
func verlaufsachse(r render.Rechteck, grad float64) (x1, y1, x2, y2 float64) {
	bogen := grad * math.Pi / 180
	cos, sin := math.Cos(bogen), math.Sin(bogen)

	mitteX := r.Links + r.Breite/2
	mitteY := r.Oben + r.Hoehe/2

	laenge := math.Abs(r.Breite*cos) + math.Abs(r.Hoehe*sin)

	return mitteX - cos*laenge/2, mitteY - sin*laenge/2,
		mitteX + cos*laenge/2, mitteY + sin*laenge/2
}
