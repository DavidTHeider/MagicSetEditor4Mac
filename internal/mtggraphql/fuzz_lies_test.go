package mtggraphql_test

// **Ein GraphQL-Umschlag aus fremder Hand** — Werkstück A19 von Phase 5.3
// (#1550).
//
// # Warum gerade hier
//
// **Weil der Statuscode bei diesem Anbieter nichts entscheidet.** Nachgemessen
// am 2026-09-02: „Access denied" kommt mit **200**, ein Feldfehler mit **400**
// — und beide tragen ihre Auskunft in `errors`. Ein Leser, der den Umschlag
// falsch deutet, meldet dann einen Erfolg, wo eine Absage steht; genau diese
// Klasse sucht der Fuzzer.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - **zweimal dieselbe Eingabe, zwei Ergebnisse** — auf dem Erfolgs- **und**
//     auf dem Ablehnungsweg (Nachlauf über Phase 4.2, #1289)
//   - **Karten und ein Fehler zugleich**
//   - **ein `errors`-Umschlag, der als Erfolg durchkommt** — das ist die
//     Zusage dieses Blatts, und sie ist am teuersten zu verlieren
//
//	go test ./internal/mtggraphql -run Fuzz -fuzz FuzzLies -fuzztime 2m

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtggraphql"
)

// graphsaat sind die Gestalten, mit denen der Fuzzer beginnt.
var graphsaat = []string{
	`{}`,
	`{"data":{"cards":[]}}`,
	`{"data":{"cards":null}}`,
	`{"data":{"cards":[{}]}}`,
	`{"data":{"cards":[{"name":"A","uuid":"u1"}]}}`,
	`{"data":{"cards":[{"name":"A","uuid":"u1","manaValue":1e309}]}}`,
	`{"data":{"cards":[{"name":123,"uuid":"u1"}]}}`,
	`{"errors":[{"message":"Access denied"}]}`,
	`{"errors":[]}`,
	`{"errors":[{}]}`,
	`{"data":{"cards":[{"name":"A","uuid":"u1"}]},"errors":[{"message":"halb"}]}`,
	`[]`,
	`null`,
	``,
}

// **Der Leser bricht nicht ab, und ein `errors`-Umschlag kommt nie als Erfolg
// durch.**
func FuzzLies(f *testing.F) {
	for _, saat := range graphsaat {
		f.Add([]byte(saat))
	}

	for _, datei := range []string{
		"karten.json", "feldfehler.json", "zugriffverweigert.json",
	} {
		// **Der Pfad kommt aus einer festen Liste in dieser Datei** und
		// nicht von außen — `gosec` sieht nur die Veränderliche.
		roh, fehler := os.ReadFile(filepath.Join("testdaten", datei)) //nolint:gosec // eingecheckte Beispiele
		if fehler != nil {
			f.Fatalf("%s lesen: %v", datei, fehler)
		}

		f.Add(roh)
	}

	f.Fuzz(func(t *testing.T, roh []byte) {
		karten, fehler := mtggraphql.Lies(roh)

		if fehler != nil {
			if karten != nil {
				t.Errorf("abgelehnt **und** Karten zurück: %v", fehler)
			}

			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("der Fehler trägt keine Klasse: %v", fehler)
			}

			nochmal, zweiter := mtggraphql.Lies(roh)
			if zweiter == nil || nochmal != nil {
				t.Errorf("zweimal dieselbe Eingabe: erst %v, dann %v",
					fehler, zweiter)
			}

			return
		}

		// **Die Zusage dieses Blatts**: wo `errors` steht, gibt es keine
		// Karten. Ein gefuzzter Umschlag, der beides trägt und durchkommt,
		// ist der Fund, für den es dieses Ziel gibt.
		if bytes.Contains(roh, []byte(`"errors"`)) &&
			!bytes.Contains(roh, []byte(`"errors":[]`)) &&
			!bytes.Contains(roh, []byte(`"errors": []`)) &&
			!bytes.Contains(roh, []byte(`"errors":null`)) &&
			strings.Contains(string(roh), `"message"`) {
			t.Errorf("ein errors-Umschlag kam als Erfolg durch: %q", roh)
		}

		nochmal, zweiter := mtggraphql.Lies(bytes.Clone(roh))
		if zweiter != nil || len(nochmal) != len(karten) {
			t.Errorf("zweimal dieselbe Eingabe: %d Karten, dann %d (%v)",
				len(karten), len(nochmal), zweiter)
		}
	})
}
