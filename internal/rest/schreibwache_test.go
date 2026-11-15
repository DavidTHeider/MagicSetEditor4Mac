package rest_test

// **Die Schreibwache je Anfrage** — ID-03.AK-02
// (#1702).
// Der Dienst hält seine Verbindung über seine ganze Laufzeit; ob eine
// Anfrage schreiben darf, entscheidet sich bei jeder neu.

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine schreibende Anfrage fragt, eine lesende nicht — und die zwei
// Wechsel auch nicht**: wer nur noch lesen darf, muss hinauskommen.
func TestDieSchreibwacheStehtVorJederSchreibendenAnfrage(t *testing.T) {
	gefragt := 0

	dienst := neuerDienst(t).MitSchreibwache(func(context.Context) error {
		gefragt++

		return fmt.Errorf("%w: seit dem Start entzogen", domain.ErrNurLesend)
	})

	faelle := []struct {
		ziel     endpunkt
		verwehrt bool
	}{
		{endpunkt{name: "lesen", methode: http.MethodGet, weg: "/sets"}, false},
		{endpunkt{name: "schreiben", methode: http.MethodPost, weg: "/sets/1/karten", rumpf: `{}`}, true},
		{endpunkt{name: "Profil wechseln", methode: http.MethodPut, weg: "/profil", rumpf: `{"profilID":"2"}`}, false},
		{endpunkt{name: "Nutzer wechseln", methode: http.MethodPut, weg: "/nutzer", rumpf: `{"nutzerID":"2"}`}, false},
	}

	for _, fall := range faelle {
		antwort := ruf(t, dienst, fall.ziel, gueltig)

		verwehrt := antwort.Code == http.StatusForbidden && alsBefund(t, antwort).Code == "nurLesend"
		if verwehrt != fall.verwehrt {
			t.Errorf("%s: verwehrt %v, erwartet %v — %d %s",
				fall.ziel.name, verwehrt, fall.verwehrt, antwort.Code, antwort.Body.String())
		}
	}

	if gefragt != 1 {
		t.Errorf("die Wache wurde %d-mal gefragt, erwartet einmal", gefragt)
	}
}
