package msepaket_test

// **`statistics dimension:` — wonach ein Spielpaket auszählen lässt** —
// Werkstück C von Phase 4.11
// (#1338,
// #1349).
//
// # Was am 2026-08-31 nachgelesen wurde (§15)
//
// `src/data/statistics.hpp`, Zeilen 25–45, und
// `IMPLEMENT_REFLECTION_NO_GET_MEMBER(StatsDimension)` in
// `src/data/statistics.cpp`, Zeilen 64–78 (Fassung f9d9356 vom 2021-07-06):
//
//	REFLECT(name);            REFLECT_LOCALIZED(description);
//	REFLECT(position_hint);   REFLECT_N("icon", icon_filename);
//	REFLECT(script);          REFLECT(numeric);
//	REFLECT(bin_size);        REFLECT(show_empty);
//	REFLECT(split_list);      REFLECT(colors);
//	REFLECT(groups);
//
// Elf Angaben. `groups` und `colors` sind Mehrfachschlüssel: jede Zeile
// `group:` hängt einen Namen an, `colors:` ist ein Unterblock aus Name und
// Farbe.
//
// # Und was im Pack steht — gemessen am 2026-08-31
//
// **685 Blöcke über den ganzen Pack**, davon 650 in `magic.mse-game`: 50 je
// Sprachfassung, dreizehn Fassungen (`statistics_en`, `statistics_de`, …),
// eingebunden über `include localized file: statistics`. Die übrigen 35
// stehen unmittelbar in fünf `game`-Dateien und in `space.mse-game/statistics`.
//
// **`statistics category:` kommt im ganzen Pack null-mal vor.** MSE2 baut die
// Kategorien dann selbst aus den Dimensionen
// (`StatsCategory::StatsCategory(const StatsDimensionP&)`); ein Leser dafür
// hätte hier keinen einzigen Fall zu lesen. Die Reichweite dieser Zahl ist
// dieses Pack.

import (
	"testing"
)

// **Ein `statistics dimension:` wird mit allen elf Angaben gelesen.**
func TestEineAuszaehldimensionWirdGelesen(t *testing.T) {
	t.Parallel()

	paket := spielMitBlock(t, `statistics dimension:
	name: farbgruppe
	position hint: 0000
	icon: stats/farbe.png
	description: Wie viele Karten je Farbe im Set liegen
	split list: true
	show empty: false
	script: farbgruppe_von(card)
	group: Land
	group: Weiss
	colors:
		Weiss	: rgb(255,231,185)
		Land	: rgb(103,64,35)
`)

	dimensionen := paket.Auszaehldimensionen()

	if len(dimensionen) != 1 {
		t.Fatalf("%d Dimensionen, erwartet war eine", len(dimensionen))
	}

	eine := dimensionen[0]

	if eine.Name != "farbgruppe" {
		t.Errorf("Name %q", eine.Name)
	}

	if eine.Beschreibung != "Wie viele Karten je Farbe im Set liegen" {
		t.Errorf("Beschreibung %q", eine.Beschreibung)
	}

	if eine.Rang != "0000" {
		t.Errorf("Rang %q", eine.Rang)
	}

	if eine.Sinnbild != "stats/farbe.png" {
		t.Errorf("Sinnbild %q", eine.Sinnbild)
	}

	if eine.Skript != "farbgruppe_von(card)" {
		t.Errorf("Skript %q", eine.Skript)
	}

	if !eine.Zerlegen {
		t.Error("split list: true kam nicht an")
	}

	if eine.LeereZeigen {
		t.Error("show empty: false kam nicht an")
	}

	// **Die Reihenfolge der Gruppen ist die erklärte.** Sie ist die Antwort
	// darauf, dass „Land" vor „Weiss" stehen soll und nicht alphabetisch
	// dahinter.
	if len(eine.Gruppen) != 2 || eine.Gruppen[0] != "Land" || eine.Gruppen[1] != "Weiss" {
		t.Errorf("Gruppen %+v", eine.Gruppen)
	}

	if eine.Farben["Land"] != "rgb(103,64,35)" {
		t.Errorf("Farben %+v", eine.Farben)
	}
}

// **`numeric:` und `bin size:` kommen als Zahl an.**
//
// MSEs `bin_size` ist ein `double`; `to_bin` (`src/gui/control/graph.cpp`,
// Zeilen 67–73, Fassung f9d9356 vom 2021-07-06) rechnet damit `ceil(value / bin_size)` und beschriftet die
// Klasse als `(bin-1)*bin_size+1 – bin*bin_size`.
func TestEineNumerischeAuszaehldimensionTraegtIhreKlassenbreite(t *testing.T) {
	t.Parallel()

	paket := spielMitBlock(t, `statistics dimension:
	name: manawert
	numeric: true
	bin size: 5
	script: card.manawert
`)

	dimensionen := paket.Auszaehldimensionen()
	if len(dimensionen) != 1 {
		t.Fatalf("%d Dimensionen", len(dimensionen))
	}

	if !dimensionen[0].Numerisch {
		t.Error("numeric: true kam nicht an")
	}

	if dimensionen[0].Klassenbreite != 5 {
		t.Errorf("Klassenbreite %v", dimensionen[0].Klassenbreite)
	}
}

// **Die Vorgaben sind MSEs Vorgaben** — der Konstruktor
// `StatsDimension::StatsDimension()` (`statistics.cpp`, Zeilen 18–25) setzt
// `numeric(false)`, `bin_size(0)`, `show_empty(false)`, `split_list(false)`.
func TestEineAuszaehldimensionOhneAngabenTraegtMSEsVorgaben(t *testing.T) {
	t.Parallel()

	paket := spielMitBlock(t, `statistics dimension:
	name: schlicht
	script: card.name
`)

	dimensionen := paket.Auszaehldimensionen()
	if len(dimensionen) != 1 {
		t.Fatalf("%d Dimensionen", len(dimensionen))
	}

	eine := dimensionen[0]

	if eine.Numerisch || eine.Zerlegen || eine.LeereZeigen || eine.Klassenbreite != 0 {
		t.Errorf("die Vorgaben lauten %+v", eine)
	}

	// **Eine leere Tabelle und kein `nil`** — dieselbe Zusage wie überall
	// sonst (§3): wer sie liest, prüft auf eines statt auf zwei.
	if eine.Farben == nil || eine.Gruppen == nil {
		t.Errorf("Farben %v, Gruppen %v", eine.Farben, eine.Gruppen)
	}
}

// **Ein Spielpaket ohne Dimensionen gibt eine leere Liste.**
func TestOhneAuszaehldimensionenKommtEineLeereListe(t *testing.T) {
	t.Parallel()

	if dimensionen := spielMitBlock(t, "").Auszaehldimensionen(); len(dimensionen) != 0 {
		t.Errorf("%d Dimensionen, erwartet war keine", len(dimensionen))
	}
}

// **Eine unbrauchbare Klassenbreite gibt null und keinen Befund.**
//
// MSE liest `bin_size` über seinen Reader: der meldet eine unbrauchbare Zahl
// und lässt den Vorgabewert stehen. Eine Dimension, die deshalb ganz ausfiele,
// wäre strenger als das Original (§15, Satz 4).
func TestEineUnbrauchbareKlassenbreiteGibtNull(t *testing.T) {
	t.Parallel()

	paket := spielMitBlock(t, `statistics dimension:
	name: krumm
	bin size: fünf
	script: card.name
`)

	dimensionen := paket.Auszaehldimensionen()
	if len(dimensionen) != 1 || dimensionen[0].Klassenbreite != 0 {
		t.Errorf("die Dimensionen lauten %+v", dimensionen)
	}
}

// **Eine Leerzeile im `colors:`-Block ist keine Farbe.**
//
// Im Pack stehen die Farbtafeln mit Leerzeilen und Kommentaren dazwischen —
// `statistics_en` trennt die Blöcke so.
//
// **Und diese Zusage hält der Zerteiler, nicht der Leser**: Leerzeilen,
// Kommentare und Fortsetzungen werden gar nicht erst als Kind eines Blocks
// abgelegt (gemessen am 2026-08-31). Der Fall steht hier, damit dieselbe
// Zusage bemerkt wird, wenn sie sich ändert — dann bekäme die Farbtafel einen
// Eintrag mit leerem Namen.
func TestEineLeerzeileImFarbblockIstKeineFarbe(t *testing.T) {
	t.Parallel()

	paket := spielMitBlock(t, `statistics dimension:
	name: farbe
	script: card.farbe
	colors:
		rot	: rgb(156,15,15)

		# eine Anmerkung
		blau	: rgb(15,80,166)
`)

	dimensionen := paket.Auszaehldimensionen()
	if len(dimensionen) != 1 {
		t.Fatalf("%d Dimensionen", len(dimensionen))
	}

	if len(dimensionen[0].Farben) != 2 {
		t.Errorf("die Farbtafel lautet %+v", dimensionen[0].Farben)
	}
}
