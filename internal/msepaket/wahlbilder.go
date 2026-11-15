package msepaket

// **Der `choice images:`-Block** — je Wahl ein Bild.
//
// Getrennt von `stil.go`, weil es eine eigene Frage ist: dort steht, welche
// Angaben ein Feldstil trägt — hier, wie eine davon aufgebaut ist.

// Wahlbild ist ein Eintrag eines `choice images:`-Blocks.
//
// **Eine Liste und keine Abbildung**, denn die Reihenfolge steht in der Datei
// und gehört zum Rundlauf (§2, dieselbe Begründung wie bei
// [msescript.Eintrag]).
type Wahlbild struct {
	Wahl string
	Bild string
}

// wahlbildwert liest einen Eintrag — inline oder als `script:`-Block.
//
// # Beide Formen stehen im Pack
//
// `magic-m15.mse-style:163` schreibt den Pfad hin:
//
//	choice images:
//	     yes: /magic-mana-small.mse-symbol-font/mana_guild_rg.png
//
// `magic-modules.mse-include/rarities/choice_images` schachtelt ein Skript:
//
//	choice images:
//	     basic land:
//	          script:
//	               if use_main_rarity() then …
//
// **Der erste Anlauf las nur die erste Form** und ließ die Seltenheitssymbole
// leer — genau die 342 Felder, um die es geht. Ein Wahlbild ist eine
// Bildangabe wie jede andere, und die darf ein Skript sein.
func wahlbildwert(knoten *Knoten) string {
	if skript := knoten.Ersten("script"); skript != nil {
		return "{ " + skript.Skripttext() + " }"
	}

	return knoten.Wert
}

// wahlbilderAus liest die Einträge eines `choice images:`-Blocks.
func wahlbilderAus(knoten *Knoten) []Wahlbild {
	bilder := make([]Wahlbild, 0, len(knoten.Kinder))

	for _, kind := range knoten.Kinder {
		if kind.Zeilenart != ArtEintrag {
			continue
		}

		bilder = append(bilder, Wahlbild{Wahl: kind.Schluessel, Bild: wahlbildwert(kind)})
	}

	return bilder
}
