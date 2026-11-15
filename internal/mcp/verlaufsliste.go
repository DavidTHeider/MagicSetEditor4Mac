package mcp

import (
	"context"
	"encoding/json"
)

// **Der Verlauf über MCP** — Werkstück U6 von Phase 5.1
// (#1417).
// Ein Modell, das etwas zurücknehmen soll, braucht den **Vorgang** — bis
// hierher kannte es ihn nur aus der Antwort seiner eigenen Änderung.

// Verlaufsschritt ist ein Schritt des Verlaufs, wie der Dienst ihn nennt —
// dieselbe Gestalt wie am Fenster.
type Verlaufsschritt struct {
	ID              string `json:"id"`
	Zeitpunkt       string `json:"zeitpunkt"`
	Handlung        string `json:"handlung"`
	KarteID         string `json:"karteID"`
	Feld            string `json:"feld"`
	Vorher          string `json:"vorher"`
	GabVorher       bool   `json:"gabVorher"`
	Nachher         string `json:"nachher"`
	GabNachher      bool   `json:"gabNachher"`
	Zurueckgenommen bool   `json:"zurueckgenommen"`
	Urheber         string `json:"urheber"`
	Herkunft        string `json:"herkunft"`
	Vorgang         string `json:"vorgang"`
	Umfang          int    `json:"umfang"`
}

// verlaufswerkzeug gibt das lesende Werkzeug zum Verlauf.
func verlaufswerkzeug(quelle Bestandsquelle) []Werkzeug {
	return []Werkzeug{{
		Name: "lesen.verlauf", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die Schritte des Verlaufs eines Sets — der jüngste zuerst. " +
			"Je Schritt: Zeitpunkt, Handlung (feld, karte, reihenfolge, schluesselwort), " +
			"Karte, Feld, Vorher und Nachher (mit »gabVorher«/»gabNachher«: eine leere " +
			"Zeile ist nicht dasselbe wie keine), ob er zurückgenommen ist, Urheber, " +
			"Herkunft, »umfang« (wie viele Karten) und »vorgang« — die Kennung, die " +
			"»schreiben.zuruecknehmen« nimmt, um GENAU diesen Schritt zurückzunehmen. " +
			"Eine Massenänderung ist EIN Schritt; die Zeile zeigt ihre jüngste Karte.",
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
		}, "setID"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			var wahl struct {
				SetID int64 `json:"setID"`
			}
			if err := lies(argumente, &wahl); err != nil {
				return "", err
			}

			if err := pruefeKennung("setID", wahl.SetID); err != nil {
				return "", err
			}

			schritte, err := quelle.Verlauf(ctx, wahl.SetID)
			if err != nil {
				return "", err
			}

			return alsJSON(map[string]any{"schritte": liste(schritte)})
		},
	}}
}
