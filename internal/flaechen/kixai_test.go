// **xAI über jede Fläche ohne Fenster** — Werkstück A10 von Phase 5.3
// (#1449).
//
// # Was hier gemessen wird, und warum es nicht dasselbe ist wie ein Lauf mehr
//
// **Dass der Anbieter an jeder Fläche *bekannt* ist.** Die beiden Ausgänge
// unterscheiden sich, und genau darin liegt die Aussage:
//
//	unbekannt         → nichtGefunden, „den Anbieter »xai« gibt es nicht"
//	bekannt und aus   → validierung,   „»xai« ist nicht eingeschaltet"
//
// **Der erste wäre der Fehler, den dieses Werkstück verhindern soll.** Es gab
// bis A10 **zwei** Composition Roots, die jeden Cloudanbieter einzeln
// aufzählten — einen für das Fenster, einen für alles ohne Fenster. Wer nur
// den ersten pflegte, bekam einen Anbieter, der am Fenster dasteht und an
// MCP, REST und der Kommandozeile nicht existiert; das ist der Fall, den §16
// mit „dieselbe Fähigkeit, überall dieselbe Antwort" meint.
//
// **Was er nicht deckt**, steht hier und nicht in einer Fußnote: einen Lauf
// gegen den **echten** Dienst. Er brauchte ein Konto mit Guthaben, und jeder
// Ruf kostete Geld — es gibt ihn deshalb nirgends, auch nicht hinter einem
// Bautag ([0101](../../docs/entscheidungen/0101-was-xai-belegt.md)).
package flaechen_test

import (
	"net/http"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ein Vorschlag von xAI scheitert überall gleich — und an derselben Stelle.**
func TestEinVorschlagVonXAIUeberJedeFlaeche(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "ein Vorschlag von xAI ohne Schalter", abweisung{
		// **Validierung und nicht nichtGefunden** — siehe den Kopf.
		klasse: klasseValidierung,
		kern:   "eingeschaltet",
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).KIVorschlagen(t.Context(), "1", app.KIVorschlagsbezug{
				KarteID: 1, Anbieter: "xai", Alias: "grok-4", Auftrag: "Schreibe den Regeltext.",
			}).Fehler
		},
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung,
				"ki-vorschlag", "1", "xai", "grok-4",
				"Schreibe den Regeltext.")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.ki.vorschlag",
				map[string]any{
					"karteId": 1, "anbieter": "xai", "alias": "grok-4",
					"auftrag": "Schreibe den Regeltext.",
				})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/karten/1/ki-vorschlag",
				`{"anbieter":"xai","alias":"grok-4",`+
					`"auftrag":"Schreibe den Regeltext."}`)
		},
	})
}

// **Und eine Messung seiner Modelle ebenso.**
//
// **Zwei Wege und nicht einer**: der Vorschlag geht über
// [app.KIVorschlagen], die Messung über [app.KIModellPruefen], und beide
// suchen ihre Quelle selbst. Ein Anbieter kann an dem einen bekannt sein und
// am anderen nicht.
func TestEineMessungVonXAIUeberJedeFlaeche(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "eine Messung von xAI ohne Schalter", abweisung{
		klasse: klasseValidierung,
		kern:   "eingeschaltet",
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).KIModellPruefen(t.Context(), "xai", "grok-4", "").Fehler
		},
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung,
				"ki-modell-pruefen", "xai", "grok-4")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.ki.modell.pruefen",
				map[string]any{"anbieter": "xai", "alias": "grok-4"})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/ki/modelle/pruefen",
				`{"anbieter":"xai","alias":"grok-4"}`)
		},
	})
}
