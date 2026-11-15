// **Profilvorgaben über MCP** — Werkstück 15 von Phase 4.12
// (#1351),
// Punkte 4 bis 7.
//
// **Zwei Werkzeuge, beide erweiternd** (§16).
//
// # Warum das lesende das Angebot mitgibt
//
// **Weil ein Modell sonst rät.** Eine Kennung wie `7` sagt nicht, welches Deck
// gemeint ist; ohne die Liste dessen, was es gibt, schriebe es eine Zahl
// hinein und bekäme ein Nein. Dieselbe Liste prüft danach — sie steht also
// ohnehin bereit.

package mcp

import (
	"context"
	"encoding/json"
)

// Vorgabengesetzt ist die Antwort auf ein Setzen — der ganze Stand und der
// Grund, aus dem er nach geglücktem Schreiben fehlen kann (#1970). Dieselbe
// Gestalt wie bisher und ein Feld dazu: erweiternd (0033).
type Vorgabengesetzt struct {
	Vorgaben       []Vorgabenzeile `json:"vorgaben"`
	Nachladefehler *Nachladefehler `json:"nachladefehler,omitempty"`
}

// Vorgabenpflege ist die Naht zu den Profilvorgaben.
type Vorgabenpflege interface {
	Vorgaben(ctx context.Context, profilID string) ([]Vorgabenzeile, error)
	VorgabeSetzen(ctx context.Context, profilID, art string, werte []string) (
		Vorgabengesetzt, error)
}

// Vorgabenzeile ist eine Art samt ihrem Stand.
type Vorgabenzeile struct {
	Art        string `json:"art"`
	Mehrwertig bool   `json:"mehrwertig"`

	Werte    []string         `json:"werte"`
	Moeglich []Vorgabeangebot `json:"moeglich"`

	// Verwaist sind gewählte Werte, die es nicht mehr gibt.
	Verwaist []string `json:"verwaist"`
}

// Vorgabeangebot ist ein Wert, den ein Profil wählen kann.
type Vorgabeangebot struct {
	Wert         string `json:"wert"`
	Beschriftung string `json:"beschriftung"`
}

// vorgabenwerkzeuge gibt das lesende und das schreibende.
func vorgabenwerkzeuge(pflege Vorgabenpflege) (lesend, schreibend []Werkzeug) {
	return []Werkzeug{vorgabenlesen(pflege)},
		[]Werkzeug{vorgabensetzen(pflege)}
}

func vorgabenlesen(pflege Vorgabenpflege) Werkzeug {
	return Werkzeug{
		Name: "lesen.profilvorgaben", Art: Lesend, Idempotent: true,
		Erklaerung: "Nennt die Vorgaben eines Profils, die auf ZEILEN im " +
			"Bestand zeigen — Vorgabeformat, Vorgabespiel, Vorgabevorlage, " +
			"Präferenzdecks und zugeordnete Lagerorte. Zu jeder Art steht " +
			"unter »moeglich«, was zur Wahl steht; ohne diese Liste ist eine " +
			"Kennung nicht zu erraten. »verwaist« nennt gewählte Werte, die " +
			"es nicht mehr gibt — sie bleiben stehen und wirken nicht. " +
			"NICHT dasselbe wie »lesen.einstellungen«: jene halten " +
			"aufzählbare Werte, diese zeigen auf den Bestand.",
		Schema: objekt(map[string]any{
			"profilID": text("Die Kennung des Profils aus »lesen.nutzer«."),
		}, "profilID"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				ProfilID string `json:"profilID"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			zeilen, err := pflege.Vorgaben(ctx, argumente.ProfilID)
			if err != nil {
				return "", err
			}

			return alsJSON(map[string]any{"vorgaben": zeilen})
		},
	}
}

func vorgabensetzen(pflege Vorgabenpflege) Werkzeug {
	return Werkzeug{
		Name: "schreiben.profilvorgabe", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: setzt die Werte EINER Art. Es SETZT " +
			"und ergänzt nicht — was hier fehlt, steht danach nicht mehr da; " +
			"ein leeres »werte« räumt die Art ab. Die Arten »format«, " +
			"»spiel« und »vorlage« halten HÖCHSTENS EINEN Wert, »deck« und " +
			"»lagerort« mehrere. Jeder Wert muss in »moeglich« aus " +
			"»lesen.profilvorgaben« stehen; alles andere wird ABGEWIESEN. " +
			"Bei »deck« ist die REIHENFOLGE die Präferenz.",
		Schema: objekt(map[string]any{
			"profilID": text("Die Kennung des Profils aus »lesen.nutzer«."),
			"art": text("»format«, »spiel«, »vorlage«, »deck« oder " +
				"»lagerort«."),
			"werte": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Die gewählten Werte aus »moeglich«. Leer räumt ab.",
			},
		}, "profilID", "art", "werte"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				ProfilID string   `json:"profilID"`
				Art      string   `json:"art"`
				Werte    []string `json:"werte"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			gesetzt, err := pflege.VorgabeSetzen(
				ctx, argumente.ProfilID, argumente.Art, argumente.Werte)
			if err != nil {
				return "", err
			}

			return alsJSON(gesetzt)
		},
	}
}
