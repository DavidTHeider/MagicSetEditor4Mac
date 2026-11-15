// Die Karten eines ganzen Satzes über `/cards/search`
// (#1033).
//
// **Eine eigene Datei, weil sie eine eigene Frage beantwortet** (§2): `Suche`
// gibt **eine** Seite und sagt, dass es mehr gibt — hier wird geblättert.
package scryfall_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// eineSeite baut eine Listenantwort mit einer Karte.
func eineSeite(name, weiter string) string {
	naechste := "null"
	if weiter != "" {
		naechste = fmt.Sprintf("%q", weiter)
	}

	return fmt.Sprintf(`{
		"object": "list",
		"total_cards": 2,
		"has_more": %t,
		"next_page": %s,
		"data": [{
			"object": "card",
			"id": "00000000-0000-4000-8000-00000000000%d",
			"name": %q,
			"layout": "normal",
			"rarity": "common",
			"scryfall_uri": "https://beispiel.invalid/karte"
		}]
	}`, weiter != "", naechste, len(name)%10, name)
}

// **Geblättert wird bis `has_more` falsch ist** —
// Abnahmekriterium aus #1033.
//
// # Warum das nicht `Suche` ist
//
// `Suche` gibt **eine** Seite und sagt, dass es mehr gibt — mit Absicht: eine
// Suche, die von selbst durch alle Seiten liefe, stellte ohne Zutun beliebig
// viele Anfragen. **Ein Satz ist der Fall, in dem man genau das will**, und er
// ist begrenzt: der größte Magic-Satz liegt unter tausend Karten, also unter
// sechs Seiten zu 175.
func TestSetkartenBlaettertBisZumEnde(t *testing.T) {
	t.Parallel()

	var gerufen []string

	var gegenstelle *httptest.Server

	gegenstelle = httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			gerufen = append(gerufen, anfrage.URL.String())

			if strings.Contains(anfrage.URL.RawQuery, "page=2") {
				schreibe(t, schreiben, []byte(eineSeite("Zweite", "")))

				return
			}

			schreibe(t, schreiben, []byte(eineSeite("Erste",
				gegenstelle.URL+"/cards/search?q=e%3Alea&page=2")))
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	satz, _, fehler := klient.Setkarten(t.Context(), "lea")
	if fehler != nil {
		t.Fatalf("Setkarten: %v", fehler)
	}

	if len(satz.Karten) != 2 {
		t.Fatalf("%d Karten, erwartet zwei — die zweite Seite fehlt", len(satz.Karten))
	}

	if satz.Karten[0].Name != "Erste" || satz.Karten[1].Name != "Zweite" {
		t.Errorf("Karten = %q, %q — die Reihenfolge der Seiten muss halten",
			satz.Karten[0].Name, satz.Karten[1].Name)
	}

	if len(gerufen) != 2 {
		t.Errorf("%d Anfragen, erwartet zwei: %v", len(gerufen), gerufen)
	}

	// **Die zweite Anfrage geht an `next_page`** und wird nicht selbst
	// zusammengebaut: Scryfall gibt eine volle Adresse, und wer sie nachbaut,
	// rät die Blätterung.
	if !strings.Contains(gerufen[1], "page=2") {
		t.Errorf("die zweite Anfrage ging an %q", gerufen[1])
	}
}

// **Ein Satz, den es nicht gibt, legt nichts an.**
//
// Scryfall antwortet auf eine leere Suche mit 404 — und das ist die richtige
// Antwort: „kein Treffer" ist hier kein leeres Ergebnis, sondern ein falscher
// Satzcode.
func TestSetkartenWeistEinenUnbekanntenSatzAb(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusNotFound)
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	if _, _, fehler := klient.Setkarten(t.Context(), "gibbesnicht"); fehler == nil {
		t.Fatal("kein Fehler bei einem unbekannten Satz")
	}
}

// **Die Warnungen der Antwort werden zu Befunden** (§8).
//
// Nachgelesen am 2026-08-24 unter `https://scryfall.com/docs/api/lists`:
// „non-fatal issues that the API discovered with your input […] they indicate
// that the List will not contain all of the information you requested."
func TestSetkartenReichtDieWarnungenDurch(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreibe(t, schreiben, []byte(`{
				"object": "list",
				"has_more": false,
				"warnings": ["Der Satzcode wurde gekürzt"],
				"data": []
			}`))
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	_, befunde, fehler := klient.Setkarten(t.Context(), "lea")
	if fehler != nil {
		t.Fatalf("Setkarten: %v", fehler)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet einer: %v", len(befunde), befunde)
	}

	if !strings.Contains(befunde[0].Meldung, "gekürzt") {
		t.Errorf("Befund = %q", befunde[0].Meldung)
	}
}

// **Die Blätterung hat eine Grenze, und sie sagt es.**
//
// Ein Dienst, der immer `has_more` sagt, liefe sonst endlos. Der größte
// Magic-Satz liegt unter tausend Karten; die Grenze ist großzügig darüber und
// deckelt trotzdem.
func TestSetkartenHatEineGrenze(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			schreibe(t, schreiben, []byte(eineSeite("Immer",
				"http://"+anfrage.Host+"/cards/search?q=e%3Alea&page=99")))
		}))
	t.Cleanup(gegenstelle.Close)

	// **Mit laufender Uhr**, sonst kostete dieser Lauf allein vierzig mal
	// 600 ms Drosselabstand.
	klient := klientOhneWarten(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	satz, befunde, fehler := klient.Setkarten(t.Context(), "lea")
	if fehler != nil {
		t.Fatalf("Setkarten: %v", fehler)
	}

	if !satz.Abgebrochen {
		t.Error("der Satz sagt nicht, dass die Grenze gegriffen hat — ein " +
			"Aufrufer legte ihn als vollständig ab (#1042)")
	}

	if len(satz.Karten) != scryfall.Seitengrenze {
		t.Errorf("%d Karten, erwartet %d — die Grenze hat nicht gegriffen",
			len(satz.Karten), scryfall.Seitengrenze)
	}

	if len(befunde) == 0 {
		t.Error("die Grenze griff still — ohne Befund hielte der Aufrufer die " +
			"Liste für vollständig")
	}
}

// **Eine Antwort, die keine Liste ist, bricht laut** (§16).
func TestEineSatzantwortOhneListeBrichtLaut(t *testing.T) {
	t.Parallel()

	for name, rumpf := range map[string]string{
		"kein JSON":          "das ist kein JSON",
		"keine Liste":        `{"object":"card","id":"x","name":"y"}`,
		"eine kaputte Karte": `{"object":"list","has_more":false,"data":[{"object":"card"}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gegenstelle := httptest.NewServer(http.HandlerFunc(
				func(schreiben http.ResponseWriter, _ *http.Request) {
					schreibe(t, schreiben, []byte(rumpf))
				}))
			t.Cleanup(gegenstelle.Close)

			klient := klientFuer(t, gegenstelle.URL,
				netz.KlientUeber(gegenstelle.Client()))

			if _, _, fehler := klient.Setkarten(t.Context(), "lea"); fehler == nil {
				t.Fatal("kein Fehler")
			}
		})
	}
}
