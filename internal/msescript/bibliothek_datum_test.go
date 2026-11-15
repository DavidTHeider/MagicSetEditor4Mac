package msescript_test

// **`to_date`** — Werkstück A von Phase 4.9
// (#1324).
//
// # Was das Orakel sagt
//
// Gemessen am 2026-09-02 (Fassung f9d9356 vom 2021-07-06):
//
//	to_string(to_date("2020-01-02 03:04:05"))  → "2020-01-02 03:04:05"
//	to_string(to_date("2020-01-02"))           → ERROR: Can't convert
//	                                             "2020-01-02" … to date
//
// **Die Kurzform geht nicht.** `wxDateTime::ParseDateTime` verlangt Datum
// **und** Zeit; das steht in keiner Dokumentation, und geraten hätte man das
// Gegenteil.
//
//	SCRIPT_FUNCTION(to_date) {
//	  try {
//	    SCRIPT_PARAM_C(wxDateTime, input);
//	    SCRIPT_RETURN(input);
//	  } catch (const ScriptError& e) { return delay_error(e); }
//	}
//
// `src/script/functions/basic.cpp`, Zeilen 217–224.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestZuDatum(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, quelle, erwartet string }{
		{"die volle Form", `to_date("2020-01-02 03:04:05")`, "2020-01-02 03:04:05"},
		{
			"ein Datum bleibt eines",
			`to_date(to_date("2020-01-02 03:04:05"))`, "2020-01-02 03:04:05",
		},

		// **`type_name` sagt »date«** — und das ist der Grund, warum es hier
		// eine eigene Art gibt und nicht bloß Text.
		{"und heißt date", `type_name(to_date("2020-01-02 03:04:05"))`, "date"},
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

// **Zwei Daten lassen sich mit `<` nicht vergleichen — hier so wenig wie
// dort.**
//
// # Der Anlauf, der falsch war
//
// Hier stand zuerst die Zusicherung, zwei Daten verglichen sich zeitlich: die
// kanonische Form ist ISO-geordnet, buchstabenweise verglichen käme dieselbe
// Reihenfolge heraus. **Das war erschlossen und falsch** — `<` rechnet in
// dieser Sprache mit Zahlen.
//
// **Am Orakel nachgefragt** (2026-09-02), und MSE2 tut dasselbe:
//
//	to_date("2020-01-02 03:04:05") < to_date("2021-01-02 03:04:05")
//	  → ERROR: Can't convert from date to integer number
//	"a" < "b"
//	  → ERROR: Can't convert "a" from string ("a") to integer number
//
// Die Übereinstimmung ist damit **gemessen** und nicht behauptet.
func TestZweiDatenLassenSichNichtMitKleinerVergleichen(t *testing.T) {
	t.Parallel()

	_, err := msescript.AuswertenQuelle(t.Context(),
		`to_date("2020-01-02 03:04:05") < to_date("2021-01-02 03:04:05")`,
		msescript.Bibliothek())

	if err == nil {
		t.Fatal("der Vergleich glückte — MSE2 lehnt ihn ab")
	}

	if !strings.Contains(err.Error(), "Zahl") {
		t.Errorf("die Meldung %q sagt nicht, dass `<` mit Zahlen rechnet", err)
	}
}

// **Was sich nicht lesen lässt, ist ein Fehler — und sagt, was dastand.**
func TestZuDatumMeldetWasNichtGeht(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, quelle, nennt string }{
		{"die Kurzform", `to_date("2020-01-02")`, "2020-01-02"},
		{"Kauderwelsch", `to_date("morgen")`, "morgen"},
		{"eine Zahl", `to_date(3)`, "to_date"},
		{"ohne Eingabe", `to_date()`, "to_date"},
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
