// **Das anonyme Konto über MCP** — Werkstück 6 von Phase 4.12
// (#1351).
//
// # Warum hier nur eines von drei Werkzeugen steht
//
// **Weil die beiden anderen ein Kennwort im Aufruf trügen.** Ein Aufruf über
// MCP steht im Verlauf eines Sprachmodells — auf dem Rechner, beim Anbieter,
// in jedem Mitschnitt. Ein Kennwort, das dort einmal steht, ist kein
// Geheimnis mehr, und keine spätere Löschung holt es zurück.
//
// **Und `konto_anmelden` wäre zusätzlich ein Orakel**: es beantwortete die
// Frage „stimmt dieses Kennwort?" beliebig oft und ohne Zeugen.
//
// **Was das Modell stattdessen kann**: ein Konto ohne jede Angabe anlegen. Es
// hat kein Kennwort, es verrät nichts, und es ist genau die Kontoart, für die
// man niemanden fragen muss.

package mcp

import (
	"context"
	"encoding/json"
)

// kontowerkzeug gibt das eine schreibende Werkzeug der Kontoarten.
func kontowerkzeug(pflege Nutzerpflege) Werkzeug {
	return Werkzeug{
		Name: "schreiben.konto.anonym", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: legt einen NEUEN Nutzer OHNE jede Angabe " +
			"an — kein Name, keine Adresse, keine Rufnummer, kein Kennwort. Er " +
			"bekommt eine Zufallskennung als Namen und ein erstes Profil. " +
			"Ein Konto mit Adresse oder Rufnummer legt dieser Server NICHT an: " +
			"dafür bräuchte es ein Kennwort, und ein Kennwort gehört nicht in " +
			"einen Werkzeugaufruf.",
		Schema: objekt(map[string]any{}),
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			angelegt, err := pflege.KontoAnonym(ctx)
			if err != nil {
				return "", err
			}

			return alsJSON(map[string]any{"nutzer": angelegt})
		},
	}
}
