// **Setangaben und Symbologie** — die beiden anderen Wege der 10/s-Klasse
// (#888).
//
// **Eine eigene Datei, weil `weitere_test.go` mit ihr über 400 Zeilen ginge**
// (§2). Der Schnitt folgt der Frage: dort stehen die Regelhinweise und die
// Drosseln, hier die beiden Wege, die eine Vorlage beschreiben statt eine
// Karte.
package scryfall_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

const setrumpf = `{"object":"set","code":"prf","name":"Prüfsatz",
	"set_type":"expansion","released_at":"2026-08-24","card_count":3,
	"digital":false,"scryfall_uri":"https://beispiel.invalid/sets/prf"}`

const symbolrumpf = `{"object":"list","data":[
	{"object":"card_symbol","symbol":"{W}","english":"one white mana",
	 "represents_mana":true,"mana_value":1,"appears_in_mana_costs":true,"colors":["W"]}]}`

func TestKlientSetangaben(t *testing.T) {
	t.Parallel()

	t.Run("holt und liest", func(t *testing.T) {
		t.Parallel()

		server, _ := gegenstelleMit(t, setrumpf)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		satz, _, befunde, fehler := klient.Setangaben(t.Context(), "prf")
		if fehler != nil {
			t.Fatalf("Setangaben: %v", fehler)
		}

		if satz.Code != "prf" || satz.Karten != 3 || len(befunde) != 0 {
			t.Errorf("Satz = %+v, Befunde = %+v", satz, befunde)
		}
	})

	t.Run("führt eine unbekannte Satzart als Befund", func(t *testing.T) {
		t.Parallel()

		server, _ := gegenstelleMit(t,
			`{"object":"set","code":"neu","name":"Neuartig","set_type":"holodeck"}`)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		satz, _, befunde, fehler := klient.Setangaben(t.Context(), "neu")
		if fehler != nil {
			t.Fatalf("Setangaben: %v", fehler)
		}

		if satz.Art != "holodeck" {
			t.Errorf("die Art wurde verändert: %q", satz.Art)
		}

		if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "holodeck") {
			t.Errorf("Befunde = %+v", befunde)
		}
	})

	t.Run("weist eine Antwort ohne Kürzel ab", func(t *testing.T) {
		t.Parallel()

		server, _ := gegenstelleMit(t, `{"object":"set","name":"Ohne Kürzel"}`)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		if _, _, _, fehler := klient.Setangaben(t.Context(), "x"); !errors.Is(
			fehler, domain.ErrValidation) {
			t.Fatalf("Fehler = %v", fehler)
		}
	})

	t.Run("meldet einen unlesbaren Rumpf", func(t *testing.T) {
		t.Parallel()

		server, _ := gegenstelleMit(t, `{kein json`)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		if _, _, _, fehler := klient.Setangaben(t.Context(), "x"); fehler == nil {
			t.Fatal("der unlesbare Rumpf kam durch")
		}
	})

	t.Run("meldet einen Ausfall der Gegenstelle", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(
			func(schreiben http.ResponseWriter, _ *http.Request) {
				schreiben.WriteHeader(http.StatusInternalServerError)
			}))
		t.Cleanup(server.Close)

		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		if _, _, _, fehler := klient.Setangaben(t.Context(), "x"); fehler == nil {
			t.Fatal("der Ausfall kam nicht durch")
		}
	})
}

func TestKlientSymbologie(t *testing.T) {
	t.Parallel()

	t.Run("holt und liest", func(t *testing.T) {
		t.Parallel()

		server, _ := gegenstelleMit(t, symbolrumpf)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		symbole, _, befunde, fehler := klient.Symbologie(t.Context())
		if fehler != nil {
			t.Fatalf("Symbologie: %v", fehler)
		}

		if len(symbole) != 1 || symbole[0].Zeichen != "{W}" || len(befunde) != 0 {
			t.Errorf("Symbole = %+v, Befunde = %+v", symbole, befunde)
		}
	})

	t.Run("führt eine unbekannte Farbe als Befund", func(t *testing.T) {
		t.Parallel()

		server, _ := gegenstelleMit(t, `{"object":"list","data":[
			{"object":"card_symbol","symbol":"{P}","english":"purple","colors":["P"]}]}`)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		symbole, _, befunde, fehler := klient.Symbologie(t.Context())
		if fehler != nil {
			t.Fatalf("Symbologie: %v", fehler)
		}

		if len(symbole) != 1 || len(befunde) != 1 ||
			!strings.Contains(befunde[0].Meldung, "\"P\"") {
			t.Errorf("Symbole = %+v, Befunde = %+v", symbole, befunde)
		}
	})

	t.Run("weist ein Symbol ohne Schreibweise ab", func(t *testing.T) {
		t.Parallel()

		server, _ := gegenstelleMit(t,
			`{"object":"list","data":[{"object":"card_symbol","english":"was?"}]}`)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		if _, _, _, fehler := klient.Symbologie(t.Context()); !errors.Is(
			fehler, domain.ErrValidation) {
			t.Fatalf("Fehler = %v", fehler)
		}
	})

	t.Run("meldet einen unlesbaren Rumpf", func(t *testing.T) {
		t.Parallel()

		server, _ := gegenstelleMit(t, `[kein objekt]`)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		if _, _, _, fehler := klient.Symbologie(t.Context()); fehler == nil {
			t.Fatal("der unlesbare Rumpf kam durch")
		}
	})

	t.Run("meldet einen Ausfall der Gegenstelle", func(t *testing.T) {
		t.Parallel()

		klient := klientFuer(t, "http://unerreichbar.invalid",
			netz.NeuerKlient(netz.UeblicheFristen()))

		if _, _, _, fehler := klient.Symbologie(t.Context()); fehler == nil {
			t.Fatal("der Ausfall kam nicht durch")
		}
	})
}

// **Und die Regelhinweise melden ihren Ausfall ebenso.**
func TestRulingsMeldenIhrenAusfall(t *testing.T) {
	t.Parallel()

	klient := klientFuer(t, "http://unerreichbar.invalid",
		netz.NeuerKlient(netz.UeblicheFristen()))

	if _, _, _, fehler := klient.Rulings(t.Context(), "x"); fehler == nil {
		t.Fatal("der Ausfall kam nicht durch")
	}
}

// **Ein unlesbarer Rumpf bei den Regelhinweisen kommt als Fehler.**
func TestRulingsMeldenEinenUnlesbarenRumpf(t *testing.T) {
	t.Parallel()

	server, _ := gegenstelleMit(t, `{kein json`)
	klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

	if _, _, _, fehler := klient.Rulings(t.Context(), "x"); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("Fehler = %v", fehler)
	}
}
