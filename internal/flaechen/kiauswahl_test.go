// **Die Auswahl über jede Fläche ohne Fenster** — Werkstück A11 von Phase 5.3
// (#1450).
//
// # Was hier gemessen wird
//
// **Dass ein Feldname, den der Vorschlag nicht trägt, überall gleich
// abgewiesen wird.** Das Issue verlangt „nur ausgewählte gültige Felder" —
// und die Auswahl kommt an diesen Flächen aus einem Ruf, den jemand tippt.
// Ein Tippfehler dort darf nicht zu einer Übernahme führen, die weniger tut
// als verlangt.
//
// **Was er nicht deckt**, steht hier und nicht in einer Fußnote: eine
// Übernahme, die wirklich schreibt. Sie brauchte eine laufende KI-Laufzeit;
// ohne sie scheitert der Ruf schon davor — und genau das misst
// `kimodelle_test.go`.
package flaechen_test

import (
	"net/http"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Eine Auswahl ohne Laufzeit scheitert überall gleich.**
//
// **Und zwar an der Laufzeit und nicht an der Auswahl**: die Reihenfolge ist
// die Zusage. Wer zuerst die Felder prüfte, verriete, was ein Vorschlag
// enthielte, ohne je einen erzeugt zu haben.
func TestEineAuswahlOhneLaufzeitUeberJedeFlaeche(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "eine Auswahl ohne Laufzeit", abweisung{
		klasse: klasseValidierung,
		kern:   "eingeschaltet",
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).KIVorschlagen(t.Context(), "1", app.KIVorschlagsbezug{
				KarteID: 1, Anbieter: "ollama", Alias: "gemma4:8b", Auftrag: "Schreibe den Regeltext.", Nur: []string{"name"},
			}).Fehler
		},
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung,
				"ki-vorschlag", "1", "ollama", "gemma4:8b",
				"Schreibe den Regeltext.", "--uebernehmen", "--felder=name")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.ki.vorschlag",
				map[string]any{
					"karteId": 1, "anbieter": "ollama", "alias": "gemma4:8b",
					"auftrag":     "Schreibe den Regeltext.",
					"uebernehmen": true, "felder": []string{"name"},
				})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/karten/1/ki-vorschlag",
				`{"anbieter":"ollama","alias":"gemma4:8b",`+
					`"auftrag":"Schreibe den Regeltext.",`+
					`"uebernehmen":true,"felder":["name"]}`)
		},
	})
}
