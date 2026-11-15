package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **`card list name:` und seine Übersetzungen kommen mit** —
// #699.
//
//	card field:
//		type: text
//		name: type
//		card list name: Type
//		localized card list name:
//			de: Typ
//			fr: Type
//
// `magic.mse-game/card_fields`. Über das Paket: 243 Kartenfelder mit
// `card list name:`, 25 mit `localized card list name:` — und **null** mit
// `caption:` oder `localized caption:`. Das ist das einzige Deutsch, das ein
// Kartenfeld dieses Packs mitbringt.
func TestEinKartenfeldTraegtSeinenListennamen(t *testing.T) {
	t.Parallel()

	pfad := probespielpaket(t, `mse version: 2.0.0
card field:
	type: text
	name: type
	card list name: Type
	localized card list name:
		de: Typ
		fr: Type
card field:
	type: text
	name: rule text
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

	spiel, _ := paket.Spielmodell()
	if len(spiel.Fields) != 2 {
		t.Fatalf("%d Felder", len(spiel.Fields))
	}

	if spiel.Fields[0].Listenname != "Type" {
		t.Errorf("Listenname %q", spiel.Fields[0].Listenname)
	}

	if spiel.Fields[0].Listenuebersetzungen["de"] != "Typ" {
		t.Errorf("Übersetzungen %v", spiel.Fields[0].Listenuebersetzungen)
	}

	if spiel.Fields[0].Beschriftet("de") != "Typ" {
		t.Errorf("Aufschrift %q", spiel.Fields[0].Beschriftet("de"))
	}

	// **Ohne Angabe bleibt es bei MSEs Aufschrift aus dem Namen.**
	if spiel.Fields[1].Beschriftet("de") != "Rule text" {
		t.Errorf("Aufschrift %q", spiel.Fields[1].Beschriftet("de"))
	}
}
