package scryfall_test

// **Fremdes JSON aus einem fremden Dienst** —
// #1289.
//
// # Warum
//
// §16 sagt es: **Pflichtfelder streng, Zusatzfelder tolerant.** Ein neues Feld
// beim Anbieter bricht nichts; ein verschwundenes Pflichtfeld bricht laut, und
// ein unbekannter Aufzählungswert wird als **Befund** geführt und nicht
// verschluckt.
//
// Das ist eine Zusage über **fremde** Daten, und fremde Daten sind das, was
// ein Fuzzer erzeugt.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - zweimal dieselbe Eingabe, zwei Ergebnisse
//   - **Erfolg gemeldet, Zustand strukturell ungültig**: eine Karte ohne
//     Kennung oder ohne Namen, obwohl `pflichtfelder` beides verlangt
//   - eine Karte **und** ein Fehler, oder keins von beidem
//
// **Nicht zugesichert**: dass ein Befund entsteht, wo einer entstehen sollte.
// Welcher Aufzählungswert bekannt ist, sagt `aufzaehlungen.go`, und das prüfen
// die Läufe daneben mit benannten Fällen.
//
//	go test ./internal/scryfall -run Fuzz -fuzz FuzzLiesKarte -fuzztime 2m

import (
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// kartensaat sind die Gestalten, mit denen der Fuzzer beginnt.
//
// **Aus der Form, die Scryfall wirklich schickt** — Pflichtfelder, ein
// unbekannter Aufzählungswert, eine mehrseitige Karte, extreme Zahlen, ein
// fehlendes Pflichtfeld.
var kartensaat = []string{
	`{}`,
	`{"id":"a","name":"Lightning Bolt","lang":"en","layout":"normal"}`,
	`{"id":"a","name":"N","lang":"en","layout":"gibtsnicht","rarity":"auchnicht"}`,
	`{"id":"a","name":"N","lang":"en","layout":"transform","card_faces":[{"name":"A"},{"name":"B"}]}`,
	`{"id":"a","name":"N","lang":"en","layout":"normal","cmc":1e309}`,
	`{"id":"a","name":"N","lang":"en","layout":"normal","collector_number":"999999999999999999999"}`,
	`{"name":"ohne Kennung","lang":"en","layout":"normal"}`,
	`{"id":"a","lang":"en","layout":"normal"}`,
	`{"id":"a","name":"Ω 🜂","lang":"de","layout":"normal","image_uris":{"png":"https://x/y.png"}}`,
	`[]`,
	`null`,
	`"text"`,
}

// **Der Kartenleser bricht nicht ab, und was er annimmt, ist vollständig.**
func FuzzLiesKarte(f *testing.F) {
	for _, saat := range kartensaat {
		f.Add([]byte(saat))
	}

	f.Fuzz(func(t *testing.T, roh []byte) {
		karte, befunde, fehler := scryfall.LiesKarte(roh)

		if fehler != nil {
			// **Abgelehnt heißt: keine Karte.** Beides zugleich wäre ein
			// halber Stand — und der Aufrufer nähme die Karte (§8).
			if karte != nil {
				t.Errorf("abgelehnt **und** eine Karte zurück: %v", fehler)
			}

			auchDieAblehnungZweimal(t, roh, fehler)

			return
		}

		if karte == nil {
			t.Fatal("kein Fehler und keine Karte — das ist keine Antwort")
		}

		// **Angenommen heißt: die Pflichtfelder stehen da.** `pflichtfelder`
		// verlangt sie; wer hier eine Karte ohne Kennung bekäme, hätte einen
		// Datensatz, den nichts wiederfindet.
		if karte.Kennung == "" {
			t.Error("eine angenommene Karte ohne Kennung")
		}

		if karte.Name == "" {
			t.Error("eine angenommene Karte ohne Namen")
		}

		// **Und ein Befund ist kein Fehler.** Ein unbekannter
		// Aufzählungswert wird geführt, nicht verschluckt (§16) — dass er
		// *entsteht*, prüfen die Läufe daneben; hier zählt, dass er die Karte
		// nicht umbringt.
		for _, befund := range befunde {
			if befund.Meldung == "" {
				t.Error("ein Befund ohne Meldung")
			}
		}

		gleichBeimZweitenMal(t, roh, karte, befunde)
	})
}

// gleichBeimZweitenMal sichert zu, dass derselbe Text dieselbe Karte gibt.
//
// **Die ganze Karte und alle Befunde**, nicht Kennung und Name.
//
// Der erste Anlauf verglich zwei Felder von vielen. Damit durfte derselbe Text
// zwei fachlich verschiedene Karten geben — andere Farben, andere Seiten,
// andere Befunde —, solange Kennung und Name stimmten; der Kopf versprach
// „zweimal dieselbe Eingabe, zwei Ergebnisse" als Fund und der Rumpf sah zwei
// Felder an. Angestrichen vom Zweitblick über Phase 4.2.
func gleichBeimZweitenMal(t *testing.T, roh []byte, erste *scryfall.Karte,
	ersteBefunde []domain.Befund,
) {
	t.Helper()

	zweite, zweiteBefunde, fehler := scryfall.LiesKarte(roh)
	if fehler != nil {
		t.Fatalf("derselbe Text wurde beim zweiten Mal abgelehnt: %v", fehler)
	}

	if !reflect.DeepEqual(erste, zweite) {
		t.Errorf("derselbe Text gab zwei Karten: %q/%q gegen %q/%q",
			erste.Kennung, erste.Name, zweite.Kennung, zweite.Name)
	}

	if !reflect.DeepEqual(ersteBefunde, zweiteBefunde) {
		t.Errorf("derselbe Text gab %d und dann %d Befunde",
			len(ersteBefunde), len(zweiteBefunde))
	}
}

// auchDieAblehnungZweimal sichert zu, dass eine Ablehnung sich wiederholt.
//
// **Zum Ergebnis gehört auch das Nein.** Der erste Anlauf kehrte bei einem
// Fehler zurück, ohne ein zweites Mal zu lesen — ein Leser, der abwechselnd
// annimmt und ablehnt oder zwei verschiedene Meldungen gibt, wäre damit
// unsichtbar geblieben. Angestrichen vom Nachlauf über Phase 4.2.
func auchDieAblehnungZweimal(t *testing.T, roh []byte, erster error) {
	t.Helper()

	karte, _, zweiter := scryfall.LiesKarte(roh)

	if zweiter == nil {
		t.Errorf("derselbe Text wurde erst abgelehnt (%v) und dann angenommen: %+v",
			erster, karte)

		return
	}

	if zweiter.Error() != erster.Error() {
		t.Errorf("derselbe Text gab zwei Meldungen: %q gegen %q",
			erster, zweiter)
	}
}
