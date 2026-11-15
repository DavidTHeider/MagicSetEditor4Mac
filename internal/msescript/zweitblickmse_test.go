package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Die Befunde des Zweitblicks über die Portierung, am Quelltext nachgelesen
// (#240).
//
// **`mod` folgt der Zahlenart** — `context.cpp:495`:
//
//	if (at == SCRIPT_DOUBLE || bt == SCRIPT_DOUBLE) fmod(…) else a % b
//
// Hier stand die Begründung, im Pack stehe `mod` „ausschließlich hinter ganzen
// Werten". Das mag stimmen und ist trotzdem keine Begründung für eine andere
// Regel: `10.5 mod 3` ergab 1 statt 1,5, und niemand hätte es gemerkt, bis eine
// Vorlage es tut.
func TestModFolgtDerZahlenart(t *testing.T) {
	faelle := []struct {
		quelle  string
		erwartt float64
	}{
		{`10 mod 3`, 1},
		{`10.5 mod 3`, 1.5},
		{`3 mod 10.5`, 3},

		// Negatives bleibt beim Vorzeichen des linken Werts — so rechnet `%`
		// in C++ wie in Go, und `fmod` ebenso.
		{`-10 mod 3`, -1},
		{`-10.5 mod 3`, -1.5},
	}

	for _, fall := range faelle {
		t.Run(fall.quelle, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.Zahl != fall.erwartt {
				t.Errorf("%v, erwartet war %v", wert.Zahl, fall.erwartt)
			}
		})
	}
}

// **Zahlen vergleichen sich mit einer Toleranz** — `approx_equal` in
// `value.cpp:92`:
//
//	return a == b || fabs(a - b) < 1e-14;
//
// # Warum das kein Feinschliff ist
//
// In den Vorlagen wird gerechnet, bevor verglichen wird: `card_style.name.width
// * 0.1 == 3.75`. Gleitkomma trifft solche Gleichungen nicht auf das letzte
// Bit, und ein Vergleich ohne Toleranz ist dann **manchmal** falsch — die
// unangenehmste Sorte, weil sie von den Eingabewerten abhängt.
func TestZahlenVergleichenSichMitToleranz(t *testing.T) {
	faelle := []struct {
		name    string
		quelle  string
		erwartt bool
	}{
		{"der klassische Fall", `0.1 + 0.2 == 0.3`, true},
		{"gleich bleibt gleich", `3 == 3`, true},
		{"verschieden bleibt verschieden", `3 == 3.5`, false},

		// **Die Toleranz ist absolut und winzig.** `1e-14` — wer sie für einen
		// Freibrief hält, sieht hier, dass sie keiner ist.
		{"knapp daneben ist daneben", `1 == 1.0001`, false},
		{"und `!=` folgt derselben Regel", `0.1 + 0.2 != 0.3`, false},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.Wahrheit != fall.erwartt {
				t.Errorf("%v, erwartet war %v", wert.Wahrheit, fall.erwartt)
			}
		})
	}
}

// **Ein Ersatztext kennt die Fluchtzeichen von sed** — nicht nur `\1` und `&`.
//
// Boost führt für den sed-Modus ausdrücklich `\a \e \f \n \r \t \v`, `\xDD`,
// `\x{DDDD}` und `\cX`; ein Rückstrich vor irgendetwas anderem gibt dieses
// Zeichen wörtlich.
//
// Hier stand „alles übrige steht wörtlich", und das machte aus `\n` den
// Buchstaben `n`. Ein Zeilenumbruch, der als Buchstabe im Kartentext landet,
// ist genau der stille Fehler, den §8 meint.
func TestEinErsatztextKenntDieFluchtzeichenVonSed(t *testing.T) {
	faelle := []struct {
		name    string
		quelle  string
		erwartt string
	}{
		{"Zeilenumbruch", `replace("a", match: "a", replace: "\\n")`, "\n"},
		{"Tabulator", `replace("a", match: "a", replace: "\\t")`, "\t"},
		{"Wagenrücklauf", `replace("a", match: "a", replace: "\\r")`, "\r"},
		{"Seitenvorschub", `replace("a", match: "a", replace: "\\f")`, "\f"},
		{"Vertikaltabulator", `replace("a", match: "a", replace: "\\v")`, "\v"},
		{"Klingel", `replace("a", match: "a", replace: "\\a")`, "\a"},
		{"Escape", `replace("a", match: "a", replace: "\\e")`, "\x1b"},

		{"zweistellig hexadezimal", `replace("a", match: "a", replace: "\\x41")`, "A"},
		{"geklammert hexadezimal", `replace("a", match: "a", replace: "\\x{20AC}")`, "€"},
		{"Steuerzeichen", `replace("a", match: "a", replace: "\\cA")`, "\x01"},

		// **Und der Auffangfall bleibt.** Ein Rückstrich vor etwas anderem gibt
		// das Zeichen — so steht es bei Boost, und `\&` hängt daran.
		{"ein unbekanntes Fluchtzeichen", `replace("a", match: "a", replace: "\\q")`, "q"},
		{"das kaufmännische Und", `replace("a", match: "a", replace: "\\&")`, "&"},

		// **Ein unvollständiges `\x` ist kein Absturz.** Fremde Daten sind der
		// Normalfall; ein Ersatztext, der auf `\x` endet, darf nicht hinter das
		// Ende greifen.
		{"abgeschnittenes Hexadezimal", `replace("a", match: "a", replace: "[\\x]")`, "[x]"},
		{"offene Klammer", `replace("a", match: "a", replace: "[\\x{41]")`, "[x{41]"},

		// **Keine Ziffern, aber eine geschlossene Klammer.** `\x{}` sieht aus
		// wie eine Angabe und ist keine.
		{"leere Klammer", `replace("a", match: "a", replace: "[\\x{}]")`, "[x{}]"},

		// **Keine Hexadezimalziffern.** `strconv` weist das ab, und der
		// Auffangfall gibt das `x`.
		{"keine Ziffern", `replace("a", match: "a", replace: "[\\xzz]")`, "[xzz]"},

		// **Über dem größten Zeichen.** `\x{110000}` liegt jenseits von
		// Unicode; ein Ersatztext aus einem bezogenen Paket ist fremde Eingabe.
		{"jenseits von Unicode", `replace("a", match: "a", replace: "[\\x{110000}]")`, "[x{110000}]"},
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
