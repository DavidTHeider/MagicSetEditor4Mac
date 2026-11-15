package rest_test

// **Ein Set vorab ansehen am Dienst** — FMT-03.AK-01 (#1692).

import (
	"net/http"
	"strings"
	"testing"
)

func TestDerDienstZeigtEineSetvorschau(t *testing.T) {
	antwort := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodPost, weg: "/sets/vorschau", rumpf: `{"pfad":"/tmp/probe.mse-set"}`,
	}, gueltig)

	if antwort.Code != http.StatusOK || !strings.Contains(antwort.Body.String(), `"vorschau":{`) ||
		!strings.Contains(antwort.Body.String(), "Mein Probeset") {
		t.Errorf("%d %s", antwort.Code, antwort.Body.String())
	}
}

func TestEineSetvorschauOhnePfadWirdAbgewiesen(t *testing.T) {
	for _, rumpf := range []string{`{}`, `{"falsch":1}`} {
		antwort := ruf(t, neuerDienst(t), endpunkt{
			methode: http.MethodPost, weg: "/sets/vorschau", rumpf: rumpf,
		}, gueltig)

		if antwort.Code != http.StatusBadRequest {
			t.Errorf("%s: %d %s", rumpf, antwort.Code, antwort.Body.String())
		}
	}
}
