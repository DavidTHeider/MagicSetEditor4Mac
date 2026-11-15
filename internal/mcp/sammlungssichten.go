// **Die Fragen an eine geordnete Sammlung, über MCP** — Werkstück 5 von
// Phase 4.5 (#1292,
// Vermerk [0057](../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md)).

package mcp

import (
	"context"
	"encoding/json"
)

// Sammlungssichten ist, was MCP zum Beantworten der vier Fragen braucht.
type Sammlungssichten interface {
	Besitz(ctx context.Context, printing string) (any, error)
	Lagerorte(ctx context.Context) (any, error)
	Dubletten(ctx context.Context, behalten int) (any, error)
	Wuensche(ctx context.Context) (any, error)
}

// sammlungssichten gibt die vier Werkzeuge — alle lesend.
func sammlungssichten(sichten Sammlungssichten) []Werkzeug {
	if sichten == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.sammlung.besitz", Art: Lesend,
		Erklaerung: "Sagt, wie viele Stücke eines Drucks in der Sammlung sind — " +
			"und wie viele davon nur GEWÜNSCHT. Ein Wunsch ist kein Besitz: " +
			"»besitzt« ist falsch, wenn es die Karte nur auf der Wunschliste gibt.",
		Schema: objekt(map[string]any{
			"printing": text("Die Scryfall-Druckkennung."),
		}, "printing"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Printing string `json:"printing"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			stand, err := sichten.Besitz(ctx, argumente.Printing)
			if err != nil {
				return "", err
			}

			return alsJSON(stand)
		},
	}, {
		Name: "lesen.sammlung.lagerorte", Art: Lesend,
		Erklaerung: "Zählt die Lagerorte auf — Freitext wie »Kiste 4« oder " +
			"»Ordner Rot« — samt Zeilen und Stücken. Wünsche zählen nicht mit: " +
			"sie liegen nirgends.",
		Schema: objekt(map[string]any{}),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			orte, err := sichten.Lagerorte(ctx)
			if err != nil {
				return "", err
			}

			return alsJSON(orte)
		},
	}, {
		Name: "lesen.sammlung.dubletten", Art: Lesend,
		Erklaerung: "Sagt, wovon MEHR da ist als gewollt. »behalten« ist die " +
			"Zahl, die bleiben soll — meist 1, bei Ländern 4. Eine Dublette ist " +
			"kein Zustand, sondern ein Verhältnis, und die Zahl bringt der " +
			"Rufende mit.",
		Schema: objekt(map[string]any{
			"behalten": zahl("Wie viele Stücke bleiben sollen — 0 oder mehr."),
		}, "behalten"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Behalten int `json:"behalten"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			dubletten, err := sichten.Dubletten(ctx, argumente.Behalten)
			if err != nil {
				return "", err
			}

			return alsJSON(dubletten)
		},
	}, {
		Name: "lesen.sammlung.wuensche", Art: Lesend,
		Erklaerung: "Gibt die Wunschliste — die Zeilen, die KEIN Besitz sind. " +
			"Sie zählen nicht ins Portfolio und nicht in »lesen.sammlung.besitz«.",
		Schema: objekt(map[string]any{}),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			wuensche, err := sichten.Wuensche(ctx)
			if err != nil {
				return "", err
			}

			return alsJSON(wuensche)
		},
	}}
}
