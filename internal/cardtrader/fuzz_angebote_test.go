package cardtrader_test

// **Eine Marktplatzantwort aus fremder Hand** — Werkstück A19 von Phase 5.3
// (#1550).
//
// # Warum gerade die Angebote
//
// **Weil dort Geld steht.** Die Stammdaten sind Namen und Kennungen; ein
// Angebot trägt Preis, Währung, Menge und Zustand — und was ein Blatt mit
// einer Zahl macht, die niemand erwartet hat, ist die Frage, die A3 teuer
// bezahlt hat ([0093](../../docs/entscheidungen/0093-eine-preisreihe-hat-sieben-teile.md)).
//
// **Und weil die Antwort eine Karte ist, keine Liste**: der Rumpf ordnet die
// Angebote nach Blueprint, und ein Schlüssel, der keine Zahl ist, ist der
// erste Fall, den niemand von Hand schreibt.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - **zweimal dieselbe Antwort, zwei Ergebnisse** — auf dem Erfolgs- **und**
//     auf dem Ablehnungsweg (Nachlauf über Phase 4.2, #1289)
//   - **Angebote und ein Fehler zugleich**
//   - **ein Angebot ohne Blueprint** — es gehörte dann zu keiner Karte
//   - **ein Befund ohne Meldung oder ohne Stelle** (§8)
//
//	go test ./internal/cardtrader -run Fuzz -fuzz FuzzAngebote -fuzztime 2m

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/cardtrader"
)

// cardtradersaat sind die Gestalten, mit denen der Fuzzer beginnt.
var cardtradersaat = []string{
	`{}`,
	`{"10050":[]}`,
	`{"10050":null}`,
	`{"10050":[{}]}`,
	`{"10050":[{"id":1,"blueprint_id":10050,"quantity":3,` +
		`"price":{"cents":250,"currency":"EUR"}}]}`,
	`{"10050":[{"id":1,"blueprint_id":10050,"price":{"cents":1e309}}]}`,
	`{"nichtszahl":[{"id":1}]}`,
	`{"10050":[{"id":1,"blueprint_id":99999}]}`,
	`[]`,
	`null`,
	``,
}

// **Der Angebotsweg bricht nicht ab, und jedes Angebot gehört zu einer
// Karte.**
func FuzzAngebote(f *testing.F) {
	for _, saat := range cardtradersaat {
		f.Add([]byte(saat), 200)
	}

	// **Der Pfad steht fest in dieser Datei** — `gosec` sieht nur die
	// Veränderliche.
	roh, fehler := os.ReadFile( //nolint:gosec // eingechecktes Beispiel
		filepath.Join("testdaten", "angebote.json"))
	if fehler != nil {
		f.Fatalf("angebote.json lesen: %v", fehler)
	}

	f.Add(roh, 200)
	f.Add(roh, 429)

	gegenstelle := &probeserver{}

	server := httptest.NewServer(gegenstelle)
	f.Cleanup(server.Close)

	f.Fuzz(func(t *testing.T, rumpf []byte, status int) {
		// **Ein Status außerhalb des Gültigen bricht `httptest` ab** — das
		// wäre ein Fund über den Testserver und keiner über dieses Blatt.
		if status < 100 || status > 599 {
			t.Skip()
		}

		gegenstelle.setze(rumpf, status)

		klient := klientFuer(server)

		angebote, befunde, fehler := klient.Angebote(t.Context(), 10050,
			cardtrader.Angebotsfilter{})

		if fehler != nil {
			if angebote != nil {
				t.Errorf("abgelehnt **und** Angebote zurück: %v", fehler)
			}

			// **Zum Ergebnis gehört auch das Nein.**
			nochmal, _, zweiter := klient.Angebote(t.Context(), 10050,
				cardtrader.Angebotsfilter{})
			if zweiter == nil || nochmal != nil {
				t.Errorf("zweimal dieselbe Antwort: erst %v, dann %v",
					fehler, zweiter)
			}

			return
		}

		for _, befund := range befunde {
			if befund.Meldung == "" {
				t.Errorf("ein Befund ohne Meldung: %+v", befund)
			}

			if befund.Datei == "" && befund.Ort == "" {
				t.Errorf("ein Befund ohne Stelle: %+v", befund)
			}
		}

		for _, angebot := range angebote {
			// **Ein Angebot ohne Blueprint gehört zu keiner Karte** — und
			// was zu keiner Karte gehört, hat in einer Preisantwort nichts
			// verloren.
			if angebot.BlueprintID == 0 {
				t.Errorf("ein Angebot ohne Blueprint: %+v", angebot)
			}
		}

		nochmal, _, zweiter := klient.Angebote(t.Context(), 10050,
			cardtrader.Angebotsfilter{})
		if zweiter != nil || len(nochmal) != len(angebote) {
			t.Errorf("zweimal dieselbe Antwort: %d Angebote, dann %d (%v)",
				len(angebote), len(nochmal), zweiter)
		}
	})
}

// gegenstelle ist **ein** Testserver je Fuzzprozess.
//
// # Warum nicht einer je Ausführung
//
// **Weil zehn Arbeiter mit hunderttausend Ausführungen je Minute die Ports
// aufbrauchen.** Gemessen am 2026-09-13: `dial tcp: connect: resource
// temporarily unavailable`, und der Lauf meldete daraufhin „zweimal dieselbe
// Antwort, zwei Ergebnisse" — ein Fund über das Betriebssystem und keiner
// über dieses Blatt. Ein geteilter Server nimmt den Druck weg **und** macht
// das Ziel um Größenordnungen schneller.
//
// **Die Ausführungen eines Arbeiters laufen nacheinander** (kein
// `t.Parallel` im Fuzzrumpf); das Schloss hält die Stelle trotzdem, denn
// `net/http` bedient jede Anfrage in einer eigenen Goroutine.
type probeserver struct {
	schloss sync.Mutex
	rumpf   []byte
	status  int
}

// setze legt fest, was die Gegenstelle als Nächstes antwortet.
func (p *probeserver) setze(rumpf []byte, status int) {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	p.rumpf, p.status = rumpf, status
}

// ServeHTTP antwortet mit dem zuletzt gesetzten Rumpf.
func (p *probeserver) ServeHTTP(schreiben http.ResponseWriter, _ *http.Request) {
	p.schloss.Lock()
	rumpf, status := p.rumpf, p.status
	p.schloss.Unlock()

	schreiben.Header().Set("Content-Type", "application/json")
	schreiben.WriteHeader(status)

	// **Nicht jeder Status darf einen Rumpf tragen.** 1xx, 204 und 304
	// tragen keinen — `net/http` weist das Schreiben ab, und das ist eine
	// Eigenschaft von HTTP und keine dieses Blatts. Der Fuzzer hat genau
	// diese Stati gefunden; was er damit prüft, ist der Fall »Status ohne
	// Rumpf«, und den gibt es wirklich (ein Anbieter darf mit 204 antworten).
	if status < 200 || status == 204 || status == 304 {
		return
	}

	_, _ = schreiben.Write(rumpf) //nolint:errcheck // die Gegenstelle
}
