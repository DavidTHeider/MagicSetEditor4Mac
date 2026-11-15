package msescript_test

// `to_title`, `to_color` und die trigonometrischen Funktionen
// (#472).

import (
	"math"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`to_title` schreibt jedes Wort groß — außer den kleinen.**
//
//	SCRIPT_RETURN(capitalize(input.Lower()));
//
// `src/script/functions/basic.cpp`, Zeilen 218–221, und
// `capitalize_in_place`, `src/util/string.cpp`, Fassung f9d9356 vom
// 2021-07-06.
func TestToTitle(t *testing.T) {
	faelle := []struct{ name, quelle, erwartet string }{
		{"jedes Wort groß", `to_title("plains of ruin")`, "Plains of Ruin"},

		// **Erst klein, dann groß** — `input.Lower()` steht vor `capitalize`.
		{"vorher alles klein", `to_title("PLAINS OF RUIN")`, "Plains of Ruin"},

		// **Die kleinen Worte**: `is`, `the`, `in`, `of`, `to`, `at`, `a`.
		{
			"die kleinen bleiben klein", `to_title("a tale of the sea")`,
			"A Tale of the Sea",
		},

		// **Am Anfang aber groß** — `it != s.begin()`.
		{"das erste Wort auch klein geschrieben", `to_title("the end")`, "The End"},

		// **Mit dem Abstand dahinter geprüft**: „Island" fängt mit `is` an
		// und wird trotzdem groß.
		{"kein halbes Wort", `to_title("the island")`, "The Island"},

		// **Der Schrägstrich trennt wie ein Leerzeichen**
		// (`*it == ' ' || *it == '/'`).
		{"am Schrägstrich", `to_title("fire/ice")`, "Fire/Ice"},

		{"leer bleibt leer", `to_title("")`, ""},
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

// **`to_color` gibt eine Farbe — und einen Fehlerwert, wenn es keine ist.**
func TestToColor(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(
		t.Context(), `to_color(rgb(1,2,3))`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if _, istFarbe := wert.AlsFarbe(); !istFarbe {
		t.Errorf("%v ist keine Farbe", wert)
	}

	// **Was keine Farbe ist, wird ein Fehlerwert und wirft nicht sofort.**
	//
	//	} catch (const ScriptError& e) { return delay_error(e); }
	andere, err := msescript.AuswertenQuelle(
		t.Context(), `to_color("kein Ton")`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if !andere.IstFehler() {
		t.Errorf("%v ist kein Fehlerwert", andere)
	}
}

// **`sin` und `cos` rechnen im Bogenmaß, die `_deg`-Geschwister in Grad.**
func TestWinkelfunktionen(t *testing.T) {
	faelle := []struct {
		quelle   string
		erwartet float64
	}{
		{"sin(0)", 0},
		{"cos(0)", 1},
		{"tan(0)", 0},
		{"sin_deg(90)", 1},
		{"cos_deg(180)", -1},
		{"tan_deg(45)", 1},
	}

	for _, fall := range faelle {
		t.Run(fall.quelle, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(
				t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("ausrechnen: %v", err)
			}

			zahl, err := wert.AlsZahl()
			if err != nil {
				t.Fatalf("keine Zahl: %v", err)
			}

			if math.Abs(zahl-fall.erwartet) > 1e-9 {
				t.Errorf("ergibt %v, erwartet %v", zahl, fall.erwartet)
			}
		})
	}

	// **Und eine Eingabe, die keine Zahl ist, wird benannt.**
	if _, err := msescript.AuswertenQuelle(
		t.Context(), `sin("x")`, msescript.Bibliothek()); err == nil {
		t.Error("sin(\"x\") hat keinen Fehler gegeben")
	}
}

// **Ohne Argument sagen sie es.**
func TestTitelFarbeUndWinkelBrauchenIhreEingabe(t *testing.T) {
	for _, quelle := range []string{"to_title()", "to_color()", "sin()"} {
		t.Run(quelle, func(t *testing.T) {
			if _, err := msescript.AuswertenQuelle(
				t.Context(), quelle, msescript.Bibliothek()); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}
