package render

import (
	"image"
	"math"
)

// **Bikubisch vergrößern** — die Rechnung hinter [GueteHoechst].
//
// # Warum überhaupt eine vierte
//
// MSE mittelt beim Vergrößern über die Fläche (`wxIMAGE_QUALITY_NORMAL`,
// gemessen in [abtasten]). Über eine Illustration, die für den Druck auf ein
// Vielfaches ihrer Kantenlänge wächst, heißt das: jeder Quellbildpunkt wird
// zu einem Block. Das ist die Parität und für den Bildschirm richtig; für
// einen Druck in voller Auflösung ist es die schlechteste Wahl von allen.
//
// # Catmull-Rom, nicht irgendein Kubik
//
// Die kubische Faltung hat einen freien Parameter; `a = -0,5` ist der Wert,
// der die Taylorreihe der Quelle bis zur zweiten Ordnung trifft — Keys 1981,
// „Cubic convolution interpolation for digital image processing". Andere Werte
// sind weicher (Mitchell) oder schärfer und klingeln stärker.
//
// **Was das kostet, gehört dazu**: sechzehn Nachbarn statt vier, und der Kern
// überschwingt an harten Kanten. Der Überschwinger wird geklemmt
// ([alsAnteil]), das Klingeln bleibt — es ist der Preis der Schärfe und der
// Grund, warum diese Güte nicht die Vorgabe ist.
const kubischeSchaerfe = -0.5

// kubischesGewicht gibt das Gewicht eines Nachbarn im Abstand d.
func kubischesGewicht(abstand float64) float64 {
	abstand = math.Abs(abstand)
	a := kubischeSchaerfe

	switch {
	case abstand < 1:
		return (a+2)*abstand*abstand*abstand - (a+3)*abstand*abstand + 1
	case abstand < 2:
		return a*abstand*abstand*abstand - 5*a*abstand*abstand + 8*a*abstand - 4*a
	}

	return 0
}

// bikubisch mischt die sechzehn Nachbarn der Abtaststelle.
//
// **Vorgewichtet wie der bilineare Weg**, und aus demselben Grund: sonst färbt
// ein unsichtbarer Bildpunkt seinen Nachbarn.
func bikubisch(
	quelle image.Image, innen image.Rectangle, waagerecht, senkrecht float64,
) Farbe {
	links := int(math.Floor(waagerecht))
	oben := int(math.Floor(senkrecht))

	var (
		summe   anteilsfarbe
		gewicht float64
	)

	for zeile := -1; zeile <= 2; zeile++ {
		gewichtY := kubischesGewicht(senkrecht - float64(oben+zeile))
		if gewichtY == 0 {
			continue
		}

		for spalte := -1; spalte <= 2; spalte++ {
			gewichtX := kubischesGewicht(waagerecht - float64(links+spalte))
			if gewichtX == 0 {
				continue
			}

			anteil := gewichtX * gewichtY
			farbe := vorgewichtet(punktfarbe(quelle, innen, links+spalte, oben+zeile))

			summe.r += farbe.r * anteil
			summe.g += farbe.g * anteil
			summe.b += farbe.b * anteil
			summe.a += farbe.a * anteil
			gewicht += anteil
		}
	}

	// **Geteilt wird trotzdem, obwohl die Summe eins ist.**
	//
	// Catmull-Rom ist eine Zerlegung der Eins: über die vier Nachbarn einer
	// Achse ergeben die Gewichte für **jede** Bruchstelle exakt 1. Nachbarn
	// fallen auch am Rand nicht weg — [punktfarbe] klemmt auf den letzten
	// gültigen Bildpunkt, statt nichts zu liefern.
	//
	// **Kein Zweig für „Summe null"**, denn kein Aufruf erreicht ihn: bei
	// jeder Bruchstelle trägt mindestens ein Nachbar ein Gewicht nahe eins.
	// Ein Zweig ohne Fall wäre eine Behauptung darüber, was im Ernstfall
	// geschieht (AGENTS.md §3). Geteilt wird gegen die **Rundung**: in
	// Gleitkomma ist die Summe 0,999… oder 1,000…, und ohne die Teilung
	// wanderte diese Abweichung in die Helligkeit.
	return entwichtet(anteilsfarbe{
		r: summe.r / gewicht,
		g: summe.g / gewicht,
		b: summe.b / gewicht,
		a: summe.a / gewicht,
	})
}
