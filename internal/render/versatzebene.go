package render

// **Der Versatz einer einzelnen Ebene** —
// #1029,
// Werkstück Ω.2.
//
// # Warum die Probe über das ganze Bild nicht genügt
//
// [ProbeVersatz] verschiebt die **ganze** Karte. Am 2026-08-25 fand sie über
// 680 kanonische Karten **null**, die dadurch gewinnen — und daraus las sich,
// der Abstand sei kein Versatz.
//
// **Die Bilder sagen etwas anderes.** Bei `Search for Azcanta` und `Callow
// Jushi` deckt sich die Illustration, und der Regeltext steht im Differenzbild
// **doppelt**, um wenige Punkte senkrecht versetzt. Ein Versatz **einer Ebene**
// verschwindet in einer Probe über alle: der Rahmen, der 93 % der Fläche
// stellt, zieht jeden Kandidaten außer (0,0) sofort ins Minus.
//
// Diese Probe misst deshalb nur die Bildpunkte **einer Art**.
//
// # Was sie nicht sagt
//
// Dasselbe wie [Herkunft]: sie kennt, **wer zuletzt gemalt hat**, nicht wer
// schuld ist. Und sie verschiebt unser Bild gegen die Referenz — wo beide
// dieselbe Ebene an verschiedenen Stellen malen, zählt der Bildpunkt in
// **unserer** Karte zur Art, in der Referenz weiß niemand es.
//
// **Deshalb wird über die Vereinigung gemessen, nicht über den Schnitt**: ein
// Textbildpunkt, den nur wir malen, gehört zum Befund und nicht in die
// Auslassung.

import (
	"fmt"
	"image"
)

// ProbeVersatzJeEbene sucht den Versatz für die Bildpunkte **einer** Art.
//
// `unser` und `referenz` müssen dasselbe Maß haben; `herkunft` gehört zu
// `unser`. Der Rückgabewert zählt in `Punkte` nur die Bildpunkte der Art —
// eine Karte ohne solche gibt `Punkte: 0` und einen Versatz von (0,0).
func ProbeVersatzJeEbene(
	unser, referenz image.Image, herkunft *Herkunft, art Ebenenart, reichweite int,
) (Versatz, error) {
	rahmen, err := pruefeBildpaar(unser, referenz)
	if err != nil {
		return Versatz{}, err
	}

	if reichweite < 0 {
		return Versatz{}, fmt.Errorf("%w: Reichweite %d", ErrPlan, reichweite)
	}

	if herkunft == nil ||
		herkunft.Breite != rahmen.Dx() || herkunft.Hoehe != rahmen.Dy() {
		return Versatz{}, fmt.Errorf(
			"%w: die Herkunft passt nicht zum Bild", ErrPlan)
	}

	innen := image.Pt(rahmen.Dx()-2*reichweite, rahmen.Dy()-2*reichweite)
	if innen.X <= 0 || innen.Y <= 0 {
		return Versatz{}, fmt.Errorf(
			"%w: %dx%d ist zu klein für die Reichweite %d",
			ErrPlan, rahmen.Dx(), rahmen.Dy(), reichweite)
	}

	stellen := stellenDerArt(herkunft, art, innen, reichweite)

	probe := Versatz{Punkte: len(stellen), Reichweite: reichweite}
	if len(stellen) == 0 {
		return probe, nil
	}

	uMin, rMin := unser.Bounds().Min, referenz.Bounds().Min

	// **(0,0) zuerst und als Ausgangswert** — dieselbe Ordnung wie in
	// [ProbeVersatz]: ein Versatz wird nur gemeldet, wenn er den Abstand
	// wirklich verkleinert.
	probe.Ohne = anStellenGemessen(unser, referenz, uMin, rMin, stellen, reichweite, 0, 0)
	probe.Bester = probe.Ohne

	for dy := -reichweite; dy <= reichweite; dy++ {
		for dx := -reichweite; dx <= reichweite; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}

			abstand := anStellenGemessen(unser, referenz, uMin, rMin,
				stellen, reichweite, dx, dy)
			if besser(abstand, dx, dy, probe) {
				probe.Bester, probe.X, probe.Y = abstand, dx, dy
			}
		}
	}

	return probe, nil
}

// stellenDerArt sammelt die Bildpunkte einer Art im inneren Rechteck.
//
// **Die Stellen sind auf das innere Rechteck bezogen**, wie in
// [versetztGemessen]: so misst jeder Kandidat über dieselbe Menge, und Zahlen
// über verschiedene Mengen entstehen gar nicht erst (§10).
func stellenDerArt(
	herkunft *Herkunft, art Ebenenart, innen image.Point, rand int,
) []image.Point {
	stellen := make([]image.Point, 0, innen.X*innen.Y/8)

	for y := range innen.Y {
		for x := range innen.X {
			if herkunft.Art(rand+x, rand+y) == art {
				stellen = append(stellen, image.Pt(x, y))
			}
		}
	}

	return stellen
}

// anStellenGemessen misst den Abstand über die gegebenen Stellen.
func anStellenGemessen(
	unser, referenz image.Image, uMin, rMin image.Point,
	stellen []image.Point, rand, dx, dy int,
) float64 {
	teil := Teilabstand{Punkte: len(stellen)}

	versetzt := uMin.Add(image.Pt(rand+dx, rand+dy))
	fest := rMin.Add(image.Pt(rand, rand))

	for _, stelle := range stellen {
		teil.Quadratsumme += punktquadrat(unser, referenz, versetzt, fest,
			stelle.X, stelle.Y)
	}

	return teil.Abstand()
}
