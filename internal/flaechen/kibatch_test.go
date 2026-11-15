package flaechen_test

// **Ein Batch über jede Fläche** — Werkstück A13.1 von Phase 5.3
// (#1452).
//
// # Was hier gemessen wird
//
// **Dass die drei Flächen ohne Fenster denselben fachlichen Fehler geben.**
// Das ist die Zusage aus §16 — „dieselbe Fähigkeit, überall dieselbe
// Antwort" —, und ein Wächter, der nur zählt, ob es die Fläche gibt, sieht
// sie nicht.
//
// # Und warum an der **Grenze** und nicht an einem Erfolg
//
// **Weil ein Erfolg eine Laufzeit bräuchte** und ein Cloudanbieter eine
// Freigabe, die nur das Fenster ausstellt
// ([0105](../../docs/entscheidungen/0105-eine-freigabe-fuer-einen-batch.md)).
// Die Grenze dagegen gilt **vor** dem ersten Ruf und ist an jeder Fläche
// dieselbe fachliche Aussage.

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// zuVieleZiele baut eine Auswahl über der Grenze.
func zuVieleZiele(anzahl int) string {
	teile := make([]string, 0, anzahl)
	for i := range anzahl {
		teile = append(teile, itoa(i+1))
	}

	return strings.Join(teile, ",")
}

// **Ein zu großer Batch wird an jeder Fläche abgewiesen.**
func TestEinZuGrosserBatchUeberJedeFlaeche(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	const zuViele = 25

	abweisungGleich(t, "ein zu großer Batch", abweisung{
		klasse: klasseValidierung,
		kern:   "höchstens",
		fenster: func() *app.Fehler {
			ziele := make([]string, 0, zuViele)
			for _, ziel := range zielzahlen(zuViele) {
				ziele = append(ziele, strconv.Itoa(ziel))
			}

			return fensterAus(t, umgebung).KIBatchVorschlagen(t.Context(), ziele, "ollama", "gemma4:8b",
				"Schreibe den Regeltext.", "").Fehler
		},
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung,
				"ki-batch", zuVieleZiele(zuViele), "ollama", "gemma4:8b",
				"Schreibe den Regeltext.")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.ki.batch",
				map[string]any{
					"karteIds": zielzahlen(zuViele), "anbieter": "ollama",
					"alias":   "gemma4:8b",
					"auftrag": "Schreibe den Regeltext.",
				})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/ki/batch",
				`{"karteIds":[`+zuVieleZiele(zuViele)+`],`+
					`"anbieter":"ollama","alias":"gemma4:8b",`+
					`"auftrag":"Schreibe den Regeltext."}`)
		},
	})
}

// zielzahlen baut die Kennungen als Zahlen.
func zielzahlen(anzahl int) []int {
	ziele := make([]int, 0, anzahl)
	for i := range anzahl {
		ziele = append(ziele, i+1)
	}

	return ziele
}
