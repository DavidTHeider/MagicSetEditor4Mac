package msescript_test

// **Die sechs Rechenfunktionen und zwei Umwandlungen** — Werkstück A von
// Phase 4.9 (#1324).
//
// # Woher die erwarteten Werte stammen
//
// **Erfragt, nicht erschlossen** (§15). Jede Zeile unten ist am Orakel
// gemessen — `magicseteditor --cli <SET>.mse-set -q --raw`, Fassung f9d9356
// vom 2021-07-06, am 2026-09-02:
//
//	to_string(abs(-3))              → "3"
//	to_string(abs(-3.5))            → "3.5"
//	to_string(pow(2, exponent: 10)) → "1024"
//	to_string(sqrt(9))              → "3"
//	to_string(exp(0))               → "1"
//	to_string(log(1))               → "0"
//	to_string(log10(100))           → "2"
//	to_string(to_boolean(0))        → "false"
//	to_string(to_boolean("yes"))    → "true"
//	to_code("a")                    → "\"a\""
//	to_code(3)                      → "3"
//
// # Die Quellen
//
// `src/script/functions/basic.cpp`, Fassung f9d9356 vom 2021-07-06:
//
//	SCRIPT_FUNCTION(log10) {
//	  SCRIPT_PARAM_C(double, input);
//	  SCRIPT_RETURN(log(input) / log(10.0));
//	}
//
// Zeilen 298–301. **Nicht `log10(x)`, sondern `log(x)/log(10)`** — bei den
// Zahlen, um die es hier geht, dasselbe; aufgeschrieben, weil es dasteht.
//
//	SCRIPT_FUNCTION(abs) {
//	  ScriptValueP input = ctx.getVariable(SCRIPT_VAR_input);
//	  ScriptType t = input->type();
//	  if (t == SCRIPT_DOUBLE) { SCRIPT_RETURN(fabs(input->toDouble())); }
//	  else                    { SCRIPT_RETURN(abs(input->toInt())); }
//	}
//
// Zeilen 238–246. **Eine ganze Zahl bleibt ganz** — deshalb `"3"` und nicht
// `"3.0"`.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestRechnenUndUmwandeln(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, quelle, erwartet string }{
		{"abs einer negativen ganzen Zahl", `abs(-3)`, "3"},
		{"abs einer negativen Bruchzahl", `abs(-3.5)`, "3.5"},
		{"abs lässt Positives", `abs(3)`, "3"},
		{"abs benannt über input", `abs(input: -7)`, "7"},

		{"pow", `pow(2, exponent: 10)`, "1024"},
		{"pow mit Bruchzahlen", `pow(2.5, exponent: 2)`, "6.25"},
		{"sqrt", `sqrt(9)`, "3"},
		{"exp von null", `exp(0)`, "1"},
		{"log von eins", `log(1)`, "0"},
		{"log10 von hundert", `log10(100)`, "2"},

		{"to_boolean von null", `to_boolean(0)`, "false"},
		{"to_boolean von eins", `to_boolean(1)`, "true"},
		{"to_boolean von yes", `to_boolean("yes")`, "true"},
		{"to_boolean von no", `to_boolean("no")`, "false"},
		{"to_boolean eines Wahrheitswerts", `to_boolean(true)`, "true"},

		// **`to_code` gibt die Schreibweise, nicht den Wert.** Eine
		// Zeichenkette bekommt ihre Anführungszeichen zurück; eine Zahl
		// steht, wie sie dasteht.
		{"to_code einer Zeichenkette", `to_code("a")`, `"a"`},
		{"to_code einer Zahl", `to_code(3)`, "3"},
		{"to_code eines Wahrheitswerts", `to_code(true)`, "true"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Und jeder Fehlerweg wird betreten** (§3).
//
// Ein Zweig, den kein Test durchläuft, ist keine Zusage — er ist eine
// Vermutung darüber, was geschähe.
func TestRechnenMeldetWasFehlt(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, quelle, nennt string }{
		{"abs ohne Eingabe", `abs()`, "abs"},
		{"abs auf einer Abbildung", `abs([1, 2])`, "abs"},
		{"pow ohne Eingabe", `pow(exponent: 2)`, "pow"},
		{"pow ohne exponent", `pow(2)`, "exponent"},
		{"pow mit unbrauchbarem exponent", `pow(2, exponent: [1])`, "exponent"},
		{"pow auf einer Abbildung", `pow([1], exponent: 2)`, "pow"},
		{"to_boolean ohne Eingabe", `to_boolean()`, "to_boolean"},
		{"to_boolean auf einer Abbildung", `to_boolean([1, 2])`, "to_boolean"},
		{"to_code ohne Eingabe", `to_code()`, "to_code"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err == nil {
				t.Fatal("der Aufruf glückte, erwartet ein Fehler")
			}

			if !strings.Contains(err.Error(), fall.nennt) {
				t.Errorf("die Meldung %q nennt %q nicht", err, fall.nennt)
			}
		})
	}
}
