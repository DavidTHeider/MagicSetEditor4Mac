// **Die Fehlerwege der Bestandssicherung am Dienst** —
// #1169.
//
// Die vier Klassen — gültig, fremde Herkunft, ohne Marke, über der Grenze —
// prüft `endpunkte_test.go` für **jeden** Endpunkt. Was hier danebensteht,
// sind die Fälle, die nur diese beiden haben: eine leere Angabe und ein Rumpf,
// der keiner ist.
package rest_test

import (
	"net/http"
	"strings"
	"testing"
)

// **Eine leere Angabe wird abgewiesen.**
//
// **Und zwar mit dem Namen des Feldes.** Ein „ungültige Anfrage" ohne ihn
// schickte den Aufrufer suchen (§8: Fehler erscheinen am verursachenden Feld).
func TestEineLeereAngabeAnDerSicherungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		methode, weg, rumpf, feld string
	}{
		"sichern ohne Ziel": {
			http.MethodPost, "/bestand/sicherung", `{"ziel":"   "}`, "ziel",
		},
		"schauen ohne Pfad": {
			http.MethodGet, "/bestand/sicherung", "", "pfad",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), endpunkt{
				methode: fall.methode, weg: fall.weg, rumpf: fall.rumpf,
			}, gueltig)

			if antwort.Code != http.StatusBadRequest {
				t.Fatalf("Status %d, erwartet war 400: %s", antwort.Code, antwort.Body.String())
			}

			if !strings.Contains(antwort.Body.String(), fall.feld) {
				t.Errorf("die Meldung nennt das Feld %q nicht: %s",
					fall.feld, antwort.Body.String())
			}
		})
	}
}

// **Ein Rumpf, der kein JSON ist, wird abgewiesen.**
func TestEinKaputterRumpfAnDerSicherungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodPost, weg: "/bestand/sicherung", rumpf: `kein JSON`,
	}, gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Errorf("Status %d, erwartet war 400: %s", antwort.Code, antwort.Body.String())
	}
}
