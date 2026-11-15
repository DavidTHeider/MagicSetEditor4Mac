// **Eine eigene Kopie eines Pakets** — FMT-02.AK-04
// (#1691).
//
// Einordnung nach §16: **erweiternd** — ein Werkzeug kommt hinzu.

package mcp

import (
	"context"
	"encoding/json"
)

// paketkopie gibt das Werkzeug, das eine eigene Kopie anlegt.
func paketkopie(pflege Paketpflege) Werkzeug {
	return Werkzeug{
		Name: "schreiben.paket.kopieren", Art: Schreibend,
		Erklaerung: "ÄNDERT den Rechner: legt eine EIGENE KOPIE eines Vorlagenpakets im " +
			"verwalteten Paketverzeichnis an — der Weg, ein mitgeliefertes oder " +
			"bezogenes Paket anzupassen, ohne das Original zu ändern. Ohne »neu« heißt " +
			"die Kopie wie das Original mit »-kopie« (dann »-kopie-2« …); ein Stil " +
			"bekommt sein Spiel vorangestellt. Ein vergebener Name wird abgewiesen und " +
			"nichts überschrieben.",
		Schema: objekt(map[string]any{
			"art":  text("Die Paketart aus »lesen.pakete«, etwa »mse-style«."),
			"name": text("Der Name des Originals aus »lesen.pakete«."),
			"neu":  text("Der Name der Kopie. Leer heißt: der nächste freie mit »-kopie«."),
		}, "art", "name"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Art  string `json:"art"`
				Name string `json:"name"`
				Neu  string `json:"neu"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			paket, err := pflege.PaketKopieren(ctx, argumente.Art, argumente.Name, argumente.Neu)
			if err != nil {
				return "", err
			}

			return alsJSON(map[string]any{"paket": paket})
		},
	}
}
