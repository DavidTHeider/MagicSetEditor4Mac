package msepaket_test

import "testing"

// **Die `pack type:`-Blöcke eines Spiels werden gelesen** —
// #1072.
//
// # Was gemessen ist
//
// `magic.mse-game/packs` führt **30** Blöcke, eingebunden über
// `include file: packs`. Die benutzten Auswahlarten: `equal` (5), `first` (4),
// `no replace` (1), `nonempty` (1), `proportional` (1) — der Rest steht auf
// `auto`.
func TestDiePacksEinesSpielsWerdenGelesen(t *testing.T) {
	paket := spielMitBlock(t, `pack type:
	name: common
	filter: card.rarity == "common"
pack type:
	name: booster pack
	item:
		name: common
		amount: 10
	item:
		name: rare
		amount: 1
		weight: 2
`)

	packs := paket.Packs()

	if len(packs) != 2 {
		t.Fatalf("%d Packs, erwartet zwei: %+v", len(packs), packs)
	}

	if packs[0].Name != "common" || packs[0].Filter != `card.rarity == "common"` {
		t.Errorf("das erste Pack ist %+v", packs[0])
	}

	// **`auto` heißt: mit Filter »no replace«, sonst »all«** —
	// `after_reading(PackType&)`, `src/data/pack.cpp`, Zeilen 43–46, Fassung (Fassung f9d9356 vom 2021-07-06)
	// f9d9356 vom 2021-07-06. Das ist keine Vorgabe, sondern eine Ableitung.
	if packs[0].Auswahl != "no replace" {
		t.Errorf("ein Pack mit Filter muss »no replace« werden, ist %q", packs[0].Auswahl)
	}

	if packs[1].Auswahl != "all" {
		t.Errorf("ein Pack mit Posten muss »all« werden, ist %q", packs[1].Auswahl)
	}

	if len(packs[1].Posten) != 2 {
		t.Fatalf("%d Posten", len(packs[1].Posten))
	}

	if packs[1].Posten[0].Name != "common" || packs[1].Posten[0].Anzahl != 10 {
		t.Errorf("der erste Posten ist %+v", packs[1].Posten[0])
	}

	// **Ohne Angabe wiegt ein Posten eins** — `PackItem::PackItem`,
	// `src/data/pack.cpp`, Zeilen 75–79, Fassung f9d9356 vom 2021-07-06.
	if packs[1].Posten[0].Gewicht != 1 {
		t.Errorf("ohne Angabe muss das Gewicht eins sein, ist %v", packs[1].Posten[0].Gewicht)
	}

	if packs[1].Posten[1].Gewicht != 2 {
		t.Errorf("»weight: 2« wurde nicht gelesen: %+v", packs[1].Posten[1])
	}
}

// **Die genannte Auswahlart bleibt stehen.**
func TestEineGenannteAuswahlartBleibtStehen(t *testing.T) {
	paket := spielMitBlock(t, `pack type:
	name: land
	select: equal
	filter: card.rarity == "basic land"
`)

	packs := paket.Packs()

	if len(packs) != 1 || packs[0].Auswahl != "equal" {
		t.Fatalf("Packs %+v", packs)
	}
}

// **Ein Pack ohne Namen wird übergangen.**
//
// Es ließe sich nicht nennen — und ein Posten, der auf nichts zeigt, ist ein
// Fehler, den niemand finden kann.
func TestEinPackOhneNamenWirdUebergangen(t *testing.T) {
	paket := spielMitBlock(t, `pack type:
	filter: card.rarity == "common"
pack type:
	name: mit
	filter: x
`)

	if packs := paket.Packs(); len(packs) != 1 || packs[0].Name != "mit" {
		t.Fatalf("Packs %+v", packs)
	}
}

// **Ein Pack ohne Filter und ohne Posten bleibt `auto`.**
//
// MSE lässt es dort ebenso stehen (`after_reading`, `src/data/pack.cpp`,
// Zeilen 43–46, Fassung f9d9356 vom 2021-07-06) — und der Zieher zieht daraus
// nichts. Eine erfundene Auflösung wäre eine Behauptung.
func TestEinPackOhneAllesBleibtAuto(t *testing.T) {
	paket := spielMitBlock(t, "pack type:\n\tname: leer\n")

	packs := paket.Packs()

	if len(packs) != 1 || packs[0].Auswahl != "auto" {
		t.Fatalf("Packs %+v", packs)
	}
}

// **Die Randfälle des Packlesens.**
func TestDieRandfaelleDesPacklesens(t *testing.T) {
	paket := spielMitBlock(t, `pack type:
	name: krumm
	select: equal
	item:
		name: ohneZahlen
		amount: viele
		weight: schwer
	item:
		amount: 2
`)

	packs := paket.Packs()

	if len(packs) != 1 {
		t.Fatalf("%d Packs", len(packs))
	}

	// **Ein Posten ohne Namen fällt weg** — er zeigte auf nichts.
	if len(packs[0].Posten) != 1 {
		t.Fatalf("%d Posten, erwartet einen: %+v", len(packs[0].Posten), packs[0].Posten)
	}

	// **Was keine Zahl ist, bekommt die Vorgabe** — und der Zieher rechnet
	// weiter, statt das ganze Pack fallen zu lassen (§8).
	if packs[0].Posten[0].Anzahl != 1 || packs[0].Posten[0].Gewicht != 1 {
		t.Errorf("der Posten ist %+v, erwartet Anzahl und Gewicht eins", packs[0].Posten[0])
	}
}
