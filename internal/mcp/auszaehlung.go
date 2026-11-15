// **Auszählen nach einer erklärten Dimension, über MCP** — Werkstück C von
// Phase 4.11 (#1338).
//
// Einordnung nach §16: **erweiternd** — zwei Werkzeuge kommen hinzu, beide
// lesend.

package mcp

import (
	"context"
	"encoding/json"
)

// Auszaehlung ist, was MCP für die Auszählung eines Sets braucht.
type Auszaehlung interface {
	Angebot(ctx context.Context, setID int64) (any, error)
	Auszaehlen(ctx context.Context, setID int64, name string) (any, error)
}

// auszaehlung gibt die beiden Werkzeuge — beide lesend.
func auszaehlung(zaehlung Auszaehlung) []Werkzeug {
	if zaehlung == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.auszaehldimensionen", Art: Lesend,
		Erklaerung: "Nennt, wonach dieses Spiel auszählen lässt — Farbgruppe, " +
			"Manawert, Seltenheit, Kartentyp und so fort. Die Dimensionen " +
			"stehen im Spielpaket und rechnen ihren Wert je Karte mit einem " +
			"Skript; ein Set ohne Spielpaket hat keine. ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"setID": zahl("Das Set, dessen Spielpaket gefragt ist."),
		}, "setID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			setID, err := kennung(roh)
			if err != nil {
				return "", err
			}

			angebot, fehler := zaehlung.Angebot(ctx, setID)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"Dimensionen": angebot})
		},
	}, {
		Name: "lesen.auszaehlung", Art: Lesend,
		Erklaerung: "Zählt ein Set nach einer Dimension aus und gibt die " +
			"Gruppen in der Reihenfolge, die das Spielpaket erklärt — samt " +
			"den leeren — und zu jeder die Karten (»kartenIDs«, aufsteigend). " +
			"Der Name ist der aus »lesen.auszaehldimensionen«. " +
			"ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"setID": zahl("Das Set, das gezählt werden soll."),
			"name": text(
				"Die Dimension, wie »lesen.auszaehldimensionen« sie nennt."),
		}, "setID", "name"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			// **Nicht [kennung]**: die deutet **alle** Argumente als Zahlen,
			// und `name` ist eine Zeichenkette. Dieselbe Gestalt wie bei
			// `schreiben.kartenskript`.
			var argumente struct {
				SetID int64  `json:"setID"`
				Name  string `json:"name"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if err := pruefeKennung("setID", argumente.SetID); err != nil {
				return "", err
			}

			gezaehlt, fehler := zaehlung.Auszaehlen(ctx, argumente.SetID, argumente.Name)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(gezaehlt)
		},
	}}
}
