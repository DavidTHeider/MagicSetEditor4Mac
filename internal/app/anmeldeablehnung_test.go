package app_test

// **Eine Ablehnung des Anbieters beendet den Fluss** — SEC-18
// (#2152).
//
// Seit der Horcher nur noch den passenden Rückruf annimmt, kommt ein fremder
// `state` gar nicht mehr bis zur Naht. Was bleibt, ist der **richtige**
// Rückruf mit `error=` statt `code=` — der Mensch hat beim Anbieter
// abgelehnt. Er darf nichts eintauschen.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anmeldung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestEineAblehnungDesAnbietersTauschtNichts(t *testing.T) {
	t.Parallel()

	getauscht := false

	anbieter := httptest.NewServer(http.HandlerFunc(
		func(http.ResponseWriter, *http.Request) { getauscht = true }))
	t.Cleanup(anbieter.Close)

	naht := app.NeueAnmeldenaht(&anmeldedienstprobe{},
		func(context.Context, string) (app.Anbieterzugang, error) {
			return app.Anbieterzugang{Kennung: "k", Geheimnis: "g"}, nil
		},
		func(ctx context.Context, adresse string) error {
			zerlegt, fehler := url.Parse(adresse)
			if fehler != nil {
				return fehler
			}

			werte := zerlegt.Query()
			rueckruf := werte.Get("redirect_uri") + "?error=access_denied&state=" +
				url.QueryEscape(werte.Get("state"))

			anfrage, baufehler := http.NewRequestWithContext(ctx, http.MethodGet, rueckruf, http.NoBody)
			if baufehler != nil {
				return baufehler
			}

			antwort, ruffehler := http.DefaultClient.Do(anfrage)
			if ruffehler != nil {
				return ruffehler
			}

			return antwort.Body.Close()
		},
		anmeldung.RuferUeber(anbieter.Client())).MitAnbietern(gegen(anbieter.URL))

	if _, fehler := naht.Anmelden(t.Context(), "github"); fehler == nil {
		t.Error("eine Ablehnung meldete einen Nutzer")
	}

	if getauscht {
		t.Error("nach der Ablehnung wurde getauscht")
	}
}
