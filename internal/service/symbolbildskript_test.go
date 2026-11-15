package service_test

// **Ein Symbolbild darf ein Skript sein**
// (#422).
//
// `SymbolInFont` hält sein Bild bei MSE2 als `ScriptableImage`
// (`src/data/symbol_font.cpp`, Zeile 93, Fassung f9d9356 vom 2021-07-06) —
// dieselbe Klasse, die auch ein Feldbild trägt. Über `Full-Magic-Pack` sind
// **2.435 von 8.110** `image:`-Zeilen in Symbolschriften Skripte; gezählt an
// der Zeile, nicht am gezeichneten Symbol (§10).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinSymbolbildDarfEinSkriptSein(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "casting cost", Links: "300", Oben: "29", Breite: "40", Hoehe: "23",
			Symbolschrift: msepaket.Symbolangabe{Name: "probe-symbole", Groesse: "15"},
			ImmerSymbole:  true,
		}},
	}

	vorlage := service.Vorlage{Stil: stil, Symbolschriften: map[string]msepaket.Symbolschrift{
		"probe-symbole": {Name: "probe-symbole", Symbole: []msepaket.Symbol{
			{Code: "W", Bild: `{ recolor_image(color: rgb(1,2,3), "w.png") }`},
			{Code: "U", Bild: "u.png"},
		}},
	}}

	plan, _, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "casting cost", Text: "WU"},
		}}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	symbole := symbolzugAus(t, plan)
	if len(symbole) != 2 {
		t.Fatalf("%d Symbole, erwartet 2", len(symbole))
	}

	// Das gerechnete Bild ist ein **Baum**, kein Pfad; sein Blatt ist die
	// Datei, die das Skript einfärbt.
	if symbole[0].Bild.Knotenart != "recolor_image" {
		t.Errorf("Knotenart %q, erwartet einen gerechneten Knoten (%s)",
			symbole[0].Bild.Knotenart, symbole[0].Bild.Beschreibung())
	}

	if blaetter := symbole[0].Bild.Blaetter(); len(blaetter) != 1 ||
		blaetter[0] != "/probe-symbole.mse-symbol-font/w.png" {
		t.Errorf("Blätter %v", blaetter)
	}

	// **Und das rohe daneben bleibt ein Blatt.**
	if !symbole[1].Bild.IstBlatt() ||
		symbole[1].Bild.Pfad != "/probe-symbole.mse-symbol-font/u.png" {
		t.Errorf("das rohe Bild ist %q", symbole[1].Bild.Beschreibung())
	}
}

// symbolzugAus holt die Symbole der einen Symbolebene eines Plans.
func symbolzugAus(t *testing.T, plan *render.Plan) []render.Symbol {
	t.Helper()

	for i := range plan.Ebenen {
		for _, abschnitt := range plan.Ebenen[i].Abschnitte {
			if len(abschnitt.Symbole) > 0 {
				return abschnitt.Symbole
			}
		}
	}

	t.Fatal("der Plan führt keinen Symbolzug")

	return nil
}

// **Eine Farbe darf als Zeichenkette ankommen.**
//
// `ScriptValue::toColor` (`src/script/value.cpp`, Zeilen 354–360, Fassung
// f9d9356 vom 2021-07-06) schickt jede Zeichenkette durch `parse_color`
// (`src/gfx/color.cpp`, Zeile 42): `rgb(…)`, `#rrggbb`, `transparent` und
// jeden Namen aus wx' Farbdatenbank.
//
// `planechase.mse-game` schreibt `symbol_font_small_color := { "black" }`, und
// damit kommt bei `recolor_image(color: …)` eine **Zeichenkette** an. Bis
// #422 fiel
// sie unter die Texte, und der Rasterer meldete
// „recolor_image braucht ein Argument `color:`".
func TestEineFarbeDarfEineZeichenketteSein(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "casting cost", Links: "300", Oben: "29", Breite: "40", Hoehe: "23",
			Symbolschrift: msepaket.Symbolangabe{Name: "probe-symbole", Groesse: "15"},
			ImmerSymbole:  true,
		}},
	}

	vorlage := service.Vorlage{Stil: stil, Symbolschriften: map[string]msepaket.Symbolschrift{
		"probe-symbole": {Name: "probe-symbole", Symbole: []msepaket.Symbol{
			{Code: "W", Bild: `{ recolor_image(color: "black", "w.png") }`},
		}},
	}}

	plan, _, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "casting cost", Text: "W"},
		}}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	symbole := symbolzugAus(t, plan)

	if farbe, da := symbole[0].Bild.Farben["color"]; !da || farbe != (render.Farbe{A: 255}) {
		t.Errorf("die Farbe `color` ist %v (da: %t), erwartet Schwarz", farbe, da)
	}
}
