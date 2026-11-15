package render_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// manaschrift ist eine Symbolschrift, wie ein `.mse-symbol-font` sie beschreibt.
type manaschrift map[string]string

func (m manaschrift) Bildpfad(zeichen string) (string, bool) {
	pfad, da := m[zeichen]

	return pfad, da
}

func magicMana() manaschrift {
	return manaschrift{
		"W": "/magic-mana-large.mse-symbol-font/w.png",
		"U": "/magic-mana-large.mse-symbol-font/u.png",
		"T": "/magic-mana-large.mse-symbol-font/tap.png",
	}
}

// kurz stellt ein Stück lesbar dar: Text, `Text→Bild` oder `Text?`.
func kurzform(stuecke []render.Textstueck) string {
	teile := make([]string, 0, len(stuecke))

	for _, stueck := range stuecke {
		switch {
		case stueck.Unbekannt:
			teile = append(teile, stueck.Text+"?")

		case stueck.Bild != "":
			teile = append(teile, stueck.Text+"→"+stueck.Bild)

		default:
			teile = append(teile, stueck.Text)
		}
	}

	return strings.Join(teile, "|")
}

func TestSymboleWerdenAusDemTextGeloest(t *testing.T) {
	faelle := []struct {
		name     string
		text     string
		erwartet string
	}{
		{
			name:     "ein Symbol mitten im Satz",
			text:     "{T}: Ziehe eine Karte.",
			erwartet: "{T}→/magic-mana-large.mse-symbol-font/tap.png|: Ziehe eine Karte.",
		},
		{
			name:     "mehrere Symbole hintereinander",
			text:     "{W}{U}",
			erwartet: "{W}→/magic-mana-large.mse-symbol-font/w.png|{U}→/magic-mana-large.mse-symbol-font/u.png",
		},
		{
			name:     "Text ohne Symbol",
			text:     "Fliegend",
			erwartet: "Fliegend",
		},
		{
			name:     "Text vor und nach einem Symbol",
			text:     "Zahle {W} dafür",
			erwartet: "Zahle |{W}→/magic-mana-large.mse-symbol-font/w.png| dafür",
		},
		{
			// **Ein unbekanntes Symbol bleibt sichtbar.** Ein Kartentext, aus
			// dem stillschweigend etwas verschwindet, sieht vollständig aus —
			// und wer ihn gegen das Original hält, findet die Lücke erst beim
			// Zählen.
			name:     "ein Symbol, das die Schrift nicht kennt",
			text:     "{Q}",
			erwartet: "{Q}?",
		},
		{
			name:     "bekannt und unbekannt gemischt",
			text:     "{W}{Q}{U}",
			erwartet: "{W}→/magic-mana-large.mse-symbol-font/w.png|{Q}?|{U}→/magic-mana-large.mse-symbol-font/u.png",
		},
		{
			// Eine Klammer ohne Ende ist keine halbe Symbolklammer, sondern eine
			// geschweifte Klammer, die jemand getippt hat.
			name:     "eine Klammer ohne Ende",
			text:     "Kosten {T",
			erwartet: "Kosten {T",
		},
		{
			name:     "eine schließende Klammer ohne öffnende",
			text:     "T}",
			erwartet: "T}",
		},
		{
			name:     "eine leere Klammer",
			text:     "{}",
			erwartet: "{}?",
		},
		{
			name:     "leerer Text",
			text:     "",
			erwartet: "",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if got := kurzform(render.ZerlegeSymbole(fall.text, magicMana())); got != fall.erwartet {
				t.Errorf("%q, erwartet war %q", got, fall.erwartet)
			}
		})
	}
}

// Eine Schrift ohne Symbole lässt jeden Text unverändert — als Text.
func TestOhneSymbolschriftBleibtAllesText(t *testing.T) {
	stuecke := render.ZerlegeSymbole("{W}{U}", manaschrift{})

	if len(stuecke) != 2 {
		t.Fatalf("%d Stücke: %s", len(stuecke), kurzform(stuecke))
	}

	for _, stueck := range stuecke {
		if !stueck.Unbekannt {
			t.Errorf("%q wurde nicht als unbekannt geführt", stueck.Text)
		}
	}
}
