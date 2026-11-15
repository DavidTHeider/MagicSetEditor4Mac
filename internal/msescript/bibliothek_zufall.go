package msescript

import (
	"context"
	"math/rand/v2"
)

// Die **Zufallsfunktionen**
// (#655).
//
// # Was MSE führt
//
//	SCRIPT_FUNCTION(random_real) {
//	  SCRIPT_PARAM_DEFAULT_C(double, begin, 0.0);
//	  SCRIPT_PARAM_DEFAULT_C(double, end,   1.0);
//	  SCRIPT_RETURN( (double)rand() / RAND_MAX * (end - begin) + begin );
//	}
//
//	SCRIPT_FUNCTION(random_int) {
//	  SCRIPT_PARAM_DEFAULT_C(int, begin, 0);
//	  SCRIPT_PARAM_C(        int, end);
//	  SCRIPT_RETURN( rand() % (end - begin) + begin );
//	}
//
//	SCRIPT_FUNCTION(random_boolean) {
//	  SCRIPT_PARAM_DEFAULT_C(double, input, 0.5);
//	  SCRIPT_RETURN( rand() < RAND_MAX * input );
//	}
//
// `src/script/functions/basic.cpp`, Zeilen 248–263, Fassung f9d9356 vom
// 2021-07-06. **`end` ist ausschließend**: `random_int(end: 3)` gibt 0, 1
// oder 2.
//
// # Wo sie stehen
//
// `magic-wagic.mse-export-template` ruft einmal `random_int`, um einer Karte
// eine Kennung zu geben, wenn niemand eine Nummer eingetragen hat. Die beiden
// anderen stehen in **keiner** Datei des vollen Packs; sie sind hier, weil sie
// im Original danebenstehen und eine halbe Familie eine Falle ist — dieselbe
// Begründung wie bei den `_deg`-Winkeln.
//
// # Die Listenfamilie steht nebenan
//
// `random_shuffle`, `random_select` und `random_select_many` (`basic.cpp`
// 602–656) arbeiten über Listen — „mische diese Liste" ist eine andere Frage
// als „gib mir eine Zahl". Sie stehen in `bibliothek_zufallsliste.go`, seit
// Phase 4.9 die Sprache vollständig trägt
// (#1324).
//
// # Wo der Zufall herkommt
//
// `math/rand/v2` und nicht `crypto/rand`. Was hier gewürfelt wird, ist eine
// Kartennummer in einer Ausfuhrdatei — kein Geheimnis, und nichts, dessen
// Vorhersagbarkeit jemandem nützte.
func init() {
	eintragen("random_int", zufallsganzzahl)
	eintragen("random_real", zufallszahl)
	eintragen("random_boolean", zufallswahrheit)
}

// zufallsganzzahl ist `random_int(begin: …, end: …)`.
//
// **Der leere Bereich meldet sich, statt abzubrechen.** `rand() % (end -
// begin)` ist in C++ bei `end == begin` undefiniert und in Go ein Abbruch;
// eine Vorlage darf das Programm nicht anhalten (§8).
func zufallsganzzahl(_ context.Context, ruf Ruf) (Wert, error) {
	anfang, err := zahlOder(ruf, "begin", 0, "random_int")
	if err != nil {
		return Nichts, err
	}

	ende, err := zahlOder(ruf, "end", 0, "random_int")
	if err != nil {
		return Nichts, err
	}

	weite := int(ende) - int(anfang)
	if weite <= 0 {
		return Nichts, ruf.fehlerAn(
			"random_int braucht ein `end:` über dem `begin:`")
	}

	return Zahl(float64(int(anfang) + rand.IntN(weite))), nil // #nosec G404 -- siehe Dateikopf
}

// zufallszahl ist `random_real(begin: …, end: …)`.
func zufallszahl(_ context.Context, ruf Ruf) (Wert, error) {
	anfang, err := zahlOder(ruf, "begin", 0, "random_real")
	if err != nil {
		return Nichts, err
	}

	ende, err := zahlOder(ruf, "end", 1, "random_real")
	if err != nil {
		return Nichts, err
	}

	return Zahl(rand.Float64()*(ende-anfang) + anfang), nil // #nosec G404 -- siehe Dateikopf
}

// zufallswahrheit ist `random_boolean(<Wahrscheinlichkeit>)`.
//
// **Die Eingabe ist die Wahrscheinlichkeit für `true`**, vorbelegt mit einer
// Hälfte. MSE nimmt sie über die Eingabekette; hier tut es dasselbe, was jede
// einstellige Funktion der Bibliothek tut.
func zufallswahrheit(_ context.Context, ruf Ruf) (Wert, error) {
	anteil := 0.5

	if len(ruf.Stellung) > 0 {
		gelesen, err := ruf.Stellung[0].AlsZahl()
		if err != nil {
			return Nichts, ruf.fehlerAn("random_boolean: " + err.Error())
		}

		anteil = gelesen
	}

	return Wahrheit(rand.Float64() < anteil), nil // #nosec G404 -- siehe Dateikopf
}

// zahlOder liest ein benanntes Argument als Zahl oder gibt die Vorgabe.
func zahlOder(ruf Ruf, name string, vorgabe float64, funktion string) (float64, error) {
	wert, da := ruf.Benannter(name)
	if !da {
		return vorgabe, nil
	}

	zahl, err := wert.AlsZahl()
	if err != nil {
		return 0, ruf.fehlerAn(funktion + ": `" + name + "` " + err.Error())
	}

	return zahl, nil
}
