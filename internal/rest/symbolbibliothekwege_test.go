// **Die Abweisungen der Bibliothekswege** — DOM-02
// (#1697).
//
// Die vier Klassen hält `endpunktliste_test.go`; hier steht, was kein
// Endpunktlauf betritt: eine Kennung, die keine ist, und ein Rumpf, der
// sich nicht lesen lässt.

package rest_test

import (
	"net/http"
	"testing"
)

func TestDieBibliothekswegeWeisenUnbrauchbaresAb(t *testing.T) {
	t.Parallel()

	faelle := []endpunkt{
		{name: "Text, Kennung", methode: http.MethodGet, weg: "/symbole/x"},
		{name: "Kopie, Kennung", methode: http.MethodPost, weg: "/symbole/x/kopie", rumpf: `{"name":"X"}`},
		{name: "Name, Kennung", methode: http.MethodPut, weg: "/symbole/x/name", rumpf: `{"name":"X"}`},
		{name: "Löschen, Kennung", methode: http.MethodDelete, weg: "/symbole/x"},
		{name: "Einsetzen, Kennung", methode: http.MethodPut, weg: "/sets/x/bibliothekssymbol", rumpf: `{"id":1}`},
		{name: "Aufnehmen, Rumpf", methode: http.MethodPost, weg: "/symbole", rumpf: `{`},
		{name: "Einlesen, Rumpf", methode: http.MethodPost, weg: "/symbole/einlesen", rumpf: `{`},
		{name: "Einlesen ohne Text", methode: http.MethodPost, weg: "/symbole/einlesen", rumpf: `{"inhalt":""}`},
		{name: "Kopie, Rumpf", methode: http.MethodPost, weg: "/symbole/1/kopie", rumpf: `{`},
		{name: "Einsetzen, Rumpf", methode: http.MethodPut, weg: "/sets/1/bibliothekssymbol", rumpf: `{`},
		{name: "Teile, Kennung", methode: http.MethodGet, weg: "/symbole/x/teile"},
		{name: "Sichern, Kennung", methode: http.MethodPut, weg: "/symbole/x/teile", rumpf: `{"teile":[]}`},
		{name: "Sichern, Rumpf", methode: http.MethodPut, weg: "/symbole/1/teile", rumpf: `{`},
		{name: "Zurücknehmen, Kennung", methode: http.MethodPost, weg: "/symbole/x/zuruecknahme"},
		{name: "Nachziehen, Kennung", methode: http.MethodPost, weg: "/symbole/x/nachzug"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), fall, gueltig)
			if antwort.Code != http.StatusBadRequest {
				t.Errorf("endete mit %d statt 400: %s", antwort.Code, antwort.Body.String())
			}
		})
	}
}
