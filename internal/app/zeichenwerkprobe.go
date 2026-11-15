package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bogenOhneFenster sind Maße und Auswahl des Bogens an den Flächen ohne
// Fenster: A4, neun Karten je Seite, 300 dpi, Schnittmarken, das ganze Set.
//
// **Eine Stelle für den Lauf und den Probelauf** (#2015): was `ProbeBogen`
// sagt, ist, was `LegeBogen` täte.
func bogenOhneFenster() bogenvorgabe {
	return bogenvorgabe{masse: service.StandardBogen()}
}

// bogenvorgabe sind Maße und Auswahl eines Bogens ohne Fenster; die Auswahl
// ist leer und heißt „das ganze Set" (`domain.Kartenauswahl`).
type bogenvorgabe struct {
	masse   service.Bogenmasse
	auswahl domain.Kartenauswahl
}

// Bogenprobe sagt, wie ein Druckbogen würde — vor dem Legen
// (#2015).
//
// **Ohne `service`-Typen**, aus demselben Grund wie [Bogenlegung]: eine Fläche
// ohne Fenster liest sie, ohne `internal/service` zu kennen (§2).
type Bogenprobe struct {
	Karten          int
	Seiten          int
	Spalten         int
	Zeilen          int
	JeSeite         int
	AufLetzterSeite int

	// Bildbreite ist die Breite einer Karte in Bildpunkten.
	Bildbreite int

	// Maße in Millimetern.
	Seitenbreite float64
	Seitenhoehe  float64
	Kartenbreite float64
	Kartenhoehe  float64
	Anschnitt    float64
	Rand         float64

	DPI           int
	Schnittmarken bool
}

// ProbeBogen rechnet den Bogen durch, den [Zeichenwerk.LegeBogen] legte — und
// legt ihn nicht.
//
// **Derselbe Dienst, dieselben Maße**: die Fehler sind die des Legens (ein Set,
// das es nicht gibt, ein Set ohne Karte), und der Schreibriegel wird nicht
// gefragt.
func (z *Zeichenwerk) ProbeBogen(ctx context.Context, setID int64) (*Bogenprobe, error) {
	vorgabe := bogenOhneFenster()
	masse, auswahl := vorgabe.masse, vorgabe.auswahl

	probe, err := z.Bogenleger.Probelauf(ctx, setID, masse, auswahl)
	if err != nil {
		return nil, err
	}

	return &Bogenprobe{
		Karten: probe.Karten, Seiten: probe.Seiten, Spalten: probe.Spalten,
		Zeilen: probe.Zeilen, JeSeite: probe.JeSeite, AufLetzterSeite: probe.AufLetzterSeite,
		Bildbreite:   probe.Bildbreite,
		Seitenbreite: probe.Masse.Seitenbreite, Seitenhoehe: probe.Masse.Seitenhoehe,
		Kartenbreite: probe.Masse.Kartenbreite, Kartenhoehe: probe.Masse.Kartenhoehe,
		Anschnitt: probe.Masse.Anschnitt, Rand: probe.Masse.Rand,
		DPI: probe.Masse.DPI, Schnittmarken: probe.Masse.Schnittmarken,
	}, nil
}
