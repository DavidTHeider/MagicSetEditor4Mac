// **`format`, `format_rule` und `to_string(format:)`** (#843).
//
// # Woher die erwarteten Werte kommen
//
// **Vom Orakel, nicht aus dem Kopf** (§15). Jede Zeile der Tabelle unten ist
// am 2026-08-22 mit
//
//	magicseteditor --cli Archenemy.mse-set -q --raw
//
// gefragt worden. Das hat gleich beim ersten Versuch etwas berichtigt: `format`
// nimmt **benannte** Argumente. `format("%d", 42)` meldet
// „Can't convert from built-in function 'format' to string" — es sucht die
// Variable `format` und findet sich selbst.
package msescript_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestFormatAntwortetWieDasOrakel(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		quelle   string
		erwartet string
	}{
		{`format(format: "%d", input: 42)`, "42"},

		// **Ohne `%` dasselbe.** `format_input` streicht jedes `%` und setzt
		// eines davor (`basic.cpp`, Zeile 77).
		{`format(format: "d", input: 42)`, "42"},

		{`format(format: "%05d", input: 42)`, "00042"},
		{`format(format: "%.2f", input: 3.14159)`, "3.14"},
		{`format(format: "%s", input: "abc")`, "abc"},

		// **Linksbündig auf sechs Zeichen** — die Leerzeichen gehören dazu.
		{`format(format: "%-6s", input: "ab")`, "ab    "},

		{`format(format: "%x", input: 255)`, "ff"},

		// **Ein Text wird zur Zahl, wenn das Format eine verlangt.**
		{`format(format: "%d", input: "42")`, "42"},

		// **Und abgeschnitten, nicht gerundet**: 3,9 gibt 3.
		{`format(format: "%d", input: 3.9)`, "3"},

		{`format(format: "%s", input: 42)`, "42"},

		// **Die Regelfassung ist dieselbe Funktion** (`basic.cpp`, Zeile 769).
		{`format_rule(format: "%05d")(42)`, "00042"},

		// **Und `to_string` liest `format:`** (`basic.cpp`, Zeilen 89–99).
		{`to_string(42, format: "%05d")`, "00042"},
	}

	for _, fall := range faelle {
		t.Run(fall.quelle, func(t *testing.T) {
			t.Parallel()

			if gegeben := auswerten(t, fall.quelle); gegeben != fall.erwartet {
				t.Errorf("%q → %q, das Orakel gibt %q",
					fall.quelle, gegeben, fall.erwartet)
			}
		})
	}
}

// **Ein Format ohne bekannten Buchstaben ist ein Fehler.**
//
// Das Orakel: `Invalid string format: '%q'`. MSE wirft dort, statt etwas zu
// raten — und das ist die richtige Antwort: welche Umwandlung gemeint war,
// weiß niemand.
func TestEinUnbekanntesFormatIstEinFehler(t *testing.T) {
	t.Parallel()

	fehler := fehlerAus(t, `format(format: "%q", input: 1)`)
	if !strings.Contains(fehler, "%q") {
		t.Errorf("die Meldung nennt das Format nicht: %s", fehler)
	}
}

// **Ein Wahrheitswert lässt sich nicht in eine Zahl formatieren.**
//
// Das Orakel: `Can't convert from boolean to integer number`.
func TestEinWahrheitswertLaesstSichNichtAlsZahlFormatieren(t *testing.T) {
	t.Parallel()

	if fehler := fehlerAus(t, `format(format: "%d", input: true)`); fehler == "" {
		t.Error("kein Fehler für einen Wahrheitswert")
	}
}

// fehlerAus wertet einen Ausdruck aus und verlangt einen Fehler.
func fehlerAus(t *testing.T, quelle string) string {
	t.Helper()

	wert, fehler := msescript.AuswertenQuelle(t.Context(), quelle, msescript.Bibliothek())
	if fehler != nil {
		return fehler.Error()
	}

	if wert.IstFehler() {
		return wert.AlsText()
	}

	t.Fatalf("%q gab %q statt eines Fehlers", quelle, wert.AlsText())

	return ""
}

// **Die drei Fehlerwege von `format`.**
//
// Sie sind je einzeln erreichbar, und ohne sie wären es drei Zweige, die kein
// Test betritt (§3).
func TestFormatWeistUnvollstaendigeRufeAb(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct{ quelle, fehlt string }{
		"ohne format":          {`format(input: 42)`, "format"},
		"ohne input":           {`format(format: "%d")`, "input"},
		"Gleitkomma ohne Zahl": {`format(format: "%.2f", input: true)`, "Zahl"},

		// **Auch über `to_string`** — dort ist es ein eigener Weg zu derselben
		// Funktion, und ein eigener Weg braucht seinen eigenen Fall.
		"to_string mit unbekanntem Format": {`to_string(1, format: "%q")`, "%q"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := fehlerAus(t, fall.quelle); fehler == "" {
				t.Errorf("%q kam durch", fall.quelle)
			}
		})
	}
}
