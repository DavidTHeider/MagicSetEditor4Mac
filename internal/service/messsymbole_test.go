package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Gemessen werden die Symbole, die gezeichnet werden** — gefunden an
// F02 (#2244), beim Kanonvergleich.
//
// MSE2 setzt `content_width` und `layout` mit demselben `TextViewer`, der
// zeichnet, Symbole eingeschlossen (`TextViewer::prepareLines`,
// `src/render/text/viewer.cpp`, Zeilen 392–426, Fassung f9d9356 vom
// 2021-07-06). Hier maß die Messung `{3}{U}{U}` als drei schmale
// Buchstaben, die Zeichnung als drei runde Symbole. Solange der kursive
// Erinnerungstext aufrecht und damit breiter gemessen wurde, glich sich das
// aus; seit er kursiv gemessen wird, fehlte „Inspired Idea" eine Zeile, und
// der Strich zum Flavortext rückte um sie nach oben.
func TestDieMessungSetztDieSymbole(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{
			{
				Feld: "text", Links: "10", Oben: "10", Hoehe: "40", Breite: "200",
				Schrift:       msepaket.Schrift{Name: "Probe", Groesse: "10"},
				Symbolschrift: msepaket.Symbolangabe{Name: "probe-symbole", Groesse: "15"},
			},
			{
				Feld: "type", Links: "10", Oben: "60", Hoehe: "23",
				Rechts:  "{ card_style.text.content_width }",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
		},
	}

	plan, _, err := service.NeuerPlanbauer().MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "text", Text: "ab<sym>W</sym>"}, {Field: "type", Text: "x"},
		}}, service.Vorlage{Stil: stil, Symbolschriften: map[string]msepaket.Symbolschrift{
			"probe-symbole": {Name: "probe-symbole", Symbole: []msepaket.Symbol{{Code: "W", Bild: "w.png"}}},
		}})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	// „ab" ist 2·5 breit, das Symbol ein Quadrat von 15: zusammen 25, ab 10
	// also 15. Als Buchstabe gemessen wären es 15, ab 10 also 5.
	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtText && ebene.Klartext() == "x" {
			if ebene.Rahmen.Breite != 15 {
				t.Errorf("die Typzeile ist %.1f breit, erwartet 15", ebene.Rahmen.Breite)
			}

			return
		}
	}

	t.Fatal("der Plan führt keine Typzeile")
}

// **Und gemessen wird verkleinert, wo verkleinert gezeichnet wird** — der
// zweite Teil desselben Befunds. `scale down to:` setzt einen zu langen Text
// kleiner (`TextViewer::prepareLinesTryScales`, `src/render/text/viewer.cpp`,
// Zeilen 447–500, Fassung f9d9356 vom 2021-07-06), und MSE2 misst den Satz,
// den es zeichnet. Hier maß die Messung in voller Größe: bei den
// Planeswalkern rückten die Treuekästen damit um bis zu sieben Punkte nach
// unten, seit die Symbole mitgemessen werden.
func TestDieMessungVerkleinertWieDieZeichnung(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{
			{
				Feld: "text", Links: "10", Oben: "10", Hoehe: "30", Breite: "100",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10", Kleinstgroesse: "4"},
			},
			{
				Feld: "type", Links: "10", Oben: "60", Hoehe: "5",
				Rechts:  "{ 15 + card_style.text.content_height }",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
		},
	}

	plan, _, err := service.NeuerPlanbauer().MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "text", Text: strings.Repeat("wort ", 60)}, {Field: "type", Text: "x"},
		}}, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtText && ebene.Klartext() == "x" {
			// Die Typzeile reicht bis 15 + content_height; verkleinert passt
			// der Text in die 30 Punkte des Kastens.
			if hoehe := ebene.Rahmen.Breite + ebene.Rahmen.Links - 15; hoehe > 30 {
				t.Errorf("gemessen %.1f hoch — der Kasten hat 30, gezeichnet wird verkleinert", hoehe)
			}

			return
		}
	}

	t.Fatal("der Plan führt keine Typzeile")
}
