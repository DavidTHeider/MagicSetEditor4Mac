package mcp

// **Einen Deckplatz in den anderen Bereich verschieben** — Befund 7 des
// Zweitblicks über Phase 5.1
// (#1405).
//
// **Eine eigene Datei**, damit `decksschreiben.go` unter der Zeilengrenze aus
// §2 bleibt.

import (
	"context"
	"encoding/json"
)

// schreibenDeckplatzVerschieben legt einen Platz in den anderen Bereich.
func schreibenDeckplatzVerschieben(naht Decks) Werkzeug {
	return Werkzeug{
		Name: "schreiben.deckplatz.verschieben", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: legt einen Platz vom Hauptdeck ins " +
			"Sideboard oder zurück — in EINEM Zug. Liegt dort schon derselbe " +
			"Bezug, gehen die Mengen zusammen.",
		Schema: objekt(map[string]any{
			"deckID":  zahl("Die Kennung des Decks."),
			"platzID": zahl("Die Kennung des Platzes aus »lesen.deck«."),
			"bereich": text("»hauptdeck« oder »sideboard«."),
		}, "deckID", "platzID", "bereich"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				DeckID  int64  `json:"deckID"`
				PlatzID int64  `json:"platzID"`
				Bereich string `json:"bereich"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			platz, fehler := naht.DeckplatzVerschieben(
				ctx, wahl.DeckID, wahl.PlatzID, wahl.Bereich)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(platz)
		},
	}
}
