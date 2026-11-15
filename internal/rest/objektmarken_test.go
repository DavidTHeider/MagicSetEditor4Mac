package rest_test

// **Die Marken an jedem Fachobjekt weisen ab, was keine Marke ist** — UX-12
// (#1684).

import (
	"net/http"
	"testing"
)

func TestEineLeereMarkeOderEinKaputterRumpfErreichtDieNahtNicht(t *testing.T) {
	t.Parallel()

	for _, ziel := range []endpunkt{
		{name: "setzen ohne Marke", methode: http.MethodPut, weg: "/objektmarken/deck/11/%20"},
		{name: "entfernen ohne Marke", methode: http.MethodDelete, weg: "/objektmarken/deck/11/%20"},
		{name: "überall entfernen ohne Marke", methode: http.MethodDelete, weg: "/marken/%20"},
		{name: "vergeben mit kaputtem Rumpf", methode: http.MethodPost, weg: "/objektmarken/deck/vergabe", rumpf: `{`},
		{name: "umbenennen mit kaputtem Rumpf", methode: http.MethodPost, weg: "/marken/umbenennung", rumpf: `{`},
	} {
		t.Run(ziel.name, func(t *testing.T) {
			t.Parallel()

			gezaehlt := &zaehlender{}

			antwort := ruf(t, dienstMit(t, gezaehlt), ziel, gueltig)
			if antwort.Code != http.StatusBadRequest || gezaehlt.rufe != 0 {
				t.Errorf("Status %d, %d Rufe: %s", antwort.Code, gezaehlt.rufe, antwort.Body.String())
			}
		})
	}
}
