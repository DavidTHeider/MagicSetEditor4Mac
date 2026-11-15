package msescript_test

// **`in_context` — ein Treffer zählt nur in seinem Zusammenhang** — #2315
// (E2315).
//
// MSE2 setzt für jeden Treffer die Zeichenkette „davor + `<match>` +
// danach" zusammen und sucht den Kontext darin
// (`ScriptRegex::matches`, `src/script/functions/regex.cpp`, Zeilen 29–51,
// Fassung f9d9356 vom 2021-07-06); `replace`, `filter_text` und `break_text`
// nehmen ihn (ebd. Zeilen 113–208). **Hier wurde er bis dahin gar nicht
// gelesen**: jede Ersetzung mit Kontext wirkte wie eine ohne. Gefunden an den
// sechs Kanonkarten aus #2315 — `magic.mse-game` macht mit
// `in_context: mana_context` aus „non-Zombie" ein Symbol „-Z", wo MSE2 es
// Text lässt.
//
// **Jede Erwartung ist vom Orakel erfragt** (`magicseteditor --cli … -q
// --raw`, Fassung f9d9356, am 2026-10-14) — die Fälle sind eigene, kleine
// Sätze und kein Inhalt eines Pakets.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestEinTrefferZaehltNurInSeinemKontext(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name, quelle, erwartet string
	}{
		{
			"ein Zeichen mitten im Wort ist kein Kostensymbol",
			`replace("non-Zombie, +1: x -2: y", match: "[+-][XYZ0-9]", in_context: "(^|[ ])<match>:", replace: "<s>&</s>")`,
			"non-Zombie, <s>+1</s>: x <s>-2</s>: y",
		},
		{
			"nur, was hinter einem y steht",
			`replace("x1 y2 y3", match: "[0-9]", in_context: "y<match>", replace: "#")`,
			"x1 y# y#",
		},
		{
			"mit einer Ersatzfunktion",
			`replace("x1 y2 y3", match: "[0-9]", in_context: "y<match>", replace: {"[" + input + "]"})`,
			"x1 y[2] y[3]",
		},
		{
			"mit Gruppen in der Vorlage",
			`replace("x1 y2 y3", match: "([a-z])([0-9])", in_context: "<match> y", replace: "\\2\\1")`,
			"1x 2y y3",
		},
		{
			"eine Gruppe, die das Muster nicht hat, ist leer",
			`replace("x1 y2", match: "[0-9]", in_context: "y<match>", replace: "<\\2>")`,
			"x1 y<>",
		},
		{
			"der Kontext darf bis ans Ende greifen",
			`replace("aXb", match: "X", in_context: "a<match>b$", replace: "Y")`,
			"aYb",
		},
		{
			"Anfang oder Ende",
			`replace("aaa", match: "a", in_context: "^<match>|<match>$", replace: "b")`,
			"bab",
		},
		{
			"filter_text behält nur die Treffer im Kontext",
			`filter_text("a1 b2 c3 b4", match: "[0-9]", in_context: "b<match>")`,
			"24",
		},
		{
			"break_text ebenso",
			`break_text("a1 b2 c3", match: "[a-z][0-9]", in_context: "<match> c")`,
			"b2",
		},
		{
			"die Regelfassung reicht den Kontext durch",
			`(replace_rule(match: "[0-9]", in_context: "y<match>", replace: "#"))("x1 y2")`,
			"x1 y#",
		},
		{
			"filter_rule ebenso",
			`(filter_rule(match: "[0-9]", in_context: "^a<match>"))("a1 b2")`,
			"1",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("%v", err)
			}

			// **Eine Liste mit einem Eintrag als dieser Eintrag** — so gibt
			// `break_text` zurück, und das Orakel zeigt `["b2"]`.
			gesehen := wert.AlsText()
			if wert.Art == msescript.ArtAbbildung && len(wert.Eintraege) == 1 {
				gesehen = wert.Eintraege[0].Wert.AlsText()
			}

			if gesehen != fall.erwartet {
				t.Errorf("ergibt %q, MSE2 gibt %q", gesehen, fall.erwartet)
			}
		})
	}
}
