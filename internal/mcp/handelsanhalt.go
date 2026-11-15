package mcp

// **Was die Seiten eines Tauschs ungefähr wert sind** — TRADE-02b
// (#1951).

import (
	"context"
	"encoding/json"
)

// handelsanhaltwerkzeug ist das lesende.
func handelsanhaltwerkzeug(auskunft Handelsauskunft) Werkzeug {
	return Werkzeug{
		Name: "lesen.handelsanhalt", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt, was die Seiten eines Tauschs ungefähr wert sind: " +
			"je Position der Preis mal Menge, je Seite die Summe und die Zahl " +
			"der Positionen ohne Preis, je Währung die Differenz aus Sicht dieses " +
			"Profils (erhalten weniger gegeben) und ob jede Position einen Preis " +
			"hat. Beträge in der kleinsten Einheit (EUR: Cent), nicht umgerechnet. " +
			"LIEST NUR — die Preise sind ein Anhalt aus dem eigenen Preisbestand " +
			"und gehen weder in den Vertrag noch über die Leitung.",
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

			anhalt, err := auskunft.Handelsanhalt(ctx, argumente.Vorgang)
			if err != nil {
				return "", err
			}

			return alsJSON(anhalt)
		},
	}
}
