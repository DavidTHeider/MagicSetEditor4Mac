// **Ein Bild nachzeichnen: vereinfachen** —
// #1340,
// Werkstück F von Phase 4.11.
//
// Portiert aus `simplify_symbol_shape` und den sieben Schritten darunter
// (`src/data/format/image_to_symbol.cpp`, Zeilen 241–470, Fassung `f9d9356`
// vom 2021-07-06).
//
// # Warum es sieben Schritte sind und nicht einer
//
// Der Rundgang gibt einen Punkt **je Bildpunkt** der Kante — für ein Bild von
// 100 × 100 sind das schnell dreihundert. Eine Zeichnung mit dreihundert
// Punkten ist keine, die man weiterbearbeiten kann, und genau das ist die
// Zusage aus dem Issue.
//
// Die Reihenfolge ist MSEs, und sie ist keine Geschmacksfrage: geglättet wird
// **vor** dem Umwandeln in Kurven, weil die Griffe sonst aus der ungeglätteten
// Lage entstünden; begradigt wird **nach** dem Entfernen, weil erst dann
// feststeht, welche Strecken lang genug sind, um gerade zu heißen.

package service

import (
	"math"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// sperreFrei und sperreRichtung sind MSEs `LOCK_FREE` und `LOCK_DIR`.
//
// **Sie stehen als Zeichenketten in der Datei** und heißen dort `free` und
// `direction` (`docs/FORMATE.md`).
const (
	sperreFrei     = "free"
	sperreRichtung = "direction"
)

// punktZyklisch greift einen Punkt, auch mit negativem oder zu großem Index.
//
// **Weil der Zug geschlossen ist**: MSEs `getPoint(i)` rechnet `mod` und
// wickelt um (`src/data/symbol.hpp`, Fassung `f9d9356` vom 2021-07-06). Die Nachbarschaftsrechnungen darunter
// greifen bis vier Punkte weit und liefen sonst am Anfang und am Ende ins
// Leere — und ein Umriss hat weder das eine noch das andere.
func punktZyklisch(punkte []msepaket.Kontrollpunkt, i int) *msepaket.Kontrollpunkt {
	n := len(punkte)

	return &punkte[((i%n)+n)%n]
}

// **`plus`, `minus` und `mal` stehen in `symbolpunkte.go`** — eine zweite
// Fassung daneben liefe mit jener auseinander (§2).

func laenge(a msepaket.Vektor) float64 {
	return math.Hypot(a.X, a.Y)
}

// genormt gibt den Vektor mit Länge eins — und den Nullvektor für den
// Nullvektor.
//
// **MSEs `normalized()` teilt ohne Prüfung** und gäbe `NaN`. Hier steht die
// Prüfung, weil ein `NaN` sich still über die ganze Form ausbreitete und die
// Zeichnung erst beim Zeichnen als leer auffiele.
func genormt(a msepaket.Vektor) msepaket.Vektor {
	l := laenge(a)
	if l == 0 {
		return msepaket.Vektor{}
	}

	return mal(a, 1/l)
}

func punktprodukt(a, b msepaket.Vektor) float64 {
	return a.X*b.X + a.Y*b.Y
}

func kreuzprodukt(a, b msepaket.Vektor) float64 {
	return a.X*b.Y - a.Y*b.X
}

// eckenMarkieren setzt `free` an Ecken und `direction` sonst.
//
// Portiert aus `mark_corners` (Zeilen 245–259). **Die Gewichte sind MSEs**:
// `.6/.2/.1/.1` über die vier Nachbarn je Seite, und die Schwelle ist
// `dot >= -0.25`. Wer sie herleitet statt nachzulesen, bekommt eine andere
// Menge Ecken — und eine Ecke zu viel ist eine Delle im Umriss.
func eckenMarkieren(punkte []msepaket.Kontrollpunkt) {
	gewichte := [4]float64{0.6, 0.2, 0.1, 0.1}

	for i := range punkte {
		hier := punkte[i].Stelle

		var davor, danach msepaket.Vektor

		for j, gewicht := range gewichte {
			davor = plus(davor, mal(punktZyklisch(punkte, i-1-j).Stelle, gewicht))
			danach = plus(danach, mal(punktZyklisch(punkte, i+1+j).Stelle, gewicht))
		}

		davor = genormt(minus(davor, hier))
		danach = genormt(minus(danach, hier))

		if punktprodukt(davor, danach) >= -0.25 {
			punkte[i].Sperre = sperreFrei
		} else {
			punkte[i].Sperre = sperreRichtung
		}
	}
}

// eckenVerschmelzen zieht zwei benachbarte Ecken auf ihren Schnittpunkt.
//
// Portiert aus `merge_corners` (Zeilen 275–318). **Der Grund steht in MSEs
// eigenem Kommentar**: ein Dreieck gibt an seiner Spitze zwei benachbarte
// Ecken, eine flache Kante mit einer Spitze darüber aber auch — und die
// zweiten gehören nicht zusammen. Entschieden wird es über die Tangenten:
// treffen sie sich nah genug, ist es **eine** Ecke.
func eckenVerschmelzen(punkte []msepaket.Kontrollpunkt) []msepaket.Kontrollpunkt {
	for i := 0; i < len(punkte); i++ {
		if len(punkte) < 3 {
			break
		}

		hier := punktZyklisch(punkte, i)
		vorher := punktZyklisch(punkte, i-1)

		if vorher.Sperre != sperreFrei || hier.Sperre != sperreFrei {
			continue
		}

		ab := minus(hier.Stelle, vorher.Stelle)

		kleinstesA, kleinstesB := math.Inf(1), math.Inf(1)

		var a, b msepaket.Vektor

		for j := range 4 {
			bewerberA := genormt(minus(punktZyklisch(punkte, i-j-1).Stelle, vorher.Stelle))
			bewerberB := genormt(minus(punktZyklisch(punkte, i+j).Stelle, hier.Stelle))

			if wert := punktprodukt(bewerberA, ab); wert < kleinstesA {
				kleinstesA, a = wert, bewerberA
			}

			if wert := -punktprodukt(bewerberB, ab); wert < kleinstesB {
				kleinstesB, b = wert, bewerberB
			}
		}

		// **`t a + ab = u b`, aufgelöst nach `t`.** Der Nenner wird nach unten
		// geklemmt, weil zwei parallele Tangenten ihn null machten — MSEs
		// `max(0.00000001, cross(b,a))`.
		nenner := math.Max(0.00000001, kreuzprodukt(b, a))
		t := kreuzprodukt(ab, b) / nenner

		if t >= 0 && t < 20.0 {
			vorher.Stelle = plus(vorher.Stelle, mal(a, -t))
			punkte = append(punkte[:i], punkte[i+1:]...)
			i--
		}
	}

	return punkte
}

// glaetten mittelt jeden Nicht-Eckpunkt mit seinen Nachbarn.
//
// Portiert aus `avarage` (Zeilen 321–335). **Über eine Kopie**, weil sonst der
// schon geglättete Vorgänger in die Rechnung des nächsten einginge — und aus
// einem Mittelwert würde ein Nachlauf.
func glaetten(punkte []msepaket.Kontrollpunkt) {
	alte := make([]msepaket.Vektor, len(punkte))
	for i := range punkte {
		alte[i] = punkte[i].Stelle
	}

	n := len(alte)

	for i := range punkte {
		if punkte[i].Sperre != sperreRichtung {
			continue
		}

		davor := alte[((i-1)%n+n)%n]
		danach := alte[(i+1)%n]

		punkte[i].Stelle = plus(
			plus(mal(davor, 0.25), mal(punkte[i].Stelle, 0.5)),
			mal(danach, 0.25),
		)
	}
}
