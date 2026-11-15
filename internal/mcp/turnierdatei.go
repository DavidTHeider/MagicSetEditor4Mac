// **Ein Turnier als Datei und als Bericht** — DOM-04.AK-03
// (#1699),
// Vermerk 0140.
//
// Einordnung nach §16: **erweiternd** — drei Werkzeuge kommen hinzu, zwei
// lesend und eines schreibend. **Als Text und nicht als Datei**: ein
// MCP-Server steht an keiner Platte, die das Modell sieht.

package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// lesenAusgabe gibt ein Turnier als Datei im Text.
func lesenAusgabe(naht Turniere) Werkzeug {
	return Werkzeug{
		Name: "lesen.turnier.ausgabe", Art: Lesend, Idempotent: true,
		Erklaerung: "Gibt ein Turnier samt Verlauf als Datei im Text (JSON, Fassung 1): " +
			"Kopf, Teilnehmende mit eingereichter Liste, Runden, Paarungen, Ergebnisse, " +
			"Abschluss. Das Format steht mit seinem NAMEN da; KEINE Kennung dieses " +
			"Bestands — die Paarungen zählen die Teilnehmenden ab 1. ÄNDERT NICHTS.",
		Schema: turnierkennung(),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			turnierID, err := liesTurnierkennung(roh)
			if err != nil {
				return "", err
			}

			angaben, fehler := naht.TurnierAusgabe(ctx, turnierID)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(angaben)
		},
	}
}

// lesenBericht gibt den Abschlussbericht.
func lesenBericht(naht Turniere) Werkzeug {
	return Werkzeug{
		Name: "lesen.turnier.bericht", Art: Lesend, Idempotent: true,
		Erklaerung: "Gibt den Bericht über ein Turnier als Markdown im Feld »inhalt«: " +
			"Kopf mit Stand, Rangliste mit den Tiebreakern der MTR und jede Runde mit " +
			"ihren Ergebnissen. Bei einem laufenden Turnier ist es ein Zwischenstand. " +
			"ÄNDERT NICHTS.",
		Schema: turnierkennung(),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			turnierID, err := liesTurnierkennung(roh)
			if err != nil {
				return "", err
			}

			angaben, fehler := naht.TurnierBericht(ctx, turnierID)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(angaben)
		},
	}
}

// schreibenEinlesen legt ein Turnier aus dem Text einer Ausgabe an.
func schreibenEinlesen(naht Turniere) Werkzeug {
	return Werkzeug{
		Name: "schreiben.turnier.einlesen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: legt ein NEUES Turnier samt Verlauf aus dem Text " +
			"von »lesen.turnier.ausgabe« an — in einem Zug, oder gar nicht. Ein Name, " +
			"den es im Spiel schon gibt, ist ein Konflikt. Ein Format, das es hier nicht " +
			"gibt, steht als Befund in der Antwort, und das Turnier kommt ohne.",
		Schema: objekt(map[string]any{
			"inhalt": text("Der Text aus »lesen.turnier.ausgabe«."),
		}, "inhalt"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				Inhalt string `json:"inhalt"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if argumente.Inhalt == "" {
				return "", fmt.Errorf("%w: es braucht den Text des Turniers", domain.ErrValidation)
			}

			angaben, fehler := naht.TurnierEinlesen(ctx, argumente.Inhalt)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(angaben)
		},
	}
}
