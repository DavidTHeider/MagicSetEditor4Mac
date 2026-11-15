package anthropic_test

// **Was ein Fehler von Anthropic heißt** — Werkstück A8 von Phase 5.3
// (#1447).

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anthropic"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jeder Status bekommt seine Fehlerart.**
//
// **Nachgelesen am 2026-09-11** unter
// <https://platform.claude.com/docs/en/api/errors>.
func TestJederStatusBekommtSeineFehlerart(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		status   int
		rumpf    string
		erwartet error
	}{
		"unbrauchbare Anfrage": {
			400, `{"error":{"type":"invalid_request_error","message":"x"}}`,
			domain.ErrValidation,
		},
		"falscher Schlüssel": {
			401, `{"error":{"type":"authentication_error","message":"x"}}`,
			domain.ErrErlaubnis,
		},
		"offene Rechnung": {
			402, `{"error":{"type":"billing_error","message":"x"}}`,
			domain.ErrErlaubnis,
		},
		"kein Zugriff": {
			403, `{"error":{"type":"permission_error","message":"x"}}`,
			domain.ErrErlaubnis,
		},
		"gibt es nicht": {
			404, `{"error":{"type":"not_found_error","message":"x"}}`,
			domain.ErrNotFound,
		},
		"Widerspruch": {
			409, `{"error":{"type":"conflict_error","message":"x"}}`,
			domain.ErrConflict,
		},
		"zu groß": {
			413, `{"error":{"type":"request_too_large","message":"x"}}`,
			domain.ErrValidation,
		},
		"gedrosselt": {
			429, `{"error":{"type":"rate_limit_error","message":"x"}}`,
			domain.ErrConflict,
		},
		"Ausfall": {
			500, `{"error":{"type":"api_error","message":"x"}}`,
			domain.ErrValidation,
		},
		"überlastet": {
			529, `{"error":{"type":"overloaded_error","message":"x"}}`,
			domain.ErrValidation,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := redeGegen(t, fall.status, fall.rumpf, nil)

			if !errors.Is(fehler, fall.erwartet) {
				t.Errorf("%d gibt %v, erwartet %v", fall.status, fehler, fall.erwartet)
			}
		})
	}
}

// **Die Sperre am Ausgabendeckel ist keine Drossel.**
//
// # Warum das der teuerste Fund der Recherche war
//
// **Beide sind 429 und beide sagen `rate_limit_error`.** Die Dokumentation
// sagt am 2026-09-11 den Unterschied: „A tier spend-cap 429 has no
// `retry-after` header and **keeps failing until access resumes**."
//
// Ein blindes Wiederholen liefe dort ins Leere — Minuten lang, und dann wäre
// immer noch Monatsanfang nötig. Erkannt wird es an
// `error.details.error_code`.
func TestDieAusgabensperreIstKeineDrossel(t *testing.T) {
	t.Parallel()

	fehler := redeGegen(t, 429, `{"error":{"type":"rate_limit_error",
		"message":"You have reached your API usage limits",
		"details":{"error_code":"enforced_spend_limit_reached"}}}`, nil)

	// **Erlaubnis und nicht Konflikt**: ein Konflikt sagt „gleich noch
	// einmal", und genau das ist hier falsch.
	if !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	// **Die Gegenprobe**: dasselbe ohne den Schlüssel bleibt eine Drossel.
	if ohne := redeGegen(t, 429,
		`{"error":{"type":"rate_limit_error","message":"x"}}`, nil); !errors.Is(
		ohne, domain.ErrConflict) {
		t.Errorf("die gewöhnliche Drossel gibt %v", ohne)
	}
}

// **Die Vorgangskennung reist mit — aus dem Rumpf oder der Kopfzeile.**
//
// **Beide Wege, weil beide vorkommen**: der Fehlerrumpf trägt `request_id`,
// jede Antwort die Kopfzeile `request-id`. Wer nur eine läse, stünde bei
// einem unlesbaren Rumpf ohne die Kennung da, die der Support verlangt.
func TestDieVorgangskennungReistMit(t *testing.T) {
	t.Parallel()

	ausRumpf := redeGegen(t, 400, `{"error":{"type":"invalid_request_error",
		"message":"x"},"request_id":"req_ausdemrumpf"}`, nil)

	if !strings.Contains(grundtext(ausRumpf), "req_ausdemrumpf") {
		t.Errorf("der Grund ist %q", grundtext(ausRumpf))
	}

	ausKopf := redeGegen(t, 400, `kein JSON`,
		http.Header{"Request-Id": {"req_ausdemkopf"}})

	if !strings.Contains(grundtext(ausKopf), "req_ausdemkopf") {
		t.Errorf("der Grund ist %q", grundtext(ausKopf))
	}

	// **Und ohne beides steht, dass es sie nicht gibt** — statt einer Lücke.
	ohne := redeGegen(t, 400, `kein JSON`, nil)

	if !strings.Contains(grundtext(ohne), "ohne Vorgangskennung") {
		t.Errorf("der Grund ist %q", grundtext(ohne))
	}
}

// **Die Meldung des Dienstes reist als Grund und nicht als Satz.**
//
// **§8**: sie ist englisch, nennt Kontingente und Fristen und gehört ins
// Protokoll — nicht in ein Fenster.
func TestDieMeldungDesDienstesReistAlsGrund(t *testing.T) {
	t.Parallel()

	fehler := redeGegen(t, 400, `{"error":{"type":"invalid_request_error",
		"message":"max_tokens: Field required"}}`, nil)

	// **Der Grund steht im Protokoll und nicht im Satz** ([domain.GrundVon]).
	if !strings.Contains(grundtext(fehler), "max_tokens: Field required") {
		t.Errorf("der Grund fehlt: %q", grundtext(fehler))
	}

	// **Und der Satz für Menschen ist deutsch und nennt die Meldung nicht.**
	if !strings.Contains(fehler.Error(), "abgelehnt") {
		t.Errorf("der Satz fehlt: %v", fehler)
	}

	if strings.Contains(fehler.Error(), "Field required") {
		t.Errorf("die fremde Meldung steht im Satz: %v", fehler)
	}
}

// redeGegen stellt eine Frage an eine Gegenstelle, die ablehnt.
func redeGegen(t *testing.T, status int, rumpf string, kopf http.Header) error {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			for name, werte := range kopf {
				schreiben.Header()[name] = werte
			}

			schreiben.WriteHeader(status)
			_, _ = schreiben.Write([]byte(rumpf)) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	_, fehler := klientFuer(server).Rede(t.Context(), anthropic.Frage{
		Alias: "claude-opus-5", Auftrag: "x",
	})

	if fehler == nil {
		t.Fatalf("Status %d kam ohne Fehler durch", status)
	}

	return fehler
}

// grundtext gibt die verdeckte Ursache als Text — oder einen Hinweis darauf,
// dass es keine gibt.
func grundtext(fehler error) string {
	grund := domain.GrundVon(fehler)
	if grund == nil {
		return "ohne Grund: " + fehler.Error()
	}

	return grund.Error()
}

// **Und die selbst gesetzte Ausgabengrenze ist keine unbrauchbare Anfrage.**
//
// # Warum sie hier eigens steht
//
// **Nachgelesen am 2026-09-12** unter
// <https://platform.claude.com/docs/en/api/rate-limits>: wer sich im Konto
// eine eigene Grenze unterhalb des Tarifdeckels setzt, bekommt beim
// Erreichen **400** mit `invalid_request_error` — nicht 429. Die Meldung
// beginnt mit „You have reached your specified API usage limits", bei einer
// Arbeitsfläche mit „…your specified workspace API usage limits".
//
// **Derselbe Status wie ein Tippfehler im Schema**, und eine völlig andere
// Abhilfe: nicht „berichtige die Anfrage", sondern „hebe die Grenze an oder
// warte auf den Monatsanfang". Wer beides als [domain.ErrValidation] zeigte,
// schickte jemanden auf die Suche nach einem Fehler, den es nicht gibt (§8).
func TestDieEigeneAusgabengrenzeIstKeineUnbrauchbareAnfrage(t *testing.T) {
	t.Parallel()

	for name, meldung := range map[string]string{
		"für das Konto": "You have reached your specified API usage limits. " +
			"Access resumes on 2026-10-01.",
		"für die Arbeitsfläche": "You have reached your specified workspace " +
			"API usage limits. Access resumes on 2026-10-01.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := redeGegen(t, 400, `{"error":{"type":"invalid_request_error",
				"message":"`+meldung+`"}}`, nil)

			if !errors.Is(fehler, domain.ErrErlaubnis) {
				t.Fatalf("der Fehler ist %v", fehler)
			}
		})
	}

	// **Die Gegenprobe**: ein gewöhnlicher 400 bleibt eine Validierung.
	if ohne := redeGegen(t, 400,
		`{"error":{"type":"invalid_request_error","message":"max_tokens: required"}}`,
		nil); !errors.Is(ohne, domain.ErrValidation) {
		t.Errorf("die gewöhnliche unbrauchbare Anfrage gibt %v", ohne)
	}
}
