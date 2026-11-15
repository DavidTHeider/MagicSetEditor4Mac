package render

import (
	"image"
	"math"
)

// **Wie ein Bild auf die Größe seines Feldes kommt** — die Filterwahl und ihre
// Herkunft (#532).
//
// Eigene Datei, weil `bild.go` mit ihr über 400 Zeilen ginge (§2).

// abtasten liest ein Bild an der Stelle, die einem Bildpunkt des Rahmens
// entspricht — bilinear zwischen den vier Nachbarn.
//
// **Bilinear und nicht nächster Nachbar**, weil ein Kartenrahmen bei jeder
// anderen Vergrößerung anders aussähe: eine Vorschau mit Treppen an den Kanten
// und ein Druckbild ohne sie sind nicht mehr dasselbe Bild, und der
// Goldbildvergleich aus #35 könnte beide nicht mit einer Toleranz fassen.
//
// Von Hand statt über `golang.org/x/image/draw`: es sind dreißig Zeilen, und
// eine Abhängigkeit mehr wäre eine Zeile in `DRITTANBIETER.md`, eine Pinnung
// und ein Aktualisierungspfad — für dreißig Zeilen.
func abtasten(
	quelle image.Image, rahmen image.Rectangle, x, y int, guete Guete,
) Farbe {
	innen := quelle.Bounds()
	if innen.Empty() || rahmen.Empty() {
		return Farbe{}
	}

	// Die Mitte des Zielbildpunkts, umgerechnet in Quellkoordinaten.
	//
	// **Samt `innen.Min`.** Ein `image.Image` muss nicht bei (0,0) beginnen: ein
	// Teilbild aus `SubImage` trägt seinen Ursprung mit. Der erste Anlauf
	// rechnete ab null und klemmte danach fast alles auf die erste Zeile und
	// Spalte — ein Teilbild wurde zu einer Fläche seiner linken oberen Ecke.
	// Gefunden vom Nachlauf des Zweitblicks über Phase 3.
	waagerecht := float64(innen.Min.X) +
		(float64(x-rahmen.Min.X)+0.5)/float64(rahmen.Dx())*float64(innen.Dx()) - 0.5
	senkrecht := float64(innen.Min.Y) +
		(float64(y-rahmen.Min.Y)+0.5)/float64(rahmen.Dy())*float64(innen.Dy()) - 0.5

	// **Bilinear beim Verkleinern, Flächenmittel sonst** — MSEs Regel,
	// gemessen (#532).
	//
	// **Die Regel, nicht das Ergebnis.** Was hier steht, ist die Filterwahl;
	// dass daraus dasselbe Bild würde, steht nirgends und wäre falsch —
	// weiter unten sind die 2,305 beziffert, die die bilineare Rechnung
	// danebenliegt, und oberhalb von Faktor zwei ist wx' Rechnung ungeklärt.
	// Angestrichen vom Nachlauf des Zweitblicks über Phase Ω, weil die
	// Einschränkung nur an einer Stelle stand.
	//
	// MSE bringt ein Bild mit `image.Rescale(w, h)` auf die Feldgröße
	// (`src/render/value/image.cpp`, Zeilen 33–40, Fassung f9d9356 vom
	// 2021-07-06). Das ist `wxImage::Scale` mit dem Vorgabewert
	// `wxIMAGE_QUALITY_NORMAL` — und **was dieser Wert rechnet, steht in
	// keiner Kopfdatei**; sie sagt nur „default quality, suitable for most
	// icons".
	//
	// Am 2026-08-19 gegen **wx 3.3.3** gemessen, der Fassung aus
	// `mse2-quelle/build/CMakeCache.txt`, an einer echten Illustration
	// (626 × 457):
	//
	//	beide Achsen kleiner, Faktor 1,04 … 1,98   BILINEAR       RMS 0,0000
	//	nur eine Achse kleiner (316 × 457)         BILINEAR       RMS 0,0000
	//	beide Achsen größer                        BOX_AVERAGE    RMS 0,0000
	//	gemischt (300×900 und 900×300)             BOX_AVERAGE    RMS 0,0000
	//
	// # Hier stand bis dahin das Gegenteil
	//
	// Verkleinert wurde gemittelt und vergrößert bilinear — **genau
	// vertauscht**. Der Grund war selbst gemessen
	// (#113):
	// bilinear lässt feine Linien beim starken Verkleinern flimmern, und bei
	// achtfacher Verkleinerung streuten die Zielzeilen um 127 von 255 Stufen.
	//
	// **Das bleibt richtig und wird trotzdem aufgegeben.** Es ist eine Aussage
	// über Bildqualität, nicht über Parität; MSE flimmert ebenso, und das Ziel
	// dieses Projekts ist dasselbe Bild. Was es kostet, gehört ausgesprochen:
	// die Vorschau in kleinen Größen und der Druckbogen bekommen MSEs Flimmern
	// zurück.
	//
	// # Und die Filterwahl ist nicht die ganze Antwort
	//
	// **Angeglichen fällt das Kunstfenster von `A-Cloister Gargoyle` auf
	// 2,362, nicht auf null.** Erwartet war null: wx' `NORMAL`-Ergebnis trifft
	// MSE2s fertige Karte exakt. Was übrig bleibt, ist **wx' bilineare
	// Rechnung selbst** — unsere Abtastung liegt 2,305 daneben, und das ist
	// fast genau der Rest. Nachgemessen: eckbündig statt mittenbündig ist
	// **schlechter** (4,55), die Abbildung stimmt also; es ist die Arithmetik.
	//
	// Das ist eine eigene, kleinere Frage und steht als solche an
	// #532.
	//
	// # Was oberhalb von Faktor 2 gilt, ist **nicht** geklärt
	//
	// Dort trifft `NORMAL` keinen der reinen Filter: bei 3,13-facher
	// Verkleinerung liegt es 1,66 vom Flächenmittel und 4,75 von bilinear, bei
	// 6,26-facher 1,50 gegen 9,31. Das Flächenmittel ist damit die beste
	// verfügbare Wahl **und** die bessere für #113 — beides zugleich, und
	// deshalb steht es hier. Was wx wirklich rechnet, ist offen.
	breit := float64(innen.Dx()) / float64(rahmen.Dx())
	hoch := float64(innen.Dy()) / float64(rahmen.Dy())

	// **Welche Rechnung, sagt die Güte** — Werkstück Ω.4 aus
	// #1029.
	// Die Vorgabe [GueteMittel] rechnet weiter wie MSE; die drei anderen
	// weichen ab und sagen es an der Ausgabe ([Guete.Hinweis]).
	switch guete {
	case GueteSchnell:
		return punktfarbe(quelle, innen,
			int(math.Round(waagerecht)), int(math.Round(senkrecht)))

	case GueteHoch:
		if verkleinert(breit, hoch) {
			return flaechenmittel(quelle, innen, waagerecht, senkrecht, breit, hoch)
		}

		return bilinear(quelle, innen, waagerecht, senkrecht)

	case GueteHoechst:
		if verkleinert(breit, hoch) {
			return flaechenmittel(quelle, innen, waagerecht, senkrecht, breit, hoch)
		}

		return bikubisch(quelle, innen, waagerecht, senkrecht)

	case GueteMittel:
	}

	if !bilinearWieMSE(breit, hoch) {
		return flaechenmittel(quelle, innen, waagerecht, senkrecht, breit, hoch)
	}

	return bilinear(quelle, innen, waagerecht, senkrecht)
}

// verkleinert sagt, ob irgendeine Achse schrumpft.
//
// **Irgendeine, nicht beide.** Ein Bild, das nur waagerecht schrumpft, verliert
// dort Bildpunkte, und ein Mittel über die Fläche ist die Antwort darauf —
// dass die andere Achse wächst, ändert daran nichts.
func verkleinert(breit, hoch float64) bool {
	return breit > 1 || hoch > 1
}

// bilinear mischt die vier Nachbarn der Abtaststelle.
//
// **Vorgewichtet mischen, danach einmal zurückrechnen.**
//
// Der erste Anlauf mischte die entwichteten Kanäle und das Alpha getrennt.
// An einer Kante zwischen durchsichtigem Rot und deckendem Blau ergibt das
// einen violetten Saum statt einer anteilig durchsichtigen blauen Kante —
// die Farbe eines Bildpunkts, den man gar nicht sieht, färbt seinen
// Nachbarn. Gefunden vom Zweitblick über Phase 4.
func bilinear(
	quelle image.Image, innen image.Rectangle, waagerecht, senkrecht float64,
) Farbe {
	links := int(math.Floor(waagerecht))
	oben := int(math.Floor(senkrecht))
	anteilRechts := waagerecht - float64(links)
	anteilUnten := senkrecht - float64(oben)

	obenLinks := vorgewichtet(punktfarbe(quelle, innen, links, oben))
	obenRechts := vorgewichtet(punktfarbe(quelle, innen, links+1, oben))
	untenLinks := vorgewichtet(punktfarbe(quelle, innen, links, oben+1))
	untenRechts := vorgewichtet(punktfarbe(quelle, innen, links+1, oben+1))

	return entwichtet(mischeAnteile(
		mischeAnteile(obenLinks, obenRechts, anteilRechts),
		mischeAnteile(untenLinks, untenRechts, anteilRechts),
		anteilUnten,
	))
}

// Verkleinerungsgrenze ist der Faktor, ab dem `wxIMAGE_QUALITY_NORMAL` nicht
// mehr bilinear rechnet — gemessen, nicht erschlossen (siehe [abtasten]).
const Verkleinerungsgrenze = 2.0

// bilinearWieMSE sagt, ob MSE für diese Skalierung bilinear rechnet.
//
// **Keine Achse darf vergrößern**, und keine um mehr als das Doppelte
// verkleinern. Eine unveränderte Achse steht dem nicht entgegen — gemessen an
// 626 × 457 → 316 × 457, wo `NORMAL` bilinear rechnet (RMS 0,0000).
//
// **Bei gemischter Richtung nimmt wx das Flächenmittel** — für den ganzen
// Aufruf, nicht je Achse.
func bilinearWieMSE(breit, hoch float64) bool {
	return breit >= 1 && hoch >= 1 &&
		breit < Verkleinerungsgrenze && hoch < Verkleinerungsgrenze
}

// spanne gibt die Quellpunkte, die ein Zielbildpunkt abdeckt.
//
// **Deckt er weniger als einen ab, ist es genau der, in dem seine Mitte
// liegt.** Beim Vergrößern ist das der ganze Filter — und genau das rechnet
// `wxIMAGE_QUALITY_BOX_AVERAGE` dort: jeder Zielpunkt liegt in **einem**
// Quellpunkt, also ist das Mittel dieser eine Wert.
//
// Ohne diese Unterscheidung mittelte der Filter über zwei Nachbarn und ergäbe
// weiche Kanten, wo MSE harte zeichnet
// (#532).
func spanne(mitte, breite float64) (von, bis int) {
	if breite <= 1 {
		stelle := int(math.Floor(mitte + 0.5))

		return stelle, stelle
	}

	return int(math.Floor(mitte - breite/2 + 0.5)),
		int(math.Ceil(mitte + breite/2 - 0.5))
}

// flaechenmittel mittelt über den Quellbereich, den ein Zielbildpunkt abdeckt.
//
// **Ungewichtet über ganze Quellpunkte.** Ein Kastenfilter mit Teilgewichten an
// den Rändern wäre genauer und teurer; für die Frage, um die es geht — flimmern
// feine Linien —, entscheidet die Zahl der einbezogenen Punkte, nicht ihr
// Randgewicht.
func flaechenmittel(
	quelle image.Image, innen image.Rectangle, mitteX, mitteY, breit, hoch float64,
) Farbe {
	vonX, bisX := spanne(mitteX, breit)
	vonY, bisY := spanne(mitteY, hoch)

	var (
		summe  anteilsfarbe
		punkte float64
	)

	for y := vonY; y <= bisY; y++ {
		for x := vonX; x <= bisX; x++ {
			farbe := vorgewichtet(punktfarbe(quelle, innen, x, y))

			summe.r += farbe.r
			summe.g += farbe.g
			summe.b += farbe.b
			summe.a += farbe.a

			punkte++
		}
	}

	// **Ohne Prüfung auf null.** [spanne] gibt immer mindestens einen Punkt,
	// und die Schleife zählt ihn. Ein Zweig, den kein Aufruf betreten kann,
	// wäre eine Behauptung darüber, was im Ernstfall geschieht (§3).
	return entwichtet(anteilsfarbe{
		r: summe.r / punkte, g: summe.g / punkte,
		b: summe.b / punkte, a: summe.a / punkte,
	})
}

// anteilsfarbe ist eine Farbe in 0…1 mit vorgewichteten Kanälen.
type anteilsfarbe struct{ r, g, b, a float64 }

func vorgewichtet(farbe Farbe) anteilsfarbe {
	a := float64(farbe.A) / 255

	return anteilsfarbe{
		r: float64(farbe.R) / 255 * a,
		g: float64(farbe.G) / 255 * a,
		b: float64(farbe.B) / 255 * a,
		a: a,
	}
}

func entwichtet(farbe anteilsfarbe) Farbe {
	if farbe.a <= 0 {
		return Farbe{}
	}

	return Farbe{
		R: alsAnteil(farbe.r / farbe.a * 255),
		G: alsAnteil(farbe.g / farbe.a * 255),
		B: alsAnteil(farbe.b / farbe.a * 255),
		A: alsAnteil(farbe.a * 255),
	}
}

func mischeAnteile(von, bis anteilsfarbe, anteil float64) anteilsfarbe {
	kanal := func(a, b float64) float64 { return a + (b-a)*anteil }

	return anteilsfarbe{
		r: kanal(von.r, bis.r),
		g: kanal(von.g, bis.g),
		b: kanal(von.b, bis.b),
		a: kanal(von.a, bis.a),
	}
}
