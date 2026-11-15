package msescript_test

// **Jedes bloße Argument heißt `input`, und das erste gewinnt**
// (#438).
//
//	} else {
//	  // implicit "input" argument
//	  arguments.push_back(SCRIPT_VAR_input);
//	  parseOper(input, script, PREC_SEQ);
//	}
//
// `src/script/parser.cpp`, Zeilen 890–894, Fassung f9d9356 vom 2021-07-06.
// **Jedes** unbenannte Argument bekommt denselben Namen; welches am Ende
// dasteht, entscheidet die Zuweisungsreihenfolge:
//
//	for (unsigned int j = 0 ; j < i.data ; ++j) {
//	  setVariable((Variable)instr[i.data - j - 1].data, stack.back());
//	  stack.pop_back();
//	}
//
// `src/script/context.cpp`, Zeilen 152–155. Die Schleife läuft vom **letzten**
// Argument zum ersten, und die letzte Zuweisung bleibt stehen — also gewinnt
// das **erste**.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestBlosseArgumenteHeissenAlleInput(t *testing.T) {
	for _, fall := range []struct {
		name     string
		quelle   string
		erwartet string
	}{
		{
			// **Der Fall aus dem Pack**: `magic-m15-mainframe-dfc.mse-style`
			// schreibt `to_number(pull_comma_array(…, cell:2), 152)` — die 152
			// war als Ersatzwert gemeint und wird von MSE verworfen.
			name: "das zweite Argument wird verworfen", quelle: `to_number("7", 152)`,
			erwartet: "7",
		},
		{
			name: "auch bei to_int", quelle: `to_int("7", 152)`, erwartet: "7",
		},
		{
			name: "und bei einer Textfunktion", quelle: `to_upper("ab", "cd")`,
			erwartet: "AB",
		},
		{
			// **Die Gegenprobe:** ohne Argument bleibt es ein Fehler. Ohne
			// diesen Fall wäre eine Prüfung, die gar nichts mehr verlangt,
			// ebenso grün.
			name: "ohne Argument bleibt es ein Fehler", quelle: `to_int()`,
			erwartet: "",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, msescript.Bibliothek())

			if fall.erwartet == "" {
				if err == nil && !wert.IstFehler() {
					t.Errorf("kein Fehler, sondern %q", wert.AlsText())
				}

				return
			}

			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Ohne bloßes Argument bleibt jede von ihnen ein Fehler.**
//
// Die Gegenprobe zur Regel oben: eine Prüfung, die nach
// #438 gar
// nichts mehr verlangte, wäre an den Fällen dort ebenso grün. Hier steht, was
// sie weiterhin abweist — und `rgb`/`rgba` daneben, die als Parser-Konstrukte
// wirklich mehrere Stellungen zählen (`src/script/parser.cpp`, Zeilen 610 und
// 620, Fassung f9d9356 vom 2021-07-06).
func TestOhneEingabeBleibtEsEinFehler(t *testing.T) {
	for _, quelle := range []string{
		`break_text(match:"x")`, `chosen(choice:"x")`, `count_chosen(choices:"x")`,
		`match(match:"x")`, `regex_escape()`, `remove_tag(tag:"<x")`,
		`split_text(match:"x")`, `tag_contents(tag:"<b>", contents:{""})`,
		`rgb()`, `rgb(1)`, `rgba(1, 2, 3)`,
	} {
		t.Run(quelle, func(t *testing.T) {
			if _, err := msescript.AuswertenQuelle(t.Context(), quelle,
				msescript.Bibliothek()); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// **Ein Dateibild ist ein Blatt.**
//
// [msescript.Dateibild] steht für Aufrufer außerhalb der Bibliothek —
// `internal/service` legt damit den Wert eines Bildfeldes in `card`
// (#440).
func TestEinDateibildIstEinBlatt(t *testing.T) {
	baum, ist := msescript.Dateibild("setpaket:/wo/Probe.mse-set!image16").AlsBild()
	if !ist {
		t.Fatal("kein Bild")
	}

	if baum.Knotenart != "datei" || baum.Pfad != "setpaket:/wo/Probe.mse-set!image16" {
		t.Errorf("Knoten %q, Pfad %q", baum.Knotenart, baum.Pfad)
	}
}
