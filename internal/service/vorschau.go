package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Planbauer macht aus einer Karte einen Zeichenplan.
//
// # Die Naht, die AGENTS.md §2 verlangt
//
// `render` sieht kein Skript, `msescript` kennt keine Karte. Zwischen beiden
// steht diese Stelle: sie nimmt die Kartenwerte, wertet die Feldskripte des
// Stils aus und legt das Ergebnis als Plan hin.
//
// # Der Behelfsplan bleibt der Rückfall
//
// [Kartenplan] liest den Stil über die eingerichtete [Stilquelle]. Fehlt sie
// oder trägt der Stil nicht, entsteht ein **Behelfsplan**: eine Kartenfläche,
// ein Namensfeld, ein Textfeld. Der Befund daneben macht sichtbar, dass dies
// nicht die Kartenausgabe des Stils ist.
type Planbauer struct {
	// stile findet das Stilpaket eines Sets — oder ist nil, dann behilft er.
	stile Stilquelle

	// schriften misst Textzüge für `content_width` — oder ist nil, dann bleibt
	// es bei der deklarierten Breite ([inhaltsmasse]).
	schriften render.Schriftquelle

	// bilder misst Symbolzüge für `content_width` — oder ist nil, dann gilt
	// das Quadrat der Symbolgröße.
	bilder render.Bildquelle
}

// NeuerPlanbauer baut den Planbauer.
func NeuerPlanbauer() *Planbauer { return &Planbauer{} }

// MitSchriften gibt dem Planbauer eine Schriftquelle für den **ersten** Schritt
// des zweistufigen Satzes.
//
// # Warum der Planbauer Schriften braucht, obwohl er nicht zeichnet
//
// `card_style.<feld>.content_width` fragt nach der Breite des gesetzten
// Inhalts, und Feldangaben rechnen damit, **bevor** der Rahmen feststeht. Ohne
// Schriftmaße lässt sich ein Textzug nicht messen — und ohne die Messung
// bekommen 500 Rechnungen in `M15-Magic-Pack` still ein falsches Ergebnis
// (#291).
//
// **Es ist eine Schnittstelle und keine Plattform.** `render.Schriftquelle` ist
// dieselbe, die der Rasterer bekommt; wo die Dateien liegen, weiß
// `internal/schriften`, und verdrahtet wird im Composition Root (§2).
//
// Ohne sie bleibt es bei der deklarierten Breite — dieselbe Antwort wie bisher,
// und keine erfundene.
func (p *Planbauer) MitSchriften(schriften render.Schriftquelle) *Planbauer {
	p.schriften = schriften

	return p
}

// MitBildern gibt dem Planbauer eine Bildquelle für den **ersten** Schritt des
// zweistufigen Satzes.
//
// # Warum der Planbauer Bilder braucht, obwohl er nicht zeichnet
//
// Aus demselben Grund wie bei [Planbauer.MitSchriften]: `content_width` ist
// eine **Messung**, und wie breit ein Symbolzug ist, hängt an den Bildern
// (Vermerk 0022). `magic-m15` setzt die rechte Kante des Kartennamens auf
//
//	right: { 341 - card_style.casting_cost.content_width }
//
// Ein zu schmal gemessener Manakostenzug macht den Namenskasten zu breit — auf
// **jeder** Karte
// (#291).
//
// **Es ist eine Schnittstelle und kein Paketleser.** `render.Bildquelle` ist
// dieselbe, die der Rasterer bekommt; wo die Dateien liegen, weiß `msepaket`,
// und verdrahtet wird im Composition Root (§2).
func (p *Planbauer) MitBildern(bilder render.Bildquelle) *Planbauer {
	p.bilder = bilder

	return p
}

// Kartenmasse sind die Maße einer Karte in Vorlagenpunkten.
//
// **Aus keiner Vorlage gelesen**, sondern das übliche Verhältnis einer
// Spielkarte (63 × 88 mm). Sobald ein Stil gelesen wird, kommen sie von dort.
const (
	Kartenbreite = 375.0
	Kartenhoehe  = 523.0
)

// Behelfsschrift ist die Schrift des Behelfsplans.
//
// **Eine, die sich auf macOS auch wirklich lesen lässt.** Sie liegt dort als
// Sammlung (`Helvetica.ttc`), und die Schriftsuche liest daraus die erste
// Schnittform. `Helvetica Neue` ginge nicht: `x/image` weist deren Sammlung mit
// „unsupported cmap encodings" ab — nachgesehen, nicht vermutet.
//
// Sobald ein Stil gelesen wird (#120), kommt der Name von dort.
const Behelfsschrift = "Helvetica"

// Behelfsplan legt die Werte einer Karte als Plan hin.
//
// Der `context` wird durchgereicht, damit der Aufrufer abbrechen kann — eine
// Vorschau wird abgebrochen, nicht abgewartet (§7).
func (p *Planbauer) Behelfsplan(ctx context.Context, karte *domain.CardRow) (*render.Plan, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	plan := &render.Plan{
		Breite: Kartenbreite,
		Hoehe:  Kartenhoehe,
		Ebenen: []render.Ebene{
			{
				Art:     render.ArtFlaeche,
				Rahmen:  render.Rechteck{Breite: Kartenbreite, Hoehe: Kartenhoehe},
				Farbe:   render.Farbe{R: 236, G: 232, B: 222, A: 255},
				Deckung: 1,
			},
			{
				Art:    render.ArtFlaeche,
				Rahmen: render.Rechteck{Links: 16, Oben: 16, Breite: Kartenbreite - 32, Hoehe: 40},
				Farbe:  render.Farbe{R: 202, G: 190, B: 160, A: 255},
				Verlauf: &render.Verlauf{Winkel: 90, Stufen: []render.Verlaufsstufe{
					{Stelle: 0, Farbe: render.Farbe{R: 214, G: 203, B: 174, A: 255}},
					{Stelle: 1, Farbe: render.Farbe{R: 178, G: 164, B: 130, A: 255}},
				}},
				Deckung: 1,
			},
		},
	}

	// Jeder Feldwert bekommt eine Zeile. **Alle**, nicht die bekannten: ein
	// fremdes Spiel führt Felder, die wir nicht kennen, und sie wegzulassen
	// hieße, die Vorschau ausgerechnet dort zu beschneiden, wo sie gebraucht
	// wird.
	oben := 76.0

	for _, wert := range karte.Values {
		zeile := render.Rechteck{Links: 24, Oben: oben, Breite: Kartenbreite - 48, Hoehe: 22}

		// **Auch der Behelf sagt, wo ein Feld liegt.** Ohne diese Zeile böte
		// der Editor auf einer Karte ohne Stil kein einziges Feld an — und
		// ohne Stil ist genau die Lage, in der jemand nachsehen will, was in
		// den Feldern steht.
		plan.Feldflaechen = append(plan.Feldflaechen,
			render.Feldflaeche{Feld: wert.Field, Rahmen: zeile})

		plan.Ebenen = append(plan.Ebenen, render.Ebene{
			Art: render.ArtText,

			// **Auch der Behelf ist anklickbar.** Er zeigt je Feldwert eine
			// Zeile; wer darauf klickt, meint dieses Feld — genauso wie auf
			// einer richtig gezeichneten Karte
			// (#246).
			Feld:       wert.Field,
			Rahmen:     zeile,
			Farbe:      render.Farbe{A: 255},
			Abschnitte: render.Aufrecht(wert.Field + ": " + wert.Text),
			Satz: render.Satz{
				Schrift: Behelfsschrift, Groesse: 13, Ausrichtung: render.LinksMitte,
			},
			Deckung: 1,
		})

		oben += 24
	}

	return plan, nil
}
