package service_test

// **Der Texturvertrag** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Wogegen hier geprüft wird
//
// **Dass derselbe Inhalt dasselbe Bild trifft und jede Änderung es verwirft.**
// Das Issue nennt die Klassen einzeln — Seite, Bild, Font, Stil, Pixelgröße,
// Renderer —, und jede steht hier als eigener Fall: „irgendetwas ändert den
// Schlüssel" wäre keine Aussage.
//
// **Und dass dieselbe lokale Kartenkennung mit anderem Inhalt nicht trifft.**
// Das verlangt das Issue wörtlich, und es ist der Fall, an dem ein Cache über
// der Zeilennummer zerbricht.

import (
	"context"
	"image"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// texturplan gibt einen Plan, der etwas zeichnet.
func texturplan(bildpfad, schrift string) *render.Plan {
	return &render.Plan{
		Breite: 375, Hoehe: 523, DPI: 150,
		Ebenen: []render.Ebene{
			{
				Art:  render.ArtBild,
				Bild: render.Bildbaum{Knotenart: "datei", Pfad: bildpfad},
			},
			{Art: render.ArtText, Feld: "name", Satz: render.Satz{Schrift: schrift}},
		},
	}
}

// planprobe gibt immer denselben Plan — oder klagt.
type planprobe struct {
	plan    *render.Plan
	befunde []domain.Befund
	klagt   error

	gefragt int
}

func (p *planprobe) Kartenplan(_ context.Context, _ *domain.CardRow) (
	*render.Plan, []domain.Befund, error,
) {
	p.gefragt++

	if p.klagt != nil {
		return nil, nil, p.klagt
	}

	return p.plan, p.befunde, nil
}

// rasterprobe zählt, wie oft wirklich gezeichnet wurde.
type rasterprobe struct {
	klagt   error
	befunde []render.Befund

	gezeichnet     int
	breite         int
	hoehe          int
	vergroesserung float64
	guete          render.Guete
}

func (r *rasterprobe) Zeichne(
	_ context.Context, plan *render.Plan, vergroesserung float64,
	guete render.Guete,
) (*image.NRGBA, []render.Befund, error) {
	r.gezeichnet++
	r.guete = guete
	r.vergroesserung = vergroesserung

	if r.klagt != nil {
		return nil, nil, r.klagt
	}

	// **Breite *und Höhe* folgen dem Plan** — wie beim echten Rasterer.
	//
	// **Hier stand ein Quadrat**, und das hat eine Zusage unsichtbar
	// gemacht: der Fehltreffer meldete die Maße des Bildes, der Treffer
	// rechnete sie aus dem Plan, und beide waren gleich, solange das Bild
	// quadratisch war. Eine Vorrichtung, die anders geformt ist als der
	// Betrieb, prüft die Vorrichtung (§3).
	r.breite = int(plan.Breite * vergroesserung)
	r.hoehe = int(plan.Hoehe * vergroesserung)

	// **Die linke Hälfte rot, die rechte blau.**
	//
	// **Ein leeres Bild machte die eigentliche Aussage unsichtbar**: dass
	// die zweite Seite die **rechte** Hälfte zeigt, ist eine Aussage über
	// Bildpunkte. Mit lauter Nullen bleibt ein Lauf grün, der beide Seiten
	// links ausschneidet — gemessen am 2026-09-15 als Gegenprobe (§10).
	gemalt := image.NewNRGBA(image.Rect(0, 0, r.breite, r.hoehe))

	for x := range r.breite {
		farbe := service.LinkeFarbe
		if x >= r.breite/2 {
			farbe = service.RechteFarbe
		}

		for y := range r.hoehe {
			gemalt.SetNRGBA(x, y, farbe)
		}
	}

	return gemalt, r.befunde, nil
}

// ressourcenprobe gibt Abdrücke, die sich stellen lassen.
// **Sie klagt nicht, sie gibt einen Wert.**
//
// Eine fehlende Datei ist ein Befund und kein Abbruch — der Grund steht an
// [service.Texturschluessel]. Die Vorrichtung bildet das ab: was sie nicht
// kennt, bekommt einen **anderen** Abdruck und damit einen anderen Schlüssel.
type ressourcenprobe struct {
	bilder    map[string]string
	schriften map[string]string

	// ruecken sagt je Spiel, wo sein Kartenrücken liegt.
	ruecken map[string]string
}

func (r ressourcenprobe) Bildabdruck(pfad string) string {
	abdruck, da := r.bilder[pfad]
	if !da {
		return "fehlt:" + pfad
	}

	return abdruck
}

func (r ressourcenprobe) Rueckenbild(spiel string) string {
	return r.ruecken[spiel]
}

func (r ressourcenprobe) Schriftabdruck(name string) string {
	abdruck, da := r.schriften[name]
	if !da {
		return "fehlt:" + name
	}

	return abdruck
}

// eineWerkstatt baut den Dienst über Doppeln.
func eineWerkstatt(t *testing.T) (
	*service.Spieltexturen, *planprobe, *rasterprobe, *service.Texturfachprobe,
) {
	t.Helper()

	plaene := &planprobe{plan: texturplan("rahmen.png", "MPlantin")}
	rasterer := &rasterprobe{}
	ablage := service.NeueTexturfachprobe()

	quelle := ressourcenprobe{
		bilder: map[string]string{
			"rahmen.png":                   "bild-1",
			"magic.mse-game/card-back.png": "ruecken-1",
		},
		schriften: map[string]string{"MPlantin": "schrift-1"},

		// **Nur `magic` bringt einen mit** — `saat` nicht, und genau der
		// Unterschied wird gemessen.
		ruecken: map[string]string{
			"magic": "/magic.mse-game/card-back.png",
		},
	}

	return service.NeueSpieltexturen(plaene, rasterer, quelle, ablage),
		plaene, rasterer, ablage
}

// eineTexturkarte gibt eine Karte mit Kennung.
//
// **Eigen und nicht [eineKarte] daneben**: jene gibt eine Liste für die
// Bildausfuhr, hier zählt die Kennung — und zwei Vorrichtungen mit demselben
// Namen wären in jedem Review eine Rückfrage.
func eineTexturkarte(id int64) *domain.CardRow {
	return &domain.CardRow{ID: id}
}
