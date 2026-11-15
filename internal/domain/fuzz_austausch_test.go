package domain_test

// **Ein Einstellungsaustausch aus fremder Hand** — Werkstück O11 von Phase 5.2
// (#1435).
//
// # Warum
//
// `domain.Zerlege` liest eine Datei, die jemand **mitbringt** — aus einem
// anderen Profil, von einem anderen Rechner, aus einem Ticket. Sie kann alt
// sein, halb geschrieben, aus einer Fassung, die es noch nicht gibt. Die
// Zusage aus O3 ([0079](../../docs/entscheidungen/0079-einstellungsebenen-und-portabler-import.md))
// lautet: **ein unbekanntes Feld ist ein Befund, keine Ablehnung** — die Datei
// bleibt lesbar —, und **eine fremde Fassung ist eine Ablehnung**, weil
// niemand weiß, was darin steht.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - zweimal dieselbe Eingabe, zwei Ergebnisse — auf **beiden** Wegen, auch
//     dem der Ablehnung (§10: zum Ergebnis gehört das Nein)
//   - **ein Austausch und ein Fehler**, oder keins von beidem
//   - **ein Befund ohne Schlüssel und ohne Grund** — er sagt sonst nicht, was
//     an der Datei auffiel (§8)
//   - **ein angenommenes Paar ohne Schlüssel oder mit fremder Ebene**: was
//     `PruefeAustausch` durchlässt, wird geschrieben — und eine Einstellung
//     ohne Namen fände niemand wieder
//
// # Warum der Lauf **beide** Schritte fährt
//
// `Zerlege` ist der Leser und `PruefeAustausch` das Urteil; erst zusammen sind
// sie die Zusage „was von außen kommt, wird geprüft, bevor es gilt". Der erste
// Anlauf prüfte nur den Leser und strich einen leeren Schlüssel an, den der
// zweite Schritt längst verwirft — ein Fund an der falschen Stelle (§10).
//
//	go test ./internal/domain -run Fuzz -fuzz FuzzZerlegeAustausch -fuzztime 2m

import (
	"reflect"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// austauschsaat sind die Gestalten, mit denen der Fuzzer beginnt.
//
// **Aus der Form, die ein Austausch wirklich hat**, dazu die Fälle, die weh
// tun: keine Fassung, eine aus der Zukunft, ein unbekannter Bereich, ein Wert,
// der kein Text ist, tiefe Verschachtelung und ein leerer Schlüssel.
var austauschsaat = []string{
	`{}`,
	`{"fassung":1}`,
	`{"fassung":1,"bereiche":{}}`,
	`{"fassung":1,"bereiche":{"geraet":{"sprache":"de"}}}`,
	`{"fassung":1,"bereiche":{"nutzer":{"sprache":"en"}}}`,
	`{"fassung":1,"bereiche":{"profil":{"listensortierung":"name"}}}`,
	`{"fassung":1,"bereiche":{"arbeitsplatz":{"seitenleiste":"zu"}}}`,
	`{"fassung":0,"bereiche":{}}`,
	`{"fassung":99,"bereiche":{}}`,
	`{"fassung":"eins","bereiche":{}}`,
	`{"fassung":1,"bereiche":{"unbekannt":{"a":"b"}}}`,
	`{"fassung":1,"bereiche":{"geraet":{"":""}}}`,
	`{"fassung":1,"bereiche":{"geraet":{"  ":"x"}}}`,
	`{"fassung":1,"bereiche":{"geraet":{"sprache":42}}}`,
	`{"fassung":1,"bereiche":{"geraet":{"sprache":"de"}},"mehr":1}`,
	`{"fassung":1,"bereiche":{"nutzer":{"sprache":"  "}}}`,
	`{"fassung":1,"bereiche":{"profil":{"sprache":"de"}}}`,
	`[]`,
	`null`,
	``,
}

// **Der Leser bricht nicht ab, und jede Antwort ist eine.**
func FuzzZerlegeAustausch(f *testing.F) {
	for _, saat := range austauschsaat {
		f.Add([]byte(saat))
	}

	f.Fuzz(func(t *testing.T, roh []byte) {
		austausch, befunde, fehler := domain.Zerlege(roh)

		nochmal, zweiteBefunde, zweiterFehler := domain.Zerlege(roh)

		if (fehler == nil) != (zweiterFehler == nil) {
			t.Fatalf("dieselbe Datei einmal so und einmal anders: %v gegen %v",
				fehler, zweiterFehler)
		}

		if fehler != nil {
			// **Abgelehnt heißt: nichts angenommen.**
			if len(austausch.Bereiche) > 0 {
				t.Errorf("abgelehnt (%v) und trotzdem %d Bereiche zurück",
					fehler, len(austausch.Bereiche))
			}

			if fehler.Error() != zweiterFehler.Error() {
				t.Errorf("dieselbe Datei gab zwei Meldungen: %q gegen %q",
					fehler, zweiterFehler)
			}

			return
		}

		if !reflect.DeepEqual(austausch, nochmal) ||
			!reflect.DeepEqual(befunde, zweiteBefunde) {
			t.Errorf("dieselbe Datei gab zwei verschiedene Ergebnisse: "+
				"%d gegen %d Bereiche, %d gegen %d Befunde",
				len(austausch.Bereiche), len(nochmal.Bereiche),
				len(befunde), len(zweiteBefunde))
		}

		for _, befund := range befunde {
			if strings.TrimSpace(befund.Grund) == "" {
				t.Errorf("ein Befund ohne Grund: %+v", befund)
			}

			// **Wo es einen Ort gibt, nennt der Befund ihn** — und wo es
			// keinen gibt, sagt er das.
			//
			// JSON erlaubt `""` als Feldnamen; ein Befund darüber kann keinen
			// Schlüssel nennen, den es gibt. Statt einen zu erfinden, spricht
			// er von der **Datei** — und diese Zusicherung hält beides
			// zusammen, damit ein Befund ohne beides nicht durchrutscht.
			if strings.TrimSpace(befund.Schluessel) == "" &&
				strings.TrimSpace(befund.Ebene) == "" &&
				!strings.Contains(befund.Grund, "Diese Datei") {
				t.Errorf("ein Befund ohne Ort und ohne Wort darüber: %+v", befund)
			}
		}

		// **Das Urteil gehört zum Lesen** — geprüft wird, was danach gilt.
		paare, ausDenZeilen := domain.PruefeAustausch(austausch)

		for _, befund := range ausDenZeilen {
			if strings.TrimSpace(befund.Grund) == "" {
				t.Errorf("ein Befund ohne Grund: %+v", befund)
			}
		}

		for _, paar := range paare {
			if strings.TrimSpace(paar.Schluessel) == "" {
				t.Errorf("ein angenommenes Paar ohne Schlüssel: %+v", paar)
			}

			if !domain.KenntEbene(string(paar.Ebene)) {
				t.Errorf("ein angenommenes Paar auf fremder Ebene: %+v", paar)
			}
		}
	})
}
