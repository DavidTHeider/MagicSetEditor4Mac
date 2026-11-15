package render

import (
	"fmt"
	"image"
)

// Versatz ist das Ergebnis einer Versatzprobe.
//
// # Wozu
//
// Ein Abstand von 27 kann zweierlei heißen: überall steht die falsche Farbe —
// oder alles steht um einen Bildpunkt daneben. **Das sind zwei verschiedene
// Arbeiten**, und der Mittelwert allein trennt sie nicht.
//
// Der Kommentar an [Toleranz] sagt es seit Phase 4: „Was sie nicht deckt: eine
// Verschiebung. Ein um zwei Bildpunkte verrutschter Rahmen ändert an seinen
// Kanten viel und in der Fläche nichts."
//
// # Wie weit die Aussage reicht
//
// **Sie gilt für einen Versatz des ganzen Bildes, nicht für einen einzelner
// Ebenen.** Ein Textkasten, der allein um dreizehn Bildpunkte danebensteht,
// bleibt hier unsichtbar: ihn zu verschieben verbesserte seine Fläche und
// verschlechterte jede andere Kante, und der Mittelwert entschiede gegen ihn.
//
// **Das ist kein gedachter Fall.** Genau dieser Versatz ist am 2026-08-19 an
// zwei Karten gemessen worden
// (#528) —
// während diese Probe über 510 Karten „kein Versatz" meldete. Wer daraus
// schließt, eine ganze Klasse von Platzierungsfehlern sei ausgeschlossen,
// überdehnt sie (§10). Angestrichen vom Zweitblick über Phase I.
//
// Und nur bis [Versatz.Reichweite]: liegt das Beste am Rand des Suchbereichs
// ([Versatz.AmRand]), sagt die Zahl nicht, wie groß der Versatz **ist**,
// sondern nur, dass er mindestens so groß ist. **Ebenso unsichtbar** bleibt
// ein Versatz, der nur den abgeschnittenen Rand trifft, und einer in
// gleichförmigem oder regelmäßig gemustertem Inhalt.
type Versatz struct {
	// X und Y sagen, **um wie viel unser Inhalt neben dem der Referenz sitzt**
	// — positiv nach rechts und nach unten.
	//
	// (−1, 0) heißt: was die Referenz an der Stelle x zeigt, steht bei uns an
	// x−1. Bei Gleichstand gewinnt (0,0): eine Probe, die einen Versatz
	// meldet, wo keiner nötig ist, wäre eine Auskunft über die Reihenfolge der
	// Schleife.
	X, Y int

	// Ohne ist der Abstand bei (0,0), Bester der beim gefundenen Versatz.
	//
	// **Beide über dieselbe Fläche.** Sonst wären sie nicht vergleichbar.
	Ohne, Bester float64

	// Punkte ist die Zahl der je Kandidat verglichenen Bildpunkte.
	Punkte int

	// Reichweite ist der Suchbereich, in dem gesucht wurde.
	Reichweite int
}

// Gewinn sagt, wie viel Abstand der Versatz wegnimmt.
func (v Versatz) Gewinn() float64 { return v.Ohne - v.Bester }

// AmRand sagt, ob der beste Versatz auf dem Rand des Suchbereichs liegt.
//
// **Dann ist der Bereich zu klein**, und die Zahl ist eine Untergrenze. Bei
// Reichweite null gibt es keinen Rand zu treffen.
func (v Versatz) AmRand() bool {
	if v.Reichweite == 0 {
		return false
	}

	return abs(v.X) == v.Reichweite || abs(v.Y) == v.Reichweite
}

// ProbeVersatz sucht den Versatz, bei dem unser Bild der Referenz am nächsten
// kommt.
//
// # Jeder Kandidat über dieselbe Fläche
//
// Ein Versatz verkleinert die Überlappung zweier Bilder. Wer jeden Kandidaten
// über **seine** Überlappung misst, vergleicht Zahlen über verschiedene Mengen
// — der Fehler, den §10 anstreicht. Gemessen wird deshalb für alle über das um
// `reichweite` eingerückte innere Rechteck.
func ProbeVersatz(unser, referenz image.Image, reichweite int) (Versatz, error) {
	rahmen, err := pruefeBildpaar(unser, referenz)
	if err != nil {
		return Versatz{}, err
	}

	if reichweite < 0 {
		return Versatz{}, fmt.Errorf("%w: Reichweite %d", ErrPlan, reichweite)
	}

	innen := image.Pt(rahmen.Dx()-2*reichweite, rahmen.Dy()-2*reichweite)
	if innen.X <= 0 || innen.Y <= 0 {
		return Versatz{}, fmt.Errorf(
			"%w: %dx%d ist zu klein für die Reichweite %d",
			ErrPlan, rahmen.Dx(), rahmen.Dy(), reichweite)
	}

	probe := Versatz{Punkte: innen.X * innen.Y, Reichweite: reichweite}

	uMin, rMin := unser.Bounds().Min, referenz.Bounds().Min

	// **(0,0) zuerst und als Ausgangswert.** Damit gewinnt es jeden
	// Gleichstand — und ein Versatz wird nur gemeldet, wenn er den Abstand
	// wirklich verkleinert.
	probe.Ohne = versetztGemessen(unser, referenz, uMin, rMin, innen, reichweite, 0, 0)
	probe.Bester = probe.Ohne

	for dy := -reichweite; dy <= reichweite; dy++ {
		for dx := -reichweite; dx <= reichweite; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}

			abstand := versetztGemessen(unser, referenz, uMin, rMin,
				innen, reichweite, dx, dy)
			if besser(abstand, dx, dy, probe) {
				probe.Bester, probe.X, probe.Y = abstand, dx, dy
			}
		}
	}

	return probe, nil
}

// besser sagt, ob ein Kandidat den bisher besten ablöst.
//
// # Bei Gleichstand gewinnt der kleinere Versatz
//
// **Und ohne diese Regel gewönne die Reihenfolge der Schleife.** Ein Bild mit
// senkrechten Streifen ändert sich durch eine Verschiebung nach oben gar
// nicht: (−1,−1), (−1,0) und (−1,1) messen dasselbe. Wer den ersten nimmt,
// meldet einen senkrechten Versatz, den es nicht gibt — und die Tabelle
// darüber sagte etwas über `for`-Schleifen statt über Karten.
func besser(abstand float64, dx, dy int, bisher Versatz) bool {
	if abstand < bisher.Bester {
		return true
	}

	return abstand == bisher.Bester &&
		abs(dx)+abs(dy) < abs(bisher.X)+abs(bisher.Y)
}

// versetztGemessen misst den Abstand bei einem Versatz über das innere
// Rechteck.
func versetztGemessen(
	unser, referenz image.Image, uMin, rMin image.Point,
	innen image.Point, rand, dx, dy int,
) float64 {
	teil := Teilabstand{Punkte: innen.X * innen.Y}

	for y := range innen.Y {
		for x := range innen.X {
			teil.Quadratsumme += punktquadrat(unser, referenz,
				uMin.Add(image.Pt(rand+dx, rand+dy)),
				rMin.Add(image.Pt(rand, rand)), x, y)
		}
	}

	return teil.Abstand()
}
