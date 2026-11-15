package mcp

// **Aus einem Set ein Deck** — FMT-04
// (#1693).

import (
	"context"
	"encoding/json"
)

// schreibenDeckAusSet legt aus den Karten eines Sets ein Deck an.
func schreibenDeckAusSet(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "schreiben.deck.ausset", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: legt aus den Karten eines Sets ein neues Deck an — " +
			"jede Karte ein Platz im Hauptdeck mit derselben Menge, in einem Zug. Die " +
			"Plätze verweisen auf die Karten im Set, sie werden nicht kopiert; in der " +
			"Sammlung entsteht nichts. Ohne Namen heißt das Deck wie das Set.",
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets."),
			"name":  text("Der Name des Decks; leer heißt: wie das Set."),
			"menge": zahl("Wie oft jede Karte ins Deck kommt; mindestens 1."),
		}, "setID", "menge"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				SetID int64  `json:"setID"`
				Name  string `json:"name"`
				Menge int    `json:"menge"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			deck, fehler := naht.DeckAusSet(ctx, wahl.SetID, wahl.Name, wahl.Menge)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(deck)
		},
	}
}
