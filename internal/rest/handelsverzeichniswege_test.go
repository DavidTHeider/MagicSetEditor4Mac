package rest_test

// **Die Liste der Handelsvorgänge im Dienst** — TRADE-01
// (#1715):
// was nur dieser Weg hat, ist ein Wahrheitswert in der Abfrage — und ein
// krummer wird abgewiesen, statt still `false` zu heißen.

import (
	"net/http"
	"testing"
)

func TestEinKrummesArchivInDerAbfrageWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodGet, weg: "/handel?archiv=vielleicht",
	}, gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Errorf("Status %d, erwartet war 400: %s", antwort.Code, antwort.Body.String())
	}
}

// **Ein Entwurf, der kein JSON ist, wird abgewiesen** — bevor die Naht
// gefragt wird.
func TestEinKaputterEntwurfWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodPut, weg: "/handel/v-1/entwurf", rumpf: `kein JSON`,
	}, gueltig)

	if antwort.Code != http.StatusBadRequest {
		t.Errorf("Status %d, erwartet war 400: %s", antwort.Code, antwort.Body.String())
	}
}
