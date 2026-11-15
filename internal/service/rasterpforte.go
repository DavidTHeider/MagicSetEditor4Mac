package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Die Pforte, durch die jede Rasterung geht** — Folgebefund F3 des
// Lastlaufs QA-08
// (#2020),
// Marke `new`: MSE2 hat weder Dienst noch Werkzeugserver.
//
// # Der Befund
//
// `domain.RessourceRendern` stand in den Vorgabegrenzen, `GET /grenzen` nannte
// zwei Plätze, und **kein Weg belegte je einen**. Acht Klienten, die zeichnen,
// bekamen deshalb nicht „zwei Plätze", wie der Lastlauf es deutete, sondern
// acht gleichzeitige Rasterungen. **Was zwei statt acht am Arbeitsspeicher
// ändern, ist nicht gemessen** — die 2,7 bis 3,7 GiB aus
// `docs/LASTMESSUNG.md` sind die von damals.
//
// # Warum eine Pforte am Rasterer und nicht ein Platz in jedem Dienst
//
// Andere schwere Arbeit belegt ihren Platz am Anfang des Dienstes
// ([haushaltswahl]). Für das Zeichnen gäbe das fünf Stellen — Kartenbild,
// Druckbogen, Vorschau im Fenster, Spieltextur, die Rasterteile einer SVG —,
// und jede neue, die rastert, müsste an die sechste denken. Alle gehen durch
// [render.Rasterer.Zeichne], und alle Flächen bauen **denselben** Rasterer
// (`app.NeuesZeichenwerk`): dort hängt die Pforte, und kein Weg kommt daran
// vorbei.
//
// **Ein Platz je Rasterung und nicht je Auftrag.** Ein Druckbogen hält für
// die Dauer seines Laufs den Platz der Klasse `ausfuhr` und belegt für jede
// Karte, die er zeichnet, einen der zwei Plätze `rendern` — die Reihenfolge
// ist immer *erst `ausfuhr`, dann `rendern`*, und `rendern` wartet auf nichts
// weiter. Damit gibt es keinen Kreis, in dem zwei Läufe aufeinander warten.
//
// # Was nicht durch die Pforte geht
//
// **Das Rastern eines einzelnen Symbols oder Bildbaums** (`SymbolAlsBild`,
// `BildAusBaum`, die Miniaturen der Symbolbibliothek): es ist klein, es kommt
// innerhalb anderer Arbeit vor, und eine Pforte davor hielte eine Rasterung
// auf, die selbst schon einen Platz hat.
//
// Dem Haushalt nach sind die Plätze **je Prozess**: ein zweites `mse-rest`
// hat seine eigenen. Das gilt für jede Klasse, nicht nur für diese.

// Rasterpforte lässt eine Rasterung erst beginnen, wenn ein Platz der Klasse
// [domain.RessourceRendern] frei ist.
type Rasterpforte struct {
	haushaltswahl
}

// Der Rasterer fragt sie, ohne `service` zu kennen.
var _ render.Pforte = (*Rasterpforte)(nil)

// NeueRasterpforte baut die Pforte über dem Haushalt der Maschine; ein
// anderer kommt über `MitHaushalt` — für Tests und für einen Prozess mit
// anderen Grenzen.
func NeueRasterpforte() *Rasterpforte { return &Rasterpforte{} }

// Belegen wartet auf einen Platz zum Rastern — oder bis der Kontext endet.
func (p *Rasterpforte) Belegen(ctx context.Context) (func(), error) {
	return p.platz(ctx, domain.RessourceRendern)
}
