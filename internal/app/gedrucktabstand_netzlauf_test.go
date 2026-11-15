//go:build netzlauf

// **Wie der Abstand zur gedruckten Karte gerechnet wird** — die Rechnung zu
// `scryfallabstand_netzlauf_test.go`.
//
// Eigene Datei, weil jener Lauf mit ihr über 400 Zeilen ginge (§2). Der
// Schnitt folgt der Frage: dort steht, **was verglichen wird**, hier, **wie
// gerechnet wird**.
package app

import (
	"image"
	"math"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// vermisseGegenGedruckte rechnet den Abstand nach Bereichen.
//
// # Was ausgespart wird und warum
//
// **Die Illustration.** Sie ist bei uns das Bild, das im Set steht, und bei
// Scryfall die gedruckte — zwei verschiedene Bilder derselben Karte, und ihr
// Abstand sagt nichts über den Rasterer. Das Werkstück verlangt die
// Aussparung ausdrücklich.
//
// # Was der Abstand **nicht** ist
//
// Kein Maß für Richtigkeit. Er trägt jeden Unterschied zwischen dem echten
// Rahmen und `magic-m15`s Nachbau — andere Schrift, andere Kantenglättung,
// andere Farbprofile. **Er ist eine Karte des Geländes, keine Note.**
func vermisseGegenGedruckte(
	name string, unser *image.NRGBA, herkunft *render.Herkunft,
	gedruckt image.Image, kunst image.Rectangle,
) gedrucktervergleich {
	// **Nur die gemeinsame Fläche.** 375 × 523 auf 488 Punkte Breite gibt 681
	// Zeilen, Scryfall liefert 680 — die letzte Zeile hat kein Gegenüber, und
	// sie mitzuzählen verglich unsere Karte mit dem Rand des Bildes.
	breite := min(unser.Bounds().Dx(), gedruckt.Bounds().Dx())
	hoehe := min(unser.Bounds().Dy(), gedruckt.Bounds().Dy())

	var text, rahmen, gesamt summenquadrat

	uMin, gMin := unser.Bounds().Min, gedruckt.Bounds().Min

	for y := range hoehe {
		for x := range breite {
			if (image.Point{X: x, Y: y}).In(kunst) {
				continue
			}

			quadrat := punktquadratGegen(unser, gedruckt, uMin, gMin, x, y)

			gesamt.nimm(quadrat)

			if herkunft.Art(x, y) == render.ArtText {
				text.nimm(quadrat)
			} else {
				rahmen.nimm(quadrat)
			}
		}
	}

	return gedrucktervergleich{
		name:   name,
		gesamt: gesamt.wurzel(),
		text:   text.wurzel(),
		rahmen: rahmen.wurzel(),
	}
}

// summenquadrat sammelt Quadrate und gibt ihre Wurzel je Kanal.
type summenquadrat struct {
	summe  float64
	punkte int
}

func (s *summenquadrat) nimm(quadrat float64) {
	s.summe += quadrat
	s.punkte++
}

// wurzel gibt den quadratischen Mittelwert — oder null ohne Punkte.
//
// **Null ohne Punkte heißt „nichts gemessen" und nicht „kein Abstand".** Wer
// die Zahl liest, liest daneben, über wie viele Punkte sie geht.
func (s *summenquadrat) wurzel() float64 {
	if s.punkte == 0 {
		return 0
	}

	return math.Sqrt(s.summe / float64(s.punkte*3))
}

// punktquadratGegen gibt die Summe der Kanalquadrate an einer Stelle.
//
// **Ohne Alpha**, wie im Kanonlauf: ein gedrucktes Bild ist deckend, und ein
// Alphakanal, der überall gleich ist, verdünnte den Abstand um ein Viertel.
func punktquadratGegen(
	unser *image.NRGBA, gedruckt image.Image, uMin, gMin image.Point, x, y int,
) float64 {
	a := unser.NRGBAAt(uMin.X+x, uMin.Y+y)

	rot, gruen, blau, _ := gedruckt.At(gMin.X+x, gMin.Y+y).RGBA()

	var summe float64

	for _, paar := range [3][2]float64{
		{float64(a.R), float64(rot >> 8)},
		{float64(a.G), float64(gruen >> 8)},
		{float64(a.B), float64(blau >> 8)},
	} {
		abweichung := paar[0] - paar[1]
		summe += abweichung * abweichung
	}

	return summe
}
