package msescript_test

import "testing"

// **`@` bindet auch das Stellungsargument.**
//
// # Der Befund
//
// Angestrichen vom Zweitblick über Phase K am 2026-08-20: `binden` reichte
// nur die **benannten** Argumente weiter, und `f@(x)` verlor sein `x`.
//
//	replace@("aaa", match:"a", replace:"b")()   MSE2: "bbb"   hier: Fehler
//	to_upper@("abc")()                          MSE2: "ABC"   hier: Fehler
//
// Am Orakel gefragt, nicht erschlossen (`magicseteditor --cli … -q --raw`,
// gegen `White.mse-set`).
//
// # Warum MSE das so tut
//
// **Es kennt kein Stellungsargument.** `f(x)` setzt `x` als Kontextvariable
// `input` (`src/script/parser.cpp`, Zeilen 881–893, Fassung f9d9356 vom
// 2021-07-06), und `Context::makeClosure` nimmt die Namen der innersten Ebene
// mit — `input` darunter (`src/script/context.cpp`, Zeilen 318–333). Zwischen
// `f@(x)` und `f@(input: x)` ist bei MSE also gar kein Unterschied.
//
// # Wie er aufgefallen ist
//
// **An einer Ungleichheit im eigenen Haus.** [regelbinder] band das
// Stellungsargument von Anfang an; `@` tat es nicht. Dieselbe Sprache gab auf
// dieselbe Frage zwei Antworten, und erst der Vergleich der beiden Wege hat
// es sichtbar gemacht.
func TestEineBindungNimmtDasStellungsargumentMit(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, quelle, erwartet string }{
		{
			name:     "eine eingebaute Funktion mit einem Stellungsargument",
			quelle:   `to_string(to_upper@("abc")())`,
			erwartet: "ABC",
		},
		{
			name:     "und eine mit benannten daneben",
			quelle:   `to_string(replace@("aaa", match: "a", replace: "b")())`,
			erwartet: "bbb",
		},
		{
			name:     "`@` und die Regelfassung geben dasselbe",
			quelle:   `to_string(replace_rule("aaa", match: "a", replace: "b")())`,
			erwartet: "bbb",
		},
		{
			name:     "ein späteres Argument schlägt das gebundene",
			quelle:   `to_string(to_upper@("abc")("xyz"))`,
			erwartet: "XYZ",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if ergebnis := auswerten(t, fall.quelle); ergebnis != fall.erwartet {
				t.Errorf("%s ergibt %q, erwartet %q", fall.quelle, ergebnis, fall.erwartet)
			}
		})
	}
}
