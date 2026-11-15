// **Der zweite Faktor über MCP** — Werkstück 7 von Phase 4.12
// (#1351).
//
// # Warum hier nur gelesen und gelöscht wird
//
// **Weil das Einrichten ein Geheimnis ausgibt.** `faktor_neu` gäbe die
// `otpauth://`-Adresse samt Geheimnis und zehn Wiederherstellungscodes
// zurück — und die stünden danach im Verlauf eines Sprachmodells: auf dem
// Rechner, beim Anbieter, in jedem Mitschnitt. Wer das Geheimnis hat, macht
// jeden Code; der zweite Faktor wäre keiner mehr.
//
// **Und `faktor_bestaetigen` wäre ein Orakel** — dieselbe Überlegung wie bei
// `konto_anmelden` in Werkstück 6.
//
// **Was ein Modell kann**: nachsehen, welche Faktoren jemand hat, und einen
// herausnehmen. Beides trägt kein Geheimnis.

package mcp

import (
	"context"
	"encoding/json"
)

// Faktorzeile ist ein zweiter Faktor, wie MCP ihn ausgibt.
type Faktorzeile struct {
	Art string `json:"art"`

	// Bezeichnung sagt, welches Gerät gemeint ist — leer erlaubt.
	Bezeichnung string `json:"bezeichnung"`
}

// faktorwerkzeuge gibt das lesende und das schreibende.
func faktorwerkzeuge(pflege Nutzerpflege) (lesend, schreibend Werkzeug) {
	lesend = Werkzeug{
		Name: "lesen.faktoren", Art: Lesend,
		Erklaerung: "Nennt die zweiten Faktoren eines Nutzers — »totp« für einen " +
			"Code aus einer Authenticator-App, »biometrie« für Touch ID oder " +
			"Face ID. Wer keinen hat, kommt ohne zweiten Schritt herein. " +
			"EINRICHTEN geht über diesen Server NICHT: dabei entstünde ein " +
			"Geheimnis, und das gehört nicht in einen Werkzeugaufruf.",
		Schema: objekt(map[string]any{
			"nutzerID": text("Die Kennung des Nutzers aus »lesen.nutzer«."),
		}, "nutzerID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				NutzerID string `json:"nutzerID"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			gelesen, err := pflege.Faktoren(ctx, argumente.NutzerID)
			if err != nil {
				return "", err
			}

			// **Nie `nil`** (§3): eine leere Liste ist eine Antwort.
			if gelesen == nil {
				gelesen = []Faktorzeile{}
			}

			return alsJSON(map[string]any{"faktoren": gelesen})
		},
	}

	schreibend = Werkzeug{
		Name: "schreiben.faktor.loeschen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: NIMMT einen zweiten Faktor von einem " +
			"Nutzer. Danach kommt er ohne zweiten Schritt herein — das ist " +
			"WENIGER Sicherheit, nicht mehr. Der Nutzer bleibt, seine Daten " +
			"bleiben. Ein TOTP-Faktor nimmt sein Geheimnis mit; wer ihn wieder " +
			"will, richtet ihn NEU ein und scannt ein neues Muster.",
		Schema: objekt(map[string]any{
			"nutzerID": text("Die Kennung des Nutzers aus »lesen.nutzer«."),
			"art":      text("»totp« oder »biometrie«, aus »lesen.faktoren«."),
		}, "nutzerID", "art"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				NutzerID string `json:"nutzerID"`
				Art      string `json:"art"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			if err := pflege.FaktorEntfernen(
				ctx, argumente.NutzerID, argumente.Art); err != nil {
				return "", err
			}

			return alsJSON(map[string]any{"entfernt": argumente.Art})
		},
	}

	return lesend, schreibend
}
