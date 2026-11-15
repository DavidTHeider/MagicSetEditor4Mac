// **Die Randfälle der Rückkehren und der Pflege am Dienst** — Werkstück O7
// (#1431)
// und O8 (#1432)
// von Phase 5.2.
//
// Die vier Klassen — gültig, fremde Herkunft, ohne Marke, über der Grenze —
// prüft `endpunkte_test.go` für **jeden** Endpunkt. Was hier danebensteht,
// sind die Fälle, die nur diese Wege haben: eine fehlende Dateiangabe, eine
// Kennung, die keine ist, und eine Pflege ohne Zeitpunkt der Vorschau.
package rest_test

import (
	"net/http"
	"strings"
	"testing"
)

func TestEineKrummeAngabeAnDenRueckkehrenWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		methode, weg, rumpf, teil string
	}{
		// **Ohne Zeitpunkt gibt es keine Pflege** (O8, #1432) — er ist die
		// Bindung an die Vorschau, und ohne ihn führte jemand etwas aus, das
		// er nicht angesehen hat.
		"die Pflege ohne Zeitpunkt der Vorschau": {
			http.MethodPost, "/pflege", `{"wahl":[]}`, "erstellt",
		},
		// **Und ein Rumpf, der keiner ist** — ein Fehler der Anfrage, kein
		// leerer Plan.
		"die Pflege mit einem Rumpf, der keiner ist": {
			http.MethodPost, "/pflege", `{`, "",
		},
		// **Ohne Datei gibt es keine Lage** — und die Meldung nennt das Feld
		// (§8: Fehler erscheinen am verursachenden Feld).
		"die Lage ohne Datei": {
			http.MethodGet, "/ruecksprung/lage", "", "datei",
		},
		// **Und eine Kennung, die keine Zahl ist, ist ein Fehler der
		// Anfrage** — kein „nicht gefunden".
		"eine Quittung auf eine krumme Kennung": {
			http.MethodPost, "/rueckkehren/sieben/quittung", "", "Kennung",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), endpunkt{
				methode: fall.methode, weg: fall.weg, rumpf: fall.rumpf,
			}, gueltig)

			if antwort.Code != http.StatusBadRequest {
				t.Fatalf("Status %d, erwartet war 400: %s",
					antwort.Code, antwort.Body.String())
			}

			if !strings.Contains(antwort.Body.String(), fall.teil) {
				t.Errorf("die Meldung nennt %q nicht: %s",
					fall.teil, antwort.Body.String())
			}
		})
	}
}
