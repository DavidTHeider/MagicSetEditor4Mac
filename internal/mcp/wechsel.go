// **Wer arbeitet und worin** — die Wechselwerkzeuge des MCP-Servers.
//
// # Warum sie neben `nutzer.go` stehen und nicht darin
//
// **Weil es eine andere Frage ist** (§2): dort steht, **welche** Nutzer und
// Profile es gibt; hier, **wer von ihnen gerade arbeitet und worin**. Die
// Zeilengrenze war der Anlass — Werkstück 17 von Phase 4.12 hat die Datei über
// 400 Zeilen gebracht —, der Schnitt folgt der Frage.
//
// **Und die beiden Wechsel gehören zusammen**, weil ihr Unterschied der Inhalt
// von Werkstück 17 ist: der eine sagt „ich sehe da hinein" und prüft **meine**
// Freigaben, der andere „ich bin jetzt jemand anderes" und verlangt **dessen**
// zweiten Faktor
// (#1400).

package mcp

import (
	"context"
	"encoding/json"
)

// wechselwerkzeuge baut die vier — zwei lesende, zwei schreibende.
func wechselwerkzeuge(pflege Nutzerpflege) (lesend, schreibend []Werkzeug) {
	lesend = []Werkzeug{
		{
			Name: "lesen.profil", Art: Lesend,
			Erklaerung: "Nennt das Profil, in dem gerade gearbeitet wird. Jede " +
				"andere Antwort dieses Servers — Sets, Decks, Sammlung, " +
				"Einstellungen — gilt für DIESES Profil und kein anderes.",
			Schema: objekt(map[string]any{}),
			Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
				gezeigt, err := pflege.GezeigtesProfil(ctx)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"profil": gezeigt})
			},
		},
		{
			Name: "lesen.werarbeitet", Art: Lesend, Idempotent: true,
			Erklaerung: "Nennt die Kennung des Menschen, der gerade arbeitet. Das " +
				"ist NICHT dasselbe wie der Eigner aus »lesen.profil«: wer in ein " +
				"fremdes Profil gewechselt ist, bleibt, wer er ist — und seine " +
				"eigenen Freigaben entscheiden, wohin er weiter darf.",
			Schema: objekt(map[string]any{}),
			Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
				nutzerID, err := pflege.WerArbeitet(ctx)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"nutzerID": nutzerID})
			},
		},
		{
			Name: "lesen.kontext", Art: Lesend, Idempotent: true,
			Erklaerung: "Nennt in EINEM Zug, wer gerade in welchem Profil als was " +
				"arbeitet: Nutzer, Profil, Eigner des Profils, wirksame Rolle, ob " +
				"nur gelesen werden darf, die Fläche des Rufs und eine Revision, " +
				"die mit jedem Wechsel und jeder Freigabe steigt. Zwei Antworten " +
				"mit derselben Revision beschreiben denselben Stand.",
			Schema: objekt(map[string]any{}),
			Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
				kontext, err := pflege.Arbeitskontext(ctx)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"kontext": kontext})
			},
		},
	}

	schreibend = []Werkzeug{
		{
			Name: "schreiben.profil.wechseln", Art: Schreibend, Zeigerwechsel: true,
			Erklaerung: "ÄNDERT, in welchem Profil gearbeitet wird — an ALLEN " +
				"Flächen, auch im Fenster. Jede Antwort danach gilt für das neue " +
				"Profil. Es wird nichts gelöscht: was im alten Profil liegt, ist " +
				"nach dem Zurückwechseln wieder da.",
			Schema: objekt(map[string]any{
				"profilID": text("Die Kennung des Profils aus »lesen.nutzer«."),
				"nachweis": text("Der zweite Faktor, wenn der Nutzer einen verlangt: " +
					"der Code aus seiner Authenticator-App oder ein " +
					"Wiederherstellungscode. Leer lassen, wenn er keinen hat."),
			}, "profilID"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					ProfilID string `json:"profilID"`
					Nachweis string `json:"nachweis"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				gewechselt, err := pflege.ProfilWechseln(
					ctx, argumente.ProfilID, argumente.Nachweis)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"profil": gewechselt})
			},
		},
		{
			Name: "schreiben.nutzer.wechseln", Art: Schreibend, Zeigerwechsel: true,
			Erklaerung: "ÄNDERT, WER arbeitet — an ALLEN Flächen. Das ist etwas " +
				"anderes als »schreiben.profil.wechseln«: jenes sagt »ich sehe da " +
				"hinein« und prüft MEINE Freigaben, dieses sagt »ich bin jetzt " +
				"jemand anderes« und verlangt DESSEN zweiten Faktor. Der Zeiger " +
				"zieht mit: danach wird im ersten Profil des neuen Menschen " +
				"gearbeitet. Es wird nichts gelöscht.",
			Schema: objekt(map[string]any{
				"nutzerID": text("Die Kennung des Nutzers aus »lesen.nutzer«."),
				"nachweis": text("Der zweite Faktor DIESES Nutzers, wenn er einen " +
					"verlangt — nicht der eigene. Leer lassen, wenn er keinen hat."),
			}, "nutzerID"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					NutzerID string `json:"nutzerID"`
					Nachweis string `json:"nachweis"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				gewechselt, err := pflege.NutzerWechseln(
					ctx, argumente.NutzerID, argumente.Nachweis)
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"profil": gewechselt})
			},
		},
	}

	return lesend, schreibend
}
