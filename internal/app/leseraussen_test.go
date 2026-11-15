//go:build darwin

// **Ein Leser schickt nichts nach draußen** — Werkstück O1 von Phase 5.2
// (#1425),
// nach [0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md).
//
// **Gemessen am Server, nicht am Fehler**: ein Nein, das nach der Anfrage
// käme, wäre keins — der Name der Karte wäre schon unterwegs.

package app_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestEinLeserSchicktNichtsNachDraussen(t *testing.T) {
	var anfragen atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiber http.ResponseWriter, _ *http.Request) {
			anfragen.Add(1)
			schreiber.Header().Set("Content-Type", "application/json")
			schreiben(t, schreiber, `{"object":"list","total_cards":0,"has_more":false,"data":[]}`)
		}))
	defer server.Close()

	vorher := app.MitScryfalladresse(server.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	anwendung := alsLeserHinein(t)

	gesucht := anwendung.ScryfallSuche(t.Context(), "Lightning Bolt")
	if gesucht.Fehler == nil {
		t.Fatal("der Leser suchte bei Scryfall")
	}

	if gesucht.Fehler.Code != app.CodeNurLesend {
		t.Errorf("Code %q, erwartet %q — %s", gesucht.Fehler.Code, app.CodeNurLesend, gesucht.Fehler.Meldung)
	}

	if gezaehlt := anfragen.Load(); gezaehlt != 0 {
		t.Errorf("der Server sah %d Anfragen, erwartet keine", gezaehlt)
	}

	// **Gegenprobe im eigenen Profil**: dort geht die Anfrage hinaus — sonst
	// bewiese der Lauf einen Riegel, der alles sperrt (§10).
	ich := anwendung.WerArbeitet(t.Context())
	if ich.Fehler != nil {
		t.Fatalf("wer arbeitet: %+v", ich.Fehler)
	}

	for _, nutzer := range anwendung.Nutzer(t.Context()).Nutzer {
		if nutzer.ID == ich.NutzerID {
			if zurueck := anwendung.ProfilWechseln(t.Context(), nutzer.Profile[0].ID, ""); zurueck.Fehler != nil {
				t.Fatalf("zurück: %+v", zurueck.Fehler)
			}
		}
	}

	if wieder := anwendung.ScryfallSuche(t.Context(), "Lightning Bolt"); wieder.Fehler != nil {
		t.Fatalf("im eigenen Profil: %+v", wieder.Fehler)
	}

	if gezaehlt := anfragen.Load(); gezaehlt != 1 {
		t.Errorf("der Server sah %d Anfragen, erwartet genau eine", gezaehlt)
	}
}

// schreiben schreibt eine Antwort und meldet, wenn das nicht geht.
func schreiben(t *testing.T, schreiber http.ResponseWriter, rumpf string) {
	t.Helper()

	if _, err := schreiber.Write([]byte(rumpf)); err != nil {
		t.Errorf("antworten: %v", err)
	}
}
