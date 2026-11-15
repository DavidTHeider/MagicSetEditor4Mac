package scrydex_test

// **Eine Kartenantwort aus fremder Hand** — Werkstück A19 von Phase 5.3
// (#1550).
//
// # Warum gerade hier
//
// **Weil dieses Blatt zwei Absagewege hat und einer im Rumpf steht.** Ein
// Fehler in `error` schlägt den Statuscode — „Invalid credentials" ist eine
// Auskunft, „401" ist eine Zahl. Was geschieht, wenn beide etwas Widersprüch-
// liches sagen, sagt erst ein Fuzzer; und Scrydex ist **nie gegen sein
// Original gelaufen** ([0056](../../docs/entscheidungen/0056-drei-weitere-kartendienste.md)),
// also ist die eingecheckte Gestalt alles, was feststeht.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - **zweimal dieselbe Eingabe, zwei Ergebnisse** — auf dem Erfolgs- **und**
//     auf dem Ablehnungsweg (Nachlauf über Phase 4.2, #1289)
//   - **Karten und ein Fehler zugleich**
//   - **ein Statuscode, der nicht 200 ist, und trotzdem Karten** — das ist die
//     Zusage dieses Blatts
//
//	go test ./internal/scrydex -run Fuzz -fuzz FuzzLiesKarten -fuzztime 2m

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scrydex"
)

// scrydexsaat sind die Gestalten, mit denen der Fuzzer beginnt.
var scrydexsaat = []string{
	`{}`,
	`{"data":[]}`,
	`{"data":null}`,
	`{"data":[{}]}`,
	`{"data":[{"id":"i1","name":"A","expansion":{"code":"PRB"}}]}`,
	`{"data":[{"id":"i1","name":"A","mana_value":1e309}]}`,
	`{"data":[{"id":123,"name":"A"}]}`,
	`{"data":[{"id":"i1","name":"Ω 🜂","rules":["a\nb"]}]}`,
	`{"error":{"message":"Invalid credentials"}}`,
	`{"error":{}}`,
	`{"status":200,"data":[{"id":"i1","name":"A"}],"error":{"message":"halb"}}`,
	`[]`,
	`null`,
	``,
}

// **Der Leser bricht nicht ab, und ein Fehlerstatus bringt keine Karten.**
func FuzzLiesKarten(f *testing.F) {
	for _, saat := range scrydexsaat {
		for _, status := range []int{200, 401, 500} {
			f.Add([]byte(saat), status)
		}
	}

	for _, datei := range []string{"karten.json", "ohnezugang.json"} {
		// **Der Pfad kommt aus einer festen Liste in dieser Datei** und
		// nicht von außen — `gosec` sieht nur die Veränderliche.
		roh, fehler := os.ReadFile(filepath.Join("testdaten", datei)) //nolint:gosec // eingecheckte Beispiele
		if fehler != nil {
			f.Fatalf("%s lesen: %v", datei, fehler)
		}

		f.Add(roh, 200)
	}

	f.Fuzz(func(t *testing.T, roh []byte, status int) {
		karten, fehler := scrydex.LiesKarten(roh, status)

		if fehler != nil {
			if karten != nil {
				t.Errorf("abgelehnt **und** Karten zurück: %v", fehler)
			}

			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("der Fehler trägt keine Klasse: %v", fehler)
			}

			// **Zum Ergebnis gehört auch das Nein.**
			nochmal, zweiter := scrydex.LiesKarten(roh, status)
			if zweiter == nil || nochmal != nil {
				t.Errorf("zweimal dieselbe Eingabe: erst %v, dann %v",
					fehler, zweiter)
			}

			return
		}

		// **Die Zusage dieses Blatts**: ein Status, der nicht 200 ist, kommt
		// nie als Erfolg durch — gleich was im Rumpf steht.
		if status != 200 {
			t.Errorf("Status %d kam als Erfolg durch: %q", status, roh)
		}

		nochmal, zweiter := scrydex.LiesKarten(bytes.Clone(roh), status)
		if zweiter != nil || len(nochmal) != len(karten) {
			t.Errorf("zweimal dieselbe Eingabe: %d Karten, dann %d (%v)",
				len(karten), len(nochmal), zweiter)
		}
	})
}
