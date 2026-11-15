package service

import (
	"context"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// schriftschatten rechnet den Schatten einer Schrift aus — F02
// (#2244).
//
// **Die Farbe ist ein Skript, mit durchsichtiger Vorgabe.** `shadow_color`
// ist ein `Scriptable<Color>`, und sein leerer Wert ist `Color()` mit Alpha
// null (`src/data/font.hpp`, Zeilen 44–46, `src/gfx/color.hpp`, Zeile 32,
// Fassung f9d9356 vom 2021-07-06) — anders als die Schriftfarbe, deren
// Vorgabe Schwarz ist. **Versatz und Weichheit sind Zahlen**, gelesen über
// `REFLECT_N` und `REFLECT` ohne Skript (`src/data/font.cpp`, Zeilen 118–134).
//
// Im Pack steht die Farbe meist als Skript, das für gewöhnliche Rahmen
// `rgba(0,0,0,0)` gibt — der Schatten gehört dann zu einem besonderen Rahmen
// (`magic-m15-altered`: planeshifted, FNM, DKA).
func schriftschatten(
	ctx context.Context, feld *msepaket.Feldstil, umgebung *msescript.Umgebung,
) (render.Textschatten, []domain.Befund) {
	var (
		schatten render.Textschatten
		befunde  []domain.Befund
	)

	if strings.TrimSpace(feld.Schrift.SchattenFarbe) != "" {
		schatten.Farbe, befunde = farbangabe(ctx, feld.Feld, "shadow color", feld.Schrift.SchattenFarbe, umgebung)
	}

	for _, angabe := range []struct {
		name, wert string
		ziel       *float64
	}{
		{"shadow displacement x", feld.Schrift.SchattenX, &schatten.X},
		{"shadow displacement y", feld.Schrift.SchattenY, &schatten.Y},
		{"shadow blur", feld.Schrift.SchattenWeichheit, &schatten.Weichheit},
	} {
		roh := strings.TrimSpace(angabe.wert)
		if roh == "" {
			continue
		}

		zahl, err := strconv.ParseFloat(roh, 64)
		if err != nil {
			befunde = append(befunde, domain.Befund{
				Ort:     feld.Feld + " → " + angabe.name,
				Meldung: "die Angabe ist keine Zahl: " + roh,
			})

			continue
		}

		*angabe.ziel = zahl
	}

	return schatten, befunde
}
