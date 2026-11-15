// Die Wortlistenstellen eines Feldwerts
// (#1074).
//
// # Wofür
//
// Ein Typfeld trägt nicht bloß `Legendary Creature`, sondern
// `<word-list-type-en>Legendary Creature</word-list-type-en>` — das
// Spielpaket setzt die Marken in `sub_type_filter` und `super_type_filter`.
// Sie sagen, **welche Auswahlliste zu welchem Stück des Werts gehört**.
//
// Ohne sie gäbe es keine Auswahl: 51 Listen mit 432 Wörtern liegen im Paket,
// und keine Stelle wüsste, welche gemeint ist.
//
// # Was hier nicht steht
//
// **Ob es die genannte Liste gibt.** Das weiß erst, wer das Spielpaket
// gelesen hat; MSE meldet dort `word list type not found`
// (`src/gui/value/text.cpp`, Zeile 1372, Fassung f9d9356 vom 2021-07-06). Hier
// wird gefunden, nicht zugeordnet.

package service

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Wortlistenstelle ist ein Bereich eines rohen Feldwerts mit seiner Liste.
type Wortlistenstelle struct {
	// Liste ist der Name hinter `<word-list-` — `type-en`, `land`, `class-de`.
	Liste string

	// Anfang und Ende spannen den **ganzen** Bereich auf, Marken eingerechnet:
	// `Anfang` steht auf dem `<` der öffnenden, `Ende` hinter dem `>` der
	// schließenden Marke (`WordListPos`, `src/gui/value/text.cpp`, Zeilen
	// 67–78, Fassung f9d9356 vom 2021-07-06).
	Anfang int
	Ende   int

	// Textanfang und Textende spannen den **Inhalt** auf — das, was eine Wahl
	// ersetzt.
	//
	// **MSE rechnet dieselbe Grenze anders aus**: es setzt die Auswahl auf
	// `[start, end)` und lässt `fixSelection` sie über
	// `index_to_cursor`/`cursor_to_index` auf die nächste Textstelle rücken,
	// womit die Marken herausfallen (`TextValueEditor::fixSelection`, Zeilen
	// 1104–1112). Das Ergebnis ist dasselbe, und hier steht es direkt da.
	Textanfang int
	Textende   int

	// Text ist der Inhalt ohne Auszeichnung — `untag(…)`, wie `selection()`
	// ihn zum Vergleich nimmt (Zeile 252).
	Text string
}

// wortlistenmarke ist der Anfang jeder Wortlistenmarke.
const wortlistenmarke = "<word-list-"

// Wortlistenstellen sucht die Wortlistenbereiche eines rohen Feldwerts.
//
// Portiert aus `TextValueEditor::findWordLists`, `src/gui/value/text.cpp`,
// Zeilen 1349–1377, Fassung f9d9356 vom 2021-07-06.
//
// **Eine unvollständige Marke beendet die Suche.** `if (type_end == npos ||
// end == npos) return;` — nicht „überspringen": was danach käme, wird nicht
// mehr angesehen. Das ist streng und richtig: ein Wert mit einer offenen Marke
// ist kaputt, und die Stellen dahinter stünden auf falschen Grenzen.
func Wortlistenstellen(roh string) []Wortlistenstelle {
	var stellen []Wortlistenstelle

	for stelle := strings.Index(roh, wortlistenmarke); stelle >= 0; {
		artende := strings.IndexByte(roh[stelle:], '>')
		schluss := markenschluss(roh, stelle)

		if artende < 0 || schluss < 0 {
			return stellen
		}

		artende += stelle
		ende := schluss + strings.IndexByte(roh[schluss:], '>') + 1

		stellen = append(stellen, Wortlistenstelle{
			Liste:      roh[stelle+len(wortlistenmarke) : artende],
			Anfang:     stelle,
			Ende:       ende,
			Textanfang: artende + 1,
			Textende:   schluss,
			Text:       ohneMarken(roh[artende+1 : schluss]),
		})

		weiter := strings.Index(roh[ende:], wortlistenmarke)
		if weiter < 0 {
			return stellen
		}

		stelle = ende + weiter
	}

	return stellen
}

// markenschluss gibt die Stelle des `<` der zugehörigen schließenden Marke.
//
// Portiert aus `match_close_tag`, `src/util/tagged_string.cpp`, Zeilen 235–253,
// Fassung f9d9356 vom 2021-07-06.
//
// **Verglichen wird die Markenart, nicht der volle Name** ([wortmarkenart]):
// für `<word-list-land>` ist sie `word`. Ein `<word-list-x>` darin öffnet damit
// eine zweite Ebene — so zählt MSE, und wer nur auf `</word-list-land>` sähe,
// fände bei Verschachtelung die falsche.
func markenschluss(roh string, anfang int) int {
	const art = wortmarkenart

	ebene := 1

	for stelle := anfang + len(art) + 2; stelle < len(roh); stelle++ {
		if roh[stelle] != '<' {
			continue
		}

		rest := roh[stelle+1:]

		switch {
		case strings.HasPrefix(rest, "/"+art):
			ebene--
			if ebene == 0 {
				return stelle
			}

			stelle += len(art) + 2
		case strings.HasPrefix(rest, art):
			ebene++
			stelle += len(art) + 1
		}
	}

	return -1
}

// wortmarkenart ist die Markenart jeder Wortlistenmarke.
//
// **`tag_type_at` schneidet am ersten `>` oder `-` ab** (Zeilen 224–228): für
// `<word-list-…>` ist die Art damit immer `word`. Hier steht deshalb die
// Konstante und keine Funktion — die hätte einen Zweig für »keine Art
// gefunden«, den von dieser Stelle aus kein Aufruf betreten kann, und ein
// solcher Zweig ist eine Behauptung darüber, was im Ernstfall geschieht (§3).
const wortmarkenart = "word"

// ohneMarken nimmt die Auszeichnungen aus einem Stück Kartentext.
//
// Portiert aus `untag`, `src/util/tagged_string.cpp`, Zeilen 33–46, Fassung
// f9d9356 vom 2021-07-06.
//
// **Jedes `<` beginnt eine Marke** — anders als in [ohneAuszeichnungen], das
// ein einzelnes `1 < 2` stehen lässt. Beides stimmt an seiner Stelle: dort
// geht es um einen Text, der gezeichnet wird, hier um einen Vergleich mit
// einem Wortlisteneintrag, und MSE vergleicht mit `untag`.
func ohneMarken(text string) string {
	return domain.OhneMarken(text)
}
