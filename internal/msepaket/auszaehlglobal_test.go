package msepaket_test

// **`global_script:` an einer Auszählungsdimension** —
// #1355.
//
// # Was nachgelesen ist, und was nicht
//
// **MSE 2.1.2 kennt diesen Schlüssel nicht.** `IMPLEMENT_REFLECTION_NO_GET_MEMBER(StatsDimension)`
// (`src/data/statistics.cpp`, Zeilen 64–77, Fassung `f9d9356` vom 2021-07-06)
// führt `name`, `description`, `position_hint`, `icon`, `script`, `numeric`,
// `bin_size`, `show_empty`, `split_list`, `colors` und `groups` — und sonst
// nichts.
//
// **Gefragt statt erschlossen**: das Orakel scheitert an derselben Stelle wie
// dieses Programm.
//
//	$ echo 'average_opening_hand_statistic(card: set.cards.0)' \
//	    | magicseteditor --cli Colorless.mse-set -q --raw
//	ERROR: Variable not set: global_value
//	         in function average_opening_hand_statistic
//
// `magic.mse-game` schreibt `mse version: 2.5.8`; unser Quelltext und unser
// Orakel sind **2.1.2**. Der Schlüssel gehört einem neueren MSE, dessen
// Quelltext hier nicht vorliegt — nachgesehen wurde in `/Users/david/mse2-quelle`,
// in `/Users/david/magic-set-generator` und im ganzen Pack.
//
// **Gelesen wird deshalb die Datei**, und die ist eindeutig: `global_script:`
// steht neben `script:` und trägt einen Skripttext.

import (
	"testing"
)

// TestEineDimensionLiestIhrGlobalesSkript ist die Grundzusage.
func TestEineDimensionLiestIhrGlobalesSkript(t *testing.T) {
	t.Parallel()

	paket := spielMitBlock(t, `statistics dimension:
	name: durchschnitt
	script: je_karte()
	global_script: einmal_fuers_set()
`)

	dimensionen := paket.Auszaehldimensionen()
	if len(dimensionen) != 1 {
		t.Fatalf("%d Dimensionen, erwartet 1", len(dimensionen))
	}

	if gefunden := dimensionen[0].Globalskript; gefunden != "einmal_fuers_set()" {
		t.Errorf("Globalskript %q", gefunden)
	}

	// **Das Skript je Karte bleibt daneben stehen.**
	if gefunden := dimensionen[0].Skript; gefunden != "je_karte()" {
		t.Errorf("Skript %q", gefunden)
	}
}

// **Ohne den Schlüssel ist es leer** — und das ist der Normalfall: von den
// fünfzig Dimensionen in `magic.mse-game` trägt genau **eine** ein globales
// Skript.
func TestOhneGlobalesSkriptIstEsLeer(t *testing.T) {
	t.Parallel()

	paket := spielMitBlock(t, `statistics dimension:
	name: schlicht
	script: je_karte()
`)

	dimensionen := paket.Auszaehldimensionen()
	if len(dimensionen) != 1 {
		t.Fatalf("%d Dimensionen, erwartet 1", len(dimensionen))
	}

	if dimensionen[0].Globalskript != "" {
		t.Errorf("Globalskript %q, erwartet leer", dimensionen[0].Globalskript)
	}
}
