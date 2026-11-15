package service

import (
	"context"
	"math"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Bogenprobelauf sagt, wie ein Bogen würde — ohne dass er entsteht
// (#2015,
// UXR-F7).
//
// # Was darin steht, und was nicht
//
// Es sind die Zahlen, die der echte Lauf aus demselben Plan rechnet: wie viele
// Karten, wie viele Plätze je Seite, wie viele Seiten. **Nicht darin steht,
// was erst das Zeichnen zeigt** — ein Plan, der sich nicht bauen lässt, eine
// Vorlage im falschen Verhältnis, eine fehlende Schrift. Das sind Befunde des
// Laufs; der Probelauf kommt nie dorthin, und was er nicht gesehen hat, sagt
// er nicht.
type Bogenprobelauf struct {
	// Karten sind die Karten, die auf den Bogen kämen — nach der Auswahl.
	Karten int

	// Seiten ist die Zahl der Bögen.
	Seiten int

	// Spalten und Zeilen sind das Raster einer Seite, JeSeite die Plätze.
	Spalten int
	Zeilen  int
	JeSeite int

	// AufLetzterSeite sagt, wie viele Karten auf der letzten Seite stehen —
	// eine halbe Seite sieht, wer sie druckt, erst am Papier.
	AufLetzterSeite int

	// Bildbreite ist die Breite einer Karte in Bildpunkten, mit der der Lauf
	// rastert: die gedruckte Breite samt Anschnitt bei der gewählten
	// Auflösung.
	Bildbreite int

	// Masse sind die Maße, nach denen gerechnet wurde.
	Masse Bogenmasse
}

// Probelauf rechnet einen Bogen durch, ohne ihn zu zeichnen oder zu schreiben.
//
// # Ein Weg, zwei Ausgänge
//
// **Derselbe Plan wie [Bogenleger.Lege]** ([bogenplan]): Maße, Set, Auswahl
// und das leere Set werden hier wie dort geprüft und mit demselben Wortlaut
// abgewiesen. Ein Test fährt beide mit derselben fehlerhaften Eingabe und
// vergleicht.
//
// # Was er nicht fragt
//
// **Den Schreibriegel nicht und den Platz im Haushalt nicht**: der Probelauf
// schreibt nichts und rastert nichts. Ein Leser, der einen Bogen nicht legen
// darf, darf trotzdem wissen, wie er würde — der Riegel fragt dort, wo etwas
// auf die Platte käme.
func (b *Bogenleger) Probelauf(
	ctx context.Context, setID int64, masse Bogenmasse, auswahl domain.Kartenauswahl,
) (*Bogenprobelauf, error) {
	plan, err := b.plane(ctx, setID, masse, auswahl)
	if err != nil {
		return nil, err
	}

	// **Die Seitenzahl ist MSE2s Rechnung** — adapted: das Raster entsteht hier
	// aus Papier, Rand und Anschnitt, die Teilung bleibt dieselbe.
	//
	//	int cards_per_page = max(1,layout.cards_per_page());
	//	return ((int)cards.size() + cards_per_page - 1) / cards_per_page;
	//
	// `PrintJob::num_pages`, `src/gui/print_window.hpp`, Zeilen 55–57, Fassung
	// f9d9356 vom 2021-07-06. Dort steht sie hinter der **Druckvorschau**
	// (`SetWindow::onFilePrintPreview`, `src/gui/set/window.cpp`, Zeilen
	// 709–713), die die Seiten zeichnet; der Probelauf zählt sie nur.
	//
	// **Eine volle letzte Seite trägt `jeSeite`, nicht null.**
	rest := len(plan.karten) % plan.jeSeite
	if rest == 0 {
		rest = plan.jeSeite
	}

	return &Bogenprobelauf{
		Karten:          len(plan.karten),
		Seiten:          (len(plan.karten) + plan.jeSeite - 1) / plan.jeSeite,
		Spalten:         plan.raster.spalten,
		Zeilen:          plan.raster.zeilen,
		JeSeite:         plan.jeSeite,
		AufLetzterSeite: rest,
		Bildbreite:      int(math.Round(bildbreite(masse))),
		Masse:           masse,
	}, nil
}

// bildbreite ist die Zielbreite einer Karte in Bildpunkten: so viele, wie die
// gedruckte Breite samt Anschnitt bei dieser Auflösung trägt.
//
// **Eine Rechnung für beide Wege**: [Bogenleger.zeichne] rastert mit ihr, der
// Probelauf rundet sie nur.
func bildbreite(masse Bogenmasse) float64 {
	return (masse.Kartenbreite + 2*masse.Anschnitt) * float64(masse.DPI) / 25.4
}
