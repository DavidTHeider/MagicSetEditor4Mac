package mcp

// **Ein Set vorab ansehen** — FMT-03.AK-01
// (#1692).
//
// Was `schreiben.set.einfahren` mit einer Datei täte — Name, Spiel, Stil,
// Karten, Befunde, ob sie schon im Bestand liegt und welche Pakete fehlen —,
// **ohne** zu schreiben. Über HTTP gilt der Pfadvertrag wie überall.

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Setvorschau ist, was das Werkzeug braucht.
type Setvorschau interface {
	SetVorschau(ctx context.Context, pfad string) (any, error)
}

// setvorschauwerkzeuge baut das eine.
func setvorschauwerkzeuge(quelle Setvorschau) []Werkzeug {
	if quelle == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.set.vorschau", Art: Lesend, Idempotent: true,
		Erklaerung: "Sagt, was das Einfahren einer Setdatei täte: Name, Spiel, Stil, Kartenzahl, " +
			"was sich nicht lesen ließ, ob dieselbe Datei schon im Bestand liegt (dann gäbe das " +
			"Einfahren dieses Set zurück) und welche Pakete fehlen. GESCHRIEBEN wird nichts; eine " +
			"Datei, die das Einfahren abwiese, weist die Vorschau mit derselben Meldung ab.",
		Schema: objekt(map[string]any{"pfad": text("Der ganze Pfad der .mse-set-Datei.")}, "pfad"),
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

			vorschau, err := quelle.SetVorschau(ctx, argumente.Pfad)
			if err != nil {
				return "", err
			}

			return alsJSON(vorschau)
		},
	}}
}
