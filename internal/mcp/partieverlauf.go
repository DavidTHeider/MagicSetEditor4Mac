package mcp

// **Der Verlauf einer Partie über MCP** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum ein Sprachmodell ihn bekommt
//
// **Weil es dieselbe Fähigkeit ist** (§16): was das Fenster zeigt, zeigen auch
// Kommandozeile, MCP und Dienst. Ein Verlauf, den nur das Fenster kennte, wäre
// eine Fähigkeit mit einer Fläche.
//
// **Und weil die Grenze dieselbe bleibt.** Er ist gefiltert, bevor er hier
// ankommt — `service` rechnet `domain.RueckschauFuer`, und diese Fläche sieht
// den ungefilterten Strom nie. Wer die Sicht falsch angibt, bekommt weniger,
// nicht mehr.

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Verlaufswerkstatt ist der Weg zum Verlauf.
type Verlaufswerkstatt interface {
	PartieVerlauf(ctx context.Context, kennung, fuer string) (
		[]Partieereignis, error)
}

// partieverlaufswerkzeug gibt das eine lesende Werkzeug — oder keines.
//
// **Nicht `verlaufswerkzeug`** — das gibt es schon für den Verlauf einer
// Karte (`verlaufsliste.go`), und zwei gleichnamige Werkzeugbauer wären zwei
// verschiedene Dinge unter einem Namen.
func partieverlaufswerkzeug(quelle Verlaufswerkstatt) []Werkzeug {
	if quelle == nil {
		return nil
	}

	return []Werkzeug{{
		Name: "lesen.partieverlauf", Art: Lesend, Idempotent: true,
		Erklaerung: "Zeigt den Verlauf einer Partie, wie EIN Sitz ihn sehen " +
			"darf. ÄNDERT NICHTS. Was ein anderer Sitz gezogen hat, steht " +
			"NICHT darin — auch nicht nachträglich.",
		Schema: objekt(map[string]any{
			"partieID": text("Die Kennung der Partie."),
			"fuer": text("Die Sitzkennung, für die gerechnet wird — " +
				"eine unbekannte sieht so viel wie ein Zuschauer."),
		}, "partieID", "fuer"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				PartieID string `json:"partieID"`
				Fuer     string `json:"fuer"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			// **Dieselbe Abweisung wie an den anderen Flächen** (§16): die
			// Kommandozeile verlangt beide Argumente, der Dienst verlangt
			// `fuer` in der Abfrage. Ohne diese Zeile sähe ein Modell mit
			// leerer Sitzkennung so viel wie ein Zuschauer und hielte das
			// für seine eigene Sicht.
			if argumente.PartieID == "" || argumente.Fuer == "" {
				return "", fmt.Errorf(
					"%w: es braucht die Partie und eine Sitzkennung",
					domain.ErrValidation)
			}

			verlauf, err := quelle.PartieVerlauf(
				ctx, argumente.PartieID, argumente.Fuer)
			if err != nil {
				return "", err
			}

			// **Eine leere Liste und kein `nil`** — sonst stünde im Ergebnis
			// `null`, und ein Modell läse daraus „kein Verlauf" statt „noch
			// nichts geschehen".
			if verlauf == nil {
				verlauf = []Partieereignis{}
			}

			return alsJSON(verlauf)
		},
	}}
}
