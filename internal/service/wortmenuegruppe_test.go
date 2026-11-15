package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Gruppe öffnet ein Untermenü** — im Pack gibt es 103 davon (gemessen
// am 2026-08-24 am bezogenen `M15-Magic-Pack`), etwa `Nichtstandard` mit den
// selteneren Typen darin.
//
// `DropDownWordList::submenu`, `src/gui/value/text.cpp`, Zeilen 231–247,
// Fassung f9d9356 vom 2021-07-06.
func TestEineGruppeOeffnetEinUntermenue(t *testing.T) {
	woerter := []msepaket.Wortlisteneintrag{
		{Name: "Creature"},
		{Name: "Nichtstandard", Woerter: []msepaket.Wortlisteneintrag{
			{Name: "Plane"},
			{Name: "Scheme"},
		}},
	}

	menue := service.NeuesWortmenue(woerter, "Scheme", ohneSkript(t))

	if !menue.Eintraege[1].IstGruppe {
		t.Fatalf("der zweite Eintrag ist keine Gruppe: %+v", menue.Eintraege[1])
	}

	unten := menue.Eintraege[1].Untermenue
	if unten == nil {
		t.Fatal("das Untermenü fehlt")
	}

	if len(unten.Eintraege) != 2 {
		t.Fatalf("%d Einträge im Untermenü", len(unten.Eintraege))
	}

	// **Es steht auf demselben Text** — `Scheme` ist darin gewählt, obwohl das
	// Menü darüber nichts davon weiß.
	if unten.Gewaehlt != 1 {
		t.Errorf("gewählt ist %d, erwartet der zweite Eintrag", unten.Gewaehlt)
	}

	neu, gibts := menue.WaehlenAmWeg([]int{1, 0})
	if !gibts || neu != "Plane" {
		t.Errorf("eine Wahl im Untermenü gibt %q (%v)", neu, gibts)
	}
}

// **Ein Weg, der ins Leere zeigt, ändert nichts.**
//
// Die Zahlen kommen von außen — aus einer Bindung, aus einem Menü, das
// inzwischen ein anderes ist. Ein Weg auf einen Eintrag ohne Untermenü, ein
// Weg über das Ende hinaus und der leere Weg geben alle nichts.
func TestEinWegInsLeereAendertNichts(t *testing.T) {
	woerter := []msepaket.Wortlisteneintrag{
		{Name: "Creature"},
		{Name: "Nichtstandard", Woerter: []msepaket.Wortlisteneintrag{{Name: "Plane"}}},
	}

	menue := service.NeuesWortmenue(woerter, "Creature", ohneSkript(t))

	wege := [][]int{
		nil,
		{-1},
		{2},
		{0, 0},
		{-1, 0},
		{2, 0},
		{1, 5},
	}

	for _, weg := range wege {
		if neu, gibts := menue.WaehlenAmWeg(weg); gibts {
			t.Errorf("der Weg %v gab %q", weg, neu)
		}
	}
}

// **Ein doppelter Name fällt heraus, eine gleichnamige Gruppe nicht** —
// `if (!already_added || w->isGroup())`, `src/gui/value/text.cpp`, Zeile 172,
// Fassung f9d9356 vom 2021-07-06.
//
// Im Pack steht dasselbe Wort in `type-zhs` **zweimal** in derselben Liste;
// zwei gleiche Zeilen wären für niemanden zu unterscheiden.
func TestEinDoppelterNameFaelltHeraus(t *testing.T) {
	woerter := []msepaket.Wortlisteneintrag{
		{Name: "Creature"},
		{Name: "Creature"},
		{Name: "Creature", Woerter: []msepaket.Wortlisteneintrag{{Name: "Elf"}}},
	}

	menue := service.NeuesWortmenue(woerter, "", ohneSkript(t))

	if len(menue.Eintraege) != 2 {
		t.Fatalf("%d Einträge: %+v", len(menue.Eintraege), menue.Eintraege)
	}

	if menue.Eintraege[0].IstGruppe || !menue.Eintraege[1].IstGruppe {
		t.Errorf("übrig sind %+v", menue.Eintraege)
	}
}

// TestEinGruppenkopfIstKeineWahl ist Befund 5 des Zweitblicks über Phase W.
//
//	// don't hide if there is a child menu
//	if (selected_item != NO_SELECTION && submenu(selected_item)) return;
//
// `src/gui/drop_down_list.cpp`, Zeilen 373–374, Fassung f9d9356 vom
// 2021-07-06. MSE **beendet die Wahl nicht** auf einem Eintrag mit Untermenü;
// er öffnet es.
//
// Hier wies nur das **Fenster** den Gruppenkopf ab — Kommandozeile, MCP und
// REST nahmen ihn an und schrieben den Gruppennamen wörtlich ins Kartenfeld.
// Eine Zusage, die eine von vier Flächen hält, ist keine (§16).
func TestEinGruppenkopfIstKeineWahl(t *testing.T) {
	t.Parallel()

	menue := service.NeuesWortmenue([]msepaket.Wortlisteneintrag{
		{Name: "Plane"},
		{Name: "Nichtstandard", Woerter: []msepaket.Wortlisteneintrag{{Name: "Schema"}}},
	}, "", nil)

	if _, gewaehlt := menue.WaehlenAmWeg([]int{1}); gewaehlt {
		t.Error("der Gruppenkopf ließ sich wählen — MSE öffnet dort ein Untermenü")
	}

	if _, gewaehlt := menue.WaehlenAmWeg([]int{1, 0}); !gewaehlt {
		t.Error("der Eintrag im Untermenü ließ sich nicht wählen")
	}

	if _, gewaehlt := menue.WaehlenAmWeg([]int{0}); !gewaehlt {
		t.Error("der gewöhnliche Eintrag ließ sich nicht wählen")
	}
}

// **Auch tief unten ist ein Gruppenkopf keine Wahl.**
//
// Der Produktcode konnte das schon: ein langer Weg steigt rekursiv ins
// Untermenü, und dort trifft der letzte Schritt dieselbe Prüfung. **Geprüft war
// es nicht** — angestrichen vom Nachlauf des Zweitblicks über Phase W.
//
// Der Unterschied ist nicht akademisch: `M15-Magic-Pack` verschachtelt bis in
// die vierte Ebene (`race-de`), und eine Zusage, die nur die erste hält, wäre
// an genau den Stellen wertlos, wo es Untermenüs überhaupt gibt.
func TestAuchTiefUntenIstEinGruppenkopfKeineWahl(t *testing.T) {
	t.Parallel()

	menue := service.NeuesWortmenue([]msepaket.Wortlisteneintrag{
		{Name: "Oben", Woerter: []msepaket.Wortlisteneintrag{
			{Name: "Mitte", Woerter: []msepaket.Wortlisteneintrag{
				{Name: "Unten"},
			}},
		}},
	}, "", nil)

	for _, weg := range [][]int{{0}, {0, 0}} {
		if _, gewaehlt := menue.WaehlenAmWeg(weg); gewaehlt {
			t.Errorf("der Gruppenkopf am Weg %v ließ sich wählen", weg)
		}
	}

	wert, gewaehlt := menue.WaehlenAmWeg([]int{0, 0, 0})
	if !gewaehlt {
		t.Fatal("der Blattknoten ließ sich nicht wählen")
	}

	if wert != "Unten" {
		t.Errorf("gewählt wurde %q", wert)
	}
}
