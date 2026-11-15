package rest_test

import (
	"net/http"
	"testing"
)

// **Ein unbrauchbarer Rumpf ist 400** — und ruft den Dienst nicht.
func TestDieSprachwegeWeisenUnbrauchbaresAb(t *testing.T) {
	t.Parallel()

	for _, fall := range []endpunkt{
		{name: "Anlegen, Rumpf", methode: http.MethodPost, weg: "/sprachpakete", rumpf: `{`},
		{name: "Ändern, Rumpf", methode: http.MethodPatch, weg: "/sprachpakete/da", rumpf: `{`},
		{name: "Einlesen, Rumpf", methode: http.MethodPost, weg: "/sprachpakete/einlesen", rumpf: `{`},
		{name: "Einlesen ohne Text", methode: http.MethodPost, weg: "/sprachpakete/einlesen", rumpf: `{"inhalt":""}`},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), fall, gueltig)
			if antwort.Code != http.StatusBadRequest {
				t.Errorf("endete mit %d statt 400: %s", antwort.Code, antwort.Body.String())
			}
		})
	}
}
