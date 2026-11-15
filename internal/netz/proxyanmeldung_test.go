package netz_test

// **Ein Proxy, der eine Anmeldung verlangt, heißt so** — D02
// (#2233).
//
// Zwei Wege, und beide kommen vor: für `https://` fragt der Transport den
// Proxy mit `CONNECT`, und ein `407` darauf war bis dahin ein Fehler mit dem
// Statustext als einziger Auskunft; für `http://` kommt das `407` als
// **Antwort** zurück, und der Aufrufer hielt es für die des Dienstes.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

func TestEinProxyMitAnmeldungHeisstSo(t *testing.T) {
	t.Setenv(netz.SchleifeUmgebung, "an")
	t.Setenv("NO_PROXY", "")

	proxy := httptest.NewServer(http.HandlerFunc(func(schreiben http.ResponseWriter, _ *http.Request) {
		schreiben.Header().Set("Proxy-Authenticate", `Basic realm="firma"`)
		schreiben.WriteHeader(http.StatusProxyAuthRequired)
	}))
	t.Cleanup(proxy.Close)

	ziel, _ := url.Parse(proxy.URL) //nolint:errcheck // die Adresse des Testservers
	ziel.User = url.UserPassword("anna", "geheim")

	klient := netz.NeuerKlientMit(netz.UeblicheFristen(), func(context.Context) netz.Proxyvorgabe {
		return netz.Proxyvorgabe{HTTP: ziel.String(), HTTPS: ziel.String()}
	})

	for _, adresse := range []string{"https://dienst.invalid/karten", "http://dienst.invalid/karten"} {
		t.Run(adresse, func(t *testing.T) {
			_, fehler := netz.Hole(t.Context(), klient, nil, netz.Kenner{Name: "test", Annahme: "*/*"}, adresse, 1<<20)

			var gegenstelle *domain.Gegenstellenfehler
			if !errors.Is(fehler, domain.ErrProxyAnmeldung) || !errors.As(fehler, &gegenstelle) {
				t.Fatalf("der Fehler %v nennt die Anmeldung am Proxy nicht", fehler)
			}

			if gegenstelle.Wirt != ziel.Host {
				t.Errorf("Wirt %q, erwartet der Proxy %q", gegenstelle.Wirt, ziel.Host)
			}

			if strings.Contains(fehler.Error(), "geheim") {
				t.Errorf("die Meldung trägt das Kennwort des Proxys: %v", fehler)
			}
		})
	}
}
