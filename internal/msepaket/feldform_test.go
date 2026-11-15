package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein Kartenfeld sagt, welche Form es hat** —
// #560.
//
// # Wofür
//
// Der Karteneditor stellte jedes Feld als einzeiliges Textfeld dar: den
// Regeltext, die Seltenheit, die Manakosten, die Kennung. Was ein Feld ist,
// steht im Spielpaket und wurde nur zur Hälfte gelesen.
//
//	card field:
//		type: text
//		name: flavor text
//		description: The flavor text of the card.
//		multi line: true
//
// `magic.mse-game/card_fields`, Zeilen 1289–1295. Über das Paket tragen **68**
// Kartenfelder `multi line: true`.
func TestEinKartenfeldNenntSeineForm(t *testing.T) {
	t.Parallel()

	pfad := probespielpaket(t, `mse version: 2.0.0
card field:
	type: text
	name: regeltext
	description: Was die Karte tut.
	multi line: true
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

	if !spiel.Fields[0].Mehrzeilig {
		t.Error("regeltext ist nicht mehrzeilig")
	}

	if spiel.Fields[0].Beschreibung != "Was die Karte tut." {
		t.Errorf("Beschreibung %q", spiel.Fields[0].Beschreibung)
	}

	// **Ohne Angabe einzeilig** — MSEs Vorgabe:
	//
	//	TextField::TextField()
	//	  : multi_line(false) …
	//
	// `src/data/field/text.cpp`, Zeilen 16–24, Fassung f9d9356 vom
	// 2021-07-06.
	if spiel.Fields[1].Mehrzeilig {
		t.Error("name ist mehrzeilig, obwohl nichts dasteht")
	}
}
