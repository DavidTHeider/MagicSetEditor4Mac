package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Mehrfachwahl trägt ihren Leernamen**
// (#637).
//
// **ported.** `normalForm` schreibt `empty_choice`, wenn nichts gewählt ist
// (`src/data/field/multiple_choice.cpp`, Zeilen 110–112, Fassung f9d9356 vom
// 2021-07-06). Ohne Angabe ist er die **leere Zeichenkette** — so steht das
// Feld im Kopf der Klasse (`multiple_choice.hpp`, Zeile 27).
//
// **Über das Pack kommt `empty choice:` nirgends vor** — keine der 33
// Mehrfachwahlen erklärt eine
// (#630).
// Gelesen wird sie trotzdem: eine Angabe, die man nicht liest, kann man nicht
// einhalten.
func TestEineMehrfachwahlTraegtIhrenLeernamen(t *testing.T) {
	optionen := stiloptionenDesPruefstuecks(t)

	for name, erwartet := range map[string]string{
		"rahmen":  "",
		"leerbar": "nichts",
	} {
		option, da := optionen[name]
		if !da {
			t.Fatalf("die Stiloption %q fehlt", name)
		}

		if option.Kind != domain.KindMultipleChoice {
			t.Errorf("%s: Art %q, erwartet %q", name, option.Kind, domain.KindMultipleChoice)
		}

		if option.Leerwahl != erwartet {
			t.Errorf("%s: Leerwahl %q, erwartet %q", name, option.Leerwahl, erwartet)
		}
	}
}

// **Und ihre Wahlen in der Reihenfolge der Erklärung** — an ihr hängt die
// Normalform.
func TestEineMehrfachwahlTraegtIhreWahlenInDerErklaerungsreihenfolge(t *testing.T) {
	rahmen := stiloptionenDesPruefstuecks(t)["rahmen"]

	erwartet := []string{"nyx", "schnee", "fahrzeug"}

	if len(rahmen.Choices) != len(erwartet) {
		t.Fatalf("%+v, erwartet %+v", rahmen.Choices, erwartet)
	}

	for i := range erwartet {
		if rahmen.Choices[i] != erwartet[i] {
			t.Errorf("Stelle %d: %q, erwartet %q", i, rahmen.Choices[i], erwartet[i])
		}
	}
}
