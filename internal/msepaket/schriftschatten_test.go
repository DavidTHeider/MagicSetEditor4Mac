package msepaket_test

// **Der Schatten einer Schrift steht im `font:`-Block** — F02
// (#2244).
//
//	REFLECT_N("shadow_displacement_x", shadow_displacement.width);
//	REFLECT_N("shadow_displacement_y", shadow_displacement.height);
//	REFLECT(shadow_color);
//	REFLECT(shadow_blur);
//
// `IMPLEMENT_REFLECTION_NO_SCRIPT(Font)`, `src/data/font.cpp`, Zeilen
// 118–134, Fassung f9d9356 vom 2021-07-06. **Nur die Farbe ist ein
// `Scriptable`** (`src/data/font.hpp`, Zeilen 44–46); Versatz und Weichheit
// sind Zahlen. Im vollen Pack: 498 `shadow color:` in 107 Stilen, 392
// Versätze, 223 `shadow blur:`.

import "testing"

func TestEineSchriftNenntIhrenSchatten(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name, block            string
		farbe, x, y, weichheit string
	}{
		{
			"ganz",
			"\t\t\tshadow color: black\n\t\t\tshadow displacement x: -1\n\t\t\tshadow displacement y: 1\n\t\t\tshadow blur: 2\n",
			"black", "-1", "1", "2",
		},
		{
			"als Skript",
			"\t\t\tshadow color: {if x then rgba(0,0,0,0) else \"black\"}\n",
			`{if x then rgba(0,0,0,0) else "black"}`, "", "", "",
		},
		{
			"mit Unterstrichen",
			"\t\t\tshadow_color: white\n\t\t\tshadow_displacement_x: 2\n\t\t\tshadow_displacement_y: 3\n\t\t\tshadow_blur: 1\n",
			"white", "2", "3", "1",
		},
		{"ohne Angabe", "", "", "", "", ""},
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
			s := feld.Schrift

			if s.SchattenFarbe != fall.farbe || s.SchattenX != fall.x || s.SchattenY != fall.y || s.SchattenWeichheit != fall.weichheit {
				t.Errorf("Schatten %q, %q, %q, %q — erwartet %q, %q, %q, %q",
					s.SchattenFarbe, s.SchattenX, s.SchattenY, s.SchattenWeichheit,
					fall.farbe, fall.x, fall.y, fall.weichheit)
			}
		})
	}
}
