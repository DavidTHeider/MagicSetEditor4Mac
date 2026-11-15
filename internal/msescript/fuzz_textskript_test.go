package msescript_test

// **Ein Textskript aus einem Paket** — QA-06c
// (#1976),
// Fläche 15 aus `docs/FUZZZIELE.md`.
//
// # Warum ein eigenes Ziel neben `FuzzAuswerten`
//
// **Ein Textskript wird anders zerteilt**: im Zeichenkettenmodus, in dem ein
// `"` ein Zeichen ist und nur `{…}` rechnet (`AuswertenTextskript`, MSEs
// `StringScript`). `reminder:` eines Schlüsselworts und `default:` eines
// Feldes sind solche — rund jede Karte führt einen. `FuzzAuswerten` erreicht
// diesen Zerteiler nie.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger über die Frist
//   - ein Fehler **und** ein Wert
//   - zweimal dieselbe Quelle, zwei Ergebnisse
//   - **ein Text ohne `{` und ohne `\` kommt verändert zurück** — dort gibt
//     es nichts zu rechnen und nichts zu entwerten
//
// **Hier stand die letzte Zusage ohne Einschränkung, und die erste Kampagne
// zeigte, dass sie zu eng war**: `"\xfa"` kam als `"\uFFFD"` zurück. Das ist
// kein Fund — ein Textskript kommt aus einer Paketdatei, und die weist
// ungültiges UTF-8 schon beim Lesen mit Stelle ab (`msepaket.ErrKodierung`),
// wie MSE2 (`read_utf8_line`, `src/util/io/reader.cpp`, Zeilen 179–183,
// Fassung f9d9356 vom 2021-07-06). Die Zusage gilt deshalb für gültiges
// UTF-8; die Eingabe bleibt als Saat unter `testdata/fuzz/`.
//
//	go test ./internal/msescript -run Fuzz -fuzz FuzzTextskript -fuzztime 2m -fuzzminimizetime 5s

import (
	"context"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func FuzzTextskript(f *testing.F) {
	for _, saat := range []string{
		"",
		"ein Text ohne Ausdruck",
		`{card.name} kann nicht geblockt werden`,
		`{this_or_that("creature", upper:true)} can’t be blocked`,
		`Ein "Zitat" mitten im {card.rarity}`,
		`\{ bleibt Text \}`,
		"{",
		"}{",
		"{{card.name}}",
		"{1/0}",
	} {
		f.Add(saat)
	}

	f.Fuzz(func(t *testing.T, quelle string) {
		wert, fehler := textskript(t, quelle)
		if errors.Is(fehler, context.DeadlineExceeded) {
			return
		}

		nochmal, zweiterFehler := textskript(t, quelle)
		if errors.Is(zweiterFehler, context.DeadlineExceeded) {
			return
		}

		if (fehler == nil) != (zweiterFehler == nil) || !gleicherWert(wert, nochmal) {
			t.Fatalf("dieselbe Quelle, zwei Ergebnisse: %v/%+v gegen %v/%+v",
				fehler, wert.Art, zweiterFehler, nochmal.Art)
		}

		if fehler != nil {
			if wert.Art != msescript.Nichts.Art {
				t.Errorf("Fehler **und** Wert: %v / %+v", fehler, wert.Art)
			}

			return
		}

		if utf8.ValidString(quelle) && !strings.ContainsAny(quelle, `{\`) && wert.AlsText() != quelle {
			t.Errorf("ein Text ohne Ausdruck kam verändert zurück: %q → %q", quelle, wert.AlsText())
		}
	})
}

// textskript wertet einen Textskript mit Frist aus.
func textskript(t *testing.T, quelle string) (msescript.Wert, error) {
	t.Helper()

	ctx, ende := context.WithTimeout(context.Background(), fuzzfrist)
	defer ende()

	return msescript.AuswertenTextskript(ctx, quelle, fuzzumgebung())
}
