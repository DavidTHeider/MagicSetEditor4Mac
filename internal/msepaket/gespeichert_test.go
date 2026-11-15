package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **`save value: false` heißt: dieser Wert gehört nicht in die Datei** —
// #696.
//
// # Wofür
//
//	card field:
//		type: text
//		name: text
//		save value: false
//		script: combined_editor(field1: card.rule_text, …)
//
// `magic.mse-game/card_fields`. Das Feld `text` ist eine **Sicht** auf
// Regeltext und Flavortext; wer es speicherte, legte einen Wert ab, den MSE
// nie wieder liest — und der beim nächsten Öffnen still verschwindet.
func TestEinFeldSagtObEsGespeichertWird(t *testing.T) {
	t.Parallel()

	pfad := probespielpaket(t, `mse version: 2.0.0
card field:
	type: text
	name: text
	save value: false
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

	if spiel.Fields[0].Gespeichert {
		t.Error("`text` wird gespeichert, obwohl `save value: false` dasteht")
	}

	// **Ohne Angabe wahr** — `Field::Field` setzt `save_value(true)`
	// (`src/data/field.cpp`, Fassung f9d9356 vom 2021-07-06).
	if !spiel.Fields[1].Gespeichert {
		t.Error("`rule text` wird nicht gespeichert, obwohl nichts dasteht")
	}
}
