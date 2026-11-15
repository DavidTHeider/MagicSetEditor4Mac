package cardtrader_test

// **Das eigene Konto, nur lesend** — Werkstück `T11` von Phase 5.6
// (#1476).
//
// # Was hier geprüft wird
//
//   - Die drei Wege kommen durch, mit ihren Parametern.
//   - **Das `shared_secret` aus `/info` kommt nicht mit.** Der Anbieter
//     schreibt „Do not share this information"; ein Feld, das dieses Programm
//     gar nicht erst liest, kann auch nicht in einen Export geraten.
//   - Ein unbekannter Bestellzustand ist ein **Befund** und kein stilles Ja.
//   - **Und es wird nur gelesen**: der Spion hält fest, dass alle drei Rufe
//     `GET` sind.

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/cardtrader"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// kontoklient baut den Klienten gegen eine eigene Gegenstelle.
func kontoklient(t *testing.T, adresse string) *cardtrader.Klient {
	t.Helper()

	return cardtrader.NeuerKlient(adresse,
		netz.NeuerKlient(netz.UeblicheFristen()),
		cardtrader.Zugang{Marke: "geheim"})
}

// fmtKonto macht aus dem Konto eine Zeichenkette — für die Suche nach dem,
// was nicht darin stehen darf.
func fmtKonto(konto domain.Marktkonto) string {
	return fmt.Sprintf("%+v", konto)
}

// dieKontoantworten sind die Gestalten, die CardTrader schickt — selbst
// geschrieben (§11), auf das Nötige gekürzt und **nachgelesen am
// 2026-09-19**.
const (
	dasInfo = `{"id":3,"name":"Test App","shared_secret":"5acf05564",` +
		`"user_id":51242,"webhook_url":""}`

	dieAngebote = `[{"id":108455069,"name_en":"Tranquil Cove","quantity":2,` +
		`"description":"","price_cents":250,"price_currency":"EUR",` +
		`"game_id":1,"category_id":1,"blueprint_id":7480,"user_id":51242,` +
		`"graded":false,"tag":"","user_data_field":"Kiste 4",` +
		`"bundle_size":1,"bundled_quantity":2}]`

	dieBestellungen = `[{"id":1234567,"code":"AB12","state":"hub_pending",` +
		`"via_cardtrader_zero":true,"order_as":"seller","size":3,` +
		`"paid_at":"2026-09-18T12:00:00.000Z",` +
		`"seller_total":{"cents":1250,"currency":"EUR"}}]`
)

// kontogegenstelle antwortet auf die drei Wege und schreibt mit.
func kontogegenstelle(t *testing.T, bestellungen string) (
	server *httptest.Server, gesehen *[]string,
) {
	t.Helper()

	gesehen = &[]string{}

	server = httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			*gesehen = append(*gesehen,
				anfrage.Method+" "+anfrage.URL.RequestURI())

			switch {
			case strings.HasPrefix(anfrage.URL.Path, "/info"):
				schreibeKonto(t, schreiben, dasInfo)

			case strings.HasPrefix(anfrage.URL.Path, "/products/export"):
				schreibeKonto(t, schreiben, dieAngebote)

			default:
				schreibeKonto(t, schreiben, bestellungen)
			}
		}))

	t.Cleanup(server.Close)

	return server, gesehen
}

// schreibeKonto schreibt und meldet, wenn es nicht geht.
func schreibeKonto(t *testing.T, schreiben http.ResponseWriter, roh string) {
	t.Helper()

	if _, fehler := schreiben.Write([]byte(roh)); fehler != nil {
		t.Errorf("schreiben: %v", fehler)
	}
}

// **Die drei Wege kommen durch — und sie sind alle drei `GET`.**
func TestDieDreiKontowegeKommenDurch(t *testing.T) {
	t.Parallel()

	server, gesehen := kontogegenstelle(t, dieBestellungen)
	klient := kontoklient(t, server.URL)

	konto, fehler := klient.Konto(t.Context())
	if fehler != nil {
		t.Fatalf("Konto: %v", fehler)
	}

	if konto.NutzerID != 51242 || konto.Name != "Test App" {
		t.Errorf("das Konto kam als %+v an", konto)
	}

	angebote, fehler := klient.EigeneAngebote(t.Context(), 0)
	if fehler != nil {
		t.Fatalf("EigeneAngebote: %v", fehler)
	}

	if len(angebote) != 1 || angebote[0].Menge != 2 ||
		angebote[0].PreisCent != 250 || angebote[0].Waehrung != "EUR" {
		t.Errorf("die Angebote kamen als %+v an", angebote)
	}

	// **Die eigene Notiz kommt mit** — sie ist der Faden zum eigenen Lager.
	if angebote[0].Eigennotiz != "Kiste 4" {
		t.Errorf("die Eigennotiz lautet %q", angebote[0].Eigennotiz)
	}

	bestellungen, befunde, fehler := klient.Bestellungen(t.Context(), 1, 20)
	if fehler != nil {
		t.Fatalf("Bestellungen: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("es gab Befunde: %+v", befunde)
	}

	if len(bestellungen) != 1 ||
		bestellungen[0].Zustand != domain.BestellungHubOffen ||
		!bestellungen[0].UeberNullkonto {
		t.Errorf("die Bestellungen kamen als %+v an", bestellungen)
	}

	// **Alle drei Rufe sind `GET`** — und die Seitenangabe reist mit.
	for _, ruf := range *gesehen {
		if !strings.HasPrefix(ruf, "GET ") {
			t.Errorf("es ging %q hinaus", ruf)
		}
	}

	if !strings.Contains(strings.Join(*gesehen, " "), "page=1") {
		t.Errorf("die Seitenangabe fehlt: %v", *gesehen)
	}
}

// **Das `shared_secret` wird gar nicht erst gelesen.**
//
// **Der Anbieter schreibt „Do not share this information"**, und die
// billigste Art, es nicht weiterzugeben, ist, es nicht zu haben: es steht in
// keinem Feld dieses Programms, und damit auch in keinem Export.
func TestDasGeteilteGeheimnisWirdNichtGelesen(t *testing.T) {
	t.Parallel()

	server, _ := kontogegenstelle(t, dieBestellungen)

	konto, fehler := kontoklient(t, server.URL).Konto(t.Context())
	if fehler != nil {
		t.Fatalf("Konto: %v", fehler)
	}

	if strings.Contains(fmtKonto(konto), "5acf05564") {
		t.Errorf("das Geheimnis steht im Konto: %s", fmtKonto(konto))
	}
}

// **Ein unbekannter Bestellzustand ist ein Befund.**
//
// **Und die Bestellung fällt trotzdem nicht weg** (§8): was lesbar ist, wird
// gelesen, und was nicht aufgeht, wird benannt — sonst fehlte eine Zeile, und
// niemand wüsste warum.
func TestEinUnbekannterBestellzustandWirdBenannt(t *testing.T) {
	t.Parallel()

	server, _ := kontogegenstelle(t, `[{"id":9,"code":"C","state":"verschollen",`+
		`"via_cardtrader_zero":false,"order_as":"seller","size":1}]`)

	bestellungen, befunde, fehler := kontoklient(t, server.URL).
		Bestellungen(t.Context(), 0, 0)
	if fehler != nil {
		t.Fatalf("Bestellungen: %v", fehler)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde: %+v", len(befunde), befunde)
	}

	if !strings.Contains(befunde[0].Meldung, "verschollen") {
		t.Errorf("der Befund nennt den Zustand nicht: %+v", befunde[0])
	}

	if len(bestellungen) != 1 {
		t.Errorf("die Bestellung fiel weg: %+v", bestellungen)
	}
}

// **Die drei Wege stehen in der Positivliste.**
//
// **Sonst wäre die Aufzählung zu kurz**, und der Wächter darüber sähe nicht,
// dass dieses Paket seit T11 drei Adressen mehr rufen kann.
func TestDieKontowegeStehenInDerPositivliste(t *testing.T) {
	t.Parallel()

	wege := strings.Join(cardtrader.Wege(), " ")

	for _, weg := range []string{"/info", "/products/export", "/orders"} {
		if !strings.Contains(wege, weg) {
			t.Errorf("%q fehlt in der Positivliste: %v", weg, cardtrader.Wege())
		}
	}
}

// **Jeder der drei Wege reicht seinen Fehler durch.**
//
// **Und zwar vor der Zuordnung**: eine Antwort, die nicht aufgeht, darf keine
// halbe Liste ergeben — sonst sähe ein Konto leerer aus, als es ist (§8).
func TestJederKontowegReichtSeinenFehlerDurch(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusUnauthorized)
		}))

	t.Cleanup(server.Close)

	klient := kontoklient(t, server.URL)

	if _, fehler := klient.Konto(t.Context()); fehler == nil {
		t.Error("Konto kam ohne Fehler zurück")
	}

	if _, fehler := klient.EigeneAngebote(t.Context(), 0); fehler == nil {
		t.Error("EigeneAngebote kam ohne Fehler zurück")
	}

	if _, _, fehler := klient.Bestellungen(t.Context(), 0, 0); fehler == nil {
		t.Error("Bestellungen kam ohne Fehler zurück")
	}
}
