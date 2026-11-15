package anmeldung_test

// **Die Fehlerwege des Tauschs** — abgetrennt von `tausch_test.go`, als die
// Datei über die Zeilengrenze aus §2 kam.
//
// **Nach der Frage geschnitten**: nebenan steht, was **glückt**; hier steht,
// was dabei schiefgehen kann — eine abgeschnittene Antwort, ein Fehlerstatus
// ohne Fehler im Rumpf.

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anmeldung"
)

// **Jeder Fehlerweg des Tauschs meldet sich.**
func TestJederFehlerwegDesTauschsMeldetSich(t *testing.T) {
	t.Parallel()

	faelle := map[string]http.HandlerFunc{
		"der Tausch antwortet mit einem Fehler": func(s http.ResponseWriter, _ *http.Request) {
			s.WriteHeader(http.StatusBadRequest)
			antworte(t, s, `{"error":"invalid_grant"}`)
		},
		"die Antwort ist kein JSON": func(s http.ResponseWriter, _ *http.Request) {
			antworte(t, s, `kein JSON`)
		},
		"die Antwort trägt kein Zugangswort": func(s http.ResponseWriter, _ *http.Request) {
			antworte(t, s, `{}`)
		},
	}

	for name, behandler := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gegenstelle := httptest.NewServer(behandler)
			t.Cleanup(gegenstelle.Close)

			anbieter := anmeldung.GitHub("kennung", "geheim")
			anbieter.Tauschadresse = gegenstelle.URL + "/token"
			anbieter.Nutzeradresse = gegenstelle.URL + "/user"

			lauf := mussBeginnen(t, anbieter)

			if _, fehler := lauf.Einloesen(
				t.Context(), anmeldung.RuferUeber(gegenstelle.Client()), "c"); fehler == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Und jeder des zweiten Rufs.**
func TestJederFehlerwegDesZweitenRufsMeldetSich(t *testing.T) {
	t.Parallel()

	faelle := map[string]http.HandlerFunc{
		"die Nutzerabfrage klagt": func(s http.ResponseWriter, _ *http.Request) {
			s.WriteHeader(http.StatusUnauthorized)
		},
		"ihre Antwort ist kein JSON": func(s http.ResponseWriter, _ *http.Request) {
			antworte(t, s, `kein JSON`)
		},
		"sie nennt keine Kennung": func(s http.ResponseWriter, _ *http.Request) {
			antworte(t, s, `{"login":"wer"}`)
		},
	}

	for name, behandler := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gegenstelle := httptest.NewServer(http.HandlerFunc(
				func(schreiben http.ResponseWriter, anfrage *http.Request) {
					if strings.HasSuffix(anfrage.URL.Path, "/token") {
						antworte(t, schreiben, `{"access_token":"abc"}`)

						return
					}

					behandler(schreiben, anfrage)
				}))
			t.Cleanup(gegenstelle.Close)

			anbieter := anmeldung.GitHub("kennung", "geheim")
			anbieter.Tauschadresse = gegenstelle.URL + "/token"
			anbieter.Nutzeradresse = gegenstelle.URL + "/user"

			lauf := mussBeginnen(t, anbieter)

			if _, fehler := lauf.Einloesen(
				t.Context(), anmeldung.RuferUeber(gegenstelle.Client()), "c"); fehler == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Ein `id_token`, das keines ist, wird abgewiesen.**
func TestEinUnbrauchbaresIDTokenWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"gar keines":      `{"access_token":"abc"}`,
		"zu wenige Teile": `{"id_token":"eins.zwei"}`,
		"kein base64":     `{"id_token":"a.!!!.c"}`,
		"kein JSON darin": `{"id_token":"a.` + base64.RawURLEncoding.EncodeToString([]byte("kein JSON")) + `.c"}`,
		"ohne sub":        `{"id_token":"a.` + base64.RawURLEncoding.EncodeToString([]byte(`{"name":"wer"}`)) + `.c"}`,
	}

	for name, antwort := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gegenstelle := httptest.NewServer(http.HandlerFunc(
				func(schreiben http.ResponseWriter, _ *http.Request) {
					antworte(t, schreiben, antwort)
				}))
			t.Cleanup(gegenstelle.Close)

			anbieter := anmeldung.Google("kennung", "")
			anbieter.Tauschadresse = gegenstelle.URL + "/token"

			lauf := mussBeginnen(t, anbieter)

			if _, fehler := lauf.Einloesen(
				t.Context(), anmeldung.RuferUeber(gegenstelle.Client()), "c"); fehler == nil {
				t.Error("das unbrauchbare Token kam durch")
			}
		})
	}
}
