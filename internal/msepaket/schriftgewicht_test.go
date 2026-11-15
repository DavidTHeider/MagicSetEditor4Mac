package msepaket_test

// **Gewicht und Lage einer Schrift stehen im `font:`-Block** — F02
// (#2244).
//
// MSE2 liest beide und macht daraus die Schalter der Schrift:
//
//	REFLECT(weight);
//	REFLECT(style);
//
// `IMPLEMENT_REFLECTION_NO_SCRIPT(Font)`, und in `Font::update`
// `weight() == "bold"` zu `FONT_BOLD`, `style() == "italic"` zu `FONT_ITALIC`
// (`src/data/font.cpp`, Zeilen 26–41 und 118–134, Fassung f9d9356 vom
// 2021-07-06). Beide sind `Scriptable<String>` — im Pack stehen sie 768- und
// 16-mal, darunter `{if loy_leveler() then "" else "bold"}`.
//
// Bis hierher las dieses Programm keines von beiden: der Name jeder
// Archenemy-Karte (`name: Matrix`, `weight: bold`) stand aufrecht da.

import "testing"

func TestEineSchriftNenntGewichtUndLage(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name, block, gewicht, lage string
	}{
		{"fett", "\t\t\tweight: bold\n", "bold", ""},
		{"kursiv", "\t\t\tstyle: italic\n", "", "italic"},
		{"beides", "\t\t\tweight: bold\n\t\t\tstyle: italic\n", "bold", "italic"},
		{"als Skript", "\t\t\tweight: {if x then \"\" else \"bold\"}\n", "{if x then \"\" else \"bold\"}", ""},
		{"als eingerücktes Skript", "\t\t\tweight:\n\t\t\t\tscript: \"bold\"\n", "{\n\"bold\"\n}", ""},
		{"ohne Angabe", "", "", ""},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			paket := paketAusText("card style:\n\tname:\n\t\tfont:\n\t\t\tname: Matrix\n"+fall.block,
				"probe.mse-style")
			if paket == nil {
				t.Fatal("der Stil ließ sich nicht zerteilen")
			}

			stil, _ := paket.Stilmodell()
			feld, _ := stil.Feld("name")

			if feld.Schrift.Gewicht != fall.gewicht || feld.Schrift.Lage != fall.lage {
				t.Errorf("weight %q, style %q — erwartet %q und %q",
					feld.Schrift.Gewicht, feld.Schrift.Lage, fall.gewicht, fall.lage)
			}
		})
	}
}
