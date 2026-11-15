// **Die Abweisungen der drei Umschlagwege** — Werkstück N4 von Phase 5.4
// (#1457).
//
// **Der Fehlerweg neben dem Erfolg** (§10): `endpunktliste_test.go` fährt
// jeden Weg einmal glücklich, und was ein leeres Feld oder ein kaputter Rumpf
// gibt, stünde ohne diese Läufe nirgends.
package rest_test

import (
	"net/http"
	"strings"
	"testing"
)

// **Eine leere Angabe wird abgewiesen — und sie nennt ihr Feld.**
func TestEineLeereAngabeAmUmschlagWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		methode, weg, rumpf, feld string
	}{
		"schnüren ohne Set": {
			http.MethodPost, "/umschlag", `{"setID":"  ","ziel":"/tmp/x"}`, "setID",
		},
		"schnüren ohne Ziel": {
			http.MethodPost, "/umschlag", `{"setID":"1","ziel":"   "}`, "ziel",
		},
		"schauen ohne Pfad": {
			http.MethodGet, "/umschlag", "", "pfad",
		},
		"übernehmen ohne Pfad": {
			http.MethodPost, "/umschlag/uebernahme", `{"pfad":"  "}`, "pfad",
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
func TestEinKaputterRumpfAmUmschlagWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, weg := range map[string]string{
		"schnüren":   "/umschlag",
		"übernehmen": "/umschlag/uebernahme",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), endpunkt{
				methode: http.MethodPost, weg: weg, rumpf: `kein JSON`,
			}, gueltig)

			if antwort.Code != http.StatusBadRequest {
				t.Errorf("Status %d, erwartet war 400: %s",
					antwort.Code, antwort.Body.String())
			}
		})
	}
}
