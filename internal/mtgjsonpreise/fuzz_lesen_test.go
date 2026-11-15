package mtgjsonpreise_test

// **Eine Preisdatei aus fremder Hand** — Werkstück A19 von Phase 5.3
// (#1550).
//
// # Warum gerade hier
//
// **Weil dieses Blatt strömend liest und zwei Deckel hält.** `AllPrices.json`
// wiegt über ein Gigabyte; die Zusage dagegen sind eine Byte- und eine
// Ständegrenze, und beide sind der **Missbrauchsfall** dieses Blatts. Eine
// Packbombe und ein abgeschnittener Strom sind keine Ausnahme, sondern das,
// wogegen die Grenzen stehen.
//
// **Und das Geld ist Text.** `UseNumber` gibt die Ziffernfolge, wie sie
// dasteht; ein `float64` dazwischen wäre Rundungsdrift in einer Geldrechnung.
// Was bei einer Ziffernfolge geschieht, die niemand erwartet hat, sagt erst
// ein Fuzzer.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - **zweimal dieselbe Eingabe, zwei Ergebnisse** — auf dem Erfolgs- **und**
//     auf dem Ablehnungsweg (Nachlauf über Phase 4.2, #1289)
//   - **Stände und ein Fehler zugleich**
//   - **ein Befund ohne Datei oder ohne Meldung** (§8)
//   - **mehr Stände, als die Grenze erlaubt** — der Deckel ist die Zusage
//
//	go test ./internal/mtgjsonpreise -run Fuzz -fuzz FuzzLies -fuzztime 2m

import (
	"bytes"
	"compress/gzip"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjsonpreise"
)

// diePreisgrenzen sind eng gesetzt — der Fuzzer soll den Deckel treffen.
//
// **Und nicht die Vorgabe**: zwei Gigabyte erreicht kein gefuzzter Strom, und
// eine Grenze, die nie greift, ist keine Zusage, sondern eine Zahl.
var diePreisgrenzen = mtgjsonpreise.Grenzen{
	HoechstensBytes: 4096, HoechstensStaende: 32,
}

// preissaat sind die Gestalten, mit denen der Fuzzer beginnt.
var preissaat = []string{
	`{}`,
	`{"data":{}}`,
	`{"data":null}`,
	`{"meta":{"date":"2026-09-11"},"data":{}}`,
	`{"data":{"u1":{}}}`,
	`{"data":{"u1":{"paper":{}}}}`,
	`{"data":{"u1":{"paper":{"cardkingdom":{"currency":"USD"}}}}}`,
	`{"data":{"u1":{"paper":{"cardkingdom":{"currency":"USD",` +
		`"retail":{"normal":{"2026-09-10":4.17}}}}}}}`,
	`{"data":{"u1":{"paper":{"cardkingdom":{"currency":"USD",` +
		`"retail":{"normal":{"2026-09-10":1e309}}}}}}}`,
	`{"data":{"u1":{"paper":{"cardkingdom":{"currency":"USD",` +
		`"retail":{"normal":{"nichtsdatum":"vier"}}}}}}}`,
	`{"data":{"u1":{"paper":{"cardkingdom":{"currency":123}}}}}`,
	`[]`,
	`null`,
	``,
}

// derCrosswalk ist die Zuordnung, gegen die gefiltert wird.
func derCrosswalk() map[string]string {
	return map[string]string{"u1": "scry-u1", "u2": "scry-u2"}
}

// **Der Leser bricht nicht ab, und der Deckel hält.**
func FuzzLies(f *testing.F) {
	for _, saat := range preissaat {
		f.Add([]byte(saat))
	}

	f.Add([]byte(dieGestalt))

	f.Fuzz(func(t *testing.T, roh []byte) {
		staende, befunde, fehler := mtgjsonpreise.Lies(
			bytes.NewReader(roh), derCrosswalk(), diePreisgrenzen)

		if fehler != nil {
			if staende != nil {
				t.Errorf("abgelehnt **und** Stände zurück: %v", fehler)
			}

			// **Zum Ergebnis gehört auch das Nein.**
			nochmal, _, zweiter := mtgjsonpreise.Lies(
				bytes.NewReader(roh), derCrosswalk(), diePreisgrenzen)
			if zweiter == nil || nochmal != nil {
				t.Errorf("zweimal dieselbe Eingabe: erst %v, dann %v",
					fehler, zweiter)
			}

			return
		}

		if len(staende) > diePreisgrenzen.HoechstensStaende {
			t.Errorf("%d Stände über der Grenze von %d",
				len(staende), diePreisgrenzen.HoechstensStaende)
		}

		for _, befund := range befunde {
			if befund.Datei == "" || befund.Meldung == "" {
				t.Errorf("ein Befund ohne Datei oder Meldung: %+v", befund)
			}
		}

		nochmal, _, zweiter := mtgjsonpreise.Lies(
			bytes.NewReader(roh), derCrosswalk(), diePreisgrenzen)
		if zweiter != nil || len(nochmal) != len(staende) {
			t.Errorf("zweimal dieselbe Eingabe: %d Stände, dann %d (%v)",
				len(staende), len(nochmal), zweiter)
		}
	})
}

// **Und derselbe Strom gepackt** — mit dem Missbrauchsfall dieses Blatts.
//
// # Warum ein eigenes Ziel
//
// **Weil der Deckel auf dem AUSGEPACKTEN Strom liegt** und eine Packbombe
// genau daran scheitern soll: zehn Kilobyte gzip können ein Gigabyte werden,
// und eine Grenze auf der gepackten Seite wäre keine.
//
//	go test ./internal/mtgjsonpreise -run Fuzz -fuzz FuzzLiesGepackt -fuzztime 2m
func FuzzLiesGepackt(f *testing.F) {
	for _, saat := range preissaat {
		f.Add(packe(f, saat))
	}

	f.Add(packe(f, dieGestalt))

	// **Die Packbombe**: fünfzigtausend Nullbytes packen sich auf wenige
	// hundert zusammen und sprengen ausgepackt jede enge Grenze.
	f.Add(packe(f, strings.Repeat("\x00", 50_000)))

	// **Und der abgeschnittene Strom**: ein halbes gzip ist kein gzip.
	ganz := packe(f, dieGestalt)
	f.Add(ganz[:len(ganz)/2])

	f.Fuzz(func(t *testing.T, roh []byte) {
		staende, befunde, fehler := mtgjsonpreise.LiesGepackt(
			bytes.NewReader(roh), derCrosswalk(), diePreisgrenzen)

		if fehler != nil {
			if staende != nil {
				t.Errorf("abgelehnt **und** Stände zurück: %v", fehler)
			}

			return
		}

		if len(staende) > diePreisgrenzen.HoechstensStaende {
			t.Errorf("%d Stände über der Grenze von %d",
				len(staende), diePreisgrenzen.HoechstensStaende)
		}

		for _, befund := range befunde {
			if befund.Datei == "" || befund.Meldung == "" {
				t.Errorf("ein Befund ohne Datei oder Meldung: %+v", befund)
			}
		}
	})
}

// packe macht aus einem Text einen gzip-Strom.
func packe(f *testing.F, text string) []byte {
	f.Helper()

	var puffer bytes.Buffer

	schreiber := gzip.NewWriter(&puffer)

	if _, fehler := schreiber.Write([]byte(text)); fehler != nil {
		f.Fatalf("packen: %v", fehler)
	}

	if fehler := schreiber.Close(); fehler != nil {
		f.Fatalf("schließen: %v", fehler)
	}

	return puffer.Bytes()
}
