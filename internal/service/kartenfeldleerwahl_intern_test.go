package service

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Eine leere Mehrfachwahl am Kartenfeld schreibt ihren Leernamen** —
// #2006.
//
// `normalForm` setzt `empty_choice`, wenn nichts gewählt ist
// (`src/data/field/multiple_choice.cpp`, Zeilen 110–113, Fassung f9d9356 vom
// 2021-07-06) — an `card color` von `magic.mse-game` heißt er `colorless`.
// Hier stand „ein Kartenfeld hat keinen Leernamen".
func TestEineLeereMehrfachwahlEinesKartenfeldsSchreibtIhrenLeernamen(t *testing.T) {
	t.Parallel()

	felder := []domain.Field{
		{
			Name: "card color", Kind: domain.KindMultipleChoice,
			Choices: []string{"white", "blue"}, Leerwahl: "colorless",
		},
		{
			Name: "indicator", Kind: domain.KindMultipleChoice,
			Choices: []string{"white", "blue"},
		},
	}

	for _, fall := range []struct{ name, wert, erwartet string }{
		{"nichts gewählt", "", "colorless"},
		{"eine Wahl gewählt", "blue", "blue"},
		{"nur Unbekanntes", "grün", "colorless"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			gesehen := normalisiereMehrfachwahlen([]msescript.Eintrag{
				{Name: "card_color", Wert: msescript.Text(fall.wert)},
				{Name: "indicator", Wert: msescript.Text(fall.wert)},
			}, felder)

			if gesehen[0].Wert.Text != fall.erwartet {
				t.Errorf("card_color %q, erwartet %q", gesehen[0].Wert.Text, fall.erwartet)
			}

			// **Ohne Angabe die leere Zeichenkette** — das Feld im Kopf von
			// `MultipleChoiceField` (`multiple_choice.hpp`, Zeile 27).
			if fall.wert != "blue" && gesehen[1].Wert.Text != "" {
				t.Errorf("indicator %q, erwartet leer", gesehen[1].Wert.Text)
			}
		})
	}
}
