// **`format` rechnet in 32 Bit, so wie MSE2** —
// #850.
//
// # Warum eine eigene Datei
//
// Weil es eine andere Frage ist als in `bibliothek_format_test.go`. Dort steht,
// **welche** Umwandlung ein Formatwort auswählt; hier, **womit** sie rechnet.
// Der Zweitblick über Phase P hat die zweite angestrichen: MSE übergibt
// `input.toInt()`, und das ist ein C++-`int`.
//
// # Woher die erwarteten Werte kommen
//
// **Vom Orakel** (§15), jeder Ausdruck einzeln am 2026-08-22 gefragt. Das ist
// hier nicht Zierde: `%D` und `%I` haben dabei gezeigt, dass ein Teil der
// Ausgabe von MSE2 gar keine Semantik von MSE-Script ist, sondern die ABI des
// Rechners — siehe [TestDreiFormatbuchstabenSindPlattformzufall].
package msescript_test

import (
	"testing"
)

// TestFormatRechnetInTeilundVierzigBit hält die 32-Bit-Rechnung von C fest.
func TestFormatRechnetInZweiunddreissigBit(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		quelle   string
		erwartet string
		warum    string
	}{
		// **`%x`, `%X` und `%o` nehmen in C ein `unsigned int`.** Ein
		// negativer Wert erscheint deshalb als sein Bitmuster und nicht mit
		// einem Minus davor. Go schreibt für `int64(-42)` „-2a".
		{`format(format: "%x", input: -42)`, "ffffffd6", "das Bitmuster, kein Minus"},
		{`format(format: "%X", input: -42)`, "FFFFFFD6", "dasselbe in Großbuchstaben"},
		{`format(format: "%o", input: -42)`, "37777777726", "und oktal"},
		{`format(format: "%x", input: -1)`, "ffffffff", "alle 32 Bit gesetzt"},

		// **`%d` und `%i` nehmen ein `int` — vorzeichenbehaftet.**
		{`format(format: "%d", input: -42)`, "-42", "mit Vorzeichen"},

		// **Und was nicht hineinpasst, läuft um.** Kein Fehler, kein Klemmen:
		// C verengt, und MSE gibt das Ergebnis weiter.
		{`format(format: "%d", input: 2147483648)`, "-2147483648", "eins über der Grenze"},
		{`format(format: "%d", input: 4294967295)`, "-1", "die volle Breite"},
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

// TestFormatKenntDasCUebliche_i prüft `%i`.
//
// **Go kennt kein `%i`.** MSE zählt `Ii` zur Ganzzahlklasse
// (`find_first_of("DdIiOoXx")`, `src/script/functions/basic.cpp`, Zeile 77,
// Fassung f9d9356 vom 2021-07-06), und `%i` ist die geläufige C-Schreibweise
// für `%d`. Ohne Übersetzung schreibt `fmt.Sprintf` die Zeichenfolge
// `%!i(int64=-42)` — mitten in eine Karte.
//
// **Ein eigener Test und keine Zeile in der Tabelle oben**, weil er eine
// andere Klasse trifft: dort geht es um die Breite der Zahl, hier darum, dass
// der Buchstabe überhaupt ankommt.
func TestFormatKenntDenCBuchstabenI(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ quelle, erwartet string }{
		{`format(format: "%i", input: -42)`, "-42"},
		{`format(format: "%i", input: 42)`, "42"},
		{`format(format: "%05i", input: -42)`, "-0042"},
	} {
		t.Run(fall.quelle, func(t *testing.T) {
			t.Parallel()

			if gegeben := auswerten(t, fall.quelle); gegeben != fall.erwartet {
				t.Errorf("%s → %q, das Orakel gibt %q",
					fall.quelle, gegeben, fall.erwartet)
			}
		})
	}
}

// TestJenseitsVonZweiHochDreiundfuenfzigWeichtEsAb hält eine **Abweichung**
// fest, keine Übereinstimmung.
//
// **Der Werttyp verliert die Zahl, bevor das Format sie sieht.** [Wert] führt
// eine Zahl als `float64`; `9223372036854775807` kommt als
// 9223372036854775808 an, und die unteren Bits, aus denen MSE seine `-1`
// bildet, gibt es hier nicht mehr. Verengt ergibt das **0**.
//
// **Warum das hier steht und nicht überdeckt wird** (§15, Punkt 7): eine
// Abweichung vom Altverhalten wird dokumentiert und begründet. Ein Sonderfall,
// der für diese eine Zahl `-1` zurückgäbe, behandelte den Fall und nicht die
// Klasse — die Klasse ist „ganze Zahlen jenseits von 2⁵³", und die trägt
// dieses Programm nicht.
//
// Gemessen am Orakel am 2026-08-22: MSE2 gibt `-1`.
func TestJenseitsVonZweiHochDreiundfuenfzigWeichtEsAb(t *testing.T) {
	t.Parallel()

	const groesserAlsDieGenauigkeit = `format(format: "%d", input: 9223372036854775807)`

	if gegeben := auswerten(t, groesserAlsDieGenauigkeit); gegeben != "0" {
		t.Errorf("%s → %q, erwartet %q — MSE2 gibt hier %q",
			groesserAlsDieGenauigkeit, gegeben, "0", "-1")
	}
}

// TestDreiFormatbuchstabenSindPlattformzufall hält fest, was **nicht**
// portiert wird.
//
// `D`, `I` und `O` gehören bei MSE zur Ganzzahlklasse — `find_first_of` kennt
// sie. Was die C-Bibliothek daraus macht, ist die ABI dieses Rechners und
// keine Regel von MSE-Script; am Orakel gemessen am 2026-08-22:
//
//	%D mit −42 → 4294967254   (ein nullerweiterter int, gelesen als %ld)
//	%I mit −42 → "I"          (ein Flag ohne Umwandlung dahinter)
//	%O mit −42 → 37777777726  (wie %o — hier zufällig dasselbe)
//
// **Die Reichweite einer Messung ist die Reichweite ihrer Aussage** (§10): auf
// x86-64 stünden im oberen Wort andere Bits. Portiert wird deshalb die
// **Klassenwahl** und nicht die Ausgabe; `D` und `I` schreiben wie `d`, `O`
// wie `o`.
func TestDreiFormatbuchstabenSindPlattformzufall(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ quelle, erwartet, mse string }{
		{`format(format: "%D", input: -42)`, "-42", "4294967254"},
		{`format(format: "%I", input: -42)`, "-42", "I"},
		{`format(format: "%O", input: -42)`, "37777777726", "37777777726"},
	} {
		t.Run(fall.quelle, func(t *testing.T) {
			t.Parallel()

			if gegeben := auswerten(t, fall.quelle); gegeben != fall.erwartet {
				t.Errorf("%s → %q, erwartet %q — MSE2 gibt hier %q",
					fall.quelle, gegeben, fall.erwartet, fall.mse)
			}
		})
	}
}
