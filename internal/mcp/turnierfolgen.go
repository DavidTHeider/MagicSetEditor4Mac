// **Die Folgen einer Berichtigung** — DOM-04.AK-02
// (#1699).
//
// Einordnung nach §16: **erweiternd** — ein lesendes Werkzeug kommt hinzu.

package mcp

import (
	"context"
	"encoding/json"
)

// lesenKorrekturfolgen sagt, was eine Berichtigung verschöbe — **ohne sie
// einzutragen**. Dieselben Felder wie `schreiben.turnier.ergebnis`.
func lesenKorrekturfolgen(naht Turniere) Werkzeug {
	return Werkzeug{
		Name: "lesen.turnier.folgen", Art: Lesend,
		Erklaerung: "Sagt, was ein neues Ergebnis an einer Paarung VERSCHÖBE: " +
			"wer Platz oder Punkte ändert und welche späteren Runden nach dem " +
			"alten Stand gepaart wurden — sie bleiben, wie sie sind. Dieselben " +
			"Felder wie »schreiben.turnier.ergebnis«. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"turnierID":     zahl("Die Kennung des Turniers aus »lesen.turniere«."),
			"paarungID":     zahl("Die Kennung der Paarung aus »lesen.turnier«."),
			"spieleA":       zahl("Gewonnene Spiele des Erstgenannten."),
			"spieleB":       zahl("Gewonnene Spiele des Zweitgenannten."),
			"unentschieden": zahl("Unentschiedene Spiele."),
			"zuruecknehmen": jaNein("Wahr fragt nach den Folgen des Zurücknehmens."),
		}, "turnierID", "paarungID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				TurnierID     int64 `json:"turnierID"`
				PaarungID     int64 `json:"paarungID"`
				SpieleA       int   `json:"spieleA"`
				SpieleB       int   `json:"spieleB"`
				Unentschieden int   `json:"unentschieden"`
				Zuruecknehmen bool  `json:"zuruecknehmen"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			var stand map[string]any

			if !wahl.Zuruecknehmen {
				stand = map[string]any{
					"spieleA": wahl.SpieleA, "spieleB": wahl.SpieleB,
					"unentschieden": wahl.Unentschieden,
				}
			}

			folgen, fehler := naht.TurnierKorrekturfolgen(ctx, wahl.TurnierID, wahl.PaarungID, stand)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(folgen)
		},
	}
}
