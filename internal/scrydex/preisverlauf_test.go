package scrydex_test

// **Der Preisverlauf bei Scrydex** — Werkstück A5 von Phase 5.3
// (#1444).

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scrydex"
)

// **Der Verlauf kommt an — und jeder Preis bleibt eine ganze Zahl.**
//
// `4.17` sind **417** Cent und nicht `4.17` in einem `float64`: Geld, das
// durch eine Gleitkommazahl geht, kommt anders zurück
// ([0093](../../docs/entscheidungen/0093-eine-preisreihe-hat-sieben-teile.md)).
func TestDerPreisverlaufKommtAn(t *testing.T) {
	t.Parallel()

	server := kontengegenstelle(t, "preisverlauf.json")

	punkte, befunde, fehler := klientFuerKonto(server).
		Preisverlauf(t.Context(), "abc", scrydex.Verlaufswunsch{})
	if fehler != nil {
		t.Fatalf("Preisverlauf: %v", fehler)
	}

	// **Drei von vier**: der vierte nennt keine Währung.
	if len(punkte) != 3 {
		t.Fatalf("%d Punkte, erwartet 3: %+v", len(punkte), punkte)
	}

	erster := punkte[0]
	if erster.Tag != "2026-09-10" || erster.Ausfuehrung != "normal" ||
		erster.Zustand != "NM" || erster.Art != "raw" {
		t.Errorf("der erste Punkt ist %+v", erster)
	}

	if erster.Niedrigst != 417 || erster.Markt != 502 || erster.Waehrung != "USD" {
		t.Errorf("die Preise sind %d und %d %s",
			erster.Niedrigst, erster.Markt, erster.Waehrung)
	}

	zweiter := punkte[1]
	if !zweiter.Makellos || !zweiter.Signiert || !zweiter.Fehldruck ||
		zweiter.Art != "graded" || zweiter.Niedrigst != 1900 {
		t.Errorf("der zweite Punkt ist %+v", zweiter)
	}

	// **Ein Preis ohne Währung ist ein Befund** — Pflichtfelder streng (§16).
	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %+v", len(befunde), befunde)
	}
}

// **Ein Punkt wird zu zwei Preisständen — und sie sind zwei Reihen.**
//
// `low` und `market` sind **beides** Forderungen: das niedrigste Angebot und
// der Marktpreis. Wer sie in eine Reihe legte, mischte zwei Fragen; getrennt
// werden sie über den Quellentrenner, dieselbe Gestalt wie
// `mtgjson:cardkingdom` (0093, 0096).
func TestEinPunktWirdZweiStaende(t *testing.T) {
	t.Parallel()

	punkt := scrydex.Preispunkt{
		Tag: "2026-09-10", Ausfuehrung: "normal", Zustand: "NM", Art: "raw",
		Niedrigst: 417, Markt: 502, Waehrung: "USD",
	}

	staende := punkt.AlsStaende("11111111-1111-4111-8111-111111111111")

	if len(staende) != 2 {
		t.Fatalf("%d Stände, erwartet 2", len(staende))
	}

	if staende[0].Quelle != "scrydex:low" || staende[1].Quelle != "scrydex:market" {
		t.Errorf("die Quellen sind %q und %q", staende[0].Quelle, staende[1].Quelle)
	}

	// **Und beide gehören zur Familie** — für den Widerruf (0093).
	for _, stand := range staende {
		if !stand.Quelle.GehoertZu(scrydex.Familie) {
			t.Errorf("%q gehört nicht zu %q", stand.Quelle, scrydex.Familie)
		}

		if stand.Markt != domain.MarktPapier || stand.Art != domain.PreisartVerkauf {
			t.Errorf("Markt und Art sind %q und %q", stand.Markt, stand.Art)
		}

		if stand.Zustand != "NM" || stand.Waehrung != "USD" ||
			stand.Lage != domain.PreisGenannt {
			t.Errorf("der Stand ist %+v", stand)
		}
	}

	if staende[0].Wert != 417 || staende[1].Wert != 502 {
		t.Errorf("die Werte sind %d und %d", staende[0].Wert, staende[1].Wert)
	}

	// **Der Tag der Quelle wird zum Beobachtungstag** — wann gefragt wurde,
	// weiß der Dienst; welchen Tag die Quelle meint, nur sie (0093).
	if staende[0].Beobachtet.Format("2006-01-02") != "2026-09-10" {
		t.Errorf("beobachtet am %v", staende[0].Beobachtet)
	}
}

// **Ein Tag, den niemand lesen kann, ist ein Befund und kein Nullzeitpunkt.**
func TestEinUnlesbarerTagIstEinBefund(t *testing.T) {
	t.Parallel()

	server := gegenstelleMitRumpf(t,
		`{"data":[{"date":"vorgestern","prices":[
		   {"variant":"normal","condition":"NM","type":"raw",
		    "low":1.00,"market":2.00,"currency":"USD"}]}]}`)

	punkte, befunde, fehler := klientFuerKonto(server).
		Preisverlauf(t.Context(), "abc", scrydex.Verlaufswunsch{})
	if fehler != nil {
		t.Fatalf("Preisverlauf: %v", fehler)
	}

	if len(punkte) != 0 {
		t.Errorf("ein unlesbarer Tag kam durch: %+v", punkte)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1", len(befunde))
	}
}

// **Nur der gesetzte Wunsch steht in der Abfrage.**
func TestNurDerGesetzteVerlaufswunschStehtInDerAbfrage(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name     string
		wunsch   scrydex.Verlaufswunsch
		erwartet string
	}{
		{"ohne Wunsch", scrydex.Verlaufswunsch{}, ""},
		{"nur Tage", scrydex.Verlaufswunsch{Tage: 30}, "days=30"},
		{
			"Zeitraum",
			scrydex.Verlaufswunsch{Von: "2026-01-01", Bis: "2026-02-01"},
			"end_date=2026-02-01&start_date=2026-01-01",
		},
		{
			"Zustand und Ausführung",
			scrydex.Verlaufswunsch{Zustand: "NM", Ausfuehrung: "foil"},
			"condition=NM&variant=foil",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			gesehen := make(chan string, 1)

			server := httptest.NewServer(http.HandlerFunc(
				func(schreiben http.ResponseWriter, gestellt *http.Request) {
					gesehen <- gestellt.URL.RawQuery
					_, _ = schreiben.Write([]byte(`{"data":[]}`)) //nolint:errcheck // die Gegenstelle
				}))
			t.Cleanup(server.Close)

			if _, _, fehler := klientFuerKonto(server).
				Preisverlauf(t.Context(), "abc", fall.wunsch); fehler != nil {
				t.Fatalf("Preisverlauf: %v", fehler)
			}

			if abfrage := <-gesehen; abfrage != fall.erwartet {
				t.Errorf("die Abfrage ist %q, erwartet %q", abfrage, fall.erwartet)
			}
		})
	}
}

// **Ohne Kennung wird nicht gefragt** — der Weg trägt sie.
func TestOhneKennungKeinVerlauf(t *testing.T) {
	t.Parallel()

	var gefragt bool

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gefragt = true
			_, _ = schreiben.Write([]byte(`{"data":[]}`)) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	if _, _, fehler := klientFuerKonto(server).
		Preisverlauf(t.Context(), "  ", scrydex.Verlaufswunsch{}); fehler == nil {
		t.Fatal("ohne Kennung kam ein Verlauf")
	}

	if gefragt {
		t.Error("der Server wurde ohne Kennung gefragt")
	}
}

// **Und die Kennung steht im Weg, nicht in der Abfrage.**
func TestDieKennungStehtImWeg(t *testing.T) {
	t.Parallel()

	gesehen := make(chan string, 1)

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			// **Der gesendete Weg und nicht der gedeutete**: `URL.Path` ist
			// schon wieder entschlüsselt, und darin steht der Schrägstrich
			// wieder als Trenner.
			gesehen <- gestellt.URL.EscapedPath()
			_, _ = schreiben.Write([]byte(`{"data":[]}`)) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	if _, _, fehler := klientFuerKonto(server).
		Preisverlauf(t.Context(), "ab/c", scrydex.Verlaufswunsch{}); fehler != nil {
		t.Fatalf("Preisverlauf: %v", fehler)
	}

	// **Ein Schrägstrich in der Kennung darf keinen neuen Weg aufmachen.**
	if weg := <-gesehen; weg != scrydex.Kartenweg+"/ab%2Fc/price_history" {
		t.Errorf("gefragt wurde %q", weg)
	}
}

// **Eine kaputte Antwort ist ein Fehler.**
func TestEinKaputterVerlaufIstEinFehler(t *testing.T) {
	t.Parallel()

	server := gegenstelleMitRumpf(t, "{nicht json")

	if _, _, fehler := klientFuerKonto(server).
		Preisverlauf(t.Context(), "abc", scrydex.Verlaufswunsch{}); fehler == nil {
		t.Fatal("eine kaputte Antwort kam als Verlauf zurück")
	}
}

// gegenstelleMitRumpf antwortet mit einem festen Rumpf.
func gegenstelleMitRumpf(t *testing.T, rumpf string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write([]byte(rumpf)) //nolint:errcheck // die Gegenstelle
		}))

	t.Cleanup(server.Close)

	return server
}

// **Ein Preis, der keine Zahl ist, ist ein Befund.**
//
// # Wie so etwas entsteht
//
// **Durch `null`.** Die Referenz sagt „numeric", und `low` fehlt in keinem
// dokumentierten Beispiel — aber ein Feld, das der Anbieter für einen Tag
// ohne Angebot leer lässt, käme so an. Ein `float64` machte daraus eine
// stumme Null, und die sähe aus wie „umsonst".
func TestEinPreisDerKeineZahlIstIstEinBefund(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, rumpf string }{
		{
			"das niedrigste Angebot fehlt",
			`{"data":[{"date":"2026-09-10","prices":[
			   {"variant":"normal","condition":"NM","type":"raw",
			    "market":2.00,"currency":"USD"}]}]}`,
		},
		{
			"der Marktpreis fehlt",
			`{"data":[{"date":"2026-09-10","prices":[
			   {"variant":"normal","condition":"NM","type":"raw",
			    "low":1.00,"currency":"USD"}]}]}`,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			punkte, befunde, fehler := klientFuerKonto(gegenstelleMitRumpf(t, fall.rumpf)).
				Preisverlauf(t.Context(), "abc", scrydex.Verlaufswunsch{})
			if fehler != nil {
				t.Fatalf("Preisverlauf: %v", fehler)
			}

			if len(punkte) != 0 {
				t.Errorf("ein Preis ohne Zahl kam durch: %+v", punkte)
			}

			if len(befunde) != 1 {
				t.Fatalf("%d Befunde, erwartet 1", len(befunde))
			}
		})
	}
}

// **Ein Verbindungsfehler reist weiter.**
func TestEineSchweigendeGegenstelleBeimVerlauf(t *testing.T) {
	t.Parallel()

	server := gegenstelleMitRumpf(t, "{}")
	adresse := server.URL

	server.Close()

	klient := scrydex.NeuerKlient(adresse, netz.NeuerKlient(netz.UeblicheFristen()),
		scrydex.Zugang{Schluessel: "schluessel", Team: "team"}).
		MitDrossel(netz.NeueDrossel(0, time.Now, func(time.Duration) {}))

	if _, _, fehler := klient.Preisverlauf(
		t.Context(), "abc", scrydex.Verlaufswunsch{}); fehler == nil {
		t.Fatal("eine geschlossene Gegenstelle kam als Verlauf zurück")
	}
}

// **Und ein Fehler ohne Grund bleibt lesbar.**
//
// Die Gegenprobe zu [TestEinAbgelehnterKontostandIstEinFehler]: dort trägt
// der Rumpf einen Satz, hier keinen. Ohne diesen Lauf bliebe offen, ob die
// Meldung ohne ihn überhaupt zustande kommt.
func TestEinFehlerOhneGrundBleibtLesbar(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusInternalServerError)
			_, _ = schreiben.Write([]byte("kein JSON")) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	_, fehler := klientFuerKonto(server).Kontostand(t.Context())
	if fehler == nil {
		t.Fatal("ein 500 kam als Stand zurück")
	}

	if !strings.Contains(fehler.Error(), "ohne Angabe eines Grundes") {
		t.Errorf("der Fehler ist %v", fehler)
	}
}

// **Eine Karte trägt ihre Sammlernummer.**
//
// # Wofür es sie braucht
//
// **Für die Zuordnung.** Ein Name und ein Setcode nennen oft mehrere Drucke —
// dieselbe Karte kann in einem Set zweimal stehen. Ohne die Nummer bliebe
// eine Zuordnung mehrdeutig, und mehrdeutig ist ein Konflikt (A2, #1441).
//
// **Nachgelesen am 2026-09-11**: `number` steht im Kartenobjekt von Scrydex.
func TestEineKarteTraegtIhreSammlernummer(t *testing.T) {
	t.Parallel()

	karten, fehler := scrydex.LiesKarten([]byte(
		`{"data":[{"id":"mtg-m11-149","name":"Lightning Bolt","number":"149",
		           "expansion":{"code":"m11","name":"Magic 2011"}}]}`),
		http.StatusOK)
	if fehler != nil {
		t.Fatalf("LiesKarten: %v", fehler)
	}

	if len(karten) != 1 {
		t.Fatalf("%d Karten, erwartet 1", len(karten))
	}

	if karten[0].Nummer != "149" || karten[0].Kennung != "mtg-m11-149" {
		t.Errorf("die Karte ist %+v", karten[0])
	}
}
