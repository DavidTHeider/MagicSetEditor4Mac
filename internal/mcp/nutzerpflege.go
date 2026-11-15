// **Nutzer und Profile ändern und löschen über MCP** — Werkstück 5 von
// Phase 4.12
// (#1351).
//
// **Eine eigene Datei**, damit `nutzer.go` unter der Zeilengrenze aus §2
// bleibt. Der Schnitt folgt der Frage: dort steht, wie etwas **entsteht**,
// hier, wie es sich **ändert** und verschwindet.

package mcp

import (
	"context"
	"encoding/json"
)

// pflegewerkzeuge gibt die vier schreibenden.
func pflegewerkzeuge(pflege Nutzerpflege) []Werkzeug {
	return []Werkzeug{
		{
			Name: "schreiben.nutzer.aendern", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: setzt die Angaben eines vorhandenen " +
				"Nutzers. Der Name ist eindeutig und trägt keine Leerzeichen am Rand; " +
				"alles andere ist freiwillig. Es entsteht kein neuer Nutzer. " +
				"Die »fassung« kommt aus »lesen.nutzer« und muss die zuletzt " +
				"gelesene sein — hat jemand anderes inzwischen geändert, wird der " +
				"Aufruf ABGEWIESEN, statt seine Änderung zu überschreiben. " +
				"Adresse, Rufnummer, Kennwort und zweite Faktoren ändert dieses " +
				"Werkzeug NICHT.",
			Schema: objekt(map[string]any{
				"nutzerID":     text("Die Kennung des Nutzers aus »lesen.nutzer«."),
				"name":         text("Der neue Anzeigename."),
				"bild":         text("Dateiname eines Bildes. Leer heißt: keines."),
				"anrede":       text("Freier Text — jede Liste ließe jemanden aus."),
				"namenszusatz": text("Etwa »von der Heide«."),
				"vorname":      text("Der Rufname."),
				"weiterenamen": text("Weitere Vornamen."),
				"nachname":     text("Der Familienname."),
				"slogan":       text("Eine Zeile unter dem Namen."),
				"kurzebio":     text("Ein Absatz."),
				"langebio":     text("Ein längerer Text."),
				"fassung":      text("Die zuletzt gelesene Fassung aus »lesen.nutzer«."),
			}, "nutzerID", "name", "fassung"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente Nutzerangaben

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				geaendert, err := pflege.NutzerAendern(ctx, argumente)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"nutzer": geaendert})
			},
		},
		{
			Name: "schreiben.profil.aendern", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: setzt die Angaben eines vorhandenen " +
				"Profils. Das Profil bleibt bei seinem Nutzer; es wechselt ihn " +
				"nicht. Es SETZT und verschmilzt nicht: was hier fehlt, steht " +
				"danach leer da. Der »subslogan« steht UNTER dem Slogan des " +
				"Menschen und ersetzt ihn nicht.",
			Schema: objekt(map[string]any{
				"profilID":          text("Die Kennung des Profils aus »lesen.nutzer«."),
				"name":              text("Der neue Name."),
				"subslogan":         text("Eine Zeile unter dem Profilnamen."),
				"kurzebeschreibung": text("Ein Absatz."),
				"langebeschreibung": text("Ein längerer Text."),
				"badge": text("Dateiname des Abzeichens im Bildverzeichnis. " +
					"Leer heißt: keines."),
				"technisch": map[string]any{"type": "boolean", "description": "Wahr, " +
					"wenn hinter diesem Profil kein Mensch steht, sondern ein " +
					"Automat. Es gehört trotzdem seinem Nutzer."},
				"ablauf": text("Zeitpunkt in RFC 3339, ab dem dieses Profil " +
					"abgelaufen ist. Leer heißt: läuft nicht ab. Nichts wird " +
					"dabei gelöscht."),
				"fassung": text("Die zuletzt gelesene Fassung aus »lesen.nutzer«."),
			}, "profilID", "name", "fassung"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente Profilangaben

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				geaendert, err := pflege.ProfilAendern(ctx, argumente)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"profil": geaendert})
			},
		},
		{
			Name: "schreiben.nutzer.loeschen", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: LÖSCHT einen Nutzer samt allen seinen Profilen — ENDGÜLTIG. " +
				"Der Name muss stimmen; er bestätigt, wer gemeint ist. Der LETZTE " +
				"Nutzer lässt sich nicht löschen. Was IN den Profilen liegt — Sets, " +
				"Decks, Sammlung — bleibt im Bestand, ist aber nicht mehr erreichbar.",
			Schema: objekt(map[string]any{
				"nutzerID": text("Die Kennung des Nutzers aus »lesen.nutzer«."),
				"name":     text("Sein Name — er bestätigt das Löschen."),
			}, "nutzerID", "name"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					NutzerID string `json:"nutzerID"`
					Name     string `json:"name"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if err := pflege.NutzerLoeschen(
					ctx, argumente.NutzerID, argumente.Name); err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"geloescht": argumente.NutzerID})
			},
		},
		{
			Name: "schreiben.profil.loeschen", Art: Schreibend,
			Erklaerung: "ÄNDERT den Bestand: LÖSCHT ein Profil — ENDGÜLTIG. Der Name muss stimmen. Das " +
				"LETZTE Profil eines Nutzers lässt sich nicht löschen. Was IN ihm " +
				"liegt, bleibt im Bestand, ist aber nicht mehr erreichbar.",
			Schema: objekt(map[string]any{
				"profilID": text("Die Kennung des Profils aus »lesen.nutzer«."),
				"name":     text("Sein Name — er bestätigt das Löschen."),
			}, "profilID", "name"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					ProfilID string `json:"profilID"`
					Name     string `json:"name"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if err := pflege.ProfilLoeschen(
					ctx, argumente.ProfilID, argumente.Name); err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"geloescht": argumente.ProfilID})
			},
		},
	}
}
