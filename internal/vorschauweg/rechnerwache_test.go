package vorschauweg_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// **Die Wache vor dem Servermodus** — QA-09
// (#1740),
// Befund P1 in `docs/PENTEST.md`.
//
// # Wovor sie schützt
//
// `make dev-browser` stellt **dieselbe Anwendung mit echten Daten** unter
// `http://localhost:34115` bereit, und Wails' HTTP-Weg nimmt jeden Ruf an:
// kein `Origin`, kein `Host`, nicht einmal eine Methode — ein GET mit
// `?object=0&method=0&args=…` ruft jede Bindung (`transport_http.go`,
// v3.0.0-beta.23, Rückfall für WebKitGTK). Drei Angriffe aus einer
// beliebigen Seite im Browser desselben Rechners:
//
//   - **DNS-Rebinding**: der Name des Angreifers zeigt auf `127.0.0.1`, der
//     Browser hält die Seite für gleichherkünftig — und liest jede Antwort.
//   - **Ein POST von fremder Herkunft**: kein Vorabruf, weil Wails keinen
//     Inhaltstyp verlangt; die Antwort bleibt verborgen, die Wirkung nicht.
//   - **Ein Bild auf den Rufweg**: `<img src=…/wails/runtime?…>` schickt
//     weder `Origin` noch einen eigenen Kopf.
func TestDieRechnerwacheWeistFremdeAnfragenAb(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		methode, weg, rechner string
		koepfe                map[string]string
		erwartet              int
	}{
		"Rebinding: Ruf mit dem Namen des Angreifers": {
			http.MethodPost, "/wails/runtime", "angreifer.example:34115",
			map[string]string{"Origin": "http://angreifer.example:34115"},
			http.StatusForbidden,
		},
		"Rebinding: Bild ohne Origin": {
			http.MethodGet, vorschauweg.Pfad + "1.png", "angreifer.example:34115",
			nil, http.StatusForbidden,
		},
		"Rebinding: Name ohne Port": {
			http.MethodGet, "/", "angreifer.example", nil, http.StatusForbidden,
		},
		"fremde Seite schickt POST": {
			http.MethodPost, "/wails/runtime", "localhost:34115",
			map[string]string{"Origin": "https://angreifer.example"},
			http.StatusForbidden,
		},
		"Herkunft null aus einem Rahmen": {
			http.MethodPost, "/wails/runtime", "localhost:34115",
			map[string]string{"Origin": "null"},
			http.StatusForbidden,
		},
		"eigener Rechner, anderer Port": {
			http.MethodPost, "/wails/runtime", "localhost:34115",
			map[string]string{"Origin": "http://localhost:5173"},
			http.StatusForbidden,
		},
		"Bild einer fremden Seite": {
			http.MethodGet, vorschauweg.Pfad + "1.png", "localhost:34115",
			map[string]string{"Sec-Fetch-Site": "cross-site", "Sec-Fetch-Mode": "no-cors"},
			http.StatusForbidden,
		},
		"Ruf über GET ohne Sec-Fetch": {
			http.MethodGet, "/wails/runtime?object=0&method=0", "localhost:34115",
			nil, http.StatusMethodNotAllowed,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gerufen := false
			innen := func(weiter http.Handler) http.Handler {
				return http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
					gerufen = true

					weiter.ServeHTTP(schreiben, anfrage)
				})
			}

			anfrage := httptest.NewRequestWithContext(t.Context(), fall.methode, fall.weg, http.NoBody)
			anfrage.Host = fall.rechner

			for kopf, wert := range fall.koepfe {
				anfrage.Header.Set(kopf, wert)
			}

			antwort := httptest.NewRecorder()
			vorschauweg.NurVonDiesemRechner(innen)(dahinter()).ServeHTTP(antwort, anfrage)

			if antwort.Code != fall.erwartet {
				t.Errorf("Status %d, erwartet %d", antwort.Code, fall.erwartet)
			}

			if gerufen {
				t.Error("die Anfrage erreichte die Weiche dahinter")
			}
		})
	}
}

// **Und die eigene Seite kommt durch** — ohne diesen Lauf stünde die Wache
// auch dann grün da, wenn sie alles abwiese (§10).
func TestDieRechnerwacheLaesstDieEigeneSeiteDurch(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		methode, weg, rechner string
		koepfe                map[string]string
	}{
		"Ruf der eigenen Seite": {
			http.MethodPost, "/wails/runtime", "localhost:34115",
			map[string]string{"Origin": "http://localhost:34115", "Sec-Fetch-Site": "same-origin"},
		},
		"Ruf über 127.0.0.1": {
			http.MethodPost, "/wails/runtime", "127.0.0.1:34115",
			map[string]string{"Origin": "http://127.0.0.1:34115"},
		},
		"Ruf über ::1": {
			http.MethodPost, "/wails/runtime", "[::1]:34115",
			map[string]string{"Origin": "http://[::1]:34115"},
		},
		"Seite ohne Origin": {http.MethodGet, "/", "localhost:34115", nil},
		"getippte Adresse": {
			http.MethodGet, "/", "localhost:34115", map[string]string{"Sec-Fetch-Site": "none"},
		},
		"Verweis von einer fremden Seite öffnet die Anwendung": {
			http.MethodGet, "/", "localhost:34115",
			map[string]string{"Sec-Fetch-Site": "cross-site", "Sec-Fetch-Mode": "navigate"},
		},
		"eigenes Bild": {
			http.MethodGet, vorschauweg.Pfad + "1.png", "localhost",
			map[string]string{"Sec-Fetch-Site": "same-origin", "Sec-Fetch-Mode": "no-cors"},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gerufen := false
			innen := func(weiter http.Handler) http.Handler {
				return http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
					gerufen = true

					weiter.ServeHTTP(schreiben, anfrage)
				})
			}

			anfrage := httptest.NewRequestWithContext(t.Context(), fall.methode, fall.weg, http.NoBody)
			anfrage.Host = fall.rechner

			for kopf, wert := range fall.koepfe {
				anfrage.Header.Set(kopf, wert)
			}

			antwort := httptest.NewRecorder()
			vorschauweg.NurVonDiesemRechner(innen)(dahinter()).ServeHTTP(antwort, anfrage)

			if antwort.Code != http.StatusTeapot || !gerufen {
				t.Errorf("Status %d, gerufen %v — die eigene Seite kam nicht durch",
					antwort.Code, gerufen)
			}
		})
	}
}

// **Eine Absage sagt, woran es lag** — wer am Entwicklungsrechner eine 403
// sieht, soll nicht raten müssen, ob es der Name oder die Herkunft war.
func TestDieRechnerwacheNenntIhrenGrund(t *testing.T) {
	t.Parallel()

	anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody)
	anfrage.Host = "angreifer.example:34115"

	antwort := httptest.NewRecorder()
	vorschauweg.NurVonDiesemRechner(func(weiter http.Handler) http.Handler { return weiter })(
		dahinter()).ServeHTTP(antwort, anfrage)

	if !strings.Contains(antwort.Body.String(), "angreifer.example") {
		t.Errorf("die Absage nennt den Namen nicht: %q", antwort.Body.String())
	}
}

// dahinter ist der Weg hinter der Weiche — er antwortet mit einer Zahl, die
// sonst niemand schickt.
func dahinter() http.Handler {
	return http.HandlerFunc(func(schreiben http.ResponseWriter, _ *http.Request) {
		schreiben.WriteHeader(http.StatusTeapot)
	})
}
