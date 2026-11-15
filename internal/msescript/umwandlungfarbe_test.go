package msescript_test

// **Eine Farbe ist eine Zahl, wenn man sie fragt** — und `to_real` ist
// MSEs, nicht unseres (F02,
// #2244).
//
// `magic-nouveau` schreibt `if to_int(card.border_color) < 96 then "white"
// else "black"`, und hier kam ein Fehler heraus statt einer Farbe für die
// Copyright-Zeile. MSE mittelt die drei Kanäle:
//
//	} else if (t == SCRIPT_COLOR) {
//	  Color c = input->toColor();
//	  result = (c.Red() + c.Blue() + c.Green()) / 3;
//
// in `to_int`, `to_real` (`/ 3.0`) und `to_number`
// (`src/script/functions/basic.cpp`, Zeilen 106–188, Fassung f9d9356 vom
// 2021-07-06). `to_real` kennt dazu Wahrheitswerte, nimmt den leeren Text
// als null und gibt bei Unlesbarem einen **verzögerten** Fehler.
//
// **Jede erwartete Zahl ist am Orakel gefragt** (`magicseteditor --cli … -q
// --raw`, am 2026-10-11), nicht erschlossen.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestEineFarbeWirdZurZahlIhrerKanaele(t *testing.T) {
	faelle := map[string]string{
		"to_string(to_int(rgb(30,60,91)))":          "60",
		"to_string(to_int(rgba(10,20,30,0)))":       "20",
		"to_string(to_number(rgb(1,2,4)))":          "2",
		"to_string(to_number(rgb(255,255,254)))":    "254",
		"to_string(to_real(rgb(1,2,4)) * 3)":        "7",
		"to_string(to_int(rgba(0,0,0,255)) < 96)":   "true",
		"to_string(to_real(true))":                  "1",
		"to_string(to_real(false))":                 "0",
		`to_string(to_real(""))`:                    "0",
		"to_string(to_real(nil))":                   "0",
		`to_string(to_real("x") or else 7)`:         "7",
		`to_string(to_real(" 2.5"))`:                "2.5",
		`to_string(to_real("2.5 ") or else "nein")`: "nein",
	}

	for quelle, erwartet := range faelle {
		t.Run(quelle, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("ausrechnen: %v", err)
			}

			if wert.AlsText() != erwartet {
				t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), erwartet)
			}
		})
	}
}
