package app_test

// **Der ganze Anmeldefluss, ohne Browser** — Werkstück 4 von Phase 4.12
// (#1351).
//
// # Wie ein Fluss ohne Menschen geprüft wird
//
// **Der Browseröffner ist eine Naht**, und im Test tut er das, was der Browser
// täte: er ruft die Adresse ab. Der Anbieter ist eine `httptest`-Gegenstelle,
// die den Rückruf schickt — damit läuft **derselbe** Weg wie im Betrieb, mit
// `state`, PKCE und Rückruf-Horcher.

import (
	"context"
	"errors"
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

// **Der ganze Fluss führt zu einem Nutzer.**
func TestDerGanzeAnmeldeflussFuehrtZuEinemNutzer(t *testing.T) {
	t.Parallel()

	dienst := &anmeldedienstprobe{nutzerID: 7}
	naht := nahtGegen(t, dienst, nil)

	nutzerID, fehler := naht.Anmelden(t.Context(), "github")
	if fehler != nil {
		t.Fatalf("anmelden: %v", fehler)
	}

	if nutzerID != 7 {
		t.Errorf("Nutzer %d, erwartet war 7", nutzerID)
	}

	if dienst.gesehen.Subjekt != "4711" || dienst.gesehen.Anbieter != "github" {
		t.Errorf("der Dienst sah %+v", dienst.gesehen)
	}
}

// **Ein unbekannter Anbieter kommt gar nicht bis zum Browser.**
func TestEinUnbekannterAnbieterKommtNichtBisZumBrowser(t *testing.T) {
	t.Parallel()

	geoeffnet := false
	naht := nahtGegen(t, &anmeldedienstprobe{}, func() { geoeffnet = true })

	_, fehler := naht.Anmelden(t.Context(), "apple")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
	}

	if geoeffnet {
		t.Error("der Browser ging auf")
	}
}

// **Ein Abbruch beendet den Fluss.**
//
// **Und das ist mehr als eine Höflichkeit**: ohne ihn hinge das Fenster, bis
// jemand im Browser fertig wird — und wer abbricht, hat entschieden, dass er
// das nicht tut.
func TestEinAbbruchBeendetDenAnmeldefluss(t *testing.T) {
	t.Parallel()

	ctx, brich := context.WithCancel(t.Context())

	// **Der Browser bricht ab, statt zu antworten** — so sieht es aus, wenn
	// jemand das Fenster zumacht.
	naht := nahtGegen(t, &anmeldedienstprobe{}, brich)

	if _, fehler := naht.Anmelden(ctx, "github"); !errors.Is(
		fehler, context.Canceled) {
		t.Errorf("Fehler %v, erwartet war context.Canceled", fehler)
	}
}

// **Jeder Fehlerweg des Flusses meldet sich.**
func TestJederFehlerwegDesAnmeldeflussesMeldetSich(t *testing.T) {
	t.Parallel()

	t.Run("die Zugangsdaten fehlen", func(t *testing.T) {
		t.Parallel()

		naht := app.NeueAnmeldenaht(&anmeldedienstprobe{},
			func(context.Context, string) (app.Anbieterzugang, error) {
				return app.Anbieterzugang{}, errors.New("keine Kennung")
			},
			func(context.Context, string) error { return nil },
			anmeldung.NeuerRufer())

		if _, fehler := naht.Anmelden(t.Context(), "github"); fehler == nil {
			t.Error("der Fehler ging verloren")
		}
	})

	t.Run("die Zugangsdaten taugen nicht", func(t *testing.T) {
		t.Parallel()

		// **GitHub ohne Geheimnis** — nachgelesen: dort ist es Pflicht.
		naht := app.NeueAnmeldenaht(&anmeldedienstprobe{},
			func(context.Context, string) (app.Anbieterzugang, error) {
				return app.Anbieterzugang{Kennung: "k"}, nil
			},
			func(context.Context, string) error { return nil },
			anmeldung.NeuerRufer())

		if _, fehler := naht.Anmelden(t.Context(), "github"); !errors.Is(
			fehler, domain.ErrValidation) {
			t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
		}
	})

	t.Run("der Browser geht nicht auf", func(t *testing.T) {
		t.Parallel()

		naht := app.NeueAnmeldenaht(&anmeldedienstprobe{},
			func(context.Context, string) (app.Anbieterzugang, error) {
				return app.Anbieterzugang{Kennung: "k", Geheimnis: "g"}, nil
			},
			func(context.Context, string) error { return errors.New("kein Browser") },
			anmeldung.NeuerRufer())

		if _, fehler := naht.Anmelden(t.Context(), "github"); fehler == nil {
			t.Error("der Fehler ging verloren")
		}
	})
}

// **Ein Rückruf mit fremdem `state` wird nicht eingelöst.**
//
// # Das ist die Sperre, um die es geht
//
// Ohne sie nähme die Anwendung einen Code an, den jemand anderes untergeschoben
// hat — und meldete den Angreifer als den Nutzer an. Geprüft wird **an der
// Wirkung**: der Tauschendpunkt wird gar nicht erst gerufen.
//
// **Und er verbraucht nichts** (SEC-18, #2152): bis dahin beendete er die
// Anmeldung mit `ErrValidation`; jetzt nimmt ihn der Horcher gar nicht an,
// und die Anmeldung wartet weiter auf den richtigen — hier bis zu ihrer Frist.
func TestEinRueckrufMitFremdemStateWirdNichtEingeloest(t *testing.T) {
	t.Parallel()

	getauscht := false

	anbieter := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			getauscht = true

			schreiben.Header().Set("Content-Type", "application/json")

			if _, fehler := schreiben.Write([]byte(`{"access_token":"a"}`)); fehler != nil {
				t.Errorf("antworten: %v", fehler)
			}

			_ = anfrage
		}))
	t.Cleanup(anbieter.Close)

	dienst := &anmeldedienstprobe{}

	naht := app.NeueAnmeldenaht(dienst,
		func(context.Context, string) (app.Anbieterzugang, error) {
			return app.Anbieterzugang{Kennung: "k", Geheimnis: "g"}, nil
		},
		// **Der Browser schickt einen Rückruf mit fremdem `state`.**
		func(ctx context.Context, adresse string) error {
			return rufeZurueck(ctx, adresse, "ein fremder state")
		},
		anmeldung.RuferUeber(anbieter.Client())).MitAnbietern(gegen(anbieter.URL)).
		MitWartefrist(300 * time.Millisecond)

	if _, fehler := naht.Anmelden(t.Context(), "github"); !errors.Is(
		fehler, domain.ErrAnmeldungAbgelaufen) {
		t.Errorf("Fehler %v, erwartet war der Ablauf der Wartefrist", fehler)
	}

	if getauscht {
		t.Error("der Code wurde trotzdem eingelöst")
	}
}

// nahtGegen baut eine Naht gegen eine Gegenstelle, die sich wie GitHub verhält.
//
// **`vorher` läuft, bevor der Browser antwortet** — damit ein Test abbrechen
// kann, statt zu antworten.
func nahtGegen(t *testing.T, dienst *anmeldedienstprobe, vorher func()) *app.Anmeldenaht {
	t.Helper()

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

	return app.NeueAnmeldenaht(dienst,
		func(context.Context, string) (app.Anbieterzugang, error) {
			return app.Anbieterzugang{Kennung: "k", Geheimnis: "g"}, nil
		},
		func(ctx context.Context, adresse string) error {
			if vorher != nil {
				vorher()

				return nil
			}

			return rufeZurueck(ctx, adresse, "")
		},
		anmeldung.RuferUeber(anbieter.Client())).MitAnbietern(gegen(anbieter.URL))
}

// rufeZurueck spielt den Browser: es holt den `state` aus der Anmeldeadresse
// und schickt einen Rückruf an den Horcher.
//
// **Mit `state` aus der Adresse** — genau das täte ein Anbieter. Ein Test, der
// ihn erfände, prüfte die Sperre statt den Weg.
func rufeZurueck(ctx context.Context, adresse, state string) error {
	zerlegt, fehler := url.Parse(adresse)
	if fehler != nil {
		return fehler //nolint:wrapcheck // die Naht meldet ihn unverändert
	}

	werte := zerlegt.Query()

	gewaehlt := werte.Get("state")
	if state != "" {
		gewaehlt = state
	}

	rueckruf := werte.Get("redirect_uri") + "?code=c&state=" + url.QueryEscape(gewaehlt)

	anfrage, baufehler := http.NewRequestWithContext(ctx, http.MethodGet, rueckruf, http.NoBody)
	if baufehler != nil {
		return baufehler //nolint:wrapcheck // dito
	}

	antwort, ruffehler := http.DefaultClient.Do(anfrage)
	if ruffehler != nil {
		return ruffehler //nolint:wrapcheck // dito
	}

	return antwort.Body.Close() //nolint:wrapcheck // dito
}

// anmeldedienstprobe ist ein Anmeldedienst im Arbeitsspeicher.
type anmeldedienstprobe struct {
	nutzerID int64
	gesehen  domain.Fremdkennung
	fehler   error
}

func (p *anmeldedienstprobe) Herein(_ context.Context, kennung domain.Fremdkennung) (
	int64, error,
) {
	p.gesehen = kennung

	return p.nutzerID, p.fehler
}

// gegen gibt Baupläne, deren Endpunkte auf eine Gegenstelle zeigen.
//
// **Die Anmeldeadresse zeigt ebenfalls dorthin**, obwohl sie im Test nie
// abgerufen wird: der Browseröffner liest nur ihre Werte. Eine Adresse auf
// github.com stünde da wie eine, die jemand versehentlich aufmacht.
func gegen(ort string) map[string]func(kennung, geheimnis string) anmeldung.Anbieter {
	return map[string]func(kennung, geheimnis string) anmeldung.Anbieter{
		"github": func(kennung, geheimnis string) anmeldung.Anbieter {
			anbieter := anmeldung.GitHub(kennung, geheimnis)
			anbieter.Anmeldeadresse = ort + "/authorize"
			anbieter.Tauschadresse = ort + "/token"
			anbieter.Nutzeradresse = ort + "/user"

			return anbieter
		},
	}
}

// **Ein Tausch, der scheitert, beendet den Fluss.**
//
// **Und der Nutzer entsteht nicht**: ein Fluss, der nach einem gescheiterten
// Tausch trotzdem jemanden hereinließe, wäre die Anmeldung ohne Anmeldung.
func TestEinGescheiterterTauschBeendetDenFluss(t *testing.T) {
	t.Parallel()

	anbieter := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.Header().Set("Content-Type", "application/json")

			if _, fehler := schreiben.Write([]byte(`{"error":"bad_verification_code"}`)); fehler != nil {
				t.Errorf("antworten: %v", fehler)
			}
		}))
	t.Cleanup(anbieter.Close)

	dienst := &anmeldedienstprobe{nutzerID: 7}

	naht := app.NeueAnmeldenaht(dienst,
		func(context.Context, string) (app.Anbieterzugang, error) {
			return app.Anbieterzugang{Kennung: "k", Geheimnis: "g"}, nil
		},
		func(ctx context.Context, adresse string) error {
			return rufeZurueck(ctx, adresse, "")
		},
		anmeldung.RuferUeber(anbieter.Client())).MitAnbietern(gegen(anbieter.URL))

	if _, fehler := naht.Anmelden(t.Context(), "github"); fehler == nil {
		t.Fatal("der gescheiterte Tausch kam durch")
	}

	if dienst.gesehen.Subjekt != "" {
		t.Errorf("der Dienst sah trotzdem %+v", dienst.gesehen)
	}
}
