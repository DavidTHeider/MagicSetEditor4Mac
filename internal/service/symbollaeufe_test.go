package service_test

// **Symbole und Text im selben Feld** — `always symbol:`
// (#508).
//
//	if (style.always_symbol && style.symbol_font.valid()) {
//	  // mixed symbols/text, autodetected by symbol font
//	  size_t text_pos = 0;
//	  size_t pos = 0;
//	  while (pos < end-start) {
//	    if (size_t n = style.symbol_font.font->recognizePrefix(content,pos)) {
//	      // at 'pos' there are n symbol font characters
//	      if (text_pos < pos) { … }
//
// `src/render/text/element.cpp`, Zeilen 244–256, Fassung f9d9356 vom
// 2021-07-06. **MSE mischt**: `recognizePrefix` nimmt die längste Folge
// erkannter Codes, was dazwischen steht, wird Text.
//
// # Woran es aufgefallen ist
//
// Der Lauf über alle 23.168 kanonischen Karten meldete viermal „die
// Symbolschrift … führt kein Symbol für …; `<sym>…</sym>` bleibt Text" — und
// machte aus **dem ganzen Zug** Text. Aus Manasymbolen wurden Buchstaben.

import (
	"image"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestSymboleUndTextImSelbenFeld(t *testing.T) {
	stil := msepaket.Stil{
		Paket: "probe.mse-style", Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "casting cost", Links: "0", Oben: "0",
			Breite: "300", Hoehe: "23",
			Symbolschrift: msepaket.Symbolangabe{Name: "probe-symbole", Groesse: "15"},
			ImmerSymbole:  true,
			Schrift:       msepaket.Schrift{Name: "Probe", Groesse: "12"},
		}},
	}

	vorlage := service.Vorlage{Stil: stil, Symbolschriften: map[string]msepaket.Symbolschrift{
		"probe-symbole": {
			Name: "probe-symbole", Bildgroesse: 30, AbstandWaagerecht: 2,
			Symbole: []msepaket.Symbol{{Code: "W", Bild: "w.png"}},
		},
	}}

	// `WW` sind Symbole, ` — x ` ist Text, `W` wieder ein Symbol.
	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "casting cost", Text: "WW — x W"},
	}}

	quelle := bildmasse{da: map[string]image.Point{
		"/probe-symbole.mse-symbol-font/w.png": {X: 60, Y: 30},
	}}

	plan, befunde, err := service.NeuerPlanbauer().MitBildern(quelle).
		Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde: %v — MSE mischt und meldet nichts", befunde)
	}

	var (
		symbole int
		texte   []string
	)

	for i := range plan.Ebenen {
		if plan.Ebenen[i].Art != render.ArtText {
			continue
		}

		for _, abschnitt := range plan.Ebenen[i].Abschnitte {
			if len(abschnitt.Symbole) > 0 {
				symbole += len(abschnitt.Symbole)

				continue
			}

			if abschnitt.Text != "" {
				texte = append(texte, abschnitt.Text)
			}
		}
	}

	// **Drei Symbole**, und der Rest steht als Text daneben.
	if symbole != 3 {
		t.Errorf("%d Symbole, erwartet 3", symbole)
	}

	if len(texte) != 1 || texte[0] != " — x " {
		t.Errorf("Textstücke %q, erwartet [\" — x \"]", texte)
	}
}

// **Die Ränder der Aufteilung.**
func TestDieRaenderDerSymbolaufteilung(t *testing.T) {
	stil := func() msepaket.Stil {
		return msepaket.Stil{
			Paket: "probe.mse-style", Breite: 375, Hoehe: 523,
			Felder: []msepaket.Feldstil{{
				Feld: "casting cost", Links: "0", Oben: "0",
				Breite: "300", Hoehe: "23",
				Symbolschrift: msepaket.Symbolangabe{Name: "probe-symbole", Groesse: "15"},
				ImmerSymbole:  true,
				Schrift:       msepaket.Schrift{Name: "Probe", Groesse: "12"},
			}},
		}
	}

	schriften := map[string]msepaket.Symbolschrift{
		"probe-symbole": {
			Name: "probe-symbole", Bildgroesse: 30, AbstandWaagerecht: 2,
			Symbole: []msepaket.Symbol{{Code: "W", Bild: "w.png"}},
		},
	}

	quelle := bildmasse{da: map[string]image.Point{
		"/probe-symbole.mse-symbol-font/w.png": {X: 60, Y: 30},
	}}

	faelle := []struct {
		name     string
		wert     string
		symbole  int
		abschnit int
	}{
		// **Gar kein Symbol** — ein Abschnitt, ganz Text.
		{"ohne jedes Symbol", "abc", 0, 1},

		// **Ein `<sym>`-Abschnitt bleibt, wie er ist**, und wird nicht ein
		// zweites Mal befragt.
		{"schon ausgezeichnet", "<sym>W</sym>x", 1, 2},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
				{Field: "casting cost", Text: fall.wert},
			}}

			plan, _, err := service.NeuerPlanbauer().MitBildern(quelle).Stilplan(
				t.Context(), karte,
				service.Vorlage{Stil: stil(), Symbolschriften: schriften})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			var symbole, abschnitte int

			for i := range plan.Ebenen {
				if plan.Ebenen[i].Art != render.ArtText {
					continue
				}

				for _, abschnitt := range plan.Ebenen[i].Abschnitte {
					abschnitte++
					symbole += len(abschnitt.Symbole)
				}
			}

			if symbole != fall.symbole {
				t.Errorf("%d Symbole, erwartet %d", symbole, fall.symbole)
			}

			if abschnitte != fall.abschnit {
				t.Errorf("%d Abschnitte, erwartet %d", abschnitte, fall.abschnit)
			}
		})
	}
}
