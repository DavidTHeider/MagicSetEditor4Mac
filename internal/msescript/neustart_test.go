package msescript_test

// **Nach jedem Treffer sucht MSE2 neu, und die Stelle ist dann der Anfang** —
// Befund 8 des Zweitblicks über Phase 1 von 0.9.0 (E2315, #2315).
//
// `filter_text`, `break_text`, `split_text` und `replace` mit Kontext,
// Ersatzfunktion oder `recursive` laufen bei MSE2 in einer Schleife über
// `regex_search(start, end)` — **ohne** `match_prev_avail`
// (`src/script/functions/regex.cpp`, Zeilen 72–110 und 154–232;
// `src/util/regex.hpp`, Zeile 80; Fassung f9d9356 vom 2021-07-06). Boost hält
// `start` dann für den Anfang des Textes: `^` und `\b` greifen dort wieder,
// ein Rückblick sieht nichts davor. Nur das einfache `replace` geht über
// `regex_replace` und sieht den ganzen Text (`regex.cpp`, Zeilen 135–140).
// Hier liefen alle über eine Suche im ganzen Text.
//
// **Jede Erwartung ist vom Orakel erfragt** (`magicseteditor --cli … -q
// --raw`, Fassung f9d9356, am 2026-10-14; die des Nachlaufs am Abend).

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// wieDasOrakel schreibt einen Wert so, wie MSE2 ihn ausgibt: Text als Text,
// eine Liste als JSON.
func wieDasOrakel(t *testing.T, wert msescript.Wert) string {
	t.Helper()

	if wert.Art != msescript.ArtAbbildung {
		return wert.AlsText()
	}

	teile := make([]string, 0, len(wert.Eintraege))
	for _, eintrag := range wert.Eintraege {
		teile = append(teile, eintrag.Wert.AlsText())
	}

	roh, err := json.Marshal(teile)
	if err != nil {
		t.Fatal(err)
	}

	return string(roh)
}

func TestNachJedemTrefferIstDieStelleDerAnfang(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name, quelle, erwartet string
	}{
		{"filter_text: \\b am Neubeginn", `filter_text("abc", match: "\\b[a-z]")`, "abc"},
		{"break_text: ^ am Neubeginn", `break_text("abc", match: "^.")`, `["a","b","c"]`},
		{"split_text: ^ nur am Anfang", `split_text(",a,b", match: "^,")`, `["","a,b"]`},
		{"replace mit Funktion: ^ am Neubeginn", `replace("WW", match: "^W", replace: {"#"})`, "##"},
		{"break_text: \\b am Neubeginn", `break_text("abab", match: "\\bab")`, `["ab","ab"]`},
		{"replace mit Funktion: der Rückblick sieht nichts davor", `replace("aaa", match: "(?<=a)a", replace: {"#"})`, "a#a"},
		{"filter_text mit Kontext", `filter_text("WaW", match: "^W", in_context: "<match>a")`, "W"},
		{"replace rekursiv: ^ am Neubeginn", `replace("WW", match: "^W", replace: "#", recursive: true)`, "##"},
		{"split_text: ^ nach einem Umbruch", "split_text(\"a\\nb\", match: \"^b\")", `["a\n",""]`},
		{"der Kontext sieht den ganzen Text", `filter_text("aXbX", match: "X", in_context: "^a<match>")`, "X"},
		{"replace mit Kontext: \\b am Neubeginn", `replace("abab", match: "\\bab", in_context: "<match>", replace: "#")`, "##"},
		{"break_text: der Rückblick im eigenen Stück", `break_text("a1b2", match: "(?<=[a-z])\\d")`, `["1","2"]`},
		// **Ein abgelehnter leerer Treffer rückt um ein Zeichen vor**, und am
		// Ende ist Schluss (`regex.cpp`, Zeilen 42–48).
		{"ein abgelehnter leerer Treffer am Ende", `filter_text("a", match: "x*", in_context: "nie<match>")`, ""},
		{"zwischen abgelehnten leeren Treffern", `filter_text("aXa", match: "X?", in_context: "a<match>a")`, "X"},
		// **Ein leerer Treffer hinter dem Anfang rückt vor** (Nachlauf über
		// denselben Zweitblick, Befund 6): MSE2 prüft, ob die Suche
		// weiterkommt, nicht, ob der Treffer leer ist (`regex.cpp`, Zeile
		// 164) — und nach dem Neustart fehlt dem Rückblick das `a`.
		{"filter_text: ein leerer Treffer hinter dem Anfang", `filter_text("ab", match: "(?<=a)(?=b)")`, ""},
		{"break_text: ein leerer Treffer hinter dem Anfang", `break_text("ab", match: "(?<=a)(?=b)")`, `[""]`},
		{"split_text: ein leerer Treffer hinter dem Anfang", `split_text("ab", match: "(?<=a)(?=b)")`, `["a","b"]`},
		{"replace mit Funktion: ein leerer Treffer hinter dem Anfang", `replace("ab", match: "(?<=a)(?=b)", replace: {"#"})`, "a#b"},
		{"filter_text: hinter dem leeren Treffer geht es weiter", `filter_text("abc", match: "(?<=a)(?=b)|c")`, "c"},
		// **Rekursiv mit Text füllt die Vorlage aus dem Treffer selbst**
		// (Befund 7): `results.format` über die Gruppen, erst danach geht es
		// in die Tiefe (`regex.cpp`, Zeilen 96–104).
		{"replace rekursiv: der Rückblick sieht den ganzen Text", `replace("ab", match: "(?<=a)b", replace: "#", recursive: true)`, "a#"},
		{"replace rekursiv: die Gruppen des Treffers", `replace("ab", match: "(a)(b)", replace: "\\2\\1", recursive: true)`, "ba"},
		{"replace rekursiv mit Funktion: ^ am Neubeginn", `replace("WW", match: "^W", replace: {"#"}, recursive: true)`, "##"},
		{"replace rekursiv: eine Gruppe, die es nicht gibt, ist leer", `replace("ab", match: "(a)b", replace: "[\\2]", recursive: true)`, "[]"},
		// **Das einfache `replace` bleibt beim ganzen Text** — der Gegenfall.
		{"replace einfach: ^ nur am Anfang", `replace("WW", match: "^W", replace: "#")`, "#W"},
		{"replace einfach: der Rückblick sieht alles", `replace("aaa", match: "(?<=a)a", replace: "#")`, "a##"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("%v", err)
			}

			if gesehen := wieDasOrakel(t, wert); gesehen != fall.erwartet {
				t.Errorf("ergibt %s, MSE2 gibt %s", gesehen, fall.erwartet)
			}
		})
	}
}

// **Ein angenommener leerer Treffer ist ein Fehler** — Befund 9 desselben
// Zweitblicks. `filter_text` wirft bei MSE2 „Regular expression matches
// empty string" (`regex.cpp`, Zeilen 164–167, vom Orakel bestätigt);
// `break_text`, `split_text` und die Ersatzschleife liefen dort endlos
// (`start` rückt nicht vor). **Hier ist es überall derselbe Fehler** — eine
// Schutzabweichung, die die Portierungskarte nennt.
func TestEinLeererTrefferIstEinFehler(t *testing.T) {
	t.Parallel()

	for _, quelle := range []string{
		`filter_text("a", match: "", in_context: "<match>")`,
		`filter_text("a", match: "x*")`,
		`break_text("ab", match: "x*")`,
		`split_text("ab", match: "x*")`,
		`replace("ab", match: "x*", replace: {"#"})`,
		// **Auch in der Tiefe**: das Ergebnis des ersten Treffers ist leer,
		// und dort steht die Suche — einmal bleibt auch die äußere stehen,
		// einmal nur die innere.
		`replace("aa", match: "a|^", replace: "", recursive: true)`,
		`replace("ab", match: "a|^$", replace: "", recursive: true)`,
	} {
		t.Run(quelle, func(t *testing.T) {
			t.Parallel()

			// **Am Satz und nicht an der Kette**: die Bibliothek gibt einen
			// Fehler als Satz mit Stelle an die Paketautoren weiter, ohne
			// Kette (#1178) — wie jede ihrer Meldungen.
			_, err := msescript.AuswertenQuelle(t.Context(), quelle, msescript.Bibliothek())
			if err == nil || !strings.Contains(err.Error(), msescript.ErrLeererTreffer.Error()) {
				t.Errorf("ergibt %v, erwartet den leeren Treffer als Fehler", err)
			}
		})
	}
}
