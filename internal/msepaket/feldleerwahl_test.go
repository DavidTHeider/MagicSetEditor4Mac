package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Auch ein Kartenfeld trägt seinen Leernamen** —
// #2006.
//
//	void MultipleChoiceValue::normalForm() {
//	  …
//	  // empty choice name
//	  if (val.empty()) { val = field().empty_choice; }
//	}
//
// `src/data/field/multiple_choice.cpp`, Zeilen 110–113, Fassung f9d9356 vom
// 2021-07-06. Hier stand, `empty choice:` gebe es nur an einem
// `styling field:` — `magic.mse-game` erklärt es an **sieben** Kartenfeldern
// (`card color` bis `card color 4` und `indicator` bis `indicator 3`), jedes Mal als
// `colorless`. Das Orakel gibt `set.cards.N.card_color` für eine farblose Karte
// als `"colorless"`, und `color_of_card()` fragt danach: eine einzige Karte aus
// `Black.mse-set` stand deshalb an anderer Stelle der Ausfuhr.
func TestEinKartenfeldTraegtSeinenLeernamen(t *testing.T) {
	t.Parallel()

	pfad := probespielpaket(t, `mse version: 2.0.0
card field:
	type: multiple choice
	name: farbe
	empty choice: farblos
	choice: rot
	choice: blau
card field:
	type: multiple choice
	name: form
	choice: rund
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	spiel, _ := paket.Spielmodell()
	if len(spiel.Fields) != 2 {
		t.Fatalf("%d Felder", len(spiel.Fields))
	}

	if spiel.Fields[0].Leerwahl != "farblos" {
		t.Errorf("Leerwahl %q, erwartet »farblos«", spiel.Fields[0].Leerwahl)
	}

	if spiel.Fields[1].Leerwahl != "" {
		t.Errorf("Leerwahl %q, erwartet keine", spiel.Fields[1].Leerwahl)
	}
}
