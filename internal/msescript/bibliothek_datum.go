package msescript

// **`to_date` und die Art »Datum«** — Werkstück A von Phase 4.9
// (#1324).
//
//	SCRIPT_FUNCTION(to_date) {
//	  try {
//	    SCRIPT_PARAM_C(wxDateTime, input);
//	    SCRIPT_RETURN(input);
//	  } catch (const ScriptError& e) { return delay_error(e); }
//	}
//
// `src/script/functions/basic.cpp`, Zeilen 217–224, Fassung f9d9356 vom
// 2021-07-06.
//
// # Was das Orakel dazu sagt
//
// Gemessen am 2026-09-02:
//
//	to_string(to_date("2020-01-02 03:04:05"))  → "2020-01-02 03:04:05"
//	to_string(to_date("2020-01-02"))           → ERROR: Can't convert … to date
//
// **Die Kurzform geht nicht.** `wxDateTime::ParseDateTime` verlangt Datum und
// Zeit. Geraten hätte man das Gegenteil — genau dafür gibt es das Orakel
// (§15).
//
// # Die Anpassung, und warum sie eine ist
//
// **MSE trägt eine `wxDateTime`, hier trägt ein Datum seine kanonische
// Schreibweise** (`ArtDatum`). Das ist eine Anpassung an die Plattform (§15,
// `adapted`), und sie hat zwei sichtbare Folgen:
//
//   - **`type_name` sagt weiterhin »date«.** Deshalb eine eigene Art und nicht
//     einfach Text: eine Vorlage, die den Typnamen vergleicht, bekäme sonst
//     »string« und liefe anders.
//   - **Verglichen wird nicht**, auch wenn die kanonische Form ISO-geordnet
//     ist: `<` rechnet in dieser Sprache mit Zahlen — und **MSE2 tut
//     dasselbe**, am Orakel nachgefragt: „Can't convert from date to integer
//     number". Die Übereinstimmung ist gemessen, nicht behauptet.
//
// **Was sie nicht kann**: rechnen. MSE kennt keine Datumsarithmetik in der
// Sprache, und solange das so ist, fehlt hier nichts.

import (
	"context"
	"time"
)

// ArtDatum ist die Art eines Datumswerts.
const ArtDatum WertArt = "Datum"

// Datumsform ist die Schreibweise, die MSE liest und schreibt.
//
// **Beides zusammen** — `wxDateTime::ParseDateTime` nimmt keine Kurzform, am
// Orakel gemessen.
const Datumsform = "2006-01-02 15:04:05"

func init() {
	eintragen("to_date", zuDatum)
}

// zuDatum ist `to_date(x)`.
//
// **Ein Datum bleibt eines** — im Original gibt die Funktion ihre Eingabe
// zurück, wenn sie schon eine `wxDateTime` ist.
func zuDatum(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("to_date"); err != nil {
		return Nichts, err
	}

	eingabe := ruf.Stellung[0]

	if eingabe.Art == ArtDatum {
		return eingabe, nil
	}

	if eingabe.Art != ArtTextwert {
		return Nichts, ruf.fehlerAn(
			"to_date bekam kein Datum, sondern " + string(eingabe.Art))
	}

	gelesen, err := time.Parse(Datumsform, eingabe.Text)
	if err != nil {
		return Nichts, ruf.fehlerAn(
			"to_date liest " + eingabe.Text + " nicht als Datum — erwartet " +
				"wird JJJJ-MM-TT hh:mm:ss, und die Kurzform ohne Uhrzeit " +
				"nimmt auch MSE nicht an")
	}

	return Wert{Art: ArtDatum, Text: gelesen.Format(Datumsform)}, nil
}
