package msescript

// **Die sechs Rechenfunktionen und zwei Umwandlungen** — Werkstück A von
// Phase 4.9 (#1324).
//
// Alles aus `src/script/functions/basic.cpp`, Fassung f9d9356 vom 2021-07-06,
// und jeder erwartete Wert am Orakel erfragt (§15) — die Zeilen stehen im
// Test daneben.

import (
	"context"
	"math"
)

func init() {
	eintragen("abs", betrag)
	eintragen("exp", hochE)
	eintragen("log", logarithmus)
	eintragen("log10", zehnerlogarithmus)
	eintragen("sqrt", wurzel)
	eintragen("pow", potenz)

	eintragen("to_boolean", alsWahrheitswert)
	eintragen("to_code", alsQuelltext)
}

// betrag ist `abs(x)`.
//
//	ScriptType t = input->type();
//	if (t == SCRIPT_DOUBLE) { SCRIPT_RETURN(fabs(input->toDouble())); }
//	else                    { SCRIPT_RETURN(abs(input->toInt())); }
//
// Zeilen 238–246. **Eine ganze Zahl bleibt ganz.** Hier gibt es nur einen
// Zahltyp, und `math.Abs` erhält die Ganzheit ohnehin — die Unterscheidung im
// Original steht trotzdem hier, weil sie erklärt, warum `abs(-3)` als `"3"`
// und nicht als `"3.0"` herauskommt.
func betrag(_ context.Context, ruf Ruf) (Wert, error) {
	return einZahlenweg(ruf, "abs", math.Abs)
}

// hochE ist `exp(x)` — Zeilen 290–293.
func hochE(_ context.Context, ruf Ruf) (Wert, error) {
	return einZahlenweg(ruf, "exp", math.Exp)
}

// logarithmus ist `log(x)`, der natürliche — Zeilen 294–297.
func logarithmus(_ context.Context, ruf Ruf) (Wert, error) {
	return einZahlenweg(ruf, "log", math.Log)
}

// zehnerlogarithmus ist `log10(x)`.
//
//	SCRIPT_RETURN(log(input) / log(10.0));
//
// Zeilen 298–301. **So gerechnet, wie es dasteht** — nicht über `math.Log10`.
// Für die Zahlen, um die es auf einer Karte geht, ist das dasselbe; wo es das
// nicht ist, gilt das Original.
func zehnerlogarithmus(_ context.Context, ruf Ruf) (Wert, error) {
	return einZahlenweg(ruf, "log10", func(zahl float64) float64 {
		return math.Log(zahl) / math.Log(10)
	})
}

// wurzel ist `sqrt(x)` — Zeilen 302–305.
func wurzel(_ context.Context, ruf Ruf) (Wert, error) {
	return einZahlenweg(ruf, "sqrt", math.Sqrt)
}

// einZahlenweg ist der gemeinsame Rumpf der fünf einstelligen.
//
// **Einer statt fünf**: sie unterscheiden sich in einer Zeile, und fünfmal
// dieselbe Prüfung daneben wäre fünfmal dieselbe Stelle zum Ändern.
func einZahlenweg(ruf Ruf, name string, rechne func(float64) float64) (Wert, error) {
	if err := ruf.eineEingabe(name); err != nil {
		return Nichts, err
	}

	zahl, err := ruf.Stellung[0].AlsZahl()
	if err != nil {
		return Nichts, ruf.fehlerAn(name + " braucht eine Zahl: " + err.Error())
	}

	return Zahl(rechne(zahl)), nil
}

// potenz ist `pow(x, exponent: n)`.
//
//	SCRIPT_PARAM_C(double, input);
//	SCRIPT_PARAM(double, exponent);
//	SCRIPT_RETURN(pow(input,exponent));
//
// Zeilen 306–310. **Der Exponent ist benannt und hat keine Vorgabe** — am
// Orakel gemessen: `pow(2, 10)` endet mit „Variable not set: exponent".
func potenz(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("pow"); err != nil {
		return Nichts, err
	}

	grundzahl, err := ruf.Stellung[0].AlsZahl()
	if err != nil {
		return Nichts, ruf.fehlerAn("pow braucht eine Zahl: " + err.Error())
	}

	hochwert, da := ruf.Benannter("exponent")
	if !da {
		return Nichts, ruf.fehlerAn("pow braucht exponent:")
	}

	hoch, err := hochwert.AlsZahl()
	if err != nil {
		return Nichts, ruf.fehlerAn("pow braucht für exponent: eine Zahl: " + err.Error())
	}

	return Zahl(math.Pow(grundzahl, hoch)), nil
}

// alsWahrheitswert ist `to_boolean(x)`.
//
//	if (t == SCRIPT_INT) { result = input->toInt() != 0; }
//	else                 { result = input->toBool(); }
//
// Zeilen 192–206. **Eine Zahl ist wahr, wenn sie nicht null ist**; alles
// andere geht über die gewöhnliche Umwandlung, die `"yes"` kennt und `"no"`.
func alsWahrheitswert(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("to_boolean"); err != nil {
		return Nichts, err
	}

	eingabe := ruf.Stellung[0]

	if eingabe.Art == ArtZahlwert {
		return Wahrheit(eingabe.Zahl != 0), nil
	}

	trifftZu, err := eingabe.AlsWahrheitswert()
	if err != nil {
		return Nichts, ruf.fehlerAn("to_boolean: " + err.Error())
	}

	return Wahrheit(trifftZu), nil
}

// alsQuelltext ist `to_code(x)`.
//
//	SCRIPT_PARAM_C(ScriptValueP, input);
//	SCRIPT_RETURN(input->toCode());
//
// Zeilen 226–229. **Die Schreibweise, nicht der Wert**: eine Zeichenkette
// bekommt ihre Anführungszeichen zurück, eine Zahl steht, wie sie dasteht.
// Am Orakel: `to_code("a")` → `"\"a\""`, `to_code(3)` → `"3"`.
func alsQuelltext(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("to_code"); err != nil {
		return Nichts, err
	}

	return Text(ruf.Stellung[0].AlsQuelltext()), nil
}
