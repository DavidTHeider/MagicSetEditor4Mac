// **Die Stiloptionen als Werkzeuge**
// (#1107).
//
// **Eine eigene Datei neben `stile.go`**, weil der Schnitt der Frage folgt:
// dort **welcher** Stil, hier **wie** er eingestellt ist.

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// stiloptionenLesen gibt die Optionen des gewählten Stils samt Werten.
func stiloptionenLesen(quelle Stilquelle) Werkzeug {
	return Werkzeug{
		Name: "lesen.stiloptionen", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die Optionen des gewählten Stils. »gesetzt« sagt, ob der Wert aus dem Set kommt oder aus der Vorgabe.",
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
		}, "setID"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			wahl, err := kennung(argumente)
			if err != nil {
				return "", err
			}

			optionen, lesefehler := quelle.Optionen(ctx, wahl)
			if lesefehler != nil {
				return "", lesefehler
			}

			return alsJSON(map[string]any{"optionen": liste(optionen)})
		},
	}
}

// stiloptionSetzen schreibt den Wert einer Stiloption.
func stiloptionSetzen(quelle Stilquelle) Werkzeug {
	return Werkzeug{
		Name: "schreiben.stiloption", Art: Schreibend,
		Erklaerung: "ÄNDERT eine Stiloption eines Sets. Der Name kommt aus »lesen.stiloptionen«.",
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
			"feld":  text("Der Name der Option aus »lesen.stiloptionen«."),
			"wert":  text("Der neue Wert. Zum Zurückstellen auf die Vorgabe »schreiben.stiloption.zuruecksetzen«."),
		}, "setID", "feld", "wert"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			wahl, err := optionswahl(argumente)
			if err != nil {
				return "", err
			}

			if fehler := quelle.OptionSetzen(
				ctx, wahl.SetID, wahl.Feld, wahl.Wert); fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{
				"gesetzt": wahl.Feld,
				"rueckweg": rueckweg(
					"schreiben.stiloption.zuruecksetzen mit »feld« = %q", wahl.Feld),
			})
		},
	}
}

// stiloptionZuruecksetzen stellt eine Stiloption auf ihre Vorgabe zurück.
//
// **Ein eigenes Werkzeug und kein leerer Wert an [stiloptionSetzen]**: leer ist
// bei einer Textoption ein gültiger Wert, und zwei Bedeutungen an einem Aufruf
// wären später nicht mehr auseinanderzuhalten.
func stiloptionZuruecksetzen(quelle Stilquelle) Werkzeug {
	return Werkzeug{
		Name: "schreiben.stiloption.zuruecksetzen", Art: Schreibend,
		Erklaerung: "ÄNDERT eine Stiloption zurück auf ihre Vorgabe. Nicht dasselbe wie ein leerer Wert.",
		Schema: objekt(map[string]any{
			"setID": zahl("Die Kennung des Sets aus »lesen.sets«."),
			"feld":  text("Der Name der Option aus »lesen.stiloptionen«."),
		}, "setID", "feld"),
		Ruf: func(ctx context.Context, argumente json.RawMessage) (string, error) {
			wahl, err := optionswahl(argumente)
			if err != nil {
				return "", err
			}

			if fehler := quelle.OptionZuruecksetzen(
				ctx, wahl.SetID, wahl.Feld); fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"zurueckgesetzt": wahl.Feld})
		},
	}
}

// optionswahl liest Setkennung, Feldnamen und Wert aus den Argumenten.
func optionswahl(argumente json.RawMessage) (struct {
	SetID int64  `json:"setID"`
	Feld  string `json:"feld"`
	Wert  string `json:"wert"`
}, error,
) {
	var wahl struct {
		SetID int64  `json:"setID"`
		Feld  string `json:"feld"`
		Wert  string `json:"wert"`
	}

	if err := json.Unmarshal(argumente, &wahl); err != nil {
		return wahl, fmt.Errorf("die Argumente sind kein JSON: %w", err)
	}

	return wahl, pruefeKennung("setID", wahl.SetID)
}
