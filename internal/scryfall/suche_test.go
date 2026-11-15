// Die Suche über `/cards/search`
// (#1005).
//
// **Eine eigene Datei, weil sie eine eigene Frage beantwortet** (§2) — und
// weil `klient_test.go` sonst über die Zeilengrenze liefe.
package scryfall_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Eine Suche gibt mehrere Karten und sagt, ob es mehr gibt.**
func TestEineSucheGibtMehrereKarten(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreibe(t, schreiben, beispiel(t, "suche.json"))
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	treffer, _, befunde, fehler := klient.Suche(t.Context(), "t:creature")
	if fehler != nil {
		t.Fatalf("suchen: %v", fehler)
	}

	if len(treffer.Karten) != 2 {
		t.Fatalf("%d Karten, erwartet waren zwei", len(treffer.Karten))
	}

	if !treffer.Weitere {
		t.Error("has_more wurde nicht durchgereicht")
	}

	// **Die Befunde der einzelnen Karten reisen mit.** Die zweite Karte in der
	// Probe hat eine unbekannte Seltenheit.
	if len(befunde) != 1 {
		t.Errorf("%d Befunde, erwartet war einer: %v", len(befunde), befunde)
	}
}

// **Eine Suchantwort, die keine Liste ist, bricht laut.**
func TestEineSuchantwortOhneListeBrichtLaut(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"kein JSON":                       "das ist kein JSON",
		"keine Liste":                     `{"object":"card","id":"x","name":"y"}`,
		"eine kaputte Karte in der Liste": `{"object":"list","has_more":false,"data":[{"object":"card"}]}`,
	}

	for name, rumpf := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gegenstelle := httptest.NewServer(http.HandlerFunc(
				func(schreiben http.ResponseWriter, _ *http.Request) {
					schreibe(t, schreiben, []byte(rumpf))
				}))
			t.Cleanup(gegenstelle.Close)

			klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

			if _, _, _, fehler := klient.Suche(t.Context(), "x"); fehler == nil {
				t.Error("die kaputte Antwort ging durch")
			}
		})
	}
}

// **Und eine Kartenantwort, die kaputt ist, ebenso.**
func TestEineKaputteKartenantwortBrichtLaut(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreibe(t, schreiben, []byte(`{"object":"card"}`))
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	if _, _, _, fehler := klient.KarteNachName(t.Context(), "x"); fehler == nil {
		t.Error("die kaputte Antwort ging durch")
	}
}

// **Die Druckesuche fragt anders — und der Unterschied ist die Zusage.**
//
// # Warum dieser Lauf die Adresse prüft und nicht nur die Antwort
//
// Weil der Unterschied **in der Frage** liegt: ohne `unique=prints` gibt
// Scryfall für einen genauen Namen **eine** Zeile statt aller Ausgaben, und
// eine eingefügte Kartenliste sähe damit durchweg eindeutig aus
// (#1292).
//
// **Genau das ist am 2026-08-30 passiert**, und kein Test hat es gesehen: die
// Doppel gaben mehrere Treffer, weil sie so geschrieben waren. Gefunden hat es
// der Ruf durch die laufende Brücke gegen den echten Dienst.
func TestDieDruckesucheFragtNachAusgaben(t *testing.T) {
	t.Parallel()

	gefragt := make(chan string, 2)

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			gefragt <- anfrage.URL.RawQuery

			schreibe(t, schreiben, beispiel(t, "suche.json"))
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	if _, _, _, fehler := klient.Druckesuche(t.Context(), `!"Counterspell"`); fehler != nil {
		t.Fatalf("Druckesuche: %v", fehler)
	}

	frage := <-gefragt

	for _, teil := range []string{"unique=prints", "order=released", "dir=asc"} {
		if !strings.Contains(frage, teil) {
			t.Errorf("die Anfrage %q enthält %q nicht", frage, teil)
		}
	}

	// **Und die gewöhnliche Suche fragt es nicht** — sie will eine Zeile je
	// Karte, und das ist dort richtig. Ohne diese Gegenprobe wäre ein Klient,
	// der **immer** `unique=prints` anhängt, grün.
	if _, _, _, fehler := klient.Suche(t.Context(), `!"Counterspell"`); fehler != nil {
		t.Fatalf("Suche: %v", fehler)
	}

	if gewoehnlich := <-gefragt; strings.Contains(gewoehnlich, "unique=prints") {
		t.Errorf("die gewöhnliche Suche fragt %q — sie soll je Karte eine Zeile geben",
			gewoehnlich)
	}
}
