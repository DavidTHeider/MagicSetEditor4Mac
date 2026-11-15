package mcp

// **Wie ein Handel dorthin kam** — TRADE-03.AK-03
// (#1717).

import (
	"context"
	"encoding/json"
)

// Handelsprotokollzeile ist ein Zustandswechsel.
type Handelsprotokollzeile struct {
	Revision int `json:"revision"`

	// Von ist leer, wenn die Revision neu ist.
	Von  string `json:"von"`
	Nach string `json:"nach"`
	Zeit string `json:"zeit"`
}

// handelsprotokollwerkzeug ist das lesende.
func handelsprotokollwerkzeug(auskunft Handelsauskunft) Werkzeug {
	return Werkzeug{
		Name: "lesen.handelsprotokoll", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt, wie ein Handelsvorgang dorthin kam, wo er steht: " +
			"jeden Wechsel seines Zustands mit Revision und Zeitpunkt, ältester " +
			"zuerst. LIEST NUR — das Protokoll schreibt der Bestand selbst, bei " +
			"jedem Schritt, und ändern lässt es sich auf keinem Weg.",
		Schema: objekt(map[string]any{
			"vorgang": text("Die Kennung des Handelsvorgangs."),
		}, "vorgang"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Vorgang string `json:"vorgang"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			zeilen, err := auskunft.Handelsprotokoll(ctx, argumente.Vorgang)
			if err != nil {
				return "", err
			}

			return alsJSON(map[string]any{"zeilen": zeilen})
		},
	}
}
