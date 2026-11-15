package anthropic_test

// **Eine Antwort von Anthropic aus fremder Hand** — Werkstück A19 von Phase
// 5.3 (#1550).
//
// # Warum über den Transport und nicht über einen Zerteiler
//
// **Weil dieses Blatt keinen exportierten hat** — und das ist richtig so: was
// es liest, liest es aus **seiner** Antwort, samt Kopfzeilen und Statuscode.
// Ein Ziel, das nur den Rumpf nähme, prüfte die Hälfte. Der Preis ist ein
// `httptest`-Server je Durchlauf; er kostet Geschwindigkeit und findet
// denselben Fund.
//
// **Es geht dabei nichts nach draußen**: die Adresse zeigt auf den Testserver,
// und `netz.KlientUeber` nimmt dessen Klienten.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - **zweimal dieselbe Antwort, zwei Ergebnisse** — auf dem Erfolgs- **und**
//     auf dem Ablehnungsweg (Nachlauf über Phase 4.2, #1289)
//   - **Ergebnis und Fehler zugleich**
//   - **ein Befund ohne Meldung oder ohne Stelle** (§8) — `Datei` **oder**
//     `Ort`, denn ein Befund über die ganze Liste gehört keinem Modell
//
//	go test ./internal/anthropic -run Fuzz -fuzz FuzzModelle -fuzztime 2m

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// anthropicsaat sind die Gestalten, mit denen der Fuzzer beginnt.
var anthropicsaat = []string{
	`{}`,
	`{"data":[]}`,
	`{"data":null}`,
	`{"data":[{}]}`,
	`{"data":[{"id":"claude-probe","display_name":"Probe"}]}`,
	`{"data":[{"id":"claude-probe"}],"has_more":true,"last_id":"claude-probe"}`,
	`{"data":[{"id":123}]}`,
	`{"type":"error","error":{"type":"invalid_request_error","message":"nein"}}`,
	`{"error":{}}`,
	`[]`,
	`null`,
	``,
}

// **Der Modellweg bricht nicht ab, und jeder Befund sagt, wo er herkommt.**
func FuzzModelle(f *testing.F) {
	for _, saat := range anthropicsaat {
		f.Add([]byte(saat), 200)
	}

	f.Add([]byte(`{"data":[{"id":"claude-probe"}]}`), 429)
	f.Add([]byte(`{"type":"error","error":{"message":"nein"}}`), 401)

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

		modelle, befunde, fehler := klient.Modelle(t.Context())

		if fehler != nil {
			if modelle != nil {
				t.Errorf("abgelehnt **und** Modelle zurück: %v", fehler)
			}

			// **Zum Ergebnis gehört auch das Nein.**
			nochmal, _, zweiter := klient.Modelle(t.Context())
			if zweiter == nil || nochmal != nil {
				t.Errorf("zweimal dieselbe Antwort: erst %v, dann %v",
					fehler, zweiter)
			}

			return
		}

		for _, befund := range befunde {
			// **Jeder Befund sagt, wo er herkommt — und zwar auf eine von
			// zwei Arten.** `Datei` ist die Modellkennung, `Ort` die Stelle
			// in der Antwort; einer von beiden muss dastehen, sonst sind
			// zwei Befunde über zwei Einträge wortgleich (§8).
			//
			// **Beides zugleich zu verlangen wäre zu streng**: ein Befund
			// über die **Liste** gehört keinem Modell, und eine erfundene
			// Kennung dafür wäre schlimmer als keine.
			if befund.Meldung == "" {
				t.Errorf("ein Befund ohne Meldung: %+v", befund)
			}

			if befund.Datei == "" && befund.Ort == "" {
				t.Errorf("ein Befund ohne Stelle: %+v", befund)
			}
		}

		nochmal, _, zweiter := klient.Modelle(t.Context())
		if zweiter != nil || len(nochmal) != len(modelle) {
			t.Errorf("zweimal dieselbe Antwort: %d Modelle, dann %d (%v)",
				len(modelle), len(nochmal), zweiter)
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
