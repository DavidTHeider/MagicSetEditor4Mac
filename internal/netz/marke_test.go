// **Eine Marke am Kenner**
// (#1100).
//
// Das Repositorium, dessen Freigaben zählen, ist **privat**; ohne Marke
// antwortet GitHub mit 404. Sie gehört an den Kenner, weil ein Kenner genau
// das ist: die Art, wie dieser Klient sich vorstellt.
package netz_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// gesehen sammelt die Kopfzeilen einer Anfrage.
func gesehen(t *testing.T) (*httptest.Server, *http.Header) {
	t.Helper()

	var kopf http.Header

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			kopf = anfrage.Header.Clone()

			if _, fehler := schreiben.Write([]byte("gut")); fehler != nil {
				t.Errorf("schreiben: %v", fehler)
			}
		}))

	t.Cleanup(dienst.Close)

	return dienst, &kopf
}

// **Eine gesetzte Marke geht als `Authorization` hinaus.**
func TestEineMarkeGehtAlsAuthorizationHinaus(t *testing.T) {
	t.Parallel()

	dienst, kopf := gesehen(t)

	kenner := netz.Kenner{Name: "test", Annahme: "*/*", Marke: "ghp_geheim"}

	for name, holen := range map[string]func() error{
		"Hole": func() error {
			_, fehler := netz.Hole(t.Context(), netz.KlientUeber(dienst.Client()), nil,
				kenner, dienst.URL, 1<<20)

			return fehler
		},
		"Strom": func() error {
			antwort, fehler := netz.Strom(t.Context(), netz.KlientUeber(dienst.Client()), nil,
				kenner, dienst.URL)
			if fehler != nil {
				return fehler
			}

			return antwort.Rumpf.Close()
		},
	} {
		t.Run(name, func(t *testing.T) {
			if fehler := holen(); fehler != nil {
				t.Fatalf("%s: %v", name, fehler)
			}

			if kopf.Get("Authorization") != "Bearer ghp_geheim" {
				t.Errorf("die Marke ging als %q hinaus", kopf.Get("Authorization"))
			}
		})
	}
}

// **Ohne Marke geht keine leere Kopfzeile hinaus.**
//
// Ein leeres `Authorization` ist etwas anderes als keines: manche Dienste
// weisen es ab, statt die Anfrage als unangemeldet zu behandeln.
func TestOhneMarkeGehtKeineKopfzeileHinaus(t *testing.T) {
	t.Parallel()

	dienst, kopf := gesehen(t)

	if _, fehler := netz.Hole(t.Context(), netz.KlientUeber(dienst.Client()), nil,
		netz.Kenner{Name: "test", Annahme: "*/*"}, dienst.URL, 1<<20); fehler != nil {
		t.Fatalf("Hole: %v", fehler)
	}

	if _, da := (*kopf)["Authorization"]; da {
		t.Error("ohne Marke ging eine Authorization-Kopfzeile hinaus")
	}
}
