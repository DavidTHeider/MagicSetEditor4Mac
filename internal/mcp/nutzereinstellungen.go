// **Die Einstellungen eines Menschen und seine Auskunft über MCP** —
// Werkstück 8 von Phase 4.12
// (#1351),
// Punkte 9, 10 und 11.
//
// # Warum die Auskunft als fertiges JSON hereinkommt
//
// **Weil ihre Gestalt anderswo entschieden ist.** `service.Personendaten`
// trägt die Marken, unter denen sie geschrieben wird; sie hier noch einmal zu
// erklären hieße, dieselbe Feldliste an einer vierten Stelle zu führen — und
// `internal/mcp` darf `service` gar nicht kennen (§2). Was hier ankommt, ist
// deshalb Text, und der Weg dorthin steht in `cmd/mse-mcp`.

package mcp

import (
	"context"
	"encoding/json"
)

// Nutzereinstellungszeile ist eine Einstellung samt ihrer möglichen Werte.
//
// **Die Werte kommen mit**, damit ein Sprachmodell nicht raten muss: eine
// Antwort ohne sie führte zu erfundenen Werten, die der Bestand abweist.
type Nutzereinstellungszeile struct {
	Schluessel string   `json:"schluessel"`
	Wert       string   `json:"wert"`
	Werte      []string `json:"werte"`
}

// nutzereinstellungswerkzeuge gibt das lesende und die beiden schreibenden.
func nutzereinstellungswerkzeuge(pflege Nutzerpflege) (lesend, schreibend []Werkzeug) {
	lesend = []Werkzeug{
		{
			Name: "lesen.nutzereinstellungen", Art: Lesend,
			Erklaerung: "Nennt die Einstellungen, die einem MENSCHEN gehören — er " +
				"nimmt sie in jedes seiner Profile mit. Das sind andere als " +
				"»lesen.einstellungen«: die gelten für das Profil, in dem " +
				"gearbeitet wird. Ungesetztes kommt mit seinem Vorgabewert.",
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

				zeilen, fehler := pflege.Nutzereinstellungen(ctx, argumente.NutzerID)
				if fehler != nil {
					return "", fehler
				}

				return alsJSON(map[string]any{"einstellungen": zeilen})
			},
		},
		{
			Name: "lesen.personendaten", Art: Lesend,
			Erklaerung: "Gibt heraus, was dieses Programm über einen Menschen " +
				"gespeichert hat — seine Angaben, seine Anmeldewege, die ARTEN " +
				"seiner zweiten Faktoren, seine Einstellungen und die Namen seiner " +
				"Profile, samt Stichtag. OHNE Kennwort und OHNE Geheimnis: beides " +
				"ist kein Datum über einen Menschen, sondern ein Mittel, ihn zu " +
				"sein. Sets, Decks und Sammlung stehen NICHT darin — sie gehören " +
				"einem Profil und gehen über »schreiben.set.sichern«.",
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

				return pflege.Personendaten(ctx, argumente.NutzerID)
			},
		},
	}

	schreibend = []Werkzeug{{
		Name: "schreiben.nutzereinstellung", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: setzt EINE Einstellung eines Menschen. " +
			"Erlaubt sind nur die Werte, die »lesen.nutzereinstellungen« zu " +
			"diesem Schlüssel nennt. Zurück kommen ALLE Einstellungen — der " +
			"Stand nach der Änderung.",
		Schema: objekt(map[string]any{
			"nutzerID":   text("Die Kennung des Nutzers aus »lesen.nutzer«."),
			"schluessel": text("Der Schlüssel aus »lesen.nutzereinstellungen«."),
			"wert":       text("Einer der dort genannten Werte."),
		}, "nutzerID", "schluessel", "wert"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				NutzerID   string `json:"nutzerID"`
				Schluessel string `json:"schluessel"`
				Wert       string `json:"wert"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			zeilen, fehler := pflege.NutzereinstellungSetzen(
				ctx, argumente.NutzerID, argumente.Schluessel, argumente.Wert)
			if fehler != nil {
				return "", fehler
			}

			return alsJSON(map[string]any{"einstellungen": zeilen})
		},
	}}

	return lesend, schreibend
}
