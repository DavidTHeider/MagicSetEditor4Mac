package app_test

// **Eine Anmeldung ohne Rückruf endet nach ihrer Frist** — BK-07d
// (#2118).
//
// Der Horcher wartete, bis der Kontext des Rufs abbrach; eine eigene Frist
// hatte der Fluss nicht. Kam der Rückruf nie — der Browser blieb offen, der
// Anbieter antwortete nicht —, stand der Port der Schleifenadresse so lange
// offen, wie der Ruf lebte.

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anmeldung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineAnmeldungOhneRueckrufLaeuftNachIhrerFristAb(t *testing.T) {
	t.Parallel()

	// **Der Browser geht auf und antwortet nie** — die Adresse wird gemerkt,
	// damit der Port danach geprüft werden kann.
	var rueckadresse string

	naht := app.NeueAnmeldenaht(&anmeldedienstprobe{},
		func(context.Context, string) (app.Anbieterzugang, error) {
			return app.Anbieterzugang{Kennung: "k", Geheimnis: "g"}, nil
		},
		func(_ context.Context, adresse string) error {
			zerlegt, fehler := url.Parse(adresse)
			if fehler != nil {
				return fehler //nolint:wrapcheck // die Naht meldet ihn unverändert
			}

			rueckadresse = zerlegt.Query().Get("redirect_uri")

			return nil
		},
		anmeldung.RuferUeber(nil)).MitWartefrist(200 * time.Millisecond)

	anfang := time.Now()
	_, fehler := naht.Anmelden(t.Context(), "github")

	if !errors.Is(fehler, domain.ErrAnmeldungAbgelaufen) || errors.Is(fehler, context.Canceled) {
		t.Fatalf("Fehler %v, erwartet ErrAnmeldungAbgelaufen und kein Abbruch", fehler)
	}

	if dauer := time.Since(anfang); dauer > 5*time.Second {
		t.Errorf("die Anmeldung endete erst nach %v", dauer)
	}

	// **Und der Port ist danach zu.**
	ziel, fehler := url.Parse(rueckadresse)
	if fehler != nil || ziel.Host == "" {
		t.Fatalf("keine Rückadresse: %q, %v", rueckadresse, fehler)
	}

	// **Gefragt wird der Horcher und nicht der Port**: unter Last nahm ein
	// paralleler Test denselben Port, kaum dass er frei war (gemessen am
	// 2026-10-06, `127.0.0.1:51051`), und ein bloßes Wählen hielt dessen
	// Server für diesen. Antwortet dort noch die Seite des Horchers, ist er
	// offen geblieben.
	if horcherAntwortet(t, rueckadresse) {
		t.Errorf("der Horcher unter %s steht nach dem Ablauf noch offen", ziel.Host)
	}
}

// horcherAntwortet sagt, ob unter der Rückadresse noch der Horcher einer
// Anmeldung antwortet.
func horcherAntwortet(t *testing.T, rueckadresse string) bool {
	t.Helper()

	frist, aufhoeren := context.WithTimeout(t.Context(), time.Second)
	defer aufhoeren()

	anfrage, fehler := http.NewRequestWithContext(frist, http.MethodGet, rueckadresse+"?code=c&state=s", http.NoBody)
	if fehler != nil {
		t.Fatal(fehler)
	}

	antwort, fehler := http.DefaultClient.Do(anfrage)
	if fehler != nil {
		return false
	}

	defer antwort.Body.Close() //nolint:errcheck // nur gelesen

	inhalt, _ := io.ReadAll(antwort.Body) //nolint:errcheck // ein halber Inhalt genügt zum Vergleich

	return strings.Contains(string(inhalt), "MagicSetEditor4Mac hat die Antwort bekommen")
}

// **Ein zweiter Versuch nach dem Ablauf gelingt** — an derselben Naht: der
// erste Browser antwortet nie, der zweite schon.
func TestNachEinemAblaufGelingtDerNaechsteVersuch(t *testing.T) {
	t.Parallel()

	anbieter := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			schreiben.Header().Set("Content-Type", "application/json")

			rumpf := `{"id":4711,"login":"wer","name":"Wer"}`
			if strings.HasSuffix(anfrage.URL.Path, "/token") {
				rumpf = `{"access_token":"abc"}`
			}

			if _, fehler := schreiben.Write([]byte(rumpf)); fehler != nil {
				t.Errorf("antworten: %v", fehler)
			}
		}))
	t.Cleanup(anbieter.Close)

	versuche := 0
	naht := app.NeueAnmeldenaht(&anmeldedienstprobe{},
		func(context.Context, string) (app.Anbieterzugang, error) {
			return app.Anbieterzugang{Kennung: "k", Geheimnis: "g"}, nil
		},
		func(ctx context.Context, adresse string) error {
			versuche++
			if versuche == 1 {
				return nil
			}

			return rufeZurueck(ctx, adresse, "")
		},
		anmeldung.RuferUeber(anbieter.Client())).
		MitAnbietern(gegen(anbieter.URL)).
		MitWartefrist(300 * time.Millisecond)

	if _, fehler := naht.Anmelden(t.Context(), "github"); !errors.Is(fehler, domain.ErrAnmeldungAbgelaufen) {
		t.Fatalf("erster Versuch: %v", fehler)
	}

	if _, fehler := naht.Anmelden(t.Context(), "github"); fehler != nil {
		t.Errorf("zweiter Versuch: %v", fehler)
	}
}
