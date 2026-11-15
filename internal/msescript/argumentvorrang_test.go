package msescript_test

import "testing"

// **Wer schlägt wen, wenn dasselbe Argument zweimal dasteht.**
//
// # Die Reihenfolge, am Orakel abgefragt
//
// Am 2026-08-20 gegen `White.mse-set` (`magicseteditor --cli … -q --raw`),
// sechs Fälle. Daraus liest sich eine Reihenfolge, und nur eine:
//
//	zuerst  die mit `@` vorab gebundenen Werte
//	dann    die **benannten** Argumente des Aufrufs
//	zuletzt das **Stellungsargument** des Aufrufs, als `input`
//
// # Warum das Stellungsargument zuletzt steht
//
// **`({ input })("x", input: "y")` gibt bei MSE `"x"`.** Hier stand bis zum
// 2026-08-20 das Gegenteil — „`f(x, input: y)` meint ausdrücklich `y`" —, und
// das war eine Setzung ohne Beleg.
//
// Der Grund ist MSEs Bauform: **es kennt gar kein Stellungsargument.** Ein
// bloßes Argument wird beim Zerteilen zu `input` (`src/script/parser.cpp`,
// Zeilen 881–893, Fassung f9d9356 vom 2021-07-06); `f(x, input: y)` ist dort
// also `f(input: x, input: y)`, und der **erste** gewinnt.
//
// # Und warum die Bindung zuerst steht
//
// Eine Closure-Bindung greift bei MSE nur, wenn die Variable **nicht** unter
// den Argumenten des Aufrufs steht (`src/script/value.cpp`, Zeile 608). `@`
// heißt „vorbelegen", nicht „festnageln".
//
// # Der Anlass
//
// **Der Nachlauf zum Zweitblick über Phase K.** Seit `@` das
// Stellungsargument bindet ([gebundeneArgumente]), stand ein gebundenes
// `input` in `Gebunden` — und wurde im Blockpfad **nach** dem
// Stellungsargument gesetzt. `({ input })@("alt")("neu")` gab damit `"alt"`.
//
// Zwei der drei Abweichungen unten sind so entstanden; die dritte stand schon
// vorher da.
func TestWerSchlaegtWenBeimAufruf(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, quelle, erwartet string }{
		{
			name:     "das Stellungsargument des Aufrufs schlägt die Bindung",
			quelle:   `to_string(({ input })@("alt")("neu"))`,
			erwartet: "neu",
		},
		{
			name:     "ohne Argument bleibt die Bindung stehen",
			quelle:   `to_string(({ input })@("alt")())`,
			erwartet: "alt",
		},
		{
			name:     "im selben Aufruf schlägt die Stellung das benannte `input:`",
			quelle:   `to_string(({ input })("x", input: "y"))`,
			erwartet: "x",
		},
		{
			name:     "und sie schlägt auch ein gebundenes `input:`",
			quelle:   `to_string(({ input })@(input: "a")("b"))`,
			erwartet: "b",
		},
		{
			name:     "ein benanntes Argument schlägt die Bindung",
			quelle:   `to_string(({ input })@("alt")(input: "benannt"))`,
			erwartet: "benannt",
		},
		{
			name:     "und das gilt für jeden Namen, nicht nur für `input`",
			quelle:   `to_string(({ a })@(a: 1)(a: 2))`,
			erwartet: "2",
		},

		// **Der eingebaute Weg gab schon vorher das Richtige** — er geht über
		// `ruferufen` und `mitEingabe`, nicht über den Blockpfad. Er steht
		// dabei, damit die Gegenprobe zeigt, dass beide Wege dasselbe sagen.
		{
			name:     "eine eingebaute Funktion antwortet gleich",
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
