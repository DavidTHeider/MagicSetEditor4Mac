package service_test

// **Ob ein Feld umbricht, sagt das Spiel** — J05 (#2269), zu #2299.
//
// `multi line:` steht am Kartenfeld von `magic.mse-game`, nicht am Feldstil,
// und MSE liest es beim Setzen (`style.field().multi_line`,
// `src/render/text/viewer.cpp`, Zeile 618, Fassung f9d9356 vom 2021-07-06).
// Ohne Angabe ist ein Textfeld einzeilig (`TextField::TextField`,
// `src/data/field/text.cpp`, Zeilen 16–24).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieMehrzeiligkeitDesSpielfeldsStehtImSatz(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{
			{Feld: "name", Links: "10", Oben: "10", Breite: "300", Hoehe: "40", Schrift: msepaket.Schrift{Name: "Matrix", Groesse: "12"}},
			{Feld: "rule text", Links: "10", Oben: "60", Breite: "300", Hoehe: "200", Schrift: msepaket.Schrift{Name: "Matrix", Groesse: "12"}},
			{Feld: "unbekannt", Links: "10", Oben: "300", Breite: "300", Hoehe: "40", Schrift: msepaket.Schrift{Name: "Matrix", Groesse: "12"}},
		},
	}
	vorlage := service.Vorlage{
		Stil: stil,
		Kartenfelder: []domain.Field{
			{Name: "name", Kind: domain.KindText},
			{Name: "rule text", Kind: domain.KindText, Mehrzeilig: true},
		},
	}
	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "name", Text: "Name"},
		{Field: "rule text", Text: "Regel"},
		{Field: "unbekannt", Text: "Frei"},
	}}

	plan, _, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	einzeilig := map[string]bool{}
	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtText {
			einzeilig[ebene.Abschnitte[0].Text] = ebene.Satz.Einzeilig
		}
	}

	for text, erwartet := range map[string]bool{
		"Name": true, "Regel": false,
		// **Ohne Spielfeld bleibt es beim Umbruch** — so stand es vorher,
		// und ein Feld ohne Definition gibt es in einem echten Paket nicht.
		"Frei": false,
	} {
		if gab, da := einzeilig[text]; !da || gab != erwartet {
			t.Errorf("%s: Einzeilig %v (da %v), erwartet %v", text, gab, da, erwartet)
		}
	}
}
