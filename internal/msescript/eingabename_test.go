package msescript_test

// **`input:` gilt auch bei den Wertfunktionen**
// (#472).
//
// MSE schreibt sie mit `SCRIPT_PARAM_C` — benannt **oder** stellungsweise:
//
//	SCRIPT_FUNCTION(to_title) {
//	  SCRIPT_PARAM_C(String, input);
//	  SCRIPT_RETURN(capitalize(input.Lower()));
//	}
//
// `src/script/functions/basic.cpp`, Zeilen 218–221, Fassung f9d9356 vom
// 2021-07-06.
//
// **Das ist ein Wächter, keine Änderung.** [Ruf.mitEingabe] schiebt ein
// benanntes `input:` seit
// #240 in
// die Stellung; geprüft war es aber nur für den Weg über die **Umgebung**
// (`… @(input: x)`), nicht für den geschriebenen Aufruf.
//
// **Im Pack steht kein einziger solcher Aufruf** — gezählt über 1.818
// Aufrufe von Wertfunktionen in `M15-Magic-Pack`, null davon mit `input:`.
// Der Wächter steht hier, weil dieselbe Lücke bei den **Bildfunktionen** 643
// Aufrufe ins Leere laufen ließ
// (#483) —
// dort war sie nicht geprüft, und niemand sah es.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestEineWertfunktionNimmtInputAuchBenannt(t *testing.T) {
	faelle := []struct{ name, quelle, erwartet string }{
		{"to_int stellungsweise", "to_int(3.7)", "3"},
		{"to_int benannt", "to_int(input: 3.7)", "3"},
		{"reverse benannt", `reverse(input: "abc")`, "cba"},
		{"trim benannt", `trim(input: "  ab  ")`, "ab"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(
				t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("ausrechnen: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}
