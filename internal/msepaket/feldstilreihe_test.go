package msepaket_test

// **Abstand und Richtung einer Mehrfachwahl als Reihe** — #2297.
//
//	REFLECT(direction);
//	REFLECT(spacing);
//
// `MultipleChoiceStyle`, `src/data/field/multiple_choice.cpp`, Zeilen 30–40,
// Fassung f9d9356 vom 2021-07-06 — `spacing` ist ein `Scriptable<double>`,
// `direction` ein `Scriptable<Direction>`, Vorgabe `left to right`.

import "testing"

func TestEinFeldstilNenntAbstandUndRichtung(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name, block, abstand, richtung string
	}{
		{"beides", "\t\tspacing: 10\n\t\tdirection: top to bottom\n", "10", "top to bottom"},
		{"nur der Abstand", "\t\tspacing: 4\n", "4", ""},
		{"als Skript", "\t\tspacing: { 2 * 3 }\n", "{ 2 * 3 }", ""},
		{"ohne Angabe", "", "", ""},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			paket := paketAusText("card style:\n\tkeyword icons:\n\t\tleft: 0\n"+fall.block, "probe.mse-style")
			if paket == nil {
				t.Fatal("der Stil ließ sich nicht zerteilen")
			}

			stil, _ := paket.Stilmodell()
			feld, _ := stil.Feld("keyword icons")

			if feld.Abstand != fall.abstand || feld.Richtung != fall.richtung {
				t.Errorf("spacing %q, direction %q — erwartet %q und %q", feld.Abstand, feld.Richtung, fall.abstand, fall.richtung)
			}
		})
	}
}
