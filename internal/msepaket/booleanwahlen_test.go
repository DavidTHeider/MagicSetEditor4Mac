package msepaket_test

// **Ein `boolean`-Feld hat zwei feste Wahlen**
// (#456).
//
//	BooleanField::BooleanField() {
//	  choices->choices.push_back(make_intrusive<Choice>(_("yes")));
//	  choices->choices.push_back(make_intrusive<Choice>(_("no")));
//	  choices->initIds();
//	}
//
// `src/data/field/boolean.cpp`, Zeilen 14–18, Fassung f9d9356 vom 2021-07-06.
// Ohne `initial:` gilt damit die **erste** — `yes`.

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestEinBooleanFeldHatZweiWahlen(t *testing.T) {
	pfad := probespielpaket(t, `mse version: 2.0.0
set field:
	type: boolean
	name: shorten types for rarity
set field:
	type: boolean
	name: inverted common symbol
	initial: no
set field:
	type: choice
	name: farbe
	choice: rot
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

	for _, fall := range []struct {
		name    string
		wahlen  []string
		vorgabe string
	}{
		{name: "shorten types for rarity", wahlen: []string{"yes", "no"}, vorgabe: ""},
		{name: "inverted common symbol", wahlen: []string{"yes", "no"}, vorgabe: "no"},
		// **Die Gegenprobe**: ein gewöhnliches Auswahlfeld bekommt keine
		// erfundenen Wahlen dazu.
		{name: "farbe", wahlen: []string{"rot"}, vorgabe: ""},
	} {
		t.Run(fall.name, func(t *testing.T) {
			stelle := slices.IndexFunc(spiel.SetFields, func(f domain.Field) bool {
				return f.Name == fall.name
			})
			if stelle < 0 {
				t.Fatalf("kein Feld %q", fall.name)
			}

			feld := spiel.SetFields[stelle]

			if !slices.Equal(feld.Choices, fall.wahlen) {
				t.Errorf("Wahlen %v, erwartet %v", feld.Choices, fall.wahlen)
			}

			if feld.Initial != fall.vorgabe {
				t.Errorf("initial %q, erwartet %q", feld.Initial, fall.vorgabe)
			}
		})
	}
}
