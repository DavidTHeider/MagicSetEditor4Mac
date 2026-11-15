package service

// **Der Anschnitt außen um die Karte** — I04
// (#2262).
//
// # Was bis dahin geschah
//
// Der Bogen rasterte die Vorlage auf die Breite **samt** Anschnitt und zog sie
// auf den Anschnittkasten. Gemessen bei 63 × 88 mm und 3 mm Anschnitt: der
// Karteninhalt wurde waagerecht um 9,5 % und senkrecht um 6,8 % vergrößert,
// und ihr eigener Rand lag 3 mm **außerhalb** der Schnittkante — wer schnitt,
// schnitt rundum drei Millimeter Karte ab.
//
// # Was jetzt geschieht
//
// Die Karte wird in ihrem Maß gerastert und genau auf die Schnittkante gesetzt;
// der Anschnitt **verlängert ihren Rand** nach außen, Bildpunkt um Bildpunkt.
// Ein Rahmen läuft damit über die Kante hinaus weiter, und ein Schnitt, der um
// einen Millimeter danebengeht, zeigt Rahmen statt Weiß. **Eine neue Regel**:
// MSE2 kennt keinen Anschnitt — `src/gui/print_window.cpp` legt Karten ohne
// ihn, und „bleed" kommt in `src/` nicht vor (Fassung f9d9356 vom 2021-07-06).
//
// # Und erst das Kartenmaß, dann der Anschnitt
//
// **Der Anschnitt verlängert den Rand einer Karte, die schon im Kartenmaß
// steht** — nicht den einer Vorlage, die zufällig in dieselbe Breite passt.
// Bis J05 (#2269)
// rasterte der Bogen proportional zur Kartenbreite und ließ `mitAnschnitt` die
// Höhe auffüllen: bei einer quadratischen Vorlage für 63 × 88 mm und 3 mm
// Anschnitt standen oben und unten je 15,5 mm **wiederholter Randpunkt** statt
// Karteninhalt; ohne Anschnitt streckte erst der PDF-Leser ein Bild, dessen
// Bildpunkte nicht zur Auflösung passten, und eine zu hohe Vorlage wurde im
// Kasten gestaucht, ihr Inhalt lag hinter der Schnittkante.
//
// **Das Ziehen auf das Kartenmaß ist die Zusage des Bogens**
// ([Bogenleger.Lege]) und eine **angepasste** Regel: MSE2 bestimmt die
// Kartengröße aus dem Stil (`card_width`, `card_height`, `card_dpi`) und
// setzt jede Karte in diesem Maß, ohne sie je auf ein vorgegebenes zu
// ziehen — `PageLayout::init`, `src/gui/print_window.cpp`, Zeilen 31–35, und
// `CardsPrintout::drawCard`, Zeilen 143–158, die Karte wird mit einem
// einheitlichen Maßstab gesetzt, und der Kommentar in Zeile 139 nennt
// „different sized cards in general" ein **TODO** (Fassung f9d9356 vom
// 2021-07-06). Hier ist das Kartenmaß die Vorgabe der Nutzenden
// ([Bogenmasse]); weicht die Vorlage ab, wird sie gezogen, und der Befund
// daneben sagt es.

import (
	"image"
	"math"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// kartenpunkte sind die Bildpunkte der Karte **ohne** Anschnitt bei der
// Auflösung des Bogens.
//
// **Mindestens einer je Achse**: eine Karte, die kleiner ist als ein halber
// Bildpunkt, ergäbe sonst ein leeres Bild — und `internal/pdf` schriebe es.
func kartenpunkte(masse Bogenmasse) (breit, hoch int) {
	jeMM := float64(masse.DPI) / 25.4

	return max(1, int(math.Round(masse.Kartenbreite*jeMM))),
		max(1, int(math.Round(masse.Kartenhoehe*jeMM)))
}

// aufKartenmass bringt ein gerastertes Bild auf die Bildpunkte der Karte.
//
// **Ein Bild, das schon passt, bleibt, wie es ist**: eine Vorlage im Verhältnis
// der Karte rastert auf genau diese Punkte, und eine Neuberechnung mischte
// ihre Kanten für nichts. Alles andere läuft über [render.ZuschnittAus] — den
// Filter, mit dem der Rasterer jedes Bild auf sein Feld bringt (MSEs
// `resample_and_clip`); ein zweiter daneben hieße, dass derselbe Plan auf dem
// Bogen anders aussieht als im Bild.
func aufKartenmass(karte *image.NRGBA, masse Bogenmasse) *image.NRGBA {
	breit, hoch := kartenpunkte(masse)

	flaeche := karte.Bounds()
	if flaeche.Dx() == breit && flaeche.Dy() == hoch {
		return karte
	}

	return render.ZuschnittAus(karte, flaeche, breit, hoch)
}

// mitAnschnitt legt den Anschnitt um ein Bild, bis es `breite` × `hoehe`
// Bildpunkte misst, und füllt ihn mit dem nächsten Randpunkt.
//
// **Auf das Ganze gerundet und nicht je Teil**: Karte und Anschnitt einzeln
// auf Bildpunkte zu runden, summiert drei Rundungen — gemessen 0,15 mm bei
// 150 dpi. Der Aufrufer rundet die ganze Fläche, und was zur Karte fehlt,
// teilen sich die zwei Seiten. **Kleiner wird nichts**: ein Bild, das schon
// größer ist, behält seine Bildpunkte.
func mitAnschnitt(karte *image.NRGBA, breite, hoehe int) *image.NRGBA {
	flaeche := karte.Bounds()
	kb, kh := flaeche.Dx(), flaeche.Dy()
	breite, hoehe = max(breite, kb), max(hoehe, kh)

	if breite == kb && hoehe == kh {
		return karte
	}

	links, oben := (breite-kb)/2, (hoehe-kh)/2
	mit := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))

	for y := range hoehe {
		quelle := flaeche.Min.Y + min(max(y-oben, 0), kh-1)

		for x := range breite {
			mit.SetNRGBA(x, y, karte.NRGBAAt(flaeche.Min.X+min(max(x-links, 0), kb-1), quelle))
		}
	}

	return mit
}
