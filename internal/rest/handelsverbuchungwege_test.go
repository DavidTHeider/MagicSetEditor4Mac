package rest_test

// **Die Wege der Verbuchung im Dienst** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// **Was nur dieser Weg hat**: ein Rumpf, der kein JSON ist. Der Schritt steht
// darin und nicht im Pfad — ein unlesbarer Rumpf ist damit kein unbekannter
// Schritt, sondern gar keiner, und beides bekäme sonst dieselbe Antwort.

import (
	"net/http"
	"testing"
)

// **Ein Rumpf, der kein JSON ist, wird abgewiesen.**
func TestEinKaputterRumpfAmHandelsschrittWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodPost, weg: "/handel/v-1/schritt",
		rumpf: `kein JSON`,
	}, gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Errorf("Status %d, erwartet war 400: %s",
			antwort.Code, antwort.Body.String())
	}
}

// **Und ein unbekanntes Feld ebenso** — ein Tippfehler im Feldnamen führte
// sonst zu einem leeren Schritt statt zu einer Meldung.
func TestEinFremdesFeldAmHandelsschrittWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodPost, weg: "/handel/v-1/schritt",
		rumpf: `{"schrit":"verbuchen"}`,
	}, gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Errorf("Status %d, erwartet war 400: %s",
			antwort.Code, antwort.Body.String())
	}
}
