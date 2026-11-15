package openai_test

// **Drei Dinge tragen den Status 429, und nur eines geht vorüber** —
// Werkstück A9 von Phase 5.3
// (#1448).

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/openai"
)

func TestJederStatusBekommtSeineFehlerart(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		status   int
		rumpf    string
		erwartet error
	}{
		"unbrauchbare Anfrage": {
			400, `{"error":{"type":"invalid_request_error","code":"x",
			  "param":"text.format.schema","message":"y"}}`,
			domain.ErrValidation,
		},
		"falscher Schlüssel": {
			401, `{"error":{"type":"invalid_request_error",
			  "code":"invalid_api_key","message":"x"}}`,
			domain.ErrErlaubnis,
		},
		"gesperrte Region": {
			403, `{"error":{"type":"invalid_request_error",
			  "code":"unsupported_country_region_territory","message":"x"}}`,
			domain.ErrErlaubnis,
		},
		"gibt es nicht": {
			404, `{"error":{"type":"invalid_request_error",
			  "code":"model_not_found","message":"x"}}`,
			domain.ErrNotFound,
		},
		"Widerspruch": {
			409, `{"error":{"type":"invalid_request_error",
			  "code":"conflict","message":"x"}}`,
			domain.ErrConflict,
		},
		"gedrosselt": {
			429, `{"error":{"type":"rate_limit_error","code":"slow_down",
			  "message":"x"}}`,
			domain.ErrConflict,
		},
		"Ausfall": {
			500, `{"error":{"type":"server_error","code":"x","message":"y"}}`,
			domain.ErrValidation,
		},
		"überlastet": {
			503, `{"error":{"type":"server_error","code":"x","message":"y"}}`,
			domain.ErrValidation,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := fehlerGegen(t, fall.status, fall.rumpf); !errors.Is(
				fehler, fall.erwartet) {
				t.Errorf("%d gibt %v, erwartet %v", fall.status, fehler, fall.erwartet)
			}
		})
	}
}

// **Ein leeres Konto ist keine Drossel.**
//
// # Warum das der teuerste Fund der Recherche war
//
// **Drei Dinge tragen 429**, und die Dokumentation sagt am 2026-09-12 den
// Unterschied: „Retrying billing, spend, or quota errors won't restore API
// access." Ein blindes Wiederholen liefe dort ins Leere — Minuten lang, und
// danach fehlte immer noch Geld.
func TestEinLeeresKontoIstKeineDrossel(t *testing.T) {
	t.Parallel()

	for _, code := range []string{
		"credit_balance_exhausted",
		"organization_spend_limit_exceeded",
		"project_spend_limit_exceeded",
	} {
		t.Run(code, func(t *testing.T) {
			t.Parallel()

			fehler := fehlerGegen(t, 429, `{"error":{"type":"rate_limit_error",
			  "code":"`+code+`","message":"x"}}`)

			// **Erlaubnis und nicht Konflikt**: ein Konflikt sagt „gleich
			// noch einmal", und genau das ist hier falsch.
			if !errors.Is(fehler, domain.ErrErlaubnis) {
				t.Fatalf("der Fehler ist %v", fehler)
			}
		})
	}

	// **Die Gegenprobe**: ein unbekannter Schlüssel bleibt eine Drossel.
	//
	// Sie steht hier und nicht daneben, weil sie die Aufzählung prüft: wer
	// sie durch ein Präfix ersetzte, machte aus jedem künftigen Schlüssel
	// eine Erlaubnisfrage, ohne ihn gelesen zu haben.
	if ohne := fehlerGegen(t, 429,
		`{"error":{"type":"rate_limit_error","code":"etwasneues","message":"x"}}`,
	); !errors.Is(ohne, domain.ErrConflict) {
		t.Errorf("die gewöhnliche Drossel gibt %v", ohne)
	}
}

// **Die beanstandete Angabe reist mit.**
//
// **„Unbrauchbare Anfrage" sagt nichts**, `text.format.schema` sagt alles —
// und der Grund ist fürs Protokoll, nicht fürs Fenster (§8).
func TestDieBeanstandeteAngabeReistMit(t *testing.T) {
	t.Parallel()

	fehler := fehlerGegen(t, 400, `{"error":{"type":"invalid_request_error",
	  "code":"x","param":"text.format.schema","message":"y"}}`)

	grund := domain.GrundVon(fehler)
	if grund == nil || !strings.Contains(grund.Error(), "text.format.schema") {
		t.Errorf("der Grund nennt die Angabe nicht: %v", grund)
	}
}

// fehlerGegen fragt eine Gegenstelle, die ablehnt.
func fehlerGegen(t *testing.T, status int, rumpf string) error {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(status)
			_, _ = schreiben.Write([]byte(rumpf)) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	klient := openai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		openai.Zugang{Schluessel: "geheim"})

	_, fehler := klient.Rede(t.Context(), openai.Frage{Alias: "m", Auftrag: "x"})

	return fehler
}
