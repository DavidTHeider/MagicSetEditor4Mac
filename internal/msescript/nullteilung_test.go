// **Durch null zu teilen ist bei MSE kein Fehler**
// (#555).
package msescript_test

import (
	"math"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// TestDurchNullTeilenGibtUnendlich hält MSEs Rechnung fest.
//
// # Der Altcode hat keine Prüfung
//
//	case I_FDIV:
//	  a = to_script(a->toDouble() / b->toDouble());
//	  break;
//
// `src/script/context.cpp`, Zeilen 485–487, Fassung f9d9356 vom 2021-07-06.
// Zwei `double` durcheinander zu teilen ist in C++ kein Fehler, sondern
// IEEE 754: ±unendlich, und `0/0` ergibt `nan`.
//
// # Und das Orakel bestätigt es
//
//	$ magicseteditor --cli Archenemy.mse-set -q --raw
//	1/0        → inf
//	-1/0       → -inf
//	0/0        → nan
//	"x" + 1/0  → "xinf"
//	3 mod 0    → 3
//
// **Bis hierher stand ein Fehler.** Über die Stichprobe des Kanonlaufs waren
// das **84** Portierungsbefunde — und mit ihnen fiel jedes Mal ein ganzes
// Wertskript aus, nicht bloß die eine Rechnung.
func TestDurchNullTeilenGibtUnendlich(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name, quelle string
		erwartet     float64
	}{
		{"eins durch null", `1 / 0`, math.Inf(1)},
		{"minus eins durch null", `-1 / 0`, math.Inf(-1)},
		{"null durch null", `0 / 0`, math.NaN()},
		{"eine Kommazahl durch null", `1.5 / 0`, math.Inf(1)},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, nil)
			if err != nil {
				t.Fatalf("%s: %v", fall.quelle, err)
			}

			if wert.Art != msescript.ArtZahlwert {
				t.Fatalf("%s ergab %s, erwartet eine Zahl", fall.quelle, wert.Art)
			}

			if math.IsNaN(fall.erwartet) {
				if !math.IsNaN(wert.Zahl) {
					t.Errorf("%s ergab %g, erwartet nan", fall.quelle, wert.Zahl)
				}

				return
			}

			if wert.Zahl != fall.erwartet {
				t.Errorf("%s ergab %g, erwartet %g", fall.quelle, wert.Zahl, fall.erwartet)
			}
		})
	}
}

// TestUnendlichHeisstInfImText hält die Schreibweise fest.
//
//	String toString() const override { return String() << value; }
//
// `src/script/value.cpp`, Zeile 290, Fassung f9d9356 vom 2021-07-06 —
// `wxString::operator<<(double)`. Am Orakel: `"x" + 1/0` ergibt `"xinf"`.
//
// **Go schreibt `+Inf`**, und das stünde so auf der Karte.
func TestUnendlichHeisstInfImText(t *testing.T) {
	t.Parallel()

	faelle := []struct{ quelle, erwartet string }{
		{`"x" + 1 / 0`, "xinf"},
		{`"x" + -1 / 0`, "x-inf"},
		{`"x" + 0 / 0`, "xnan"},
	}

	for _, fall := range faelle {
		t.Run(fall.quelle, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, nil)
			if err != nil {
				t.Fatalf("%v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%s ergab %q, erwartet %q",
					fall.quelle, wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// TestRestDurchNullGibtDenZaehler hält fest, was das Orakel auf **diesem**
// Rechner antwortet.
//
//	case I_MOD:
//	  if (at == SCRIPT_DOUBLE || bt == SCRIPT_DOUBLE) {
//	    a = to_script(fmod(a->toDouble(), b->toDouble()));
//	  } else {
//	    a = to_script(a->toInt() % b->toInt());
//	  }
//
// `src/script/context.cpp`, Zeilen 495–500, Fassung f9d9356 vom 2021-07-06.
//
// # Und hier ist die Reichweite kleiner als sonst
//
// `a % 0` auf ganzen Zahlen ist in C++ **undefiniert**. Auf arm64 liefert
// `sdiv` bei null null, und der Rest wird als `a − (a/b)·b = a − 0 = a`
// gerechnet — das Orakel antwortet auf `3 mod 0` deshalb **3**. Auf x86 fiele
// dasselbe Programm mit einer Ausnahme aus.
//
// **Gebaut wird für darwin/arm64** (§12), und die Referenzbilder stammen von
// genau diesem Binärprogramm. Deshalb gilt hier die arm64-Antwort — als
// **adapted** und nicht als portierte Regel: sie steht in keinem Quelltext,
// sondern in einer Befehlssatzarchitektur.
//
// `fmod(x, 0)` ist dagegen definiert und ergibt `nan`.
func TestRestDurchNullGibtDenZaehler(t *testing.T) {
	t.Parallel()

	wert, err := msescript.AuswertenQuelle(t.Context(), `3 mod 0`, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if wert.Zahl != 3 {
		t.Errorf("3 mod 0 ergab %g, erwartet 3", wert.Zahl)
	}

	kommazahl, err := msescript.AuswertenQuelle(t.Context(), `3.5 mod 0`, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !math.IsNaN(kommazahl.Zahl) {
		t.Errorf("3.5 mod 0 ergab %g, erwartet nan", kommazahl.Zahl)
	}
}
