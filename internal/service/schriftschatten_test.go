package service_test

// **Der Schatten der Schrift kommt in den Satz** — F02
// (#2244).
//
// Die Farbe ist ein Skript wie die Schriftfarbe, aber mit **anderer Vorgabe**:
// ein Feld ohne `color:` schreibt schwarz (`color(Color(0,0,0))` im
// Konstruktor), ein Feld ohne `shadow color:` hat keinen Schatten — `Color()`
// ist durchsichtig (`src/data/font.cpp`, Zeilen 14–24, und
// `src/gfx/color.hpp`, Zeile 32, Fassung f9d9356 vom 2021-07-06). Versatz und
// Weichheit sind Zahlen und keine Skripte.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestDerSchattenDerSchriftStehtImSatz(t *testing.T) {
	faelle := []struct {
		name     string
		schrift  msepaket.Schrift
		erwartet render.Textschatten
	}{
		{"ohne Angabe", msepaket.Schrift{}, render.Textschatten{}},
		{
			"ganz",
			msepaket.Schrift{SchattenFarbe: "black", SchattenX: "-1", SchattenY: "1", SchattenWeichheit: "2"},
			render.Textschatten{Farbe: render.Farbe{A: 255}, X: -1, Y: 1, Weichheit: 2},
		},
		{
			"als Skript",
			msepaket.Schrift{SchattenFarbe: `{if true then rgba(0,0,0,0) else "black"}`, SchattenX: "-1"},
			render.Textschatten{X: -1},
		},
		{
			"Versatz ohne Farbe",
			msepaket.Schrift{SchattenX: "2", SchattenY: "2"},
			render.Textschatten{X: 2, Y: 2},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			schrift := fall.schrift
			schrift.Name, schrift.Groesse = "Matrix", "12"

			ebene, befunde := textebeneMitSchrift(t, schrift, "Name")
			if len(befunde) > 0 {
				t.Fatalf("Befunde: %+v", befunde)
			}

			if ebene.Satz.Schatten != fall.erwartet {
				t.Errorf("Schatten %+v, erwartet %+v", ebene.Satz.Schatten, fall.erwartet)
			}
		})
	}
}

// **Eine Zahl, die keine ist, wird benannt** und gilt als null — MSE läse
// sie ebenso nicht.
func TestEinUnlesbarerSchattenversatzIstEinBefund(t *testing.T) {
	ebene, befunde := textebeneMitSchrift(t, msepaket.Schrift{
		Name: "Matrix", Groesse: "12", SchattenFarbe: "black", SchattenX: "links", SchattenWeichheit: "viel",
	}, "Name")

	if len(befunde) != 2 || !strings.Contains(befunde[0].Ort, "shadow displacement x") ||
		!strings.Contains(befunde[1].Ort, "shadow blur") {
		t.Fatalf("Befunde: %+v, erwartet zwei an »shadow displacement x« und »shadow blur«", befunde)
	}

	if ebene.Satz.Schatten.X != 0 || ebene.Satz.Schatten.Weichheit != 0 {
		t.Errorf("Schatten %+v, erwartet Versatz und Weichheit null", ebene.Satz.Schatten)
	}
}
