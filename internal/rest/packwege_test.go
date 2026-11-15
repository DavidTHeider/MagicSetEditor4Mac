// **Die Packwege des Dienstes** —
// #889,
// Befund 7 des Zweitblicks über Phase W.
package rest_test

import (
	"net/http"
	"strings"
	"testing"
)

// **Die Packarten kommen durch, samt Zahl und Wählbarkeit.**
func TestDiePackartenKommenDurch(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t),
		endpunkt{methode: http.MethodGet, weg: "/sets/1/packarten"}, gueltig)

	if antwort.Code != http.StatusOK {
		t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
	}

	for _, erwartet := range []string{
		`"name":"alles"`, `"waehlbar":true`, `"karten":1`, `"befunde":[]`,
	} {
		if !strings.Contains(antwort.Body.String(), erwartet) {
			t.Errorf("%s fehlt in\n%s", erwartet, antwort.Body.String())
		}
	}
}

// **Und die Ziehung samt Kennung und Namen.**
//
// **Der Same steht in der Abfrage** und nicht im Rumpf: er gehört zum Aufruf,
// nicht zum Inhalt, und ein `POST` ohne Rumpf ist einer, den jeder Klient
// senden kann.
func TestEinePackziehungKommtDurch(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t), endpunkt{
		methode: http.MethodPost, weg: "/sets/1/packarten/alles/ziehung?same=4242",
	}, gueltig)

	if antwort.Code != http.StatusOK {
		t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
	}

	for _, erwartet := range []string{`"id":"1"`, `"name":"Probe"`} {
		if !strings.Contains(antwort.Body.String(), erwartet) {
			t.Errorf("%s fehlt in\n%s", erwartet, antwort.Body.String())
		}
	}
}

// **Eine unbrauchbare Setkennung kommt gar nicht erst am Dienst an.**
func TestEinePackziehungMitUnbrauchbarerKennungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, ziel := range map[string]endpunkt{
		"Packarten": {methode: http.MethodGet, weg: "/sets/keine/packarten"},
		"Ziehung": {
			methode: http.MethodPost, weg: "/sets/keine/packarten/alles/ziehung",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), ziel, gueltig)
			if antwort.Code != http.StatusBadRequest {
				t.Errorf("Status %d: %s", antwort.Code, antwort.Body.String())
			}
		})
	}
}
