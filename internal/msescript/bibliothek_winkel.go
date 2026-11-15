package msescript

import (
	"context"
	"math"
)

// Die **trigonometrischen** Funktionen und zwei Umwandlungen
// (#472).
//
// # Was MSE führt
//
//	SCRIPT_FUNCTION(sin) { SCRIPT_PARAM_C(double, input); SCRIPT_RETURN(sin(input)); }
//	SCRIPT_FUNCTION(cos) { … }  SCRIPT_FUNCTION(tan) { … }
//	SCRIPT_FUNCTION(sin_deg) { … SCRIPT_RETURN(sin(deg_to_rad(input))); }
//	SCRIPT_FUNCTION(cos_deg) { … }  SCRIPT_FUNCTION(tan_deg) { … }
//
// `src/script/functions/basic.cpp`, Zeilen 266–288, Fassung f9d9356 vom
// 2021-07-06.
//
// **`sin` und `cos` rechnen im Bogenmaß**, die `_deg`-Geschwister in Grad.
// Wer das verwechselt, bekommt eine Zahl, die aussieht wie eine Zahl.
//
// # Wo sie stehen
//
// `magic-booster.mse-style` und `magic-booster-mini.mse-style` rufen je
// einmal `sin` und `cos` — zwei Aufrufe in beiden Packs. Die `_deg`-Formen
// kommen dort nicht vor; sie stehen hier, weil sie im Original danebenstehen
// und eine halbe Familie eine Falle ist.
func init() {
	eintragen("sin", einstellig("sin", math.Sin))
	eintragen("cos", einstellig("cos", math.Cos))
	eintragen("tan", einstellig("tan", math.Tan))

	eintragen("sin_deg", einstellig("sin_deg", inGrad(math.Sin)))
	eintragen("cos_deg", einstellig("cos_deg", inGrad(math.Cos)))
	eintragen("tan_deg", einstellig("tan_deg", inGrad(math.Tan)))
}

// einstellig baut eine Funktion, die eine Zahl nimmt und eine gibt.
func einstellig(name string, rechnung func(float64) float64) Eingebaut {
	return func(_ context.Context, ruf Ruf) (Wert, error) {
		if err := ruf.eineEingabe(name); err != nil {
			return Nichts, err
		}

		zahl, err := ruf.Stellung[0].AlsZahl()
		if err != nil {
			return Nichts, ruf.fehlerAn(name + ": " + err.Error())
		}

		return Zahl(rechnung(zahl)), nil
	}
}

// inGrad macht aus einer Rechnung im Bogenmaß eine in Grad.
func inGrad(rechnung func(float64) float64) func(float64) float64 {
	return func(grad float64) float64 { return rechnung(grad * math.Pi / 180) }
}
