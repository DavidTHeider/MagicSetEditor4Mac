package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Eine `choice:` als Block trägt ihren Namen in `name:`**
// (#556).
//
//	IMPLEMENT_REFLECTION(ChoiceField::Choice) {
//	  REFLECT_IF_READING_SINGLE_VALUE {
//	    REFLECT_NAMELESS(name);
//	  } else {
//	    // complex values are groups
//	    REFLECT(name);
//
// `src/data/field/choice.cpp`, Zeilen 145–157, Fassung f9d9356 vom
// 2021-07-06.
//
// Über `full-magic-pack` stehen **691** von 8.155 Wahlen als Block — sie
// standen alle namenlos in der Liste. `/magic-modules.mse-include/loyalty/
// choices_cost` schreibt **jede** ihrer sieben so.
func TestEineWahlAlsBlockTraegtIhrenNamen(t *testing.T) {
	t.Parallel()

	pfad := probespielpaket(t, `mse version: 2.0.0
card field:
	type: multiple choice
	name: kostenkasten
	choice: einzeilig
	choice:
		name: als block
		type: radio
	choice:
		type: radio
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

	if len(spiel.Fields) != 1 {
		t.Fatalf("%d Felder, erwartet eines", len(spiel.Fields))
	}

	feld := spiel.Fields[0]
	if feld.Kind != domain.KindMultipleChoice {
		t.Errorf("Art %q, erwartet %q", feld.Kind, domain.KindMultipleChoice)
	}

	// **Die namenlose dritte steht als leere Wahl da.** Erfunden wird kein
	// Name; die Normalform lässt eine leere Wahl nie treffen.
	erwartet := []string{"einzeilig", "als block", ""}
	if len(feld.Choices) != len(erwartet) {
		t.Fatalf("Wahlen %q, erwartet %q", feld.Choices, erwartet)
	}

	for i := range erwartet {
		if feld.Choices[i] != erwartet[i] {
			t.Errorf("Wahl %d ist %q, erwartet %q", i, feld.Choices[i], erwartet[i])
		}
	}
}
