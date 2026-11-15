package rest_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// zusatzwahl hält die Zusatzfelder einer Karte und schreibt mit.
type zusatzwahl struct {
	gesetzt []string
	fehler  error
}

func (z *zusatzwahl) Zusatzfelder(_ context.Context, _, _ int64) ([]rest.Zusatzfeld, error) {
	if z.fehler != nil {
		return nil, z.fehler
	}

	return []rest.Zusatzfeld{{Name: "vermerk", Art: "text", Wert: "alt", Gespeichert: true}}, nil
}

func (z *zusatzwahl) ZusatzwertSetzen(_ context.Context, _, _ int64, feld, wert string) error {
	if z.fehler != nil {
		return z.fehler
	}

	z.gesetzt = append(z.gesetzt, feld+"="+wert)

	return nil
}

func zusatzdienst(wahl *zusatzwahl) http.Handler {
	return rest.NeuerDienst(probemarke, time.Now, rest.Naehte{Leser: williger{}, Zusatz: wahl}).
		MitGrenzen(domain.Vorgabegrenzen()).Behandler()
}

// **Die Zusatzfelder einer Karte über REST** (#2298) — lesen und setzen,
// dieselbe Fähigkeit wie im Fenster, auf der Kommandozeile und über MCP.
func TestDieZusatzfelderStehenAlsEndpunkt(t *testing.T) {
	wahl := &zusatzwahl{}
	dienst := zusatzdienst(wahl)

	gelesen := schicke(t, dienst, http.MethodGet, "/sets/7/karten/1/zusatzfelder", "", nil)
	if gelesen.Code != http.StatusOK || !strings.Contains(gelesen.Body.String(), `"vermerk"`) {
		t.Errorf("GET: %d %s", gelesen.Code, gelesen.Body.String())
	}

	gesetzt := schicke(t, dienst, http.MethodPut, "/sets/7/karten/1/zusatzfelder/vermerk",
		"application/json", []byte(`{"wert":"neu"}`))
	if gesetzt.Code != http.StatusOK || strings.Join(wahl.gesetzt, ";") != "vermerk=neu" {
		t.Errorf("PUT: %d %s, gesetzt %v", gesetzt.Code, gesetzt.Body.String(), wahl.gesetzt)
	}

	// **Ein ausdrücklich leerer Wert leert** — er ist eine Aussage, ein
	// fehlender keine.
	geleert := schicke(t, dienst, http.MethodPut, "/sets/7/karten/1/zusatzfelder/vermerk",
		"application/json", []byte(`{"wert":""}`))
	if geleert.Code != http.StatusOK || strings.Join(wahl.gesetzt, ";") != "vermerk=neu;vermerk=" {
		t.Errorf("leerer PUT: %d %s, gesetzt %v", geleert.Code, geleert.Body.String(), wahl.gesetzt)
	}
}

// **Was nicht geht, wird benannt.**
func TestDieZusatzfelderWeisenAbWasNichtGeht(t *testing.T) {
	for _, fall := range []struct {
		name, methode, weg, rumpf string
		fehler                    error
		code                      int
	}{
		{name: "eine kaputte Kennung", methode: http.MethodGet, weg: "/sets/x/karten/1/zusatzfelder", code: http.StatusBadRequest},
		{name: "eine kaputte Karte", methode: http.MethodPut, weg: "/sets/7/karten/x/zusatzfelder/vermerk", rumpf: `{"wert":"a"}`, code: http.StatusBadRequest},
		{name: "ein kaputter Rumpf", methode: http.MethodPut, weg: "/sets/7/karten/1/zusatzfelder/vermerk", rumpf: `{`, code: http.StatusBadRequest},
		{name: "ein fehlender Wert", methode: http.MethodPut, weg: "/sets/7/karten/1/zusatzfelder/vermerk", rumpf: `{}`, code: http.StatusBadRequest},
		{name: "ein Wert null", methode: http.MethodPut, weg: "/sets/7/karten/1/zusatzfelder/vermerk", rumpf: `{"wert":null}`, code: http.StatusBadRequest},
		{name: "eine Karte, die es nicht gibt", methode: http.MethodGet, weg: "/sets/7/karten/1/zusatzfelder", fehler: domain.ErrNotFound, code: http.StatusNotFound},
		{name: "ein Feld, das es nicht gibt", methode: http.MethodPut, weg: "/sets/7/karten/1/zusatzfelder/x", rumpf: `{"wert":"a"}`, fehler: domain.NewFieldError("feld", domain.SchluesselValidierung, "kein Zusatzfeld"), code: http.StatusBadRequest},
	} {
		t.Run(fall.name, func(t *testing.T) {
			art := ""
			if fall.rumpf != "" {
				art = "application/json"
			}

			antwort := schicke(t, zusatzdienst(&zusatzwahl{fehler: fall.fehler}), fall.methode, fall.weg, art, []byte(fall.rumpf))
			if antwort.Code != fall.code {
				t.Errorf("%d statt %d: %s", antwort.Code, fall.code, antwort.Body.String())
			}
		})
	}
}
