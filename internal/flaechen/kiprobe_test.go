// **Eine Modellmessung über jede Fläche ohne Fenster** — Werkstück A9 von
// Phase 5.3
// (#1448).
//
// # Was hier gemessen wird
//
// **Derselbe fachliche Fehler an allen drei Flächen** (§16) — und zwar der
// einzige, der sich ohne einen echten Anbieter messen lässt: die Vorgabe von
// `dienst.openai` ist `aus`, und „aus" heißt, dass kein einziger Aufruf
// hinausgeht.
//
// **Was er nicht deckt**, steht hier und nicht in einer Fußnote: eine Messung
// gegen einen **echten** Anbieter. Sie brauchte ein Konto mit Guthaben, und
// jeder Ruf kostete Geld — es gibt sie deshalb nirgends, auch nicht hinter
// einem Bautag ([0100](../../docs/entscheidungen/0100-was-openai-nicht-sagt.md)).
package flaechen_test

import (
	"net/http"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Eine Messung ohne eingeschalteten Anbieter scheitert überall gleich.**
func TestEineModellmessungOhneAnbieterUeberJedeFlaeche(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "eine Messung ohne eingeschalteten Anbieter", abweisung{
		klasse: klasseValidierung,
		kern:   "eingeschaltet",
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).KIModellPruefen(t.Context(), "openai", "gpt-probe", "").Fehler
		},
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung,
				"ki-modell-pruefen", "openai", "gpt-probe")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.ki.modell.pruefen",
				map[string]any{"anbieter": "openai", "alias": "gpt-probe"})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/ki/modelle/pruefen",
				`{"anbieter":"openai","alias":"gpt-probe"}`)
		},
	})
}

// **Und eine Messung zurückzunehmen geht überall — auch wenn es keine gibt.**
//
// # Warum das der richtige Ausgang ist
//
// **Weil »vergessen« ein Zustand ist und keine Handlung an einem Gegenstand.**
// Wer eine Messung wegnimmt, die es nicht gibt, will danach dasselbe wie
// jemand, der eine wegnimmt, die es gab: dass keine mehr dasteht. Ein
// `nichtGefunden` machte daraus eine Frage nach dem vorigen Zustand — und die
// hat niemand gestellt.
func TestEineMessungZuVergessenGehtUeberJedeFlaeche(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	fuerAlleDreiMit(t, "eine Messung vergessen",
		[3]string{"vergessen", `"vergessen": true`, `"vergessen":true`},
		func() string {
			return mussKommandozeile(t, umgebung,
				"ki-modell-vergessen", "openai", "gpt-probe")
		},
		func() string {
			aus, istFehler := mcpRuf(t, umgebung,
				"schreiben.ki.modell.vergessen",
				map[string]any{"anbieter": "openai", "alias": "gpt-probe"})
			if istFehler {
				t.Fatalf("schreiben.ki.modell.vergessen: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodPost, "/ki/modelle/vergessen",
				`{"anbieter":"openai","alias":"gpt-probe"}`)
			if status != http.StatusOK {
				t.Fatalf("POST ki/modelle/vergessen: %d (%s)", status, rumpf)
			}

			return rumpf
		})
}
