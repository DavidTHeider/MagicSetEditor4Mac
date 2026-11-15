package render

import (
	"fmt"
	"image"
	"math"
)

// Toleranz ist der zulässige mittlere Abstand zweier Bilder.
//
// # Woher die Zahl kommt
//
// Sie ist **kein Gefühl**, sondern die Antwort auf eine Frage: was darf sich
// ändern, ohne dass jemand es sehen kann?
//
// Ein Bildpunkt hat vier Kanäle zu je 0…255. Der Abstand ist der mittlere
// Betrag über alle Kanäle aller Bildpunkte. Bei `1.0` dürfte sich **jeder**
// Bildpunkt um einen Schritt ändern — das ist unterhalb dessen, was ein Bildschirm
// unterscheidet, und deckt zugleich die Rundung ab, die beim Mischen entsteht
// (`mischen.go` rundet auf das nächste Byte).
//
// **Was sie nicht deckt:** eine Verschiebung. Ein um zwei Bildpunkte
// verrutschter Rahmen ändert an seinen Kanten viel und in der Fläche nichts;
// der Mittelwert darüber bleibt trotzdem weit über 1, weil eine Kante über die
// ganze Höhe geht. Die Gegenprobe in `goldbild_test.go` zeigt das.
const Toleranz = 1.0

// Spitzentoleranz begrenzt, wie **stark** ein einzelner Kanal abweichen darf.
//
// # Warum der Mittelwert allein nicht reicht
//
// Er verdünnt sich mit der Bildfläche. Eine um einen Bildpunkt verschobene
// Kante ändert an ihr viel und sonst nichts; in einer 80×112-Saat reißt das den
// Mittelwert über 1, in einem Druckbild von 1500 Punkten Breite liegt dieselbe
// Kante rechnerisch weit darunter. **Die Gegenprobe belegte damit nur die Größe
// des Testbilds** — angestrichen vom Zweitblick über Phase 4.
//
// # Warum nicht der Anteil der betroffenen Punkte
//
// Das war der erste Anlauf, und er wies genau den Fall ab, den die zweite
// Gegenprobe schützt: eine geänderte Rundung trifft **jeden** Bildpunkt um
// einen Schritt. Nach Anteil gemessen ist das die schlimmste denkbare
// Abweichung; gesehen hat sie noch nie jemand.
//
// Die Stärke trennt beides sauber und **ohne Bezug zur Bildgröße**: eine
// Rundung ändert um 1, eine verschobene Kante um 100 und mehr. Zwei Schritte
// sind unter der Wahrnehmungsschwelle und decken zugleich zwei Ebenen, die je
// einen Schritt beitragen.
const Spitzentoleranz = 2

// Vergleich ist das Ergebnis eines Goldbildvergleichs.
type Vergleich struct {
	// Abstand ist der mittlere Betrag über alle Kanäle aller Bildpunkte.
	Abstand float64

	// Groesster ist der größte Betrag an einem einzelnen Kanal. Er sagt, ob eine
	// Abweichung überall ein wenig ist oder an einer Stelle viel — und das ist
	// der Unterschied zwischen einer geänderten Rundung und einem verrutschten
	// Rahmen.
	Groesster int

	// Abweichende ist die Zahl der Bildpunkte, die sich überhaupt unterscheiden.
	Abweichende int

	// Punkte ist die Zahl der verglichenen Bildpunkte — der Nenner für den
	// Anteil. Ohne ihn wäre `Abweichende` eine Zahl ohne Maßstab.
	Punkte int
}

func (v Vergleich) String() string {
	anteil := 0.0
	if v.Punkte > 0 {
		anteil = float64(v.Abweichende) / float64(v.Punkte)
	}

	return fmt.Sprintf("Abstand %.3f, größte Abweichung %d, %d von %d Bildpunkten betroffen (%.4f)",
		v.Abstand, v.Groesster, v.Abweichende, v.Punkte, anteil)
}

// Innerhalb sagt, ob ein Vergleich **beide** Schwellen hält.
//
// Der Mittelwert fängt eine Änderung, die überall ein wenig ausmacht; die
// Spitze fängt eine, die an wenigen Stellen viel ausmacht. Eine Schwelle allein
// lässt je eine der beiden Arten durch — und die zweite ist die, die mit
// wachsender Auflösung unsichtbar würde.
func (v Vergleich) Innerhalb() bool {
	return v.Abstand <= Toleranz && v.Groesster <= Spitzentoleranz
}

// Vergleiche misst den Abstand zweier Bilder.
//
// Bilder verschiedener Größe gelten als **völlig** verschieden; einen Abstand
// zwischen ihnen zu bilden hieße, sich auf eine Streckung zu einigen, die
// niemand angeordnet hat.
func Vergleiche(gold, neu image.Image) (Vergleich, error) {
	return VergleicheAusserhalb(gold, neu, image.Rectangle{})
}

// VergleicheAusserhalb misst den Abstand zweier Bilder **ohne** eine Fläche.
//
// # Wofür die Aussparung da ist
//
// Gegen eine echte MSE2-Ausgabe fehlt uns die **Illustration**: das
// Referenzbild ist ein fertig gerechnetes JPG, die Quelldatei der Kunst liegt
// nicht bei und wird es nie. Ein Vergleich über die ganze Karte misst dort zu
// einem Drittel etwas, das nie stimmen kann — und jede Verbesserung an Rahmen,
// Schrift und Anordnung verschwände in diesem Rauschen
// ([0021](../../docs/entscheidungen/0021-das-referenzbild-kommt-vom-nutzenden.md)).
//
// **Aussparen ist nicht Wegschauen.** Der Aufrufer meldet beide Zahlen: die
// ganze Karte und die Karte ohne Bildfenster. Die erste ist der wahre Abstand
// und wird die Toleranz nie erreichen; die zweite ist die, die sich bewegt.
//
// # Die Koordinaten
//
// Die Aussparung wird **relativ zur linken oberen Ecke** gelesen, nicht in den
// absoluten Koordinaten des Goldbilds. Der Unterschied fällt erst bei einem
// [image.Image] mit versetztem Ursprung auf — einem `SubImage` etwa —, und der
// erste Anlauf verglich `image.Pt(x, y)` gegen die Aussparung, während er die
// Bildpunkte an `Bounds().Min + x` las. Bei Bounds ab `(10,10)` sparte er damit
// die falsche Fläche aus. Angestrichen vom Zweitblick über Phase A.
//
// Eine Aussparung, die nichts übrig lässt, ist ein Fehler und kein leerer
// Erfolg: ein Vergleich ohne Bildpunkte hätte den Abstand null und wäre immer
// grün.
func VergleicheAusserhalb(gold, neu image.Image, aussparung image.Rectangle) (Vergleich, error) {
	if gold == nil || neu == nil {
		return Vergleich{}, fmt.Errorf("%w: ein Bild fehlt", ErrPlan)
	}

	gRahmen, nRahmen := gold.Bounds(), neu.Bounds()
	if gRahmen.Dx() != nRahmen.Dx() || gRahmen.Dy() != nRahmen.Dy() {
		return Vergleich{}, fmt.Errorf("%w: %dx%d gegen %dx%d",
			ErrPlan, gRahmen.Dx(), gRahmen.Dy(), nRahmen.Dx(), nRahmen.Dy())
	}

	var (
		summe      float64
		groesster  int
		betroffen  int
		verglichen int
	)

	for y := range gRahmen.Dy() {
		for x := range gRahmen.Dx() {
			if !aussparung.Empty() && image.Pt(x, y).In(aussparung) {
				continue
			}

			verglichen++

			alt := punktfarbe(gold, gRahmen, gRahmen.Min.X+x, gRahmen.Min.Y+y)
			jetzt := punktfarbe(neu, nRahmen, nRahmen.Min.X+x, nRahmen.Min.Y+y)

			abstand := 0
			for _, paar := range [][2]uint8{
				{alt.R, jetzt.R}, {alt.G, jetzt.G}, {alt.B, jetzt.B}, {alt.A, jetzt.A},
			} {
				unterschied := int(paar[0]) - int(paar[1])
				if unterschied < 0 {
					unterschied = -unterschied
				}

				abstand += unterschied

				if unterschied > groesster {
					groesster = unterschied
				}
			}

			if abstand > 0 {
				betroffen++
			}

			summe += float64(abstand)
		}
	}

	if verglichen == 0 {
		return Vergleich{}, fmt.Errorf(
			"%w: die Aussparung lässt keinen Bildpunkt übrig — ein Vergleich ohne "+
				"Bildpunkte hätte den Abstand null und wäre immer grün", ErrPlan)
	}

	return Vergleich{
		Abstand:     summe / float64(verglichen*4),
		Groesster:   groesster,
		Abweichende: betroffen,
		Punkte:      verglichen,
	}, nil
}

// Differenzbild zeigt, **wo** zwei Bilder auseinandergehen.
//
// Gleiche Stellen werden dunkel, verschiedene hell — je heller, desto größer
// der Unterschied. Ein Zahlenwert sagt, **dass** etwas anders ist; dieses Bild
// sagt, **was**. Bei einer Vorlage mit 60 Ebenen ist das der Unterschied
// zwischen einer Minute und einem Nachmittag.
func Differenzbild(gold, neu image.Image) *image.NRGBA {
	rahmen := gold.Bounds()
	bild := image.NewNRGBA(image.Rect(0, 0, rahmen.Dx(), rahmen.Dy()))

	for y := range rahmen.Dy() {
		for x := range rahmen.Dx() {
			alt := punktfarbe(gold, rahmen, rahmen.Min.X+x, rahmen.Min.Y+y)
			jetzt := punktfarbe(neu, neu.Bounds(), neu.Bounds().Min.X+x, neu.Bounds().Min.Y+y)

			stelle := bild.PixOffset(x, y)
			bild.Pix[stelle] = abstand(alt.R, jetzt.R)
			bild.Pix[stelle+1] = abstand(alt.G, jetzt.G)
			bild.Pix[stelle+2] = abstand(alt.B, jetzt.B)
			bild.Pix[stelle+3] = 255
		}
	}

	return bild
}

// abstand verstärkt kleine Unterschiede, damit sie sichtbar werden.
//
// Ein Unterschied von 3 wäre in einem Differenzbild sonst schwarz auf schwarz —
// und ein Differenzbild, auf dem man nichts sieht, ist keins.
func abstand(a, b uint8) uint8 {
	unterschied := math.Abs(float64(a) - float64(b))

	return alsAnteil(math.Min(unterschied*8, 255))
}
