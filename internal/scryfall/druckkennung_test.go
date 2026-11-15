// **Die Kennung eines Drucks** — Werkstück 3 von Phase 3.4
// (#1165).
//
// **Eine eigene Datei, weil sie eine eigene Frage beantwortet** (§2): welche
// Kennung meint denselben Druck? `preise_test.go` prüft, was aus einer Antwort
// gelesen wird; hier steht, welche Anfrage überhaupt zu welcher Antwort passt.
package scryfall_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// **Dieselbe Kennung in anderer Schreibweise ist kein anderer Druck.**
//
// # Warum das neben dem Lauf darüber steht und nicht darin
//
// Er prüft, dass ein **anderer** Druck auffällt. Diese Zusage ist die
// Gegenseite und geht ihm verloren, wenn sie in seinem Fall mitliefe: Scryfall
// **nimmt** eine UUID mit großen Hexadezimalbuchstaben an und antwortet
// kanonisch klein. Ein roher Vergleich meldete dann einen Konflikt zwischen
// zwei Schreibweisen desselben Drucks.
//
// Live nachgemessen vom Nachlauf über Phase 3.4:
// `/cards/516E137A-D9D9-4BA7-89FB-E680DDB2C2D1` antwortet mit
// `516e137a-d9d9-4ba7-89fb-e680ddb2c2d1`.
//
// **Und der Weg wird mitgeprüft.** Die geglättete Kennung muss auch in der
// **Adresse** stehen — sonst führe die Anfrage weiter mit dem rohen Wert
// hinaus, und der Vergleich verstünde sich mit einer Antwort, die er gar nicht
// erfragt hat.
func TestDieselbeKennungInAndererSchreibweiseGehtDurch(t *testing.T) {
	t.Parallel()

	const klein = "516e137a-d9d9-4ba7-89fb-e680ddb2c2d1"

	gefragt := make(chan string, 1)

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			gefragt <- anfrage.URL.Path

			if _, err := schreiben.Write([]byte(
				`{"id":"` + klein + `","prices":{"eur":"1.00"}}`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	preise, _, _, err := klientFuer(t, gegenstelle.URL,
		netz.KlientUeber(gegenstelle.Client())).
		PreiseFuerDruck(t.Context(), "  516E137A-D9D9-4BA7-89FB-E680DDB2C2D1  ")
	if err != nil {
		t.Fatalf("die große Schreibweise wurde abgewiesen: %v", err)
	}

	if len(preise) != 1 {
		t.Errorf("es sind %d Preise", len(preise))
	}

	if weg := <-gefragt; !strings.HasSuffix(weg, "/cards/"+klein) {
		t.Errorf("gefragt wurde %q — will die geglättete Kennung", weg)
	}
}

// **Und die Antwort, die der Dienst heute wirklich schickt, kommt an.**
//
// # Warum das neben dem Lauf über die sieben Felder steht
//
// Der prüft eine **erfundene** Antwort mit allen sieben. Diese hier hat die
// sechs Schlüssel, die am 2026-08-27 jede von acht abgerufenen Antworten
// hatte — `eur`, `eur_foil`, `tix`, `usd`, `usd_etched`, `usd_foil` —, und
// zwei davon stehen auf `null`, wie es bei den meisten Drucken der Fall ist.
//
// **Ohne diesen Lauf prüfte die Vorrichtung eine Gestalt, die es nicht gibt**,
// und die fehlende siebte Zeile fiele als Befund oder als Lücke nie auf.
// Angestrichen vom Nachlauf über Phase 3.4 als Befund 4.
func TestDieHeutigeAntwortgestaltKommtAn(t *testing.T) {
	t.Parallel()

	_, preise, befunde, err := scryfall.LiesPreise([]byte(`{
		"id": "probe-druck",
		"prices": {
			"usd": "4.17", "usd_foil": "12.00", "usd_etched": null,
			"eur": "3.99", "eur_foil": "11.00", "tix": null
		}
	}`))
	if err != nil {
		t.Fatalf("LiesPreise: %v", err)
	}

	if len(befunde) != 0 {
		t.Errorf("ein fehlender Preis ist kein Befund: %+v", befunde)
	}

	// **Vier Preise und nicht sechs.** `null` heißt „kein Preis bekannt", und
	// eine Null wäre eine Aussage, die niemand gemacht hat.
	if len(preise) != 4 {
		t.Fatalf("es sind %d Preise: %+v", len(preise), preise)
	}

	for _, preis := range preise {
		if preis.Finish == domain.FinishEtched {
			t.Errorf("ein geätzter Preis, den die Antwort nicht nennt: %+v", preis)
		}
	}
}
