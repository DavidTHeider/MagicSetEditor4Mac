// Die zehn Auswahlarten, jede mit einem Lauf
// (#1072).
//
// **Von `packzieher_test.go` getrennt**, weil der Schnitt einer Frage folgt:
// dort steht, **dass** ein Booster zieht, hier, **wie** jede Auswahlart es
// tut.
package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Jede der zehn Auswahlarten zieht, wie MSE es beschreibt.**
//
// `PackSelectType`, `src/data/pack.hpp`, Zeilen 27–39, Fassung f9d9356 vom
// 2021-07-06. Die Namen stehen in `src/data/pack.cpp`, Zeilen 21–30 —
// **`equal nonempty` heißt dort `SELECT_NONEMPTY`**, ein Tippfehler des
// Originals, den dieses Programm nicht nachbaut: es liest den Namen und
// behandelt ihn als eigene Art.
func TestJedeAuswahlartZieht(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name     string
		auswahl  string
		posten   []msepaket.Packposten
		karten   int
		erwartet int
		grund    string
	}{
		{
			name: "all nimmt jede Karte", auswahl: "all",
			karten: 5, erwartet: 5,
			grund: "generate_all, Zeilen 316–328",
		},
		{
			name: "replace zieht so oft, wie verlangt ist", auswahl: "replace",
			karten: 5, erwartet: 1,
			grund: "eine Kopie verlangt, eine gezogen",
		},
		{
			name: "no replace mischt und schneidet ab", auswahl: "no replace",
			karten: 5, erwartet: 1,
			grund: "Zeilen 230–246",
		},
		{
			name: "proportional zieht eines", auswahl: "proportional",
			karten: 5, erwartet: 1,
		},
		{
			name: "nonempty zieht eines", auswahl: "nonempty",
			karten: 5, erwartet: 1,
		},
		{
			name: "first nimmt die erste Karte", auswahl: "first",
			karten: 5, erwartet: 1,
			grund: "Zeilen 297–311",
		},
		{
			// **Bei einer einzigen Kopie würfelt auch `equal`** — „somewhat of
			// a hack to keep things fair" (Zeile 252).
			name: "equal würfelt bei einer Kopie", auswahl: "equal",
			karten: 5, erwartet: 1,
		},
		{
			name: "equal proportional ebenso", auswahl: "equal proportional",
			karten: 5, erwartet: 1,
		},
		{
			name: "equal nonempty ebenso", auswahl: "equal nonempty",
			karten: 5, erwartet: 1,
		},
		{
			// **`auto` ohne Filter und ohne Posten zieht nichts.** MSE hat
			// dort keinen Zweig; ein erfundener wäre eine Behauptung.
			name: "auto ohne alles zieht nichts", auswahl: "auto",
			karten: 5, erwartet: 0,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			zieher := service.NeuerPackzieher(
				[]msepaket.Pack{{Name: "eines", Auswahl: fall.auswahl, Posten: fall.posten}},
				map[string][]*domain.CardRow{"eines": packkarten("k", fall.karten)})

			gezogen, fehler := zieher.Ziehen("eines", 3)
			if fehler != nil {
				t.Fatalf("Ziehen: %v", fehler)
			}

			if len(gezogen) != fall.erwartet {
				t.Errorf("%d Karten, erwartet %d (%s)", len(gezogen), fall.erwartet, fall.grund)
			}
		})
	}
}

// **`equal` verteilt, sobald mehr als eine Kopie verlangt ist.**
//
// Das ist der Unterschied zwischen „ein Booster hat meistens eine Rare" und
// „ein Booster hat eine Rare": `weighted_equal_divide` gibt jedem Posten so
// viele, dass `count/weight` möglichst gleich bleibt (Zeilen 192–213).
func TestEqualVerteiltNachGewicht(t *testing.T) {
	t.Parallel()

	zieher := service.NeuerPackzieher([]msepaket.Pack{
		{Name: "wurzel", Auswahl: "all", Posten: []msepaket.Packposten{
			{Name: "gleich", Anzahl: 8, Gewicht: 1},
		}},
		{Name: "gleich", Auswahl: "equal", Posten: []msepaket.Packposten{
			{Name: "haeufig", Anzahl: 1, Gewicht: 3},
			{Name: "selten", Anzahl: 1, Gewicht: 1},
		}},
		{Name: "haeufig", Auswahl: "no replace"},
		{Name: "selten", Auswahl: "no replace"},
	}, map[string][]*domain.CardRow{
		"haeufig": packkarten("h", 20),
		"selten":  packkarten("s", 20),
	})

	gezogen, fehler := zieher.Ziehen("wurzel", 5)
	if fehler != nil {
		t.Fatalf("Ziehen: %v", fehler)
	}

	haeufige, seltene := zaehleNachVorsilbe(gezogen)

	// **Drei zu eins bei acht Zügen: sechs und zwei.** Nicht „ungefähr" —
	// die Verteilung ist gerechnet, nicht gewürfelt.
	if haeufige != 6 || seltene != 2 {
		t.Errorf("%d häufige und %d seltene, erwartet 6 und 2", haeufige, seltene)
	}
}

// **`first` nimmt den ersten Posten, der überhaupt etwas hat.**
func TestFirstNimmtDenErstenNichtleeren(t *testing.T) {
	t.Parallel()

	zieher := service.NeuerPackzieher([]msepaket.Pack{
		{Name: "wurzel", Auswahl: "first", Posten: []msepaket.Packposten{
			{Name: "leer", Anzahl: 1, Gewicht: 1},
			{Name: "voll", Anzahl: 1, Gewicht: 1},
		}},
		{Name: "leer", Auswahl: "no replace"},
		{Name: "voll", Auswahl: "no replace"},
	}, map[string][]*domain.CardRow{"voll": packkarten("v", 3)})

	gezogen, fehler := zieher.Ziehen("wurzel", 1)
	if fehler != nil {
		t.Fatalf("Ziehen: %v", fehler)
	}

	if len(gezogen) != 1 || namenVon(gezogen[0])[:1] != "v" {
		t.Errorf("gezogen %v", namenreihe(gezogen))
	}
}

// **Ein leerer Posten wird übersprungen, nicht gezogen.**
//
// `nonempty` zählt ihn gar nicht erst mit (`total_weight`, Zeilen 129–132).
func TestNonemptyUeberspringtLeerePosten(t *testing.T) {
	t.Parallel()

	zieher := service.NeuerPackzieher([]msepaket.Pack{
		{Name: "wurzel", Auswahl: "nonempty", Posten: []msepaket.Packposten{
			{Name: "leer", Anzahl: 1, Gewicht: 1000},
			{Name: "voll", Anzahl: 1, Gewicht: 1},
		}},
		{Name: "leer", Auswahl: "no replace"},
		{Name: "voll", Auswahl: "no replace"},
	}, map[string][]*domain.CardRow{"voll": packkarten("v", 3)})

	for same := range uint64(5) {
		gezogen, fehler := zieher.Ziehen("wurzel", same)
		if fehler != nil {
			t.Fatalf("Ziehen: %v", fehler)
		}

		if len(gezogen) != 1 {
			t.Fatalf("Same %d: %d Karten", same, len(gezogen))
		}
	}
}

// zaehleNachVorsilbe zählt, wie viele Karten mit `h` und mit `s` beginnen.
func zaehleNachVorsilbe(gezogen []*domain.CardRow) (haeufige, seltene int) {
	for _, karte := range gezogen {
		switch namenVon(karte)[:1] {
		case "h":
			haeufige++
		case "s":
			seltene++
		}
	}

	return haeufige, seltene
}
