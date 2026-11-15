package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **`editable: false` kommt mit** —
// #708.
//
//	card field:
//		type: text
//		name: cmc
//		editable: false
//		save value: false
//
// `magic.mse-game/card_fields`. **24** der 249 Kartenfelder tragen es.
func TestEinKartenfeldSagtObEsBearbeitetWird(t *testing.T) {
	t.Parallel()

	pfad := probespielpaket(t, `mse version: 2.0.0
card field:
	type: text
	name: cmc
	editable: false
card field:
	type: text
	name: name
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

	if spiel.Fields[0].Editierbar {
		t.Error("`cmc` ist editierbar, obwohl `editable: false` dasteht")
	}

	// **Ohne Angabe wahr** — `Field::Field` setzt `editable(true)`
	// (`src/data/field.cpp`, Fassung f9d9356 vom 2021-07-06).
	if !spiel.Fields[1].Editierbar {
		t.Error("`name` ist nicht editierbar, obwohl nichts dasteht")
	}
}
