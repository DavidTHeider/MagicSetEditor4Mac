package mcp

// **Eine Datei ansehen** — UX-13
// (#1685).
//
// **Lesend und ohne Bilddaten** — dieselbe Auskunft wie am Dienst. Im Finder
// zeigen und öffnen gibt es hier nicht: ein Werkzeugserver öffnete auf dem
// Rechner, auf dem er läuft, und ein Sprachmodell, das Programme startet,
// wäre genau die Ausführung, die diese Ansicht ausschließt.

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Dateiansicht ist, was das Werkzeug dieser Datei braucht.
type Dateiansicht interface {
	DateiAnsehen(ctx context.Context, pfad string) (any, error)
}

// dateiwerkzeuge baut das eine.
func dateiwerkzeuge(quelle Dateiansicht) []Werkzeug {
	if quelle == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.datei.ansicht", Art: Lesend, Idempotent: true,
		Erklaerung: "Sagt, was in einer Datei liegt: Art, Größe, Zeit, bei Text den Anfang, " +
			"bei einem Paket oder Archiv die Einträge — mit einer Zahl der unsicheren " +
			"(Verweise, Namen mit ..). AUSGEPACKT, GEFOLGT oder AUSGEFÜHRT wird nichts; " +
			"Bilddaten gehen nicht mit.",
		Schema: objekt(map[string]any{"pfad": text("Der ganze Pfad der Datei.")}, "pfad"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Pfad string `json:"pfad"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if argumente.Pfad == "" {
				return "", fmt.Errorf("%w: es wurde kein Pfad genannt", domain.ErrValidation)
			}

			ansicht, err := quelle.DateiAnsehen(ctx, argumente.Pfad)
			if err != nil {
				return "", err
			}

			return alsJSON(ansicht)
		},
	}}
}
