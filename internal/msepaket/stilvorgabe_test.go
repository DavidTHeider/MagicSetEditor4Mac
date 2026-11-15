package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Eine Stiloption bringt ihren Vorgabeausdruck mit** —
// #590.
//
// # Roh und ungedeutet
//
// `default:` ist bei MSE ein `OptionalScript` (`src/data/field/color.hpp`,
// Zeile 33, Fassung f9d9356 vom 2021-07-06) — also Quelltext und nicht ein
// Wert. Hier steht deshalb, was in der Datei stand, samt Anführungszeichen;
// ausgewertet wird in `internal/service`, weil eine Fachbibliothek den
// Interpreter nicht kennen darf (§2).
//
// # Die beiden Fälle sind aus dem Pack, nicht ausgedacht
//
// Über die 376 Stilpakete von `full-magic-pack` tragen zehn von 105
// Farboptionen ein `default:`. Achtmal steht dort `"black"` in
// Anführungszeichen — die Form aus `magic-booster`. Einmal steht das nackte
// `black` neben einem `initial:`, und das ist `magic-m15-bigtext`.
func TestEineStiloptionFuehrtIhrenVorgabeausdruck(t *testing.T) {
	pfad := probestilpaket(t, `mse version: 2.0.0
game: probe
card width: 100
card height: 100
styling field:
	type: color
	name: title color
	default: "black"
styling field:
	type: color
	name: text color
	default: black
	initial: rgb(9,9,9)
styling field:
	type: color
	name: border color
	initial: rgb(1,2,3)
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

	stil, befunde := paket.Stilmodell()
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(stil.Stiloptionen) != 3 {
		t.Fatalf("%d Stiloptionen: %+v", len(stil.Stiloptionen), stil.Stiloptionen)
	}

	faelle := []struct {
		stelle          int
		vorgabe, anfang string
	}{
		{0, `"black"`, ""},
		{1, "black", "rgb(9,9,9)"},

		// **Ohne `default:` bleibt das Feld leer** — und nicht etwa gleich
		// `initial:`. Die beiden sind verschiedene Angaben, und wer sie
		// zusammenwirft, kann später nicht mehr sagen, welche galt.
		{2, "", "rgb(1,2,3)"},
	}

	for _, fall := range faelle {
		option := stil.Stiloptionen[fall.stelle]
		if option.Default != fall.vorgabe || option.Initial != fall.anfang {
			t.Errorf("%q: default %q, initial %q — erwartet %q und %q",
				option.Name, option.Default, option.Initial, fall.vorgabe, fall.anfang)
		}
	}
}
