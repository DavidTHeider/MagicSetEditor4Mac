package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **`<line>` schneidet einen Block, `\n` einen Absatz.**
//
// MSE zählt `<line>` wie `<i>` und `<b>` und gibt dem Element den Umbruch
// `LineBreak::LINE` (`src/render/text/element.cpp`, Zeilen 84–85 und 235–236,
// Fassung f9d9356 vom 2021-07-06). Im Satz endet daran ein **Block**:
// `if (l.break_after == LineBreak::LINE) { paragraph = block = nullptr; }`
// (`src/render/text/viewer.cpp`, Zeilen 380–384).
//
// **Wo das herkommt:** das Feld `text` von `magic.mse-game` wird aus zwei
// Feldern zusammengesetzt —
//
//	combined_editor(field1: card.rule_text, separator: "<line>\n</line>", field2: card.flavor_text)
//
// Regeltext, Blockgrenze, Flavortext. Genau darauf rechnen die
// Planeswalker-Stile mit `layout.blocks[i]`
// (#338).
func TestEineLinieSchneidetEinenBlock(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{
			{
				Feld: "text", Links: "0", Oben: "0", Breite: "100", Hoehe: "154",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
			{
				Feld: "melder", Links: `{ length(card_style.text.layout.blocks) }`,
				Oben: "0", Breite: "50", Hoehe: "12",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
			{
				Feld: "trenner", Links: `{ length(card_style.text.layout.separators) }`,
				Oben: "0", Breite: "50", Hoehe: "13",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
		},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "text", Text: "aaaa<line>\n</line>bbbb"},
		{Field: "melder", Text: "x"},
		{Field: "trenner", Text: "x"},
	}}

	plan, befunde, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), karte, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, befund := range befunde {
		t.Logf("Befund %s: %s", befund.Ort, befund.Meldung)
	}

	gemessen := map[float64]float64{}

	for stelle := range plan.Ebenen {
		ebene := &plan.Ebenen[stelle]
		if ebene.Art == render.ArtText && ebene.Rahmen.Hoehe < 20 {
			gemessen[ebene.Rahmen.Hoehe] = ebene.Rahmen.Links
		}
	}

	if gemessen[12] != 2 {
		t.Errorf("%g Blöcke, erwartet zwei", gemessen[12])
	}

	// **Und zwischen zwei Blöcken liegt ein Trenner.**
	if gemessen[13] != 1 {
		t.Errorf("%g Trenner, erwartet einer", gemessen[13])
	}
}

// **`<line>` verschwindet aus dem gezeichneten Text.**
//
// Es ist eine Auszeichnung, kein Zeichen — wie `<i>` und `<sym>`.
func TestEineLinieStehtNichtImText(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "0", Oben: "0", Breite: "200", Hoehe: "154",
			Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
		}},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "text", Text: "aaaa<line>\n</line>bbbb"},
	}}

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), karte, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for stelle := range plan.Ebenen {
		if plan.Ebenen[stelle].Art != render.ArtText {
			continue
		}

		if text := plan.Ebenen[stelle].Klartext(); text != "aaaa\nbbbb" {
			t.Errorf("der Text lautet %q, erwartet %q", text, "aaaa\nbbbb")
		}
	}
}
