package app

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// BogenprobelaufAntwort sagt, wie ein Druckbogen würde — vor dem Legen
// (#2015,
// UXR-F7).
//
// Die Zahlen sind die des Dienstes ([service.Bogenprobelauf]); die Maße reisen
// mit, damit die Oberfläche „A4 · 300 dpi" aus derselben Antwort sagt.
type BogenprobelaufAntwort struct {
	Karten          int `json:"karten"`
	Seiten          int `json:"seiten"`
	Spalten         int `json:"spalten"`
	Zeilen          int `json:"zeilen"`
	JeSeite         int `json:"jeSeite"`
	AufLetzterSeite int `json:"aufLetzterSeite"`

	// Bildbreite ist die Breite einer Karte in Bildpunkten.
	Bildbreite int `json:"bildbreite"`

	// Maße in Millimetern.
	Seitenbreite float64 `json:"seitenbreite"`
	Seitenhoehe  float64 `json:"seitenhoehe"`
	Kartenbreite float64 `json:"kartenbreite"`
	Kartenhoehe  float64 `json:"kartenhoehe"`
	Anschnitt    float64 `json:"anschnitt"`
	Rand         float64 `json:"rand"`

	DPI           int  `json:"dpi"`
	Schnittmarken bool `json:"schnittmarken"`

	Fehler *Fehler `json:"fehler"`
}

// BogenProbelauf rechnet einen Druckbogen durch, ohne ihn zu legen.
//
// # Ein Weg, zwei Ausgänge
//
// **Dieselben Maße und derselbe Dienst wie [App.BogenLegen]**
// ([bogenmasseAus], [service.Bogenleger.Probelauf]): was dort abgewiesen wird,
// wird hier mit demselben Satz abgewiesen. Geschrieben wird nichts, und der
// Schreibriegel (0078) wird nicht gefragt — der Probelauf liest.
func (a *App) BogenProbelauf(
	ctx context.Context, setID, papier string, anschnitt float64, dpi int,
	marken bool, auswahl []string,
) BogenprobelaufAntwort {
	ctx = a.imProfil(ctx)

	kennung, err := strconv.ParseInt(setID, 10, 64)
	if err != nil {
		return BogenprobelaufAntwort{Fehler: Uebersetze(ctx, err)}
	}

	gewaehlt, auswahlbefund := kartenauswahlAus(auswahl)
	if auswahlbefund != nil {
		return BogenprobelaufAntwort{Fehler: auswahlbefund}
	}

	probe, err := a.bogenleger.Probelauf(ctx, kennung,
		bogenmasseAus(papier, anschnitt, dpi, marken), gewaehlt)
	if err != nil {
		return BogenprobelaufAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return bogenprobelaufAlsAntwort(probe)
}

// bogenprobelaufAlsAntwort trägt den Probelauf des Dienstes in die Antwort ein.
func bogenprobelaufAlsAntwort(probe *service.Bogenprobelauf) BogenprobelaufAntwort {
	return BogenprobelaufAntwort{
		Karten: probe.Karten, Seiten: probe.Seiten, Spalten: probe.Spalten,
		Zeilen: probe.Zeilen, JeSeite: probe.JeSeite, AufLetzterSeite: probe.AufLetzterSeite,
		Bildbreite:   probe.Bildbreite,
		Seitenbreite: probe.Masse.Seitenbreite, Seitenhoehe: probe.Masse.Seitenhoehe,
		Kartenbreite: probe.Masse.Kartenbreite, Kartenhoehe: probe.Masse.Kartenhoehe,
		Anschnitt: probe.Masse.Anschnitt, Rand: probe.Masse.Rand,
		DPI: probe.Masse.DPI, Schnittmarken: probe.Masse.Schnittmarken,
	}
}
