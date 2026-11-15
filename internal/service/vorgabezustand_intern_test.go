package service

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **„Es gibt eine Vorgabe" und „sie rechnet sich aus" sind zwei Aussagen** —
// aus dem Zweitblick über Phase L (2026-08-22).
//
// # Der Befund
//
// [Stiloption.Vorgabe] trug beides: leer hieß „keine Vorgabe" **und**
// „ausgerechnet hat sie nicht". Das Fenster machte daraus die Existenz des
// Eintrags.
//
// **MSE2 fragt nach dem Skript, nicht nach seinem Ergebnis:**
//
//	bool hasDefault() const { return field().default_script; }
//
// `src/gui/value/color.cpp`, Zeile 38, Fassung f9d9356 vom 2021-07-06. Und
// `selection()` wählt den Eintrag, **sobald der Wert Vorgabe ist** — vor jeder
// Auswertung (`color.cpp`, Zeilen 104–108).
//
// # Der Fall steht im Pack
//
// `magic-m15-bigtext.mse-style` schreibt `default: black` **ohne**
// Anführungszeichen. Das ist ein Skriptfehler („der Name »black« ist hier nicht
// bekannt"), und es fällt auf `initial: black` zurück — gemessen am 2026-08-21.
//
// **MSE2 zeigt dort einen Vorgabeeintrag, dieses Programm zeigte keinen.**
func TestVorhandenUndAusgerechnetSindZweiAussagen(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		feld       domain.StyleField
		hatVorgabe bool
		vorgabe    string
	}{
		"eine Vorgabe, die sich ausrechnet": {
			feld:       domain.StyleField{Kind: domain.KindColor, Default: `"black"`},
			hatVorgabe: true,
			vorgabe:    "black",
		},
		"eine Vorgabe, die scheitert — der Fall aus magic-m15-bigtext": {
			feld:       domain.StyleField{Kind: domain.KindColor, Default: "black", Initial: "black"},
			hatVorgabe: true,
			vorgabe:    "",
		},
		"gar keine Vorgabe": {
			feld:       domain.StyleField{Kind: domain.KindColor, Initial: "rgb(1,2,3)"},
			hatVorgabe: false,
			vorgabe:    "",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			option := stiloptionAus(t.Context(), fall.feld, msescript.Nichts)

			if option.HatVorgabe != fall.hatVorgabe {
				t.Errorf("HatVorgabe %v, erwartet %v", option.HatVorgabe, fall.hatVorgabe)
			}

			if option.Vorgabe != fall.vorgabe {
				t.Errorf("Vorgabe %q, erwartet %q", option.Vorgabe, fall.vorgabe)
			}
		})
	}
}
