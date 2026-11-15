// **Eine Karte über ihre Druckkennung** — Werkstück 2 von Phase 3.6
// (#1167).
//
// # Warum es diesen Weg neben `KarteNachName` gibt
//
// `KarteNachName` führt auf **irgendeinen** Druck: Scryfall wählt bei
// `/cards/named` selbst aus. Wer in der Trefferliste einen bestimmten Druck
// angeklickt hat, meint aber genau den — mit seinem Set, seiner Nummer und
// seinem Bild.
//
// **Die Adresse ist dieselbe wie beim Preis**: `/cards/:id`, nachgelesen am
// 2026-08-27 und in `preise.go` belegt. Deshalb liegt beides unter **einem**
// Schlüssel im Zwischenspeicher — zwei Schlüssel für dieselbe Antwort wären
// zwei Anfragen an eine Gegenstelle, die eine Grenze nennt (§16, Regel 4).
package scryfall_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Der Druck kommt, nach dem gefragt wurde** — und die Kennung wird geglättet.
//
// Dieselbe Regel wie bei `PreiseFuerDruck`, und aus demselben Grund: Scryfall
// nimmt eine UUID in Großschreibung an und antwortet klein. Wer roh anfragte,
// legte denselben Druck zweimal in den Zwischenspeicher.
func TestKarteNachKennungHoltDenGenanntenDruck(t *testing.T) {
	t.Parallel()

	const klein = "516e137a-d9d9-4ba7-89fb-e680ddb2c2d1"

	gefragt := make(chan string, 1)

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			gefragt <- anfrage.URL.Path

			if _, err := schreiben.Write([]byte(`{"object":"card","id":"` + klein + `",` +
				`"name":"Lightning Bolt","set":"lea","set_name":"Limited Edition Alpha",` +
				`"collector_number":"161",` +
				`"scryfall_uri":"https://scryfall.com/card/lea/161"}`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	karte, _, _, err := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client())).
		KarteNachKennung(t.Context(), "  516E137A-D9D9-4BA7-89FB-E680DDB2C2D1  ")
	if err != nil {
		t.Fatalf("KarteNachKennung: %v", err)
	}

	if karte.Name != "Lightning Bolt" || karte.Nummer != "161" || karte.Set != "lea" {
		t.Errorf("gelesen wurde %q aus %q Nr. %q", karte.Name, karte.Set, karte.Nummer)
	}

	if weg := <-gefragt; !strings.HasSuffix(weg, "/cards/"+klein) {
		t.Errorf("gefragt wurde %q — will die geglättete Kennung", weg)
	}
}

// **Ohne Kennung ist kein Druck gemeint.**
//
// Ein leerer Wert führte sonst auf `/cards/` — eine Adresse, die etwas
// anderes bedeutet, und deren Antwort niemand geprüft hat.
func TestKarteNachKennungWeistDieLeereKennungAb(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(_ http.ResponseWriter, _ *http.Request) {
			t.Error("es ging eine Anfrage hinaus")
		}))
	t.Cleanup(gegenstelle.Close)

	_, _, _, err := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client())).
		KarteNachKennung(t.Context(), "   ")
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler war %v — will ErrValidation", err)
	}
}

// **Die Antwort muss den Druck nennen, nach dem gefragt wurde.**
//
// Dieselbe Wache wie beim Preis: ein vertauschter oder beschädigter Eintrag im
// Zwischenspeicher wäre sonst die Karte eines fremden Drucks — und die Werte
// daraus landeten in einem Kartenfeld, mit einer Herkunft, die auf den
// **falschen** Druck zeigt.
func TestKarteNachKennungWeistEinenFremdenDruckAb(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, err := schreiben.Write([]byte(
				`{"object":"card","id":"00000000-0000-0000-0000-000000000000",` +
					`"name":"Etwas anderes"}`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	_, _, _, err := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client())).
		KarteNachKennung(t.Context(), "516e137a-d9d9-4ba7-89fb-e680ddb2c2d1")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("Fehler war %v — will ErrConflict", err)
	}
}

// **Karte und Preis teilen sich die eine Antwort.**
//
// # Warum das eine Zusage ist und keine Beobachtung
//
// Beide Wege fragen `/cards/:id`. Lägen sie unter verschiedenen Schlüsseln,
// stellte das Fenster für einen angeklickten Treffer **zwei** Anfragen — und
// §16, Regel 4 sagt: eine Grenze wird eingehalten, bevor sie greift.
//
// Gezählt wird an der Gegenstelle, nicht an einer Zusicherung über den
// Speicher: die Reichweite einer Messung ist die Reichweite ihrer Aussage
// (§10).
func TestKarteUndPreisTeilenSichEineAntwort(t *testing.T) {
	t.Parallel()

	const kennung = "516e137a-d9d9-4ba7-89fb-e680ddb2c2d1"

	var anfragen atomic.Int64

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			anfragen.Add(1)

			if _, err := schreiben.Write([]byte(`{"object":"card","id":"` + kennung + `",` +
				`"name":"Lightning Bolt","prices":{"eur":"1.00"}}`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	if _, _, _, err := klient.KarteNachKennung(t.Context(), kennung); err != nil {
		t.Fatalf("KarteNachKennung: %v", err)
	}

	if _, _, _, err := klient.PreiseFuerDruck(t.Context(), kennung); err != nil {
		t.Fatalf("PreiseFuerDruck: %v", err)
	}

	if gezaehlt := anfragen.Load(); gezaehlt != 1 {
		t.Errorf("es gingen %d Anfragen hinaus — will eine", gezaehlt)
	}
}

// **Eine unlesbare Antwort ist ein Fehler und keine Karte.**
//
// Der Zweig steht zwischen Holen und Prüfen; ohne ihn käme ein `nil` heraus,
// und der Aufrufer läse eine Kennung von etwas, das es nicht gibt.
func TestKarteNachKennungWeistEineUnlesbareAntwortAb(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, err := schreiben.Write([]byte(`{kein JSON`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	_, _, _, err := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client())).
		KarteNachKennung(t.Context(), "516e137a-d9d9-4ba7-89fb-e680ddb2c2d1")
	if err == nil {
		t.Fatal("die unlesbare Antwort kam als Karte zurück")
	}
}
