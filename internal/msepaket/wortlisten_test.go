package msepaket_test

import "testing"

// **Die `word list:`-Blöcke eines Spiels werden gelesen** —
// #1074.
//
// # Was gemessen ist
//
// `magic.mse-game/word_lists` führt **51** Listen mit **432** Wörtern auf
// oberster Ebene und **445** in Untergruppen; 100 tragen `is prefix:`, 77 ein
// `script:`, 60 ein `line below:`. Gemessen am 2026-08-24 am bezogenen
// `M15-Magic-Pack`.
func TestDieWortlistenEinesSpielsWerdenGelesen(t *testing.T) {
	paket := spielMitBlock(t, `word list:
	name: type-en
	word:
		name: Legendary 
		is prefix: true
		line below: true
	word: Creature
	word: Instant
`)

	listen := paket.Wortlisten()

	if len(listen) != 1 {
		t.Fatalf("%d Listen, erwartet eine: %+v", len(listen), listen)
	}

	if listen[0].Name != "type-en" {
		t.Errorf("die Liste heißt %q", listen[0].Name)
	}

	if len(listen[0].Woerter) != 3 {
		t.Fatalf("%d Wörter: %+v", len(listen[0].Woerter), listen[0].Woerter)
	}

	// **Das Leerzeichen bleibt stehen.** `Legendary ` trägt es, weil ein
	// Präfix vor das nächste Wort gesetzt wird; wer es abschnitte, schriebe
	// `LegendaryCreature`.
	erstes := listen[0].Woerter[0]
	if erstes.Name != "Legendary " || !erstes.IstPraefix || !erstes.ZeileDarunter {
		t.Errorf("das erste Wort ist %+v", erstes)
	}

	// **Ein bloßer Wert ist der Name.** `REFLECT_IF_READING_SINGLE_VALUE {
	// REFLECT_NAMELESS(name); }` — `src/data/word_list.cpp`, Zeilen 19–29,
	// Fassung f9d9356 vom 2021-07-06.
	if listen[0].Woerter[1].Name != "Creature" {
		t.Errorf("das zweite Wort ist %+v", listen[0].Woerter[1])
	}

	// **Ohne Angabe ist beides falsch** — `WordListWord::WordListWord`,
	// `src/data/word_list.cpp`, Zeilen 14–17, Fassung f9d9356 vom 2021-07-06.
	if listen[0].Woerter[1].IstPraefix || listen[0].Woerter[1].ZeileDarunter {
		t.Errorf("ein bloßer Wert darf nichts weiter tragen: %+v", listen[0].Woerter[1])
	}
}

// **Ein Wort kann seine Namen aus einem Skript holen.**
//
// Im Pack steht `script: all_sub_types(setting: "is_land")` — 77-mal.
func TestEinWortMitSkript(t *testing.T) {
	paket := spielMitBlock(t, `word list:
	name: land
	word:
		script: all_sub_types(setting: "is_land")
		line below: true
`)

	listen := paket.Wortlisten()

	if len(listen) != 1 || len(listen[0].Woerter) != 1 {
		t.Fatalf("gelesen wurde %+v", listen)
	}

	if listen[0].Woerter[0].Skript != `all_sub_types(setting: "is_land")` {
		t.Errorf("das Skript ist %q", listen[0].Woerter[0].Skript)
	}
}

// **Ein Wort mit eigenen Wörtern ist eine Gruppe** — ein Untermenü.
//
// `inline bool isGroup() const { return !words.empty(); }` —
// `src/data/word_list.hpp`, Zeile 31, Fassung f9d9356 vom 2021-07-06.
func TestEineGruppeTraegtEigeneWoerter(t *testing.T) {
	paket := spielMitBlock(t, `word list:
	name: type-zhs
	word:
		name: Nichtstandard
		word: Plane
		word: Scheme
`)

	listen := paket.Wortlisten()

	if len(listen) != 1 || len(listen[0].Woerter) != 1 {
		t.Fatalf("gelesen wurde %+v", listen)
	}

	gruppe := listen[0].Woerter[0]
	if gruppe.Name != "Nichtstandard" || len(gruppe.Woerter) != 2 {
		t.Fatalf("die Gruppe ist %+v", gruppe)
	}

	if gruppe.Woerter[0].Name != "Plane" || gruppe.Woerter[1].Name != "Scheme" {
		t.Errorf("die Untergruppe ist %+v", gruppe.Woerter)
	}

	if !gruppe.IstGruppe() {
		t.Error("ein Eintrag mit Wörtern muss eine Gruppe sein")
	}

	if gruppe.Woerter[0].IstGruppe() {
		t.Error("ein Eintrag ohne Wörter darf keine sein")
	}
}

// **Eine Liste ohne Namen wird übergangen.**
//
// `findWordLists` sucht die Liste über ihren Namen (`src/gui/value/text.cpp`,
// Zeilen 1362–1367, Fassung f9d9356 vom 2021-07-06); eine namenlose wäre nicht
// zu finden und stünde nur im Weg.
func TestEineListeOhneNamenWirdUebergangen(t *testing.T) {
	paket := spielMitBlock(t, `word list:
	word: Creature
word list:
	name: da
	word: Instant
`)

	listen := paket.Wortlisten()

	if len(listen) != 1 || listen[0].Name != "da" {
		t.Errorf("gelesen wurde %+v", listen)
	}
}
