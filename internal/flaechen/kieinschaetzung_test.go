// **Die Einschätzung über jede Fläche ohne Fenster** — Werkstück A13.2a von
// Phase 5.3
// (#1452).
//
// # Was hier gemessen wird
//
// **Dass eine erfundene Art überall dieselbe fachliche Absage bekommt — und
// dass sie sie bekommt, bevor irgendetwas geschieht.** Die Art ist die
// billigste Prüfung des ganzen Wegs; wer sie hinter den Kontextdienst legte,
// ließe eine Recherche für eine Frage laufen, die es nicht gibt.
//
// # Und was er nicht deckt
//
// **Eine Einschätzung, die wirklich entsteht.** Sie braucht einen gebauten
// Suchindex **und** eine laufende Einbettungslaufzeit; beides steht hinter
// `netzlauf` und nicht in der Pipeline (dieselbe Grenze wie bei
// `recherche_test.go`). Was hier steht, ist die **Gleichheit der Flächen** —
// die Zusage aus §16.
package flaechen_test

import (
	"net/http"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Eine erfundene Art wird an jeder Fläche gleich abgewiesen.**
func TestEineErfundeneEinschaetzungsartUeberJedeFlaeche(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "eine erfundene Einschätzungsart", abweisung{
		klasse: klasseValidierung,
		kern:   "regelchen",
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).KIEinschaetzen(t.Context(), app.KIEinschaetzungsbezug{
				Art: "regelchen", KarteID: 1, Anbieter: "ollama", Alias: "gemma4:8b",
			}).Fehler
		},
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung,
				"ki-einschaetzung", "1", "regelchen", "ollama", "gemma4:8b",
				"Wie wirkt sie?")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.ki.einschaetzung",
				map[string]any{
					"karteId": 1, "art": "regelchen", "anbieter": "ollama",
					"alias": "gemma4:8b", "frage": "Wie wirkt sie?",
				})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/karten/1/ki-einschaetzung",
				`{"art":"regelchen","anbieter":"ollama","alias":"gemma4:8b",`+
					`"frage":"Wie wirkt sie?"}`)
		},
	})
}
