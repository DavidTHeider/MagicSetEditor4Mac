// **Wohin der Finger zeigt, wenn eine Klammer nicht schließt** — Werkstück B
// von Phase 4.9
// (#1324).
//
// Von `zerteiler_stellen_test.go` abgetrennt, weil dort steht, wohin der
// Finger bei einem **fehlenden Wert** zeigt. Hier geht es um die andere
// Klasse: der Wert steht da, die Klammer um ihn nicht.
package msescript_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Eine nicht geschlossene Klammer wird an ihrer **öffnenden** gemeldet.
//
// # Warum das die richtige Stelle ist
//
// Bis hierher meldete dieses Programm sie dort, wo der Zerteiler stehen
// blieb — am Ende der Quelle. Das ist die Stelle, an der der **Zerteiler**
// merkt, dass etwas fehlt, nicht die, an der jemand etwas zu tun hat. Bei
// einem Skript über zwanzig Zeilen sagt sie nichts: die Klammer, die niemand
// geschlossen hat, steht irgendwo darüber.
//
// MSE2 meldet an der öffnenden Klammer und nennt sie:
//
//	errors.push_back(ScriptParseError(open_pos, error_pos, …,
//	                                  opening->value, expected, peek(0).value));
//
// `src/script/parser.cpp`, Zeilen 335–336 in `TokenIterator::expected`,
// Fassung f9d9356 vom 2021-07-06. Der Satz lautet dort „Expected closing ')'
// for this '(' instead of 'end of input'", und `start` ist `open_pos`.
//
// # Die Zahlen sind erfragt, nicht erschlossen
//
// Jede Erwartung unten ist am Orakel gemessen (§15):
//
//	printf 'f(1, 2\n{1 + 2\n[1: 2\nx[3\n' | magicseteditor --cli SET -q --raw
//
// gibt `(1)`, `(0)`, `(0)`, `(1)` — Zeichenstellen ab null. Dieses Programm
// zählt Zeile und Spalte ab eins, deshalb steht hier eine Spalte mehr.
func TestEineOffeneKlammerWirdAnIhremAnfangGemeldet(t *testing.T) {
	faelle := []struct {
		name    string
		quelle  string
		zeile   int
		spalte  int
		oeffner string
		schluss string
	}{
		{
			name: "ein Aufruf ohne schließende Klammer", quelle: "f(1, 2",
			zeile: 1, spalte: 2, oeffner: "(", schluss: ")",
		},
		{
			name: "eine Gruppierung ohne schließende Klammer", quelle: "x := (1 + 2",
			zeile: 1, spalte: 6, oeffner: "(", schluss: ")",
		},
		{
			name: "ein Funktionsrumpf ohne schließende Klammer", quelle: "{1 + 2",
			zeile: 1, spalte: 1, oeffner: "{", schluss: "}",
		},
		{
			name: "eine Abbildung ohne schließende Klammer", quelle: "[1: 2",
			zeile: 1, spalte: 1, oeffner: "[", schluss: "]",
		},
		{
			name: "ein Index ohne schließende Klammer", quelle: "x[3",
			zeile: 1, spalte: 2, oeffner: "[", schluss: "]",
		},
		{
			name: "eine Liste ohne schließende Klammer", quelle: "[1, 2, 3",
			zeile: 1, spalte: 1, oeffner: "[", schluss: "]",
		},
		{
			name: "eine Zuweisung im Rumpf", quelle: "{a := 1",
			zeile: 1, spalte: 1, oeffner: "{", schluss: "}",
		},
		// **Der Fall, für den es die Änderung gibt.** Über mehrere Zeilen ist
		// die Stelle am Quellende wertlos: sie steht drei Zeilen unter der
		// Klammer, die niemand geschlossen hat.
		{
			name:   "eine Klammer drei Zeilen weiter oben",
			quelle: "f(\n  1,\n  2\n",
			zeile:  1, spalte: 2, oeffner: "(", schluss: ")",
		},
		{
			name:   "eine geschachtelte Klammer nennt die innere",
			quelle: "f(g(1)\n",
			zeile:  1, spalte: 2, oeffner: "(", schluss: ")",
		},
		// **Ein Komma am Ende ist kein fehlender Wert, sondern eine offene
		// Klammer.** Erfragt, nicht erschlossen: `f(1,` gibt am Orakel
		// `(1): Expected closing ')' for this '('`, nicht
		// `Expected 'expression'`.
		{
			name: "ein Komma am Quellende", quelle: "f(1,",
			zeile: 1, spalte: 2, oeffner: "(", schluss: ")",
		},
		{
			name: "ein Komma am Quellende einer Abbildung", quelle: "[1,",
			zeile: 1, spalte: 1, oeffner: "[", schluss: "]",
		},
		{
			name: "eine leere offene Klammer", quelle: "f(",
			zeile: 1, spalte: 2, oeffner: "(", schluss: ")",
		},
		// **Und die Stelle ist die öffnende, auch wenn die Quelle weitergeht.**
		// Am Orakel gemessen, alle drei mit der Stelle der öffnenden Klammer:
		//
		//	(1+2]   → (0): Expected closing ')' for this '(' instead of ']'
		//	(1+2}   → (0): Expected closing ')' for this '(' instead of '}'
		//	(a b)   → (0): Expected closing ')' for this '(' instead of 'b'
		//
		// Der dritte ist der lehrreiche: die Klammer **schließt** dort, und
		// trotzdem ist sie der Befund — was den Ausdruck beendet hat, ist
		// nicht das Problem, sondern dass danach kein `)` steht.
		{
			name: "die falsche schließende Klammer", quelle: "(1+2]",
			zeile: 1, spalte: 1, oeffner: "(", schluss: ")",
		},
		{
			name: "eine geschweifte statt einer runden", quelle: "(1+2}",
			zeile: 1, spalte: 1, oeffner: "(", schluss: ")",
		},
		{
			name: "zwei Werte in einer Gruppierung", quelle: "(a b)",
			zeile: 1, spalte: 1, oeffner: "(", schluss: ")",
		},
		{
			name: "zwei Werte in einem Funktionsrumpf", quelle: "{a b}",
			zeile: 1, spalte: 1, oeffner: "{", schluss: "}",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, err := msescript.Zerteile(fall.quelle)

			var befund *msescript.Fehler
			if !errors.As(err, &befund) {
				t.Fatalf("Fehler %v, erwartet war ein *msescript.Fehler", err)
			}

			if befund.Zeile != fall.zeile || befund.Spalte != fall.spalte {
				t.Errorf(
					"Stelle Zeile %d, Spalte %d, erwartet Zeile %d, Spalte %d (%s)",
					befund.Zeile, befund.Spalte, fall.zeile, fall.spalte, befund.Meldung,
				)
			}

			if befund.Schluessel != msescript.SchluesselKlammerOffen {
				t.Errorf("Schlüssel %q, erwartet %q", befund.Schluessel, msescript.SchluesselKlammerOffen)
			}

			// Die Meldung nennt **beide** Klammern: die offene, damit klar
			// ist, welche gemeint ist, und die fehlende, damit klar ist, was
			// zu tun ist.
			if !strings.Contains(befund.Meldung, "`"+fall.oeffner+"`") {
				t.Errorf("Meldung %q nennt die öffnende %q nicht", befund.Meldung, fall.oeffner)
			}

			if !strings.Contains(befund.Meldung, "`"+fall.schluss+"`") {
				t.Errorf("Meldung %q nennt die fehlende %q nicht", befund.Meldung, fall.schluss)
			}
		})
	}
}

// Und die andere Klasse bleibt, wo sie war.
//
// **`x[` ist kein Klammerfehler.** MSE2 liest hinter `[` einen Ausdruck und
// meldet dessen Fehlen, bevor es die schließende Klammer erwartet:
//
//	} else if (minPrec <= PREC_FUN && token == _("[")) {
//	  parseOper(input, script, PREC_ALL);
//	  …
//	  expectToken(input, _("]"), &token);
//
// `src/script/parser.cpp`, Zeilen 800–820, Fassung f9d9356 vom 2021-07-06.
// Am Orakel gibt `x[` entsprechend `(2): Expected 'expression' instead of
// 'end of input'` — die Stelle ist das Quellende, nicht das `[`.
//
// **Dieser Test steht hier, weil eine Änderung eine Grenze hat.** Ohne ihn
// wüsste niemand, ob die neue Meldung an der richtigen Klasse haltmacht oder
// alles verschluckt, was nach einer Klammer aussieht.
func TestWasKeinKlammerfehlerIstBleibtWoEsWar(t *testing.T) {
	faelle := []struct {
		name       string
		quelle     string
		spalte     int
		schluessel msescript.Meldungsschluessel
	}{
		{
			name: "ein Index ohne Ausdruck", quelle: "x[",
			spalte: 3, schluessel: msescript.SchluesselFehltHier,
		},
		{
			name: "ein Argument, das keines ist", quelle: "f(*)",
			spalte: 3, schluessel: msescript.SchluesselFehltHier,
		},
		{
			name:   "eine schließende Klammer, wo ein Wert stehen müsste",
			quelle: "f(1, 2]",
			spalte: 7, schluessel: msescript.SchluesselFehltHier,
		},
		{
			name: "eine leere runde Klammer", quelle: "()",
			spalte: 2, schluessel: msescript.SchluesselFehltHier,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, err := msescript.Zerteile(fall.quelle)

			var befund *msescript.Fehler
			if !errors.As(err, &befund) {
				t.Fatalf("Fehler %v, erwartet war ein *msescript.Fehler", err)
			}

			if befund.Schluessel != fall.schluessel {
				t.Errorf("Schlüssel %q, erwartet %q (%s)", befund.Schluessel, fall.schluessel, befund.Meldung)
			}

			if befund.Spalte != fall.spalte {
				t.Errorf("Spalte %d, erwartet %d (%s)", befund.Spalte, fall.spalte, befund.Meldung)
			}
		})
	}
}
