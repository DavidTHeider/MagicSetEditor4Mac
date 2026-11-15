package service_test

import (
	"errors"
	"image"
	"image/color"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// errKeinBild meldet ein Bild, das die Quelle nicht führt.
var errKeinBild = errors.New("kein Bild")

// bildmasse ist eine Bildquelle mit **genannten** Bildgrößen.
type bildmasse struct{ da map[string]image.Point }

func (b bildmasse) Bild(pfad string) (image.Image, error) {
	mass, gibtEs := b.da[pfad]
	if !gibtEs {
		return nil, errKeinBild
	}

	bild := image.NewRGBA(image.Rect(0, 0, mass.X, mass.Y))
	bild.SetRGBA(0, 0, color.RGBA{A: 255})

	return bild, nil
}

// **`content_width` eines Symbolfelds misst die Bilder, nicht Quadrate.**
//
// # Der Befund
//
// Seit [0022](../../docs/entscheidungen/0022-der-rasterer-misst-die-symbole.md)
// zeichnet der Rasterer ein Symbol in seiner **tatsächlichen** Größe. Die erste
// Stufe des Satzes rechnete weiter mit einem Quadrat — und das ist nicht
// folgenlos: `magic-m15` setzt die rechte Kante des Kartennamens auf
//
//	right: { 341 - card_style.casting_cost.content_width }
//
// Ein zu schmal gemessener Manakostenzug macht den Namenskasten zu breit, und
// zwar auf **jeder** Karte
// (#291).
//
// # Warum der Planbauer dafür Bilder bekommt
//
// Aus demselben Grund, aus dem er Schriften bekommt: `content_width` ist eine
// Messung, und messen kann nur, wer die Maße hat. Es ist eine **Schnittstelle**
// und kein Paketleser — verdrahtet wird im Composition Root (§2).
func TestDieInhaltsbreiteEinesSymbolfeldsMisstDieBilder(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld:          "casting cost",
			Links:         "0",
			Oben:          "0",
			Breite:        "{ card_style.casting_cost.content_width }",
			Hoehe:         "23",
			Symbolschrift: msepaket.Symbolangabe{Name: "probe-symbole", Groesse: "15"},
			ImmerSymbole:  true,
		}},
	}

	vorlage := service.Vorlage{Stil: stil, Symbolschriften: map[string]msepaket.Symbolschrift{
		"probe-symbole": {
			Name: "probe-symbole", Bildgroesse: 30, AbstandWaagerecht: 2,
			Symbole: []msepaket.Symbol{{Code: "W", Bild: "w.png"}},
		},
	}}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "casting cost", Text: "WW"},
	}}

	// Das Bild misst 60 × 30 bei `image font size: 30` und Größe 15: also
	// 30 × 15, plus 2 Abstand ein Kasten von 32. Zwei Symbole sind 64 breit —
	// als Quadrate wären es 30.
	quelle := bildmasse{da: map[string]image.Point{
		"/probe-symbole.mse-symbol-font/w.png": {X: 60, Y: 30},
	}}

	plan, _, err := service.NeuerPlanbauer().MitBildern(quelle).
		Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtSymbolzug {
			continue
		}

		if ebene.Rahmen.Breite != 64 {
			t.Errorf("das Feld ist %g breit, erwartet 64 — gemessen wurde mit Quadraten",
				ebene.Rahmen.Breite)
		}

		return
	}

	t.Fatal("der Plan führt keinen Symbolzug")
}
