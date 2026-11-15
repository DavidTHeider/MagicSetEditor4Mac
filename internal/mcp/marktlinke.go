// **Die Wege nach draußen über MCP** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// # Ein Werkzeug und nicht zwei
//
// **Ein Modell nennt Wege und öffnet keine.** Ein Browser, den ein Werkzeug
// aufmacht, steht auf **diesem** Bildschirm; die Entscheidung dazu trifft der
// Mensch davor, und sie überlebt den Ruf
// ([0106](../../docs/entscheidungen/0106-drei-vertrauensgrenzen-und-nur-eine-wird-gebaut.md)).
//
// **Die Adressen selbst sind unbedenklich**: sie entstehen auf diesem Rechner
// aus einem Kartennamen oder einer Anbieterkennung, sie tragen kein
// Geheimnis, und sie gehen durch dieselbe Positivliste wie am Fenster.

package mcp

import (
	"context"
	"encoding/json"
)

// Linkauskunft ist die Naht zu den Wegen nach draußen.
type Linkauskunft interface {
	Marktlinks(ctx context.Context, printing, name string) ([]Marktlink, error)
}

// Marktlink ist ein Weg nach draußen.
type Marktlink struct {
	Anbieter string `json:"anbieter"`

	// Art ist `produkt`, `suche` oder `startseite`.
	//
	// **Sie steht dabei, weil sie die Genauigkeit ist**: ein Produktlink
	// führt auf **diese** Karte, ein Suchlink auf eine Liste, in der sie
	// vermutlich steht.
	Art string `json:"art"`

	// Herkunft ist `kennung`, `name` oder `keine`.
	Herkunft string `json:"herkunft"`

	Adresse string `json:"adresse"`
}

// marktlinkwerkzeug gibt das eine lesende — oder keines.
//
// **Ohne Naht kein Werkzeug**, wie bei den Umschlägen und aus demselben
// Grund: ein Name in der Liste, hinter dem ein Ruf abstürzt, ist eine Zusage,
// die keine ist (§10).
//
// **Und es ist idempotent.** Die Adressen entstehen auf diesem Rechner aus
// einer Kennung und einem Namen; es geht nichts hinaus, und zweimal gefragt
// ist zweimal dieselbe Antwort. Daran kürzt ein vorsichtiges Gegenüber seine
// Liste.
func marktlinkwerkzeug(auskunft Linkauskunft) []Werkzeug {
	if auskunft == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.marktlinks", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt je Anbieter den genauesten Weg zu einer Karte: " +
			"Cardmarket, CardTrader, TCGplayer, Card Kingdom, Gatherer und " +
			"MTGGoldfish. »art« sagt, wie genau der Weg trifft — »produkt« " +
			"führt auf DIESE Karte, »suche« auf eine Liste, »startseite« " +
			"nur zum Anbieter. Eine eigene Karte ohne kanonischen Druck " +
			"bekommt KEINE geratenen Produktlinks. Es wird nichts geöffnet " +
			"und nichts abgerufen: die Adressen entstehen auf diesem " +
			"Rechner.",
		Schema: objekt(map[string]any{
			"printing": text("Die Druckkennung aus »lesen.karten« — leer für " +
				"eine eigene Karte ohne kanonischen Druck."),
			"name": text("Der Kartenname; er trägt die Suche. Ohne ihn und " +
				"ohne Druck bleiben die Startseiten."),
		}),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Printing string `json:"printing"`
				Name     string `json:"name"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			links, err := auskunft.Marktlinks(ctx, argumente.Printing,
				argumente.Name)
			if err != nil {
				return "", err
			}

			return alsJSON(links)
		},
	}}
}
