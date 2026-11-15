package service

import (
	"fmt"
	"math"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/pdf"
)

// bogenraster ist die Aufteilung einer Seite in Kartenplätze.
//
// **Eine eigene Datei, weil es eine eigene Frage ist** (AGENTS.md §2): dort
// steht, was ein Bogen ist, hier, wo eine Karte darauf sitzt.
type bogenraster struct {
	spalten int
	zeilen  int

	// zellbreite und zellhoehe sind Kartenmaß plus dem Abstand, den der
	// Anschnitt braucht — in Millimetern.
	zellbreite float64
	zellhoehe  float64

	// linksAussen und obenAussen sind die Ecke des Rasters auf der Seite.
	linksAussen float64
	obenAussen  float64
}

// raster teilt eine Seite in so viele Plätze, wie darauf passen.
//
// **Zentriert und nicht linksbündig.** Ein Bogen, dessen Rest rechts und unten
// liegt, sieht nach einem Fehler aus; einer mit gleichen Rändern nach Absicht —
// und beim Schneiden ist der Abstand zum Papierrand auf beiden Seiten gleich.
func (m Bogenmasse) raster() bogenraster {
	// **Der Abstand ist der doppelte Anschnitt**: jede Karte bringt ihren
	// eigenen mit, und zwei benachbarte brauchen beide ihren.
	zellbreite := m.Kartenbreite + 2*m.Anschnitt
	zellhoehe := m.Kartenhoehe + 2*m.Anschnitt

	nutzbarBreit := m.Seitenbreite - 2*m.Rand
	nutzbarHoch := m.Seitenhoehe - 2*m.Rand

	spalten := int(math.Floor(nutzbarBreit / zellbreite))
	zeilen := int(math.Floor(nutzbarHoch / zellhoehe))

	return bogenraster{
		spalten: spalten, zeilen: zeilen,
		zellbreite: zellbreite, zellhoehe: zellhoehe,
		linksAussen: (m.Seitenbreite - float64(spalten)*zellbreite) / 2,
		obenAussen:  (m.Seitenhoehe - float64(zeilen)*zellhoehe) / 2,
	}
}

// stelle gibt die linke untere Ecke der **Schnittkante** eines Platzes.
//
// In Punkten, und schon in PDF-Koordinaten: der Ursprung liegt unten links,
// gezählt wird aber von oben — die erste Karte steht oben links auf dem Papier,
// nicht unten.
func (r bogenraster) stelle(platz int, masse Bogenmasse) (links, unten float64) {
	spalte := platz % r.spalten
	zeile := platz / r.spalten

	linksMM := r.linksAussen + float64(spalte)*r.zellbreite + masse.Anschnitt
	obenMM := r.obenAussen + float64(zeile)*r.zellhoehe + masse.Anschnitt

	return pdf.MM(linksMM), pdf.MM(masse.Seitenhoehe - obenMM - masse.Kartenhoehe)
}

// markenlaenge ist die Länge einer Schnittmarke in Millimetern.
const markenlaenge = 4

// marken zieht die Schnittlinien an den Papierrand.
//
// **An den Rand und nicht auf die Karte.** Eine Marke über dem Bild wäre eine,
// die man mitdruckt und danach ansieht. Gezogen wird je Kartenkante ein kurzer
// Strich außen — vier Kanten je Karte, jede zweimal.
func (r bogenraster) marken(seite *pdf.Seite, masse Bogenmasse, karten int) {
	const staerke, grau = 0.25, 0.4

	spaltenBelegt := min(karten, r.spalten)
	zeilenBelegt := (karten + r.spalten - 1) / r.spalten

	for spalte := range spaltenBelegt {
		for _, kante := range r.senkrechteKanten(spalte, masse) {
			seite.Linie(kante, pdf.MM(0), kante, pdf.MM(markenlaenge), staerke, grau)
			seite.Linie(kante, pdf.MM(masse.Seitenhoehe-markenlaenge), kante,
				pdf.MM(masse.Seitenhoehe), staerke, grau)
		}
	}

	for zeile := range zeilenBelegt {
		for _, kante := range r.waagerechteKanten(zeile, masse) {
			seite.Linie(pdf.MM(0), kante, pdf.MM(markenlaenge), kante, staerke, grau)
			seite.Linie(pdf.MM(masse.Seitenbreite-markenlaenge), kante,
				pdf.MM(masse.Seitenbreite), kante, staerke, grau)
		}
	}
}

// senkrechteKanten gibt die beiden Schnittkanten einer Spalte in Punkten.
func (r bogenraster) senkrechteKanten(spalte int, masse Bogenmasse) [2]float64 {
	links := r.linksAussen + float64(spalte)*r.zellbreite + masse.Anschnitt

	return [2]float64{pdf.MM(links), pdf.MM(links + masse.Kartenbreite)}
}

// waagerechteKanten gibt die beiden Schnittkanten einer Zeile in Punkten.
func (r bogenraster) waagerechteKanten(zeile int, masse Bogenmasse) [2]float64 {
	oben := r.obenAussen + float64(zeile)*r.zellhoehe + masse.Anschnitt

	return [2]float64{
		pdf.MM(masse.Seitenhoehe - oben),
		pdf.MM(masse.Seitenhoehe - oben - masse.Kartenhoehe),
	}
}

// pruefe weist Maße ab, mit denen kein Bogen entstehen kann.
func (m Bogenmasse) pruefe() error {
	switch {
	case m.Kartenbreite <= 0 || m.Kartenhoehe <= 0:
		return fmt.Errorf("%w: eine Karte ohne Maß", domain.ErrValidation)

	case m.Anschnitt < 0 || m.Rand < 0:
		return fmt.Errorf("%w: Anschnitt und Rand können nicht kleiner als null sein",
			domain.ErrValidation)

	case m.DPI <= 0 || m.DPI > GrenzeDPI:
		return fmt.Errorf("%w: %d dpi liegen außerhalb von 1 bis %d",
			domain.ErrValidation, m.DPI, GrenzeDPI)
	}

	if m.raster().spalten < 1 || m.raster().zeilen < 1 {
		return fmt.Errorf("%w: auf %g × %g mm passt bei %g mm Rand keine Karte von %g × %g mm",
			domain.ErrValidation, m.Seitenbreite, m.Seitenhoehe,
			m.Rand, m.Kartenbreite, m.Kartenhoehe)
	}

	return nil
}
