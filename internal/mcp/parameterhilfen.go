// **Die Parameterhilfen der Schlüsselwörter, über MCP** — Werkstück D von
// Phase 4.11 (#1342).
//
// Einordnung nach §16: **erweiternd** — ein Werkzeug kommt hinzu, lesend.

package mcp

import (
	"context"
	"encoding/json"
)

// Parameterhilfen ist, was MCP für die Parameterhilfen braucht.
type Parameterhilfen interface {
	Parameterhilfen(ctx context.Context, setID, wortID int64) (any, error)
}

// parameterhilfen gibt das Werkzeug — lesend, es ändert nichts.
func parameterhilfen(hilfen Parameterhilfen) []Werkzeug {
	if hilfen == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.parameterhilfen", Art: Lesend,
		Erklaerung: "Nennt zu einem Schlüsselwort des Sets zweierlei: die " +
			"Parameterarten, die sich in sein MUSTER einfügen lassen " +
			"(»<atom-param>art</atom-param>«), und die Formen, in denen sein " +
			"ERINNERUNGSTEXT auf einen vorhandenen Parameter zeigen kann " +
			"(»{param1}« oder ein gerechneter Ausdruck). Beides kommt aus dem " +
			"Spielpaket. ÄNDERT NICHTS — geschrieben wird mit " +
			"»schreiben.schluesselwort.setzen«.",
		Schema: objekt(map[string]any{
			"setID":  zahl("Das Set, dem das Schlüsselwort gehört."),
			"wortID": zahl("Das Schlüsselwort — eines des Sets, kein festes."),
		}, "setID", "wortID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				SetID  int64 `json:"setID"`
				WortID int64 `json:"wortID"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if err := pruefeKennung("setID", argumente.SetID); err != nil {
				return "", err
			}

			if err := pruefeKennung("wortID", argumente.WortID); err != nil {
				return "", err
			}

			gefunden, fehler := hilfen.Parameterhilfen(
				ctx, argumente.SetID, argumente.WortID)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(gefunden)
		},
	}}
}
