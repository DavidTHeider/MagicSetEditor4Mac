package msescript_test

import "testing"

// **Welcher Operator stärker bindet — am Orakel abgefragt.**
//
// # Was MSE tut
//
// `and`, `or`, `xor` und `or else` stehen auf **einer** Stufe und werden nach
// links verkettet (`PREC_AND`); rechts von `and`, `or` und `xor` steht ein
// Vergleich, rechts von `or else` nur eine Summe — `parseOper(…, PREC_ADD, …)`,
// `src/script/parser.cpp`, Zeilen 753–778, Fassung f9d9356 vom 2021-07-06. Die
// Schleife läuft danach weiter: was auf das `or else` folgt, gilt dem ganzen
// Ausdruck davor.
//
// # Was hier stand
//
// **`and` band stärker als `or`, und `or else` las rechts einen ganzen
// Vergleich** — `a or else "" != ""` hieß `a or else ("" != "")` und ergab den
// Text `a` statt eines Wahrheitswerts. Gefunden in QA-03
// (#1656):
// `lines_of_text` in `magic.mse-game/script` schreibt `if input[x] or else ""
// != ""`, und ohne Satzergebnis — ohne Schriftwurzel — erreicht der Kanonlauf
// diese Zeile. 253 Befunde der Art Portierung, und alle waren diese.
//
// # Das Orakel
//
// Am 2026-09-23 gegen `Archenemy.mse-set` (`magicseteditor --cli … -q --raw`),
// zwölf Ausdrücke — jede erwartete Zeile unten ist MSEs Antwort.
func TestDerVorrangFolgtMSE(t *testing.T) {
	t.Parallel()

	faelle := []struct{ quelle, erwartet string }{
		{`true or false and false`, "false"},
		{`false and true or true`, "true"},
		{`false or true and false`, "false"},
		{`true and false or true`, "true"},
		{`true xor true and false`, "false"},
		{`"x" or else "" != ""`, "true"},
		{`nil or else "" != ""`, "true"},
		{`1 + 2 or else 5 == 3`, "true"},
		{`1 or else 2 + 3`, "1"},
		{`to_int("x") or else 7 + 1`, "8"},
		{`to_int("x") or else 2 * 3`, "6"},
		{`1 == 1 or else 2`, "true"},
	}

	for _, fall := range faelle {
		t.Run(fall.quelle, func(t *testing.T) {
			t.Parallel()

			if ergebnis := auswerten(t, fall.quelle); ergebnis != fall.erwartet {
				t.Errorf("%s ergibt %q, MSE sagt %q", fall.quelle, ergebnis, fall.erwartet)
			}
		})
	}
}
