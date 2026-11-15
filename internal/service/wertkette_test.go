package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Wertskript sieht, was die Wertskripte davor gerechnet haben.**
//
// # Woran es aufgefallen ist
//
// Der Flavortext einer aus MTG JSON übernommenen Karte stand aufrecht statt
// kursiv. `magic.mse-game` klammert ihn in seinem Wertskript ein:
//
//	flavor_text_filter :=
//		remove_tag@(tag: "<i-flavor>") +
//		{ "<i-flavor>" + input + "</i-flavor>" } + …
//
// `magic.mse-game/script`, Zeile 1985. Das Feld `text` setzt danach zusammen:
//
//	combined_editor(field1: card.rule_text, separator: "<line>\n</line>",
//	                field2: card.flavor_text)
//
// `magic.mse-game/card_fields`, Zeile 1354. Es liest `card.flavor_text` — und
// las den **ungefilterten** Wert, weil alle Wertskripte gegen dieselbe
// `card`-Abbildung rechneten, die vor der Schleife gesetzt worden war.
//
// # Warum der Kanonlauf das nie gesehen hat
//
// In einer `.mse-set`-Datei steht die Auszeichnung **schon drin**: MSE sichert
// den gerechneten Wert. `Red.mse-set` enthält `<i-flavor>` 4.508-mal. Nur wer
// ein Set aus MTG JSON übernimmt, ist auf das Wertskript angewiesen.
//
// # Was MSE tut
//
// `Value::update` läuft je Wert, und der `Context` trägt die schon
// gerechneten: `SetScriptManager::updateAll` geht die Felder in der
// Reihenfolge des Spiels durch und **ordnet nicht nach Abhängigkeiten**
// (`src/script/script_manager.cpp`, Zeilen 265–300, Fassung f9d9356 vom
// 2021-07-06) — hier stand bis #2315 das Gegenteil. Hier wird `card` nach
// **jedem** Feld neu gesetzt; dass die Reihenfolge die des Spiels ist, hält
// der Test darunter.
func TestEinWertskriptSiehtDieVorigen(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "zusammen", Links: "0", Oben: "0", Breite: "300", Hoehe: "100",
			Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
		}},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "roh", Text: "Funke"},
		{Field: "zusammen", Text: ""},
	}}

	plan, befunde, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), karte, service.Vorlage{
			Stil: stil,
			Kartenfelder: []domain.Field{
				// **`roh` steht zuerst** — so wie `flavor text` vor `text`.
				{Name: "roh", Kind: domain.KindText, Skript: `"[" + value + "]"`},
				{Name: "zusammen", Kind: domain.KindText, Skript: `card.roh`},

				// **Ein Farbfeld gehört dazu, und das ist der Kern.**
				//
				// Ohne es gibt [kartenabbildung] die Einträge **unverändert**
				// weiter — dieselbe Scheibe, die die Schleife beschreibt —, und
				// das zweite Wertskript sähe die Änderung durch bloßes
				// Aliasing. Der erste Anlauf dieses Tests war deshalb grün und
				// belegte nichts. Jedes gesichtete Spielpaket hat Farbfelder.
				{Name: "farbe", Kind: domain.KindColor},
			},
		})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, befund := range befunde {
		t.Errorf("Befund: %s — %s", befund.Ort, befund.Meldung)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtText {
			continue
		}

		if ebene.Klartext() != "[Funke]" {
			t.Errorf("gezeichnet steht %q, erwartet %q — das zweite Wertskript "+
				"hat den ungerechneten Wert gelesen", ebene.Klartext(), "[Funke]")
		}

		return
	}

	t.Fatal("der Plan führt keine Textebene")
}

// **Die Wertskripte laufen in der Reihenfolge, in der das Spiel die Felder
// erklärt — nicht in der, in der die Karte sie speichert** (#2396).
//
// # Was MSE tut, zweimal
//
// Beim **Laden** rechnet `SetScriptManager::updateAll` je Karte
// `FOR_EACH(v, card->data)` in der Reihenfolge der Spielfelder und zieht
// nichts nach (`src/script/script_manager.cpp`, Zeilen 265–300; `card->data`
// legt `data.init(game.card_fields)` an, `src/data/card.cpp`, Zeile 37): ein
// Feld, das ein **späteres** liest, sieht dessen gespeicherten Wert. Beim
// **Bearbeiten** zieht `updateValue` die Abhängigen nach (ebd. Zeilen
// 249–262) — und gespeichert wird der gerechnete Wert. Fassung f9d9356 vom
// 2021-07-06.
//
// # Warum hier das Erste
//
// **Seit #2396 speichert auch diese Anwendung den gerechneten Wert**
// (`Feldweg`, `wertnachzug.go`); gezeichnet wird deshalb, was MSE2 beim Laden
// zeigt. Bis dahin rechnete sie wie nach einer Bearbeitung, weil sie den
// eingegebenen Wert speicherte — und dieser Test verlangte `"[Funke]"`.
func TestDieWertskripteFolgenDerReihenfolgeDesSpiels(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "zusammen", Links: "0", Oben: "0", Breite: "300", Hoehe: "100",
			Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
		}},
	}

	// **`zusammen` ist nicht gespeichert** — es wird nachgetragen, so wie
	// `text` mit `save value: false`.
	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "spaeter", Text: "Funke"},
	}}

	plan, befunde, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), karte, service.Vorlage{
			Stil: stil,
			Kartenfelder: []domain.Field{
				{Name: "zusammen", Kind: domain.KindText, Skript: `card.spaeter`},
				{Name: "spaeter", Kind: domain.KindText, Skript: `"[" + value + "]"`},
				// Ein Farbfeld, damit `card` eine Kopie ist — siehe oben.
				{Name: "farbe", Kind: domain.KindColor},
			},
		})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, befund := range befunde {
		t.Errorf("Befund: %s — %s", befund.Ort, befund.Meldung)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtText {
			continue
		}

		if ebene.Klartext() != "Funke" {
			t.Errorf("gezeichnet steht %q, erwartet %q — das spätere Feld ist "+
				"vor dem früheren gerechnet worden", ebene.Klartext(), "Funke")
		}

		return
	}

	t.Fatal("der Plan führt keine Textebene")
}
