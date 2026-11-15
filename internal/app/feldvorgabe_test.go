package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Auswahlfeld nennt seinen Anfangswert**
// (#562).
//
// # Wofür
//
//	ChoiceValue::ChoiceValue(const ChoiceFieldP& field, bool initial_first_choice)
//	  : value( !field->initial.empty() ? field->initial
//	         : initial_first_choice    ? field->choices->choiceName(0)
//	                                   : String() )
//
// `src/data/field/choice.cpp`, Zeilen 279–283, Fassung f9d9356 vom
// 2021-07-06. Eine neue Karte trägt in MSE nicht nichts, sondern den
// Anfangswert jedes Feldes (`Card::Card`, `src/data/card.cpp`, Zeile 37).
//
// # Was ohne ihn dastand
//
// Ein `select` ohne Wert zeigt seine **erste** Option. `rarity` hat sieben
// Wahlen, beginnend mit `basic land`, und `initial: common` — der Editor
// behauptete damit an jeder neuen Karte eine Seltenheit, die weder in der
// Karte noch auf ihr steht. Gezeichnet wird `common`; die beiden liefen
// auseinander.
func TestEinAuswahlfeldNenntSeinenAnfangswert(t *testing.T) {
	waehler := &stilwaehler{felder: []domain.Field{
		{
			Name: "rarity", Kind: domain.KindChoice, Initial: "common",
			Choices: []string{"basic land", "common", "rare"},
		},
		{
			Name: "card symbol", Kind: domain.KindChoice,
			Choices: []string{"none", "tap"},
		},
		{Name: "name", Kind: domain.KindText},
	}}

	anwendung := app.NeueApp("")
	anwendung.MitStilwahl(waehler)

	antwort := anwendung.Kartenfelder(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	faelle := []struct {
		name    string
		vorgabe string
		warum   string
	}{
		{"rarity", "common", "`initial:` schlägt die erste Wahl"},
		{"card symbol", "none", "ohne `initial:` gilt die erste Wahl"},
		{"name", "", "ein Textfeld hat keinen Anfangswert"},
	}

	for stelle, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if gab := antwort.Felder[stelle].Vorgabe; gab != fall.vorgabe {
				t.Errorf("%q statt %q — %s", gab, fall.vorgabe, fall.warum)
			}
		})
	}
}

// **Die Aufschrift kommt in der gemeldeten Sprache an**
// (#560).
//
// # Wofür
//
// Der Editor beschriftete jedes Feld mit seinem rohen Namen — `casting cost`,
// `card symbol`. MSE nimmt `caption:`, und wo das fehlt, den großgeschriebenen
// Namen (`src/data/field.cpp`, Zeile 69, Fassung f9d9356 vom 2021-07-06).
//
// # Warum hier und nicht im Frontend
//
// `localized caption:` ist ein Sprachkürzel aus dem **Paket** — `ch-s`, `jp`,
// `pt-br`. Wer die Auswahl dem Frontend überließe, schickte jedem Feld seine
// zwölf Übersetzungen mit, damit elf davon verworfen werden. Die Sprache liegt
// in `App`, weil Menüleiste und Dateidialog sie ohnehin brauchen.
func TestDieAufschriftKommtInDerGemeldetenSpracheAn(t *testing.T) {
	waehler := &stilwaehler{felder: []domain.Field{
		{Name: "casting cost", Kind: domain.KindText},
		{Name: "pt", Kind: domain.KindText, Beschriftung: "P/T"},
		{
			Name: "title", Kind: domain.KindText, Beschriftung: "Title",
			Uebersetzungen: map[string]string{"de": "Titel"},
		},
	}}

	anwendung := app.NeueApp("")
	anwendung.MitStilwahl(waehler)

	if antwort := anwendung.SpracheSetzen("de"); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	antwort := anwendung.Kartenfelder(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	erwartet := []string{"Casting cost", "P/T", "Titel"}
	for stelle, soll := range erwartet {
		if gab := antwort.Felder[stelle].Aufschrift; gab != soll {
			t.Errorf("%q statt %q", gab, soll)
		}
	}

	// **Der Name bleibt daneben stehen.** Er ist, was Skripte ansprechen und
	// was `FeldSetzen` schreibt; eine Aufschrift ist keine Kennung.
	if antwort.Felder[0].Name != "casting cost" {
		t.Errorf("der Name wurde überschrieben: %q", antwort.Felder[0].Name)
	}
}
