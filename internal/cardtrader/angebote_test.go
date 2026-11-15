package cardtrader_test

// **Marktangebote** — Werkstück A4 von Phase 5.3
// (#1443).

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/cardtrader"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Antwort ist nach Blueprint geordnet, nicht flach** (0095).
//
// Wer sie als Array läse, bekäme einen Zerteilfehler — und zwar erst im
// Betrieb, weil ein Beispiel mit einem Blueprint wie ein Array aussieht,
// solange man nur die innere Klammer ansieht.
func TestAngeboteLesen(t *testing.T) {
	t.Parallel()

	server := gegenstelle(t,
		map[string]string{"/marketplace/products": "angebote.json"})

	angebote, befunde, fehler := klientFuer(server).
		Angebote(t.Context(), 10050, cardtrader.Angebotsfilter{})
	if fehler != nil {
		t.Fatalf("Angebote: %v", fehler)
	}

	if len(angebote) != 2 {
		t.Fatalf("%d Angebote, erwartet 2 — das dritte nennt keine Währung", len(angebote))
	}

	erstes := angebote[0]
	if erstes.BlueprintID != 10050 || erstes.ProduktID != 101862104 {
		t.Errorf("das erste Angebot ist %+v", erstes)
	}

	// **217 Cent sind 2,17** — und der Weg geht über Text, nicht über ein
	// `float64`: `cents/100.0` ist bei Geld der Fehler, den 0093 vermeidet.
	if erstes.Preis != 217 || erstes.Waehrung != "EUR" {
		t.Errorf("der Preis ist %d %s, erwartet 217 EUR", erstes.Preis, erstes.Waehrung)
	}

	if erstes.Menge != 3 || erstes.Buendel != 1 {
		t.Errorf("Menge und Bündel sind %d und %d", erstes.Menge, erstes.Buendel)
	}

	// **Die Erhaltung reist als Text** — sie wird nicht auf `domain.Zustand`
	// abgebildet, weil die Skalen nicht ineinander überführbar sind (0095).
	if erstes.Zustand != "Moderately Played" || erstes.Sprache != "en" {
		t.Errorf("Zustand und Sprache sind %q und %q", erstes.Zustand, erstes.Sprache)
	}

	if erstes.Finish != domain.FinishNonfoil {
		t.Errorf("die Ausführung ist %q", erstes.Finish)
	}

	zweites := angebote[1]
	if zweites.Finish != domain.FinishFoil || zweites.Sprache != "de" ||
		!zweites.Gradiert || !zweites.ImUrlaub || !zweites.Signiert {
		t.Errorf("das zweite Angebot ist %+v", zweites)
	}

	if zweites.Preis != 1900 {
		t.Errorf("der zweite Preis ist %d, erwartet 1900", zweites.Preis)
	}

	// **Ein Angebot ohne Währung ist ein Befund** — Pflichtfelder streng (§16).
	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %+v", len(befunde), befunde)
	}
}

// **Ein Angebot wird zu einer Preisbeobachtung — mit allen sieben Teilen.**
//
// Das ist die Naht zu A3 (0093): ein Marktpreis ist eine Beobachtung wie jede
// andere, und ohne Markt, Preisart, Zustand und Sprache wäre seine Reihe nicht
// eindeutig.
func TestEinAngebotWirdEinPreisstand(t *testing.T) {
	t.Parallel()

	angebot := cardtrader.Angebot{
		BlueprintID: 10050, ProduktID: 1, Preis: 217, Waehrung: "EUR",
		Finish: domain.FinishNonfoil, Zustand: "Near Mint", Sprache: "en",
	}

	stand := angebot.AlsStand("11111111-1111-4111-8111-111111111111")

	if stand.Quelle != cardtrader.Familie {
		t.Errorf("die Quelle ist %q", stand.Quelle)
	}

	if stand.Markt != domain.MarktPapier || stand.Art != domain.PreisartVerkauf {
		t.Errorf("Markt und Art sind %q und %q", stand.Markt, stand.Art)
	}

	if stand.Wert != 217 || stand.Waehrung != "EUR" ||
		stand.Zustand != "Near Mint" || stand.Sprache != "en" {
		t.Errorf("der Stand ist %+v", stand)
	}

	if stand.Lage != domain.PreisGenannt {
		t.Errorf("die Lage ist %q", stand.Lage)
	}
}

// **Der Filter steht in der Abfrage — und nur, was gesetzt ist.**
//
// **Ein `foil=false` ist nicht dasselbe wie kein `foil`**: das eine fragt nach
// Nichtfoils, das andere nach allem. Ein Nullwert, der als Filter hinausginge,
// nähme die Hälfte der Angebote weg, ohne dass jemand danach gefragt hätte.
func TestNurDerGesetzteFilterStehtInDerAbfrage(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name     string
		filter   cardtrader.Angebotsfilter
		erwartet string
	}{
		{"ohne Filter", cardtrader.Angebotsfilter{}, "blueprint_id=10050"},
		{
			"nur Sprache",
			cardtrader.Angebotsfilter{Sprache: "de"},
			"blueprint_id=10050&language=de",
		},
		{
			"Nichtfoil ist ein Filter",
			cardtrader.Angebotsfilter{Finish: domain.FinishNonfoil},
			"blueprint_id=10050&foil=false",
		},
		{
			"Foil ebenso",
			cardtrader.Angebotsfilter{Finish: domain.FinishFoil},
			"blueprint_id=10050&foil=true",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			gesehen := make(chan string, 1)

			server := httptest.NewServer(http.HandlerFunc(
				func(schreiben http.ResponseWriter, gestellt *http.Request) {
					gesehen <- gestellt.URL.RawQuery
					_, _ = schreiben.Write([]byte("{}")) //nolint:errcheck // die Gegenstelle
				}))
			t.Cleanup(server.Close)

			if _, _, fehler := klientFuer(server).
				Angebote(t.Context(), 10050, fall.filter); fehler != nil {
				t.Fatalf("Angebote: %v", fehler)
			}

			if abfrage := <-gesehen; abfrage != fall.erwartet {
				t.Errorf("die Abfrage ist %q, erwartet %q", abfrage, fall.erwartet)
			}
		})
	}
}

// **Ohne Blueprint wird nicht gefragt** — `blueprint_id` ist Pflicht (0095).
func TestOhneBlueprintWirdNichtGefragt(t *testing.T) {
	t.Parallel()

	var gefragt bool

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gefragt = true
			_, _ = schreiben.Write([]byte("{}")) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	if _, _, fehler := klientFuer(server).
		Angebote(t.Context(), 0, cardtrader.Angebotsfilter{}); fehler == nil {
		t.Fatal("ohne Blueprint kam eine Antwort")
	}

	if gefragt {
		t.Error("der Server wurde ohne Blueprint gefragt")
	}
}

// **Ein Angebot ohne Blueprint gehört zu keiner Karte** — Werkstück A19 von
// Phase 5.3
// (#1550).
//
// # Der Fund
//
// **Gefunden vom Fuzzer** bei `{"":[{"priCe":{"":0,"CurrenCY":"000"}}]}`: die
// Groß- und Kleinschreibung der Feldnamen ist JSON egal, der Blueprint fehlt,
// und heraus kam ein vollwertiges Angebot `{0, 0, "", 0, "000", …}`. Über
// [Angebot.AlsStand] wäre daraus eine Preisreihe geworden — für einen Druck,
// nach dem niemand gefragt hat.
//
// **Geprüft wurde bis dahin nur die Währung.** Das war die Hälfte: ein Preis
// ohne Währung ist keine Aussage, und ein Preis ohne Karte ebenso wenig.
func TestEinAngebotOhneBlueprintWirdUebergangen(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.Header().Set("Content-Type", "application/json")

			if _, fehler := schreiben.Write([]byte(
				`{"10050":[{"id":7,"price":{"cents":250,"currency":"EUR"}}]}`,
			)); fehler != nil {
				t.Errorf("antworten: %v", fehler)
			}
		}))

	defer server.Close()

	angebote, befunde, fehler := klientFuer(server).Angebote(t.Context(),
		10050, cardtrader.Angebotsfilter{})
	if fehler != nil {
		t.Fatalf("Angebote: %v", fehler)
	}

	if len(angebote) != 0 {
		t.Errorf("ein Angebot ohne Blueprint kam durch: %+v", angebote)
	}

	// **Und es verschwindet nicht still** (§8): der Befund nennt das Angebot
	// und sagt, warum es übergangen wurde.
	if len(befunde) != 1 {
		t.Fatalf("es gab %d Befunde, erwartet war einer", len(befunde))
	}

	if befunde[0].Ort != "7" || !strings.Contains(befunde[0].Meldung, "Blueprint") {
		t.Errorf("der Befund ist %+v", befunde[0])
	}
}
