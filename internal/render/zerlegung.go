package render

import (
	"fmt"
	"image"
	"math"
)

// ArtUnbedeckt steht für Bildpunkte, die **keine** Ebene des Plans bedeckt.
//
// Sie werden eigens geführt und nicht weggelassen: eine Zerlegung, deren Teile
// sich nicht zum Ganzen summieren, verbirgt gerade die Fläche, über die
// niemand nachgedacht hat.
const ArtUnbedeckt Ebenenart = "unbedeckt"

// Teilabstand ist der Abstand über einen Teil eines Bildes.
//
// **Zwei Zahlen, und der Abstand ist keine davon.** Er steht als Methode da
// und nicht als Feld, weil ein Aufrufer Teilabstände **zusammenzählt** — über
// viele Karten etwa. Wer dabei zwei Felder addiert und ein drittes stehen
// lässt, trägt eine Zahl weiter, die zu nichts gehört; mit [Teilabstand.Abstand]
// als Rechnung kann das nicht geschehen.
type Teilabstand struct {
	// Punkte ist die Zahl der Bildpunkte dieses Teils. Ohne sie wäre der
	// Abstand eine Zahl ohne Maßstab (§10).
	Punkte int

	// Quadratsumme ist Σ (Δ Kanal)² über diesen Teil. **Sie ist die Größe, die
	// sich addiert** — der Abstand tut das nicht, weil er eine Wurzel ist.
	Quadratsumme float64
}

// Abstand ist das **quadratische** Mittel über die Farbkanäle dieses Teils.
//
// **Dieselbe Rechnung wie `bildabstand` im Kanonlauf**: die Wurzel aus dem
// Mittel über alle Kanäle aller Bildpunkte. Ein anderes Maß wäre mit dem
// Kanonlauf nicht vergleichbar, und zwei Maße nebeneinander sind der Fehler,
// den §10 anstreicht.
func (t Teilabstand) Abstand() float64 {
	if t.Punkte == 0 {
		return 0
	}

	return math.Sqrt(t.Quadratsumme / float64(t.Punkte*kanaele))
}

// Zerlegung sagt, **wo** der Abstand zweier Bilder sitzt.
//
// # Wozu
//
// Ein Abstand von 26,573 über eine ganze Karte ist eine Zahl ohne Adresse: er
// sagt nicht, ob die Textkästen daneben liegen, die Rahmenbilder falsche
// Farben haben oder die Symbole verrutscht sind. Ohne diese Auskunft wird der
// nächste Hebel nach Gefühl gewählt
// (#521).
//
// # Wie die Flächen bestimmt werden — und wie **nicht**
//
// Aus der [Herkunft], die der Rasterer beim Zeichnen mitschreibt: sie sagt für
// jeden Bildpunkt, welche Ebene ihn **zuletzt geändert** hat.
//
// **Nicht aus den Rechtecken des Plans.** Der erste Anlauf tat das und maß am
// Kanon **100 % Bild**: eine gewöhnliche M15-Karte trägt zuoberst ein Bild
// über die ganze Fläche, das jeden Bildpunkt bedeckt und fast keinen bemalt.
// Die Begründung steht bei [Herkunft].
//
// # Was sie nicht sagt
//
// Sie sagt, **wer zuletzt gemalt hat** — nicht, wer schuld ist. Ein
// Rahmenbild, das falsch gezeichnet und danach von einem Textkasten
// überschrieben wird, erscheint als Text. Wer aus einer Zeile dieser Tabelle
// eine Ursache abliest, überdehnt sie (§10).
type Zerlegung struct {
	// Teile führt je Ebenenart ihren Teilabstand, dazu [ArtUnbedeckt].
	Teile map[Ebenenart]Teilabstand

	// Gesamt ist der Abstand über das ganze Bild. Er stimmt mit dem des
	// Kanonlaufs überein — die Zerlegung ist eine Aufteilung, keine zweite
	// Messung.
	Gesamt Teilabstand
}

// Anteil sagt, welchen Bruchteil des Gesamtquadrats eine Art trägt.
//
// **Ohne Abweichung ist der Anteil null und nicht `NaN`.** Zwei gleiche Bilder
// haben keinen Abstand, und ein Anteil daran ist keine Zahl; eine Tabelle mit
// `NaN` ist unlesbar.
func (z Zerlegung) Anteil(art Ebenenart) float64 {
	if z.Gesamt.Quadratsumme == 0 {
		return 0
	}

	return z.Teile[art].Quadratsumme / z.Gesamt.Quadratsumme
}

// kanaele ist die Zahl der verglichenen Farbkanäle — Rot, Grün, Blau.
//
// **Ohne Alpha**, wie `bildabstand` im Kanonlauf: eine MSE2-Ausgabe ist ein
// deckendes PNG, und ein Alphakanal, der überall 255 ist, verdünnte den
// Abstand um ein Viertel.
const kanaele = 3

// ZerlegeAbstand teilt den Abstand zweier Bilder nach der Ebenenart auf, die
// den jeweiligen Bildpunkt **zuletzt gemalt** hat.
//
// Die [Herkunft] kommt aus [Rasterer.ZeichneMitHerkunft] und gehört zu genau
// dem Bild, das hier verglichen wird. Eine aus einem anderen Lauf wäre eine
// Karte über eine Karte, die es nicht gibt.
func ZerlegeAbstand(
	herkunft *Herkunft, unser, referenz image.Image,
) (Zerlegung, error) {
	rahmen, err := pruefeZerlegung(herkunft, unser, referenz)
	if err != nil {
		return Zerlegung{}, err
	}

	teile := map[Ebenenart]Teilabstand{}

	var gesamt Teilabstand

	uMin, rMin := unser.Bounds().Min, referenz.Bounds().Min

	for y := range rahmen.Dy() {
		for x := range rahmen.Dx() {
			quadrat := punktquadrat(unser, referenz, uMin, rMin, x, y)

			art := herkunft.Art(x, y)

			teil := teile[art]
			teil.Punkte++
			teil.Quadratsumme += quadrat
			teile[art] = teil

			gesamt.Punkte++
			gesamt.Quadratsumme += quadrat
		}
	}

	return Zerlegung{Teile: teile, Gesamt: gesamt}, nil
}

// Ganzabstand misst den Abstand zweier Bilder **ohne** Zerlegung.
//
// **Dieselbe Rechnung, nur ohne Herkunft.** Er steht hier und nicht neben
// [Vergleiche], weil der dort ein anderes Maß hat: `Vergleiche` mittelt
// Beträge, der Kanonlauf mittelt Quadrate. Zwei Maße unter einem Namen wären
// genau der Fehler, den §10 anstreicht — deshalb tragen sie verschiedene.
func Ganzabstand(unser, referenz image.Image) (Teilabstand, error) {
	rahmen, err := pruefeBildpaar(unser, referenz)
	if err != nil {
		return Teilabstand{}, err
	}

	uMin, rMin := unser.Bounds().Min, referenz.Bounds().Min

	var ganz Teilabstand

	for y := range rahmen.Dy() {
		for x := range rahmen.Dx() {
			ganz.Punkte++
			ganz.Quadratsumme += punktquadrat(unser, referenz, uMin, rMin, x, y)
		}
	}

	return ganz, nil
}

// pruefeBildpaar weist ein unbrauchbares Bildpaar ab und gibt die Messfläche.
func pruefeBildpaar(unser, referenz image.Image) (image.Rectangle, error) {
	if unser == nil || referenz == nil {
		return image.Rectangle{}, fmt.Errorf("%w: ein Bild fehlt", ErrPlan)
	}

	uRahmen, rRahmen := unser.Bounds(), referenz.Bounds()
	if uRahmen.Dx() != rRahmen.Dx() || uRahmen.Dy() != rRahmen.Dy() {
		return image.Rectangle{}, fmt.Errorf("%w: %dx%d gegen %dx%d",
			ErrPlan, uRahmen.Dx(), uRahmen.Dy(), rRahmen.Dx(), rRahmen.Dy())
	}

	return uRahmen, nil
}

// pruefeZerlegung weist unbrauchbare Eingaben ab und gibt die Messfläche.
func pruefeZerlegung(
	herkunft *Herkunft, unser, referenz image.Image,
) (image.Rectangle, error) {
	if herkunft == nil {
		return image.Rectangle{}, fmt.Errorf("%w: keine Herkunft", ErrPlan)
	}

	rahmen, err := pruefeBildpaar(unser, referenz)
	if err != nil {
		return image.Rectangle{}, err
	}

	if herkunft.Breite != rahmen.Dx() || herkunft.Hoehe != rahmen.Dy() {
		return image.Rectangle{}, fmt.Errorf(
			"%w: Herkunft %dx%d, Bilder %dx%d — sie gehören nicht zusammen",
			ErrPlan, herkunft.Breite, herkunft.Hoehe, rahmen.Dx(), rahmen.Dy())
	}

	return rahmen, nil
}

// punktquadrat gibt Σ (Δ Kanal)² für einen Bildpunkt.
//
// **Die Bilder werden an `Bounds().Min + x` gelesen**, die Flächen aber
// relativ zur linken oberen Ecke bestimmt. Der Unterschied fällt erst bei
// einem [image.Image] mit versetztem Ursprung auf — einem `SubImage` etwa —,
// und genau dort hat sich `VergleicheAusserhalb` beim ersten Anlauf vertan
// (Befund 5 des Zweitblicks über Phase A).
func punktquadrat(unser, referenz image.Image, uMin, rMin image.Point, x, y int) float64 {
	a, b, c, _ := unser.At(uMin.X+x, uMin.Y+y).RGBA()
	d, e, f, _ := referenz.At(rMin.X+x, rMin.Y+y).RGBA()

	var quadrat float64

	for _, paar := range [kanaele][2]uint32{{a, d}, {b, e}, {c, f}} {
		unterschied := float64(int(paar[0]>>8) - int(paar[1]>>8))
		quadrat += unterschied * unterschied
	}

	return quadrat
}
