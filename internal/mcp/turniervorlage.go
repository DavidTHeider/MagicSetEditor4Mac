// **Ein Turnier nach einer Vorlage** — DOM-04.AK-01 und AK-03
// (#1699).
//
// Einordnung nach §16: **erweiternd** — ein Werkzeug kommt hinzu.

package mcp

import (
	"context"
	"encoding/json"
)

// schreibenTurnierKopieren legt ein Turnier nach einer Vorlage an.
func schreibenTurnierKopieren(naht Turniere) Werkzeug {
	return Werkzeug{
		Name: "schreiben.turnier.kopieren", Art: Schreibend,
		Erklaerung: "Legt ein NEUES Turnier nach einer Vorlage an: Spiel, Format, " +
			"Einzelspiel und Notiz kommen mit, auf Wunsch auch die NAMEN der " +
			"Teilnehmenden. NICHT mit kommen Decks und ihre Listen — eine " +
			"eingereichte Liste gilt für ein Turnier —, Runden, Ergebnisse " +
			"und wer ausgestiegen ist.",
		Schema: objekt(map[string]any{
			"turnierID": zahl("Die Kennung der Vorlage aus »lesen.turniere«."),
			"name":      text("Der Name des neuen Turniers."),
			"mitTeilnehmenden": jaNein("Die Namen der Teilnehmenden übernehmen — " +
				"ohne Decks."),
		}, "turnierID", "name"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var wahl struct {
				TurnierID        int64  `json:"turnierID"`
				Name             string `json:"name"`
				MitTeilnehmenden bool   `json:"mitTeilnehmenden"`
			}

			if err := lies(roh, &wahl); err != nil {
				return "", err
			}

			kopie, fehler := naht.TurnierKopieren(
				ctx, wahl.TurnierID, wahl.Name, wahl.MitTeilnehmenden)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(kopie)
		},
	}
}
