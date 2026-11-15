// **Das Profilbild über MCP** — Werkstück 8 von Phase 4.12
// (#1351).
//
// # Warum es diese beiden gibt und das Anzeigen nicht
//
// **Setzen und Entfernen ändern den Bestand** — das ist eine Fähigkeit, und
// sie steht an allen vier Flächen (§16). **Das Anzeigen ist keine**: ein
// Daten-URI im Mitschnitt eines Sprachmodells ist eine Zeichenkette, die
// niemand ansieht, und dieselbe Begründung trägt schon `Kartenbilder`.
//
// # Und warum ein Pfad hier vertretbar ist
//
// **Weil er nichts preisgibt, was das Modell nicht schon nennt.** Es gibt
// den Pfad an; die Anwendung liest die Datei und schreibt sie um. Das ist
// dieselbe Form wie `schreiben.karte.bild` — und der Unterschied zu einem
// Kennwort ist, dass ein Pfad keine Berechtigung ist.

package mcp

import (
	"context"
	"encoding/json"
)

// nutzerbildwerkzeuge gibt die vier schreibenden.
func nutzerbildwerkzeuge(pflege Nutzerpflege) []Werkzeug {
	return append([]Werkzeug{
		{
			Name: "schreiben.nutzer.bild", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: liest eine Bilddatei von DIESEM Rechner " +
				"und legt sie als Profilbild eines Nutzers ab. Das Bild wird als PNG " +
				"gespeichert und auf höchstens 256 Punkte Kantenlänge gebracht — das " +
				"Seitenverhältnis bleibt. Ein vorhandenes Bild wird ERSETZT; ein " +
				"Mensch hat eines. Die »fassung« kommt aus »lesen.nutzer« und muss " +
				"die zuletzt gelesene sein.",
			Schema: objekt(map[string]any{
				"nutzerID": text("Die Kennung des Nutzers aus »lesen.nutzer«."),
				"pfad":     text("Der Pfad zur Bilddatei (.png, .jpg, .gif) auf diesem Rechner."),
				"fassung":  text("Die zuletzt gelesene Fassung aus »lesen.nutzer«."),
			}, "nutzerID", "pfad", "fassung"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					NutzerID string `json:"nutzerID"`
					Pfad     string `json:"pfad"`
					Fassung  string `json:"fassung"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				geaendert, err := pflege.NutzerbildSetzen(
					ctx, argumente.NutzerID, argumente.Fassung, argumente.Pfad)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"nutzer": geaendert})
			},
		},
		{
			Name: "schreiben.nutzer.bild.loeschen", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: nimmt das Profilbild eines Nutzers weg " +
				"und LÖSCHT die Datei. Wer keines hat, bleibt ohne — das ist kein " +
				"Fehler. Die »fassung« kommt aus »lesen.nutzer«.",
			Schema: objekt(map[string]any{
				"nutzerID": text("Die Kennung des Nutzers aus »lesen.nutzer«."),
				"fassung":  text("Die zuletzt gelesene Fassung aus »lesen.nutzer«."),
			}, "nutzerID", "fassung"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					NutzerID string `json:"nutzerID"`
					Fassung  string `json:"fassung"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				geaendert, err := pflege.NutzerbildEntfernen(
					ctx, argumente.NutzerID, argumente.Fassung)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"nutzer": geaendert})
			},
		},
	}, profilbadgewerkzeuge(pflege)...)
}

// profilbadgewerkzeuge gibt die beiden am Abzeichen — Werkstück 9 (#1351).
func profilbadgewerkzeuge(pflege Nutzerpflege) []Werkzeug {
	return []Werkzeug{
		{
			Name: "schreiben.profil.badge", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: liest eine Bilddatei von DIESEM " +
				"Rechner und legt sie als ABZEICHEN eines Profils ab. Ein " +
				"Abzeichen gehört einem PROFIL und liegt rechts unten ÜBER dem " +
				"Profilbild des Menschen — es zeigt die Rolle, in der er " +
				"arbeitet. Gespeichert wird als PNG mit höchstens 256 Punkten " +
				"Kantenlänge. Ein vorhandenes wird ERSETZT.",
			Schema: objekt(map[string]any{
				"profilID": text("Die Kennung des Profils aus »lesen.nutzer«."),
				"pfad":     text("Der Pfad zur Bilddatei (.png, .jpg, .gif) auf diesem Rechner."),
				"fassung":  text("Die zuletzt gelesene Fassung des PROFILS aus »lesen.nutzer«."),
			}, "profilID", "pfad", "fassung"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					ProfilID string `json:"profilID"`
					Pfad     string `json:"pfad"`
					Fassung  string `json:"fassung"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				geaendert, err := pflege.ProfilbadgeSetzen(
					ctx, argumente.ProfilID, argumente.Fassung, argumente.Pfad)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"profil": geaendert})
			},
		},
		{
			Name: "schreiben.profil.badge.loeschen", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: nimmt das Abzeichen eines Profils weg " +
				"und LÖSCHT die Datei. Wer keines hat, bleibt ohne — das ist kein " +
				"Fehler.",
			Schema: objekt(map[string]any{
				"profilID": text("Die Kennung des Profils aus »lesen.nutzer«."),
				"fassung":  text("Die zuletzt gelesene Fassung des PROFILS."),
			}, "profilID", "fassung"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					ProfilID string `json:"profilID"`
					Fassung  string `json:"fassung"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				geaendert, err := pflege.ProfilbadgeEntfernen(
					ctx, argumente.ProfilID, argumente.Fassung)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"profil": geaendert})
			},
		},
	}
}
