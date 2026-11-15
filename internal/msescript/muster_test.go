package msescript_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Die Muster, die RE2 nicht kann — und die in den Vorlagen trotzdem stehen
// (#240).
//
// # Warum das gemessen und nicht geschätzt ist
//
// MSE benutzt `boost::regex` in der **Perl**-Fassung (`src/util/regex.hpp`,
// Zeile 27). Gos `regexp` ist RE2 und kennt weder Vorausblick noch Rückblick
// noch Rückverweise — es garantiert dafür lineare Laufzeit.
//
// Ausgezählt am 2026-08-16 über `M15-Magic-Pack`: von 1.645 verschiedenen
// Mustern brauchen **15** mehr als RE2 kann. Das ist ein Prozent — und dieses
// Prozent schlug 2.280-mal zu, weil eines davon in `magic.mse-game` an zwölf
// Zeilen steht, die jeder der 190 Stile ausführt.
func TestEinVorausblickWirdUebersetzt(t *testing.T) {
	faelle := []struct {
		name    string
		quelle  string
		erwartt bool
	}{
		// **Der gemessene Hauptfall.** So steht er in `magic.mse-game` an den
		// Zeilen 8426 bis 8437 — zwölfmal dasselbe Muster, und jeder Stil lief
		// darüber.
		{"Vorausblick trifft", `match("W|U", match: "(^|\\|)[^ ]+(?=\\|)")`, true},
		{"Vorausblick trifft nicht", `match("W", match: "(^|\\|)[^ ]+(?=\\|)")`, false},

		// `(?!/)` aus `magic-mana-*.mse-symbol-font`.
		{"verneinter Vorausblick", `match("2/W", match: "[0-9]+(?!/)")`, false},
		{"verneinter Vorausblick trifft", `match("2W", match: "[0-9]+(?!/)")`, true},

		// `(?<!/)` — Rückblick, drei Fundstellen.
		{"Rückblick", `match("/X", match: "(?<!/)[X]")`, false},
		{"Rückblick trifft", `match("aX", match: "(?<!/)[X]")`, true},

		// Ein Rückverweis, eine Fundstelle.
		{"Rückverweis", `match("abab", match: "(ab)\\1")`, true},
		{"Rückverweis trifft nicht", `match("abcd", match: "(ab)\\1")`, false},

		// **`(?ix)` — der `x`-Zusatz lässt Leerraum im Muster weg.** Vier
		// Fundstellen in den Symbolschriften, und RE2 kennt ihn nicht.
		{"erweiterte Schreibweise", `match("TK", match: "(?ix) 1/2 | TK ")`, true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.Art != msescript.ArtWahrheitswert {
				t.Fatalf("Art %s", wert.Art)
			}

			if wert.Wahrheit != fall.erwartt {
				t.Errorf("%v, erwartet war %v", wert.Wahrheit, fall.erwartt)
			}
		})
	}
}

// **Ein Ersatztext benutzt Rückverweise — `\1` bis `\9` und `&`.**
//
// Herkunft: `RegexReplacer::apply` in `src/script/functions/regex.cpp` ruft
// `results.format(replacement_string)`, und `Regex::Results::format` in
// `src/util/regex.hpp` Zeile 63 gibt `boost::format_sed` mit. In der
// sed-Fassung ist `\1` die erste Gruppe und `&` der ganze Treffer; alles andere
// steht wörtlich.
//
// # Warum das ein echter Befund ist
//
// Bis hierher ersetzte dieses Programm **wörtlich**: aus
// `replace: "\\1"` wurde die Zeichenfolge `\1` im Kartentext. Ausgezählt am
// 2026-08-16 stehen in `M15-Magic-Pack` **33** verschiedene Ersatztexte mit
// Rückverweis — darunter der Umbau der Erinnerungstexte, der aus
// `\2<blob>\4\3<moved>\1</moved></blob>` besteht.
func TestEinErsatztextKenntRueckverweise(t *testing.T) {
	faelle := []struct {
		name    string
		quelle  string
		erwartt string
	}{
		{
			"die erste Gruppe",
			`replace("abc", match: "a(b)c", replace: "\\1")`, "b",
		},
		{
			"zwei Gruppen getauscht",
			`replace("ab", match: "(a)(b)", replace: "\\2\\1")`, "ba",
		},
		{
			"der ganze Treffer",
			`replace("ab", match: "ab", replace: "[&]")`, "[ab]",
		},

		// **Ein Dollarzeichen bleibt ein Dollarzeichen.** In Go ist `$1` die
		// Gruppe; in sed ist es Text. Ein Kartentext mit `$` ginge sonst kaputt.
		{
			"ein Dollarzeichen ist Text",
			`replace("a", match: "a", replace: "$1")`, "$1",
		},

		// **Ein geflohenes `&` ist Text**, so steht es in der sed-Fassung.
		{
			"geflohenes kaufmännisches Und",
			`replace("a", match: "a", replace: "\\&")`, "&",
		},

		// Eine Gruppe, die nicht getroffen hat, ist leer statt ein Fehler.
		{
			"eine leere Gruppe",
			`replace("ac", match: "a(b)?c", replace: "[\\1]")`, "[]",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartt {
				t.Errorf("%q, erwartet war %q", wert.AlsText(), fall.erwartt)
			}
		})
	}
}

// **Ein Vorausblick geht auch durch `replace`, `split_text` und `filter_text`.**
//
// Die Übersetzung sitzt an **einer** Stelle, nicht an fünf. Ohne diesen Test
// wäre nicht belegt, dass alle fünf sie benutzen — und `match` allein
// umzustellen hätte die zwölf Zeilen aus `magic.mse-game` nicht bewegt, denn
// dort steht `replace`.
func TestJedeMusterfunktionKenntDenVorausblick(t *testing.T) {
	faelle := []struct {
		name    string
		quelle  string
		erwartt string
	}{
		{"replace", `replace("W|U", match: "(^|\\|)[^ ]+(?=\\|)", replace: "X")`, "X|U"},
		{"filter_text", `filter_text("W|U", match: "[^|]+(?=\\|)")`, "W"},
		{"split_text", `to_text(split_text("a1b", match: "[a-z](?=1)"))`, "1b"},
		{"break_text", `to_text(break_text("W|U", match: "[^|]+(?=\\|)"))`, "W"},

		// **Rückverweis im Ersatz *und* Vorausblick im Muster** — beides
		// zugleich, damit die Rückfalllinie ihre eigenen Gruppen liefert und
		// nicht nur die von RE2.
		{
			"Rückverweis über die Rückfalllinie",
			`replace("aXb", match: "(a)(z)?(X)(?=b)", replace: "[\\1|\\2|\\3]")`, "[a||X]b",
		},

		// **Mehrbytige Zeichen um den Treffer herum.** `regexp2` zählt in
		// Zeichen, `regexp` in Bytes; wer das verwechselt, schneidet einen
		// Umlaut in der Mitte durch.
		{
			"Umlaute um den Treffer",
			`replace("äöX|ü", match: "X(?=\\|)", replace: "Y")`, "äöY|ü",
		},
		{
			"Umlaute beim Zerlegen",
			`to_text(split_text("äö1ü", match: "[äö]+(?=1)"))`, "1ü",
		},

		// **Eine Gruppe, die es gar nicht gibt.** Ein Ersatztext darf auf `\9`
		// zeigen, wo das Muster zwei Gruppen hat — sed setzt dort nichts ein.
		// Ohne diesen Fall bliebe der Zweig in `Ersetze` unbetreten, und was er
		// im Ernstfall täte, wäre eine Behauptung.
		{
			"eine Gruppennummer über dem Muster",
			`replace("aXb", match: "(a)(X)(?=b)", replace: "[\\9]")`, "[]b",
		},
	}

	// **`position` steht hier nicht, und das ist nachgesehen.** Der erste
	// Anlauf führte sie mit — sie liegt in derselben Datei wie `break_text` und
	// sah nach einer Musterfunktion aus. `position_of` in
	// `src/script/functions/basic.cpp:524` von MSE2 nimmt aber `of`, `in`,
	// `order_by` und `filter`; ein `match:` kennt sie nicht. Ein Test darüber
	// prüfte eine Zusage, die weder MSE noch dieses Programm gibt.
	//
	// Fassung f9d9356 vom 2021-07-06.

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartt {
				t.Errorf("%q, erwartet war %q", wert.AlsText(), fall.erwartt)
			}
		})
	}
}

// **Ein Muster, das in keiner Sprache aufgeht, bleibt ein Fehler.**
//
// Die Gegenprobe zur Rückfalllinie: ohne sie hieße „RE2 kann es nicht, also
// nimm den anderen" in Wahrheit „nimm irgendetwas", und eine offene
// Zeichenklasse liefe stillschweigend durch.
func TestEinKaputtesMusterBleibtEinFehler(t *testing.T) {
	_, err := msescript.AuswertenQuelle(t.Context(),
		`match("x", match: "[")`, msescript.Bibliothek())
	if err == nil {
		t.Fatal("kein Fehler — ein kaputtes Muster ging durch")
	}

	if !strings.Contains(err.Error(), "[") {
		t.Errorf("die Meldung nennt das Muster nicht: %v", err)
	}
}

// **Ein POSIX-Kollationselement benennt ein Zeichen — `[[.quotation-mark.]]`.**
//
// MSE übersetzt seine Muster mit `wxRegEx` im **erweiterten POSIX**-Modus, und
// dort steht `[[.name.]]` für das Zeichen mit diesem Namen. Gos `regexp` kennt
// die Schreibweise nicht und meldet
// `[n-m] range in reverse order` — 96 Befunde über die 190 Stile von
// `M15-Magic-Pack`, alle an derselben Zeile von `magic.mse-game`:
//
//	replace(match: "[[.quotation-mark.]]|”", replace: …)
//
// **Übersetzt wird nur, was im Pack vorkommt.** Eine vollständige Tabelle der
// POSIX-Namen wäre eine Tabelle ohne Prüfung; was hier steht, ist gemessen
// (#270).
func TestEinKollationselementBenenntEinZeichen(t *testing.T) {
	muster, err := msescript.UebersetzeMuster(`[[.quotation-mark.]]|”`)
	if err != nil {
		t.Fatalf("übersetzen: %v", err)
	}

	faelle := map[string]bool{
		`sagte "so"`: true,
		"sagte ”so”": true,
		"ohne alles": false,
	}

	for text, erwartet := range faelle {
		t.Run(text, func(t *testing.T) {
			passt, fehler := muster.Passt(text)
			if fehler != nil {
				t.Fatalf("vergleichen: %v", fehler)
			}

			if passt != erwartet {
				t.Errorf("%q passt %v, erwartet %v", text, passt, erwartet)
			}
		})
	}
}

// **Ein unbekannter Name bleibt stehen — und trifft dann nichts.**
//
// Das ist die Abweichung, und sie steht hier, damit niemand die Tabelle für
// vollständig hält: die Rückfalllinie nimmt `[[.gibtsnicht.]]` an und trifft
// damit **keinen einzigen Text**. MSE träfe das Zeichen mit diesem Namen.
//
// **Nichts zu treffen ist der harmlosere Fehlschlag**: eine Ersetzung
// unterbleibt, statt an der falschen Stelle zuzuschlagen. Sichtbar ist er
// trotzdem nicht, und deshalb steht er hier und in `docs/FORMATE.md`.
//
// Eine vollständige POSIX-Tabelle wäre eine Tabelle ohne Prüfung; was in
// `musterkollation.go` steht, ist im Pack ausgezählt (§10).
func TestEinUnbekanntesKollationselementTrifftNichts(t *testing.T) {
	muster, err := msescript.UebersetzeMuster(`[[.gibtsnicht.]]`)
	if err != nil {
		t.Fatalf("übersetzen: %v", err)
	}

	for _, text := range []string{"g", "[", ".", "gibtsnicht"} {
		passt, fehler := muster.Passt(text)
		if fehler != nil {
			t.Fatalf("vergleichen: %v", fehler)
		}

		if passt {
			t.Errorf("%q wird getroffen — dann ist die Abweichung eine andere "+
				"als beschrieben", text)
		}
	}
}
