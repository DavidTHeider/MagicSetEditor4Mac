package mtgjson_test

// **MTG JSON aus fremder Hand** —
// #1289.
//
// # Warum
//
// `mtgjson` liest ein Kartenformat, das ein anderes Werkzeug schreibt. Wer
// eine Datei öffnet, hat sie irgendwoher — und §8 verlangt, dass ein Set mit
// 300 lesbaren und zwei unlesbaren Karten **mit 300 Karten** geöffnet wird und
// zwei benannten Befunden.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - zweimal dieselbe Eingabe, zwei Ergebnisse
//   - **eine Quelle und ein Fehler**, oder keins von beidem
//   - **ein Befund ohne Ort oder ohne Meldung** — ein Befund, der nicht sagt,
//     wo er herkommt, ist keiner (§8: „gibt für jeden Fehler die Stelle her —
//     Datei, Zeile, Feld"). **Und ein Ort, der mit dem Trenner beginnt, sagt
//     genauso wenig**: er nennt das Feld und verschweigt das Set. Der erste
//     Anlauf prüfte nur auf die leere Zeichenkette, und drei von vier
//     Befunderzeugern kamen daran vorbei — angestrichen vom Zweitblick über
//     Phase 4.2
//
//	go test ./internal/mtgjson -run Fuzz -fuzz FuzzLiesInhalt -fuzztime 2m

import (
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// mtgsaat sind die Gestalten, mit denen der Fuzzer beginnt.
//
// **Aus der Form, die MTG JSON wirklich hat**: ein `data`-Objekt mit `cards`,
// dazu die Fälle, die weh tun — fehlende Pflichtfelder, ein falscher Typ, eine
// extreme Zahl, ein leeres Set.
var mtgsaat = []string{
	`{}`,
	`{"data":{}}`,
	`{"data":{"PRB":{"name":"Probe","code":"PRB","cards":[]}}}`,
	`{"data":{"PRB":{"name":"Probe","code":"PRB","cards":[{"name":"A","uuid":"u1"}]}}}`,
	`{"data":{"PRB":{"cards":[{"name":"A","uuid":"u1","manaCost":"{1}{U}"}]}}}`,
	`{"data":{"PRB":{"cards":[{"name":123,"uuid":"u1"}]}}}`,
	`{"data":{"PRB":{"cards":[{"name":"A","uuid":"u1","convertedManaCost":1e309}]}}}`,
	`{"data":{"PRB":{"cards":[{"name":"Ω 🜂","uuid":"u1","text":"a\nb"}]}}}`,
	`{"meta":{"version":"5.2.1","date":"2026-08-29"},"data":{}}`,
	`{"data":{"PRB":{"cards":null}}}`,
	`[]`,
	`null`,
}

// **Der Leser bricht nicht ab, und jeder Befund sagt, wo er herkommt.**
func FuzzLiesInhalt(f *testing.F) {
	zuordnung, da := mtgjson.ZuordnungFuer("magic")
	if !da {
		f.Fatal("die Zuordnung für `magic` fehlt — der Lauf greift daneben")
	}

	for _, saat := range mtgsaat {
		f.Add([]byte(saat))
	}

	f.Fuzz(func(t *testing.T, roh []byte) {
		quelle, fehler := mtgjson.LiesInhalt(roh, "probe.json", zuordnung)

		if fehler != nil {
			if quelle != nil {
				t.Errorf("abgelehnt **und** eine Quelle zurück: %v", fehler)
			}

			// **Zum Ergebnis gehört auch das Nein** (Nachlauf über Phase 4.2).
			nochmal, zweiter := mtgjson.LiesInhalt(roh, "probe.json", zuordnung)
			switch {
			case zweiter == nil:
				t.Errorf("derselbe Inhalt wurde erst abgelehnt (%v) und dann "+
					"angenommen: %d Karten", fehler, kartenzahl(nochmal))

			case zweiter.Error() != fehler.Error():
				t.Errorf("derselbe Inhalt gab zwei Meldungen: %q gegen %q",
					fehler, zweiter)
			}

			return
		}

		if quelle == nil {
			t.Fatal("kein Fehler und keine Quelle — das ist keine Antwort")
		}

		befundePruefen(t, quelle)

		// **Zweimal derselbe Inhalt, dieselbe Quelle** — die ganze, nicht
		// ihre Kartenzahl.
		//
		// Der erste Anlauf verglich nur, wie viele Karten herauskamen. Damit
		// durfte derselbe Text zwei fachlich verschiedene Quellen geben,
		// solange die Zahl stimmte — der Kopf versprach mehr, als der Rumpf
		// hielt (Zweitblick über Phase 4.2).
		nochmal, nochmalFehler := mtgjson.LiesInhalt(roh, "probe.json", zuordnung)
		if nochmalFehler != nil {
			t.Fatalf("derselbe Inhalt wurde beim zweiten Mal abgelehnt: %v",
				nochmalFehler)
		}

		if !reflect.DeepEqual(quelle, nochmal) {
			t.Errorf("derselbe Inhalt gab zwei verschiedene Quellen: "+
				"%d gegen %d Karten, %d gegen %d Befunde",
				kartenzahl(quelle), kartenzahl(nochmal),
				len(quelle.Befunde), len(nochmal.Befunde))
		}
	})
}

// befundePruefen sichert zu, dass jeder Befund seine Stelle nennt.
func befundePruefen(t *testing.T, quelle *mtgjson.Quelle) {
	t.Helper()

	for _, befund := range quelle.Befunde {
		if befund.Meldung == "" {
			t.Error("ein Befund ohne Meldung — er sagt nicht, was los ist")
		}

		if ohneSetImOrt(befund.Ort) {
			t.Errorf("ein Befund ohne Set im Ort: %q — %q (§8)",
				befund.Ort, befund.Meldung)
		}
	}
}

// kartenzahl zählt die Karten über alle Sets einer Quelle.
func kartenzahl(quelle *mtgjson.Quelle) int {
	gezaehlt := 0
	for _, set := range quelle.Sets {
		gezaehlt += len(set.Karten)
	}

	return gezaehlt
}
