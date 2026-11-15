package rest_test

// **Was an den Token- und Vorgabenwegen schiefgehen kann** — Werkstücke 10 und
// 15 von Phase 4.12
// (#1351).
//
// **Der Klassenlauf über `alleEndpunkte` prüft vier Klassen je Weg** — er
// prüft **nicht** den krummen Rumpf: er schickt jedem Weg den Rumpf, der zu ihm
// gehört. Was hier steht, ist der Fall, den er auslässt.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// **Ein krummer Rumpf ist ein Nein und kein halbes Schreiben.**
func TestEinKrummerRumpfWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	wege := map[string]struct{ methode, weg string }{
		"Token anlegen": {http.MethodPost, "/profile/1/token"},
		"Token ändern":  {http.MethodPut, "/token/1"},
		"Vorgabe setzen": {
			http.MethodPut, "/profile/1/vorgaben/deck",
		},
		"Freigabe setzen": {
			http.MethodPut, "/profile/1/freigaben/2",
		},
	}

	for name, fall := range wege {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			weg, zeuge := mitWache(t, nil)

			anfrage := httptest.NewRequestWithContext(
				t.Context(), fall.methode, fall.weg,
				strings.NewReader("{"))
			anfrage.Host = "127.0.0.1:7777"
			anfrage.Header.Set("Authorization", "Bearer "+probemarke)

			antwort := httptest.NewRecorder()
			weg.ServeHTTP(antwort, anfrage)

			if antwort.Code != http.StatusBadRequest {
				t.Errorf("Status = %d, erwartet 400", antwort.Code)
			}

			// **Und der Anwendungsfall bleibt ungerufen** — ein halbes
			// Schreiben ist schlimmer als gar keines.
			if zeuge.rufe != 0 {
				t.Errorf("der Weg dahinter wurde %dmal gerufen", zeuge.rufe)
			}
		})
	}
}

// **Ein `PUT` ohne `werte` räumt ab.**
//
// **Es als Auslassung zu lesen hieße, einen Ruf ohne Wirkung zuzulassen** — und
// der ist bei einem Ort, der die Werte **ist**, keine sinnvolle Anfrage.
func TestEinPutOhneWerteRaeumtAb(t *testing.T) {
	t.Parallel()

	zeuge := &zaehlender{}
	weg := dienstMit(t, zeuge).Behandler()

	anfrage := httptest.NewRequestWithContext(
		t.Context(), http.MethodPut, "/profile/1/vorgaben/deck",
		strings.NewReader(`{}`))
	anfrage.Host = "127.0.0.1:7777"
	anfrage.Header.Set("Authorization", "Bearer "+probemarke)

	antwort := httptest.NewRecorder()
	weg.ServeHTTP(antwort, anfrage)

	if antwort.Code != http.StatusOK {
		t.Fatalf("Status = %d — %s", antwort.Code, antwort.Body.String())
	}

	if zeuge.werte == nil {
		t.Error("unten kam `nil` an statt einer leeren Liste")
	}

	if len(zeuge.werte) != 0 {
		t.Errorf("unten kamen %v an", zeuge.werte)
	}
}

// **Und eine zu knappe Erlaubnis wird gedrosselt wie jede Arbeit.**
//
// **Aus dem eigenen Budget und nicht aus dem der Fremden**: wer ein gültiges
// Token hat, ist kein Fremder — sein Fehlschlag zählt dort, wo seine Arbeit
// zählt. Ohne diesen Lauf bliebe der Zweig unbetreten (§3).
func TestEineZuKnappeErlaubnisGehtAufsEigeneBudget(t *testing.T) {
	t.Parallel()

	wache := &tokenwache{gilt: "mse_3_ABCDEFGH", profilID: 42, nurLesen: true}
	weg := mitHaushalt(t, dienstMit(t, &zaehlender{}).MitTokenwache(wache), engeGrenzen(), festeUhr).Behandler()

	// **Das Guthaben sind drei Einheiten** (`engeGrenzen`). Der vierte Ruf
	// muss das Guthaben treffen — und zwar mit `429`, nicht mit `403`.
	var letzter int

	for range 4 {
		antwort := httptest.NewRecorder()
		weg.ServeHTTP(antwort, anfrageMit(t, http.MethodPost,
			"Bearer mse_3_ABCDEFGH"))

		letzter = antwort.Code
	}

	if letzter != http.StatusTooManyRequests {
		t.Errorf("der vierte Ruf gab %d, erwartet 429", letzter)
	}
}
