package app_test

// **Die Wege der Scrydex-Preisnaht, die kein Erfolgsfall betritt** —
// Werkstück A5 von Phase 5.3
// (#1444).

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ein anderes Set und eine andere Nummer fallen weg.**
//
// # Warum beides in einem Lauf steht
//
// **Weil es zwei Siebe hintereinander sind** und jedes für sich durchlässig
// wirkt: dieselbe Karte gibt es in vielen Sets, und in einem Set zweimal.
// Bliebe eines von beiden aus, wäre die Zuordnung mehrdeutig — und
// mehrdeutig ist ein Konflikt (A2, #1441).
func TestFremdesSetUndFremdeNummerFallenWeg(t *testing.T) {
	db := scrydexbestand(t, scrydexgegenstelle(t, `[
	  {"id":"falsches-set","name":"Probe","number":"1",
	   "expansion":{"code":"xyz","name":"Fremd"}},
	  {"id":"falsche-nummer","name":"Probe","number":"7",
	   "expansion":{"code":"ptkdf","name":"Tarkir"}},
	  {"id":"mtg-m11-149","name":"Probe","number":"1",
	   "expansion":{"code":"PTKDF","name":"Tarkir"}}]`))

	antwort := app.AnbieterAnwendungUeber(db).
		MitMarktbund(marktbund{marke: "geheim"}).
		Fremdpreisverlauf(probekontext(t), "scrydex",
			"11111111-1111-4111-8111-111111111111", "", nil)

	if antwort.Fehler != nil {
		t.Fatalf("Fremdpreisverlauf: %+v", antwort.Fehler)
	}

	// **Der Setcode wird ohne Rücksicht auf Groß und Klein verglichen** —
	// `PTKDF` ist dasselbe Set wie `ptkdf`.
	if antwort.Fassung != "mtg-m11-149" {
		t.Errorf("zugeordnet wurde %q", antwort.Fassung)
	}
}

// **Zwei Karten in demselben Set mit derselben Nummer sind ein Konflikt.**
func TestZweiScrydexkartenSindEinKonflikt(t *testing.T) {
	db := scrydexbestand(t, scrydexgegenstelle(t, `[
	  {"id":"eins","name":"Probe","number":"1",
	   "expansion":{"code":"ptkdf","name":"Tarkir"}},
	  {"id":"zwei","name":"Probe","number":"1",
	   "expansion":{"code":"ptkdf","name":"Tarkir"}}]`))

	antwort := app.AnbieterAnwendungUeber(db).
		MitMarktbund(marktbund{marke: "geheim"}).
		Fremdpreisverlauf(probekontext(t), "scrydex",
			"11111111-1111-4111-8111-111111111111", "", nil)

	if antwort.Fehler == nil {
		t.Fatal("zwei Treffer kamen als Zuordnung durch")
	}

	// **Die Quelle steht in der Meldung** — sie sagt, wen jemand fragen muss.
	if !strings.Contains(antwort.Fehler.Meldung, "scrydex") {
		t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
	}
}

// **Ein Fehler beim Suchen reist weiter.**
func TestEinFehlerBeimScrydexSuchenReistWeiter(t *testing.T) {
	db := scrydexbestand(t, kaputteGegenstelle(t))

	antwort := app.AnbieterAnwendungUeber(db).
		MitMarktbund(marktbund{marke: "geheim"}).
		Fremdpreisverlauf(probekontext(t), "scrydex",
			"11111111-1111-4111-8111-111111111111", "", nil)

	if antwort.Fehler == nil {
		t.Fatal("ein 500 kam als Verlauf zurück")
	}
}

// **Und einer beim Holen ebenso.**
func TestEinFehlerBeimScrydexHolenReistWeiter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if strings.HasSuffix(gestellt.URL.Path, "/price_history") {
				schreiben.WriteHeader(http.StatusInternalServerError)

				return
			}

			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"data":[{"id":"mtg-m11-149","name":"Probe","number":"1",
				           "expansion":{"code":"ptkdf","name":"Tarkir"}}]}`))
		}))
	t.Cleanup(server.Close)

	antwort := app.AnbieterAnwendungUeber(scrydexbestand(t, server.URL)).
		MitMarktbund(marktbund{marke: "geheim"}).
		Fremdpreisverlauf(probekontext(t), "scrydex",
			"11111111-1111-4111-8111-111111111111", "", nil)

	if antwort.Fehler == nil {
		t.Fatal("ein 500 beim Verlauf kam als Erfolg zurück")
	}
}

// **Ein Wunsch geht hinaus und siebt die Antwort.**
//
// **Beides zusammen**, und das ist kein doppelter Weg: Scrydex nimmt **einen**
// Zustand in der Abfrage, und wer zwei nennt, bekäme sonst alle (§16, Drift).
func TestDerVerlaufswunschGehtHinausUndSiebt(t *testing.T) {
	gesehen := make(chan string, 2)

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if strings.HasSuffix(gestellt.URL.Path, "/price_history") {
				gesehen <- gestellt.URL.RawQuery

				// **Die Gegenstelle übergeht den Filter** — und genau das
				// soll die Naht auffangen.
				_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
					`{"data":[{"date":"2026-09-10","prices":[
					   {"variant":"nonfoil","condition":"NM","type":"raw",
					    "low":4.17,"market":5.02,"currency":"USD"},
					   {"variant":"foil","condition":"NM","type":"raw",
					    "low":9.00,"market":9.50,"currency":"USD"}]}]}`))

				return
			}

			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"data":[{"id":"mtg-m11-149","name":"Probe","number":"1",
				           "expansion":{"code":"ptkdf","name":"Tarkir"}}]}`))
		}))
	t.Cleanup(server.Close)

	antwort := app.AnbieterAnwendungUeber(scrydexbestand(t, server.URL)).
		MitMarktbund(marktbund{marke: "geheim"}).
		Fremdpreisverlauf(probekontext(t), "scrydex",
			"11111111-1111-4111-8111-111111111111", "nonfoil", []string{"NM"})

	if antwort.Fehler != nil {
		t.Fatalf("Fremdpreisverlauf: %+v", antwort.Fehler)
	}

	if abfrage := <-gesehen; !strings.Contains(abfrage, "condition=NM") ||
		!strings.Contains(abfrage, "variant=nonfoil") {
		t.Errorf("die Abfrage ist %q", abfrage)
	}

	// **Zwei Stände statt vier**: die Foil-Zeile ist gesiebt worden.
	if antwort.Staende != 2 {
		t.Errorf("%d Stände abgelegt, erwartet 2", antwort.Staende)
	}
}

// **Und zwei gewünschte Zustände gehen nicht in die Abfrage.**
//
// **Die Gegenprobe zum Lauf darüber**: dort steht einer und geht hinaus, hier
// stehen zwei und bleiben hier. Ohne sie sagte jener nur, dass ein Zustand
// mitreist — nicht, dass zwei es nicht tun.
func TestZweiZustaendeGehenNichtInDieAbfrage(t *testing.T) {
	gesehen := make(chan string, 2)

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if strings.HasSuffix(gestellt.URL.Path, "/price_history") {
				gesehen <- gestellt.URL.RawQuery
				_, _ = schreiben.Write([]byte(`{"data":[]}`)) //nolint:errcheck // die Gegenstelle

				return
			}

			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"data":[{"id":"mtg-m11-149","name":"Probe","number":"1",
				           "expansion":{"code":"ptkdf","name":"Tarkir"}}]}`))
		}))
	t.Cleanup(server.Close)

	if antwort := app.AnbieterAnwendungUeber(scrydexbestand(t, server.URL)).
		MitMarktbund(marktbund{marke: "geheim"}).
		Fremdpreisverlauf(probekontext(t), "scrydex",
			"11111111-1111-4111-8111-111111111111", "",
			[]string{"NM", "LP"}); antwort.Fehler != nil {
		t.Fatalf("Fremdpreisverlauf: %+v", antwort.Fehler)
	}

	if abfrage := <-gesehen; strings.Contains(abfrage, "condition") {
		t.Errorf("die Abfrage ist %q — zwei Zustände gingen hinaus", abfrage)
	}
}

// scrydexgegenstelle antwortet auf die Kartensuche mit festen Daten.
func scrydexgegenstelle(t *testing.T, karten string) string {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if strings.HasSuffix(gestellt.URL.Path, "/price_history") {
				_, _ = schreiben.Write([]byte(`{"data":[]}`)) //nolint:errcheck // die Gegenstelle

				return
			}

			_, _ = schreiben.Write([]byte(`{"data":` + karten + `}`)) //nolint:errcheck // die Gegenstelle
		}))

	t.Cleanup(server.Close)

	return server.URL
}

// kaputteGegenstelle antwortet auf alles mit einem 500.
func kaputteGegenstelle(t *testing.T) string {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusInternalServerError)
		}))

	t.Cleanup(server.Close)

	return server.URL
}

// **Ein Zustand, den es im Stand nicht gibt, siebt ihn weg.**
//
// **Die Gegenprobe zu [TestDerVerlaufswunschGehtHinausUndSiebt]**: dort passt
// der gewünschte Zustand, hier keiner. Ohne diesen Lauf bliebe offen, ob das
// Sieb je etwas zurückhält, das die Quelle geschickt hat.
func TestEinFremderZustandWirdWeggesiebt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if strings.HasSuffix(gestellt.URL.Path, "/price_history") {
				_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
					`{"data":[{"date":"2026-09-10","prices":[
					   {"variant":"nonfoil","condition":"NM","type":"raw",
					    "low":4.17,"market":5.02,"currency":"USD"}]}]}`))

				return
			}

			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"data":[{"id":"mtg-m11-149","name":"Probe","number":"1",
				           "expansion":{"code":"ptkdf","name":"Tarkir"}}]}`))
		}))
	t.Cleanup(server.Close)

	antwort := app.AnbieterAnwendungUeber(scrydexbestand(t, server.URL)).
		MitMarktbund(marktbund{marke: "geheim"}).
		Fremdpreisverlauf(probekontext(t), "scrydex",
			"11111111-1111-4111-8111-111111111111", "",
			[]string{"Poor", "Played"})

	if antwort.Fehler != nil {
		t.Fatalf("Fremdpreisverlauf: %+v", antwort.Fehler)
	}

	if antwort.Staende != 0 {
		t.Errorf("%d Stände abgelegt, erwartet keiner", antwort.Staende)
	}
}
