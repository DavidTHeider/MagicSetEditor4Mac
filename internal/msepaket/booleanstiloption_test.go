package msepaket_test

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Eine Stiloption vom Typ `boolean` hat die Wahlen `yes` und `no`**
// (#531).
//
// # Die Regel war schon portiert — und blieb auf halbem Weg stehen
//
//	BooleanField::BooleanField() {
//	  choices->choices.push_back(make_intrusive<Choice>(_("yes")));
//	  choices->choices.push_back(make_intrusive<Choice>(_("no")));
//	  choices->initIds();
//	}
//
// `src/data/field/boolean.cpp`, Zeilen 14–18, Fassung f9d9356 vom 2021-07-06.
//
// #456 hat
// das für `set field:` und `card field:` des **Spielpakets** gebaut
// (`spiel.go`). Für die `styling field:` eines **Stilpakets** fehlte es — und
// das ist die Klasse, die §10 anstreicht: nicht den Fall behandeln, sondern
// die Klasse.
//
// # Umfang
//
// Über `/Users/david/full-magic-pack/data` gezählt: **1.489** `styling field:`
// vom Typ `boolean`, davon **378 ohne `initial:`**, in **178 von 365**
// Stilpaketen. Alle standen bei uns auf „nein", wo MSE „ja" liest — still,
// ohne Befund, mit einem plausiblen Bild.
func TestEineBooleanStiloptionHatZweiWahlen(t *testing.T) {
	t.Parallel()

	pfad := probestilpaket(t, `mse version: 2.0.0
game: probe
short name: Probe
styling field:
	type: boolean
	name: default image size
styling field:
	type: boolean
	name: mit Vorgabe
	initial: no
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	defer func() { _ = paket.Schliessen() }() //nolint:errcheck // gelesen ist gelesen

	stil, _ := paket.Stilmodell()

	ohne := stiloption(t, stil.Stiloptionen, "default image size")
	if !slices.Equal(ohne.Choices, []string{"yes", "no"}) {
		t.Errorf("Wahlen %v, erwartet [yes no]", ohne.Choices)
	}

	// **Die Reihenfolge ist die Aussage.** Ohne `initial:` gilt die erste, und
	// das ist `yes`.
	if len(ohne.Choices) > 0 && ohne.Choices[0] != "yes" {
		t.Errorf("die erste Wahl ist %q, erwartet \"yes\"", ohne.Choices[0])
	}

	// **Ein `initial:` schlägt die erste Wahl** — die Wahlen stehen trotzdem da.
	mit := stiloption(t, stil.Stiloptionen, "mit Vorgabe")
	if !slices.Equal(mit.Choices, []string{"yes", "no"}) {
		t.Errorf("mit initial: Wahlen %v, erwartet [yes no]", mit.Choices)
	}

	if mit.Initial != "no" {
		t.Errorf("Initial %q, erwartet \"no\"", mit.Initial)
	}
}

// stiloption sucht eine Stiloption am Namen.
func stiloption(t *testing.T, optionen []domain.StyleField, name string) domain.StyleField {
	t.Helper()

	for _, option := range optionen {
		if option.Name == name {
			return option
		}
	}

	t.Fatalf("die Stiloption %q fehlt", name)

	return domain.StyleField{}
}
