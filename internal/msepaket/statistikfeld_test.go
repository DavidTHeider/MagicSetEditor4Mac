package msepaket_test

// **Was ein Feld der Statistik mitgibt** —
// #1985.
//
// MSE liest an jedem Feld `show statistics:` — Vorgabe **ja**
// (`src/data/field.cpp`, Zeilen 28 und 55) — und an einem Auswahlfeld
// `choice colors:` (`src/data/field/choice.cpp`, Zeile 37; Fassung f9d9356
// vom 2021-07-06). Beides braucht die automatische Dimension, die MSE für
// jedes Kartenfeld anlegt (`Game::validate`, `src/data/game.cpp`, Zeilen
// 68–77).

import "testing"

func TestEinFeldSagtObUndWieEsInDieStatistikGeht(t *testing.T) {
	t.Parallel()

	paket := spielMitBlock(t, "card field:\n"+
		"\ttype: choice\n\tname: rarity\n"+
		"\tchoice: common\n\tchoice: mythic rare\n"+
		"\tchoice colors:\n"+
		"\t\tcommon: rgb(33,33,33)\n"+
		"\t\tmythic rare: rgb(213,100,13)\n"+
		"card field:\n"+
		"\ttype: text\n\tname: notes\n"+
		"\tshow statistics: false\n")

	spiel, befunde := paket.Spielmodell()
	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(spiel.Fields) != 2 {
		t.Fatalf("%d Felder", len(spiel.Fields))
	}

	seltenheit, notizen := spiel.Fields[0], spiel.Fields[1]

	if !seltenheit.Statistik || notizen.Statistik {
		t.Errorf("Statistik: rarity %v (Vorgabe ja), notes %v (nein)", seltenheit.Statistik, notizen.Statistik)
	}

	if seltenheit.Statistikfarben["mythic rare"] != "rgb(213,100,13)" ||
		seltenheit.Statistikfarben["common"] != "rgb(33,33,33)" {
		t.Errorf("Farben %v", seltenheit.Statistikfarben)
	}
}
