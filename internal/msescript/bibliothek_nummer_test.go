// **`to_number` gibt eine Kommazahl, wenn eine dasteht** —
// #859.
//
// # Was MSE tut
//
//	if (str.ToLong(&l))        SCRIPT_RETURN((int)l);
//	else if (str.ToDouble(&d)) SCRIPT_RETURN((double)d);
//	else if (str.empty())      SCRIPT_RETURN(0);
//	else                       return delay_error(…);
//
// `src/script/functions/basic.cpp`, Zeilen 160–190, Fassung f9d9356 vom
// 2021-07-06. **Erst ganz, dann gebrochen** — und die ganze Zahl wird auf
// C++-`int` verengt.
//
// Bis zum 2026-08-22 reichte `to_number` an `to_int` durch und schnitt ab:
// `to_number("1.5")` gab **1**. Angestrichen vom Nachlauf zum Zweitblick über
// Phase P.
//
// # Woher die erwarteten Werte kommen
//
// **Vom Orakel** (§15), jeder Ausdruck einzeln am 2026-08-22 gefragt.
package msescript_test

import (
	"testing"
)

func TestToNumberAntwortetWieDasOrakel(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		quelle   string
		erwartet string
		warum    string
	}{
		// **Der Fall, um den es geht.**
		{`"" + to_number("1.5")`, "1.5", "eine Kommazahl bleibt eine"},
		{`"" + to_number("-2.75")`, "-2.75", "auch eine negative"},

		// **`ToLong` scheitert an `1e3`, `ToDouble` nicht** — und das
		// Ergebnis ist ein `double`, der sich als ganze Zahl schreibt.
		{`"" + to_number("1e3")`, "1000", "Exponentialschreibweise"},

		// **Eine ganze Zahl wird auf C-`int` verengt.**
		{`"" + to_number("2147483648")`, "-2147483648", "eins über der Grenze"},

		// **Und hier ist die Zeichenkette genauer als der Werttyp**: aus ihr
		// entsteht ein `int64`, und der verengt exakt. `format` kann das
		// nicht, weil die Zahl dort schon als `float64` ankommt.
		{`"" + to_number("9223372036854775807")`, "-1", "und weit darüber"},

		{`"" + to_number("")`, "0", "leer ist null"},
		{`"" + to_number(true)`, "1", "ein Wahrheitswert"},
		{`"" + to_number(1.5)`, "1.5", "eine Zahl bleibt sie selbst"},
		{`"" + to_number(12)`, "12", "eine ganze auch"},

		// **`nil` ist null** — MSEs `SCRIPT_NIL`-Zweig, Zeile 172.
		{`"" + to_number(nil)`, "0", "nichts ist null"},
		{`"" + to_number(false)`, "0", "und falsch auch"},
	}

	for _, fall := range faelle {
		t.Run(fall.warum, func(t *testing.T) {
			t.Parallel()

			if gegeben := auswerten(t, fall.quelle); gegeben != fall.erwartet {
				t.Errorf("%s → %q, das Orakel gibt %q",
					fall.quelle, gegeben, fall.erwartet)
			}
		})
	}
}

// **Was keine Zahl ist, bleibt ein Fehlerwert.**
//
// Das Orakel: `Can't convert "12abc" from string ("12abc") to real number`.
// **Verzögert und nicht sofort**, denn die Vorlagen schreiben
// `to_number(x) or else 0`.
func TestToNumberWeistNichtzahlenAb(t *testing.T) {
	t.Parallel()

	if gegeben := auswerten(t, `to_number("12abc") or else "gefangen"`); gegeben != "gefangen" {
		t.Errorf("→ %q, erwartet war ein Fehlerwert, den `or else` fängt", gegeben)
	}

	// **`AlsText` eines Fehlerwerts ist „Fehler"**, nicht die Meldung — das
	// ist die Gestalt, in der MSE ihn weiterreicht. Geprüft wird deshalb,
	// **dass** einer entsteht.
	if fehler := fehlerAus(t, `to_number("12abc")`); fehler == "" {
		t.Error("kein Fehlerwert")
	}
}

// **Ohne Eingabe ist es ein Fehler und keine Null.**
//
// Eine erfundene Null wäre eine Zahl, die niemand geschrieben hat — und sie
// stünde als Maß in einer Karte.
func TestToNumberBrauchtEineEingabe(t *testing.T) {
	t.Parallel()

	if fehler := fehlerAus(t, `to_number()`); fehler == "" {
		t.Error("kein Fehler für einen Aufruf ohne Eingabe")
	}
}
