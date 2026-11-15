package aktualisierung_test

// **Der Freigabekanal nach D01** — R01
// (#2333),
// Vermerk 0190: das Repositorium wird öffentlich, und dann fragt jeder ohne
// Marke. Ohne Marke gelten 60 Anfragen je Stunde, und GitHub meldet eine
// erschöpfte Grenze mit **403 oder 429** und `x-ratelimit-remaining: 0`
// (nachgelesen am 2026-10-13, „Rate limits for the REST API", Abschnitt
// „Exceeding the rate limit"). Ein 403 als Konflikt zu melden hieße, eine
// Wartezeit als Fehler dieser Anwendung auszugeben.
//
// # Was sie prüfen
//
// 404, 429, 403 mit erschöpfter Grenze und Zeitablauf je für sich — vier
// Antworten, vier Sätze, und keine davon ist „keine Freigabe". Dazu, dass
// der Klient die Quelltextbeigabe erkennt, die seit D01 jeder Freigabe
// beiliegt (`scripts/releasequelltext.sh`).

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kopfstelle antwortet mit Status, Kopfzeilen und Rumpf.
func kopfstelle(t *testing.T, status int, kopf map[string]string, rumpf ...string) *httptest.Server {
	t.Helper()

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			for name, wert := range kopf {
				schreiben.Header().Set(name, wert)
			}

			schreiben.WriteHeader(status)

			text := `{"message":"Forbidden"}`
			if len(rumpf) > 0 {
				text = rumpf[0]
			}

			if _, err := schreiben.Write([]byte(text)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(dienst.Close)

	return dienst
}

func TestEineErschoepfteGrenzeIstEineDrosselUndKeinKonflikt(t *testing.T) {
	t.Parallel()

	ruecksetzung := time.Now().Add(37 * time.Minute).Truncate(time.Second)

	for name, status := range map[string]int{"403": http.StatusForbidden, "429": http.StatusTooManyRequests} {
		t.Run("mit "+name, func(t *testing.T) {
			t.Parallel()

			dienst := kopfstelle(t, status, map[string]string{
				"X-Ratelimit-Remaining": "0",
				"X-Ratelimit-Reset":     strconv.FormatInt(ruecksetzung.Unix(), 10),
			})

			_, err := probeklient(t, dienst.URL).NeuesteFreigabe(t.Context())

			if !errors.Is(err, domain.ErrGedrosselt) {
				t.Fatalf("eine erschöpfte Grenze ist keine Drossel: %v", err)
			}

			if errors.Is(err, domain.ErrConflict) || errors.Is(err, domain.ErrNotFound) {
				t.Errorf("eine erschöpfte Grenze trägt eine falsche Klasse: %v", err)
			}
		})
	}
}

// ende liest das Ende einer Drosselung aus dem Fehler.
func ende(t *testing.T, err error) time.Time {
	t.Helper()

	var genauer *domain.Gegenstellenfehler
	if !errors.Is(err, domain.ErrGedrosselt) || !errors.As(err, &genauer) {
		t.Fatalf("keine Drosselung mit Ende: %v", err)
	}

	return genauer.Bis
}

// **403 und 429 sind dieselbe Grenze** (Befund 2 des Zweitblicks über
// Phase 0): bis dahin sperrte ein 429 nur die allgemeinen 30 Sekunden,
// ein 403 bis zur Rücksetzung.
func TestDieErschoepfteGrenzeEndetBei403Und429GleichAnDerRuecksetzung(t *testing.T) {
	t.Parallel()

	ruecksetzung := time.Now().Add(37 * time.Minute).Truncate(time.Second)

	for name, status := range map[string]int{"403": http.StatusForbidden, "429": http.StatusTooManyRequests} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dienst := kopfstelle(t, status, map[string]string{
				"X-Ratelimit-Remaining": "0",
				"X-Ratelimit-Reset":     strconv.FormatInt(ruecksetzung.Unix(), 10),
			})

			_, err := probeklient(t, dienst.URL).NeuesteFreigabe(t.Context())

			if bis := ende(t, err); !bis.Equal(ruecksetzung) {
				t.Errorf("Ende %v, erwartet %v", bis, ruecksetzung)
			}
		})
	}
}

// **Eine zweite Grenze nennt ihre eigene Wartezeit** — `retry-after` gilt
// vor einem `reset`, solange die erste Grenze nicht erschöpft ist.
func TestEineZweiteGrenzeWartetNachRetryAfter(t *testing.T) {
	t.Parallel()

	dienst := kopfstelle(t, http.StatusForbidden, map[string]string{
		"X-Ratelimit-Remaining": "41",
		"X-Ratelimit-Reset":     strconv.FormatInt(time.Now().Add(37*time.Minute).Unix(), 10),
		"Retry-After":           "120",
	})

	_, err := probeklient(t, dienst.URL).NeuesteFreigabe(t.Context())

	if bis := ende(t, err); bis.Sub(time.Now().Add(120*time.Second)).Abs() > 5*time.Second {
		t.Errorf("Ende %v, erwartet in 120 Sekunden", bis)
	}
}

// **Ohne Zeitangabe wenigstens eine Minute** — „wait for at least one
// minute before retrying“ (nachgelesen am 2026-10-13).
func TestEineGrenzeOhneZeitangabeWartetEineMinute(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		status int
		kopf   map[string]string
		rumpf  string
	}{
		"erschöpft ohne reset":   {http.StatusForbidden, map[string]string{"X-Ratelimit-Remaining": "0"}, ""},
		"zweite Grenze im Rumpf": {http.StatusForbidden, nil, `{"message":"You have exceeded a secondary rate limit."}`},
		"ein nacktes 429":        {http.StatusTooManyRequests, nil, ""},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dienst := kopfstelle(t, fall.status, fall.kopf, fall.rumpf)

			_, err := probeklient(t, dienst.URL).NeuesteFreigabe(t.Context())

			if bis := ende(t, err); bis.Sub(time.Now().Add(time.Minute)).Abs() > 5*time.Second {
				t.Errorf("Ende %v, erwartet in einer Minute", bis)
			}
		})
	}
}

func TestEinVerbotOhneErschoepfteGrenzeBleibtEinVerbot(t *testing.T) {
	t.Parallel()

	dienst := kopfstelle(t, http.StatusForbidden, map[string]string{"X-Ratelimit-Remaining": "41"})

	_, err := probeklient(t, dienst.URL).NeuesteFreigabe(t.Context())

	if errors.Is(err, domain.ErrGedrosselt) || !strings.Contains(err.Error(), "403") {
		t.Errorf("ein 403 mit verbleibenden Anfragen ist keine Drossel und nennt seine Zahl: %v", err)
	}
}

func TestEinZeitablaufIstKeinNein(t *testing.T) {
	t.Parallel()

	// **Die Antwort kommt nie**, bis der Klient aufgibt — der Dienst wartet
	// auf den Abbruch der Anfrage und hält beim Schließen nichts fest.
	dienst := httptest.NewServer(http.HandlerFunc(
		func(_ http.ResponseWriter, anfrage *http.Request) {
			select {
			case <-anfrage.Context().Done():
			case <-time.After(5 * time.Second):
			}
		}))
	t.Cleanup(dienst.Close)

	ctx, abbrechen := context.WithTimeout(t.Context(), 200*time.Millisecond)
	defer abbrechen()

	_, err := probeklient(t, dienst.URL).NeuesteFreigabe(ctx)

	if err == nil || errors.Is(err, domain.ErrNotFound) || strings.Contains(err.Error(), "keine Freigabe") {
		t.Errorf("ein Zeitablauf ist kein fachliches Nein: %v", err)
	}
}

func TestDieQuelltextbeigabeWirdErkannt(t *testing.T) {
	t.Parallel()

	rumpf := strings.Replace(echteAntwort, `"assets": [`, `"assets": [
    {"name": "MagicSetEditor4Mac-0.9.0-quelltext.tar.gz", "size": 31337,
     "url": "https://api.github.com/repos/x/y/releases/assets/7"},
    {"name": "etwas-anderes.tar.gz", "size": 1,
     "url": "https://api.github.com/repos/x/y/releases/assets/8"},`, 1)
	dienst, _ := gegenstelle(t, rumpf, http.StatusOK)

	freigabe, err := probeklient(t, dienst.URL).NeuesteFreigabe(t.Context())
	if err != nil {
		t.Fatalf("holen: %v", err)
	}

	if freigabe.Quelltext.Name != "MagicSetEditor4Mac-0.9.0-quelltext.tar.gz" {
		t.Errorf("Quelltext = %q", freigabe.Quelltext.Name)
	}
}
