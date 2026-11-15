package mcp

// **Wer einen Handel unterschrieben hat** — gelesen, ohne einen Schritt zu
// gehen (Hüllenprobe `nat04handel`,
// #1672).

import (
	"context"
	"encoding/json"
)

// handelsbelegwerkzeug ist das lesende.
func handelsbelegwerkzeug(auskunft Handelsauskunft) Werkzeug {
	return Werkzeug{
		Name: "lesen.handelsbeleg", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt Stand und Beleg eines Handelsvorgangs: seinen " +
			"Zustand und, sobald verbucht ist, Abdruck, Unterschrift, ob der " +
			"Beleg beidseitig ist und wessen geprüfter Gegenbeleg hier liegt. " +
			"LIEST NUR — verbucht und bestätigt wird über diese Fläche nicht.",
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

			beleg, err := auskunft.Handelsbeleg(ctx, argumente.Vorgang)
			if err != nil {
				return "", err
			}

			return alsJSON(beleg)
		},
	}
}
