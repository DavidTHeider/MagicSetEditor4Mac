package app_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// auskunftsdoppel ist ein Doppel für [app.Auskunftsquelle].
type auskunftsdoppel struct {
	regeln    *service.Regelauskunft
	unbekannt []string
	fehler    error
	gefragt   string
}

func (a *auskunftsdoppel) Regelhinweise(_ context.Context, name string) (
	*service.Regelauskunft, error,
) {
	a.gefragt = name

	return a.regeln, a.fehler
}

func (a *auskunftsdoppel) Satz(context.Context, string) (*service.Satzauskunft, error) {
	return nil, a.fehler
}

func (a *auskunftsdoppel) Symbole(context.Context) (*service.Symbolauskunft, error) {
	return nil, a.fehler
}

func (a *auskunftsdoppel) Manakostenpruefung(_ context.Context, _ string) ([]string, error) {
	return a.unbekannt, a.fehler
}

// **Die Fassade rechnet »offiziell« aus, nicht das Fenster.**
//
// Ein `=== 'wotc'` in TypeScript wäre eine zweite Fassung derselben Regel, und
// sie liefe auseinander, sobald Scryfall einen dritten Wert einführt.
func TestScryfallRegeln(t *testing.T) {
	t.Parallel()

	t.Run("gibt die Hinweise samt Herkunftskennzeichen", func(t *testing.T) {
		t.Parallel()

		doppel := &auskunftsdoppel{regeln: &service.Regelauskunft{
			Kartenname: "Lightning Bolt",
			Hinweise: []scryfall.Regelhinweis{
				{Quelle: scryfall.QuelleWotc, Datum: "2004-10-04", Text: "Offiziell."},
				{Quelle: scryfall.QuelleScryfall, Datum: "2021-07-23", Text: "Anmerkung."},
			},
			Adresse: "https://scryfall.invalid/x",
			Befunde: []domain.Befund{{Meldung: "etwas Unbekanntes"}},
			Alter:   3 * time.Hour,
		}}

		anwendung := app.NeueApp("")
		anwendung.MitScryfallauskunft(doppel)

		antwort := anwendung.ScryfallRegeln(t.Context(), "lightning bolt")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if doppel.gefragt != "lightning bolt" {
			t.Errorf("gefragt wurde %q", doppel.gefragt)
		}

		if len(antwort.Hinweise) != 2 {
			t.Fatalf("%d Hinweise", len(antwort.Hinweise))
		}

		if !antwort.Hinweise[0].Offiziell || antwort.Hinweise[1].Offiziell {
			t.Errorf("Kennzeichen = %v / %v",
				antwort.Hinweise[0].Offiziell, antwort.Hinweise[1].Offiziell)
		}

		if antwort.Kartenname != "Lightning Bolt" || antwort.AlterStunden != 3 {
			t.Errorf("Antwort = %+v", antwort)
		}

		if len(antwort.Befunde) != 1 ||
			!strings.Contains(antwort.Befunde[0], "etwas Unbekanntes") {
			t.Errorf("Befunde = %v", antwort.Befunde)
		}
	})

	t.Run("meldet den fehlenden Zugang", func(t *testing.T) {
		t.Parallel()

		if antwort := app.NeueApp("").ScryfallRegeln(probekontext(t), "X"); antwort.Fehler == nil {
			t.Fatal("der fehlende Zugang blieb ohne Fehler")
		}
	})

	t.Run("reicht einen Fehler übersetzt durch", func(t *testing.T) {
		t.Parallel()

		anwendung := app.NeueApp("")
		anwendung.MitScryfallauskunft(&auskunftsdoppel{fehler: domain.ErrNotFound})

		antwort := anwendung.ScryfallRegeln(t.Context(), "X")
		if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeNichtGefunden {
			t.Fatalf("Fehler = %+v", antwort.Fehler)
		}
	})
}

func TestScryfallManakosten(t *testing.T) {
	t.Parallel()

	t.Run("meldet einen Kosten ohne Beanstandung als leere Liste", func(t *testing.T) {
		t.Parallel()

		anwendung := app.NeueApp("")
		anwendung.MitScryfallauskunft(&auskunftsdoppel{})

		antwort := anwendung.ScryfallManakosten(t.Context(), "{2}{R}")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		// **Eine leere Liste und nicht `nil`** — sonst stünde im Fenster
		// `undefined`, und die Prüfung sähe aus wie ein Fehler (§3).
		if antwort.Unbekannt == nil {
			t.Error("Unbekannt ist nil — im Fenster stünde undefined")
		}

		if len(antwort.Unbekannt) != 0 {
			t.Errorf("Unbekannt = %v", antwort.Unbekannt)
		}
	})

	t.Run("nennt die unbekannten Symbole", func(t *testing.T) {
		t.Parallel()

		anwendung := app.NeueApp("")
		anwendung.MitScryfallauskunft(&auskunftsdoppel{unbekannt: []string{"{r}"}})

		antwort := anwendung.ScryfallManakosten(t.Context(), "{2}{r}")
		if len(antwort.Unbekannt) != 1 || antwort.Unbekannt[0] != "{r}" {
			t.Errorf("Unbekannt = %v", antwort.Unbekannt)
		}
	})

	t.Run("meldet den fehlenden Zugang", func(t *testing.T) {
		t.Parallel()

		if antwort := app.NeueApp("").ScryfallManakosten(
			probekontext(t), "{W}"); antwort.Fehler == nil {
			t.Fatal("der fehlende Zugang blieb ohne Fehler")
		}
	})

	t.Run("reicht einen Fehler durch", func(t *testing.T) {
		t.Parallel()

		anwendung := app.NeueApp("")
		anwendung.MitScryfallauskunft(&auskunftsdoppel{fehler: domain.ErrValidation})

		if antwort := anwendung.ScryfallManakosten(
			probekontext(t), "{W}"); antwort.Fehler == nil {
			t.Fatal("der Fehler kam nicht durch")
		}
	})
}

// **Die Naht für die Flächen ohne Fenster kommt zusammen.**
//
// **Hier stand „und sie braucht keinen Bestand"** — weil eine Regelauskunft
// nichts aus der Datenbank liest. Das stimmt für die Auskunft und nicht mehr
// für die Frage davor: seit Phase 4.8 steht dort, **ob Scryfall überhaupt
// gefragt werden darf**. Ohne den Bestand ginge ein abgeschalteter Dienst an
// dieser Fläche weiter hinaus.
func TestDieAuskunftsnahtKommtZusammen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	naht, fehler := app.Fremdauskunftsnaht(frischerBestand(t))
	if fehler != nil {
		t.Fatalf("Fremdauskunftsnaht: %v", fehler)
	}

	if naht == nil {
		t.Fatal("die Naht ist leer")
	}
}

// **Ohne Zuhause kommt sie nicht zusammen.**
func TestOhneZuhauseKommtDieAuskunftsnahtNichtZusammen(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("MSE_BESTAND", "")

	if _, fehler := app.Fremdauskunftsnaht(frischerBestand(t)); fehler == nil {
		t.Error("ohne Zuhause kam eine Naht")
	}
}

// **Und der ganze Weg bis zur Gegenstelle** — Naht, Klient, beide Drosseln.
func TestDieAuskunftsnahtGehtBisZurGegenstelle(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			schreiben.Header().Set("Content-Type", "application/json")

			switch {
			case strings.HasSuffix(anfrage.URL.Path, "/rulings"):
				sende(t, schreiben, []byte(`{"object":"list","data":[
					{"object":"ruling","oracle_id":"x","source":"wotc",
					 "published_at":"2004-10-04","comment":"Ein Hinweis."}]}`))
			case anfrage.URL.Path == "/symbology":
				sende(t, schreiben, []byte(`{"object":"list","data":[
					{"object":"card_symbol","symbol":"{R}","english":"rot",
					 "represents_mana":true,"appears_in_mana_costs":true}]}`))
			default:
				sende(t, schreiben, []byte(`{"object":"card","id":"x","name":"Prüfstein",
					"scryfall_uri":"https://beispiel.invalid/x"}`))
			}
		}))
	t.Cleanup(gegenstelle.Close)

	vorher := app.MitScryfalladresse(gegenstelle.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	naht, fehler := app.Fremdauskunftsnaht(frischerBestand(t))
	if fehler != nil {
		t.Fatalf("Fremdauskunftsnaht: %v", fehler)
	}

	geholt, holfehler := naht.Regelhinweise(probekontext(t), "Prüfstein")
	if holfehler != nil {
		t.Fatalf("Regelhinweise: %v", holfehler)
	}

	if len(geholt.Hinweise) != 1 || !geholt.Hinweise[0].Offiziell {
		t.Errorf("Hinweise = %+v", geholt.Hinweise)
	}

	if geholt.Kartenname != "Prüfstein" {
		t.Errorf("Kartenname = %q", geholt.Kartenname)
	}

	// **Und die Manakostenprüfung über denselben Weg.**
	unbekannt, prueffehler := naht.Manakosten(probekontext(t), "{R}{x}")
	if prueffehler != nil {
		t.Fatalf("Manakosten: %v", prueffehler)
	}

	if len(unbekannt) != 1 || unbekannt[0] != "{x}" {
		t.Errorf("unbekannt = %v", unbekannt)
	}

	// **Und die Setangaben** — der dritte Weg der Naht.
	if _, satzfehler := naht.Satz(probekontext(t), ""); satzfehler == nil {
		t.Error("das leere Kürzel kam durch")
	}
}

// **Ein Ausfall der Gegenstelle kommt über die Naht durch.**
//
// Der Test daneben fährt den guten Weg; dieser den schlechten. Ohne ihn bliebe
// der Fehlerzweig unbetreten, und niemand wüsste, ob eine Auskunft ohne Netz
// **scheitert** oder still eine leere Liste zurückgibt.
func TestEinAusfallKommtUeberDieAuskunftsnahtDurch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusInternalServerError)
		}))
	t.Cleanup(gegenstelle.Close)

	vorher := app.MitScryfalladresse(gegenstelle.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	naht, fehler := app.Fremdauskunftsnaht(frischerBestand(t))
	if fehler != nil {
		t.Fatalf("Fremdauskunftsnaht: %v", fehler)
	}

	if _, holfehler := naht.Regelhinweise(probekontext(t), "Prüfstein"); holfehler == nil {
		t.Error("der Ausfall kam nicht durch")
	}
}
