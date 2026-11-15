// **Karten aus einem Skript, über MCP** — Werkstück B von Phase 4.11
// (#1337).

package mcp

import (
	"context"
	"encoding/json"
)

// Kartenskripte ist, was MCP für die Kartenskripte eines Sets braucht.
type Kartenskripte interface {
	Angebot(ctx context.Context, setID int64) (any, error)
	Ausfuehren(ctx context.Context, setID int64, name string) (any, error)
}

// kartenskripte gibt die beiden Werkzeuge — eines lesend, eines schreibend.
func kartenskripte(skripte Kartenskripte) (lesend, schreibend []Werkzeug) {
	if skripte == nil {
		return nil, nil
	}

	return []Werkzeug{{
		Name: "lesen.kartenskripte", Art: Lesend,
		Erklaerung: "Nennt die Kartenskripte des Spielpakets — fertige Züge, " +
			"die auf einen Ruf mehrere Karten anlegen: Grundlande, ein " +
			"Zyklus je Farbe, ein Gerüst für einen Entwurf. ÄNDERT NICHTS. " +
			"Welche es gibt, entscheidet das Spielpaket; ein Set ohne " +
			"Spielpaket hat keine.",
		Schema: objekt(map[string]any{
			"setID": zahl("Das Set, dessen Spielpaket gefragt ist."),
		}, "setID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			setID, err := kennung(roh)
			if err != nil {
				return "", err
			}

			angebot, fehler := skripte.Angebot(ctx, setID)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"Skripte": angebot})
		},
	}}, []Werkzeug{{
		Name: "schreiben.kartenskript", Art: Schreibend,
		Erklaerung: "Führt ein Kartenskript aus und hängt seine Karten ans " +
			"Set — in EINEM Zug: ein Zurücknehmen holt alle zusammen zurück. " +
			"Der Name ist der aus »lesen.kartenskripte«.",
		Schema: objekt(map[string]any{
			"setID": zahl("Das Set, an das die Karten kommen."),
			"name":  text("Der Name des Skripts, wie »lesen.kartenskripte« ihn nennt."),
		}, "setID", "name"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			// **Nicht [kennung]**: die deutet **alle** Argumente als
			// Zahlen, und `name` ist eine Zeichenkette — der erste Anlauf
			// gab „die Argumente sind kein JSON: invalid number literal".
			// Dieselbe Gestalt wie bei `schreiben.set.sichern`.
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

			lauf, fehler := skripte.Ausfuehren(ctx, argumente.SetID, argumente.Name)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(lauf)
		},
	}}
}
