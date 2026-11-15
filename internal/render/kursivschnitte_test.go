package render_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// namensschriften kennt Schriften nur unter ihrem Namen und merkt sich, ob
// fett gefragt wurde.
type namensschriften map[string]float64

func (n namensschriften) Suche(name string, fett bool) (render.Schrift, error) {
	breite, da := n[name]
	if !da {
		return nil, errors.New("unbekannt")
	}

	if fett {
		breite *= 10
	}

	return gleichbreit{breite: breite}, nil
}

// **Ohne `italic name:` ist der kursive Schnitt der der Familie** — Befund 3
// des Zweitblicks über E/F.
//
// MSE reicht dann Familienname und `wxFONTSTYLE_ITALIC` an das
// Betriebssystem (`Font::toWxFont`, `src/data/font.cpp`, Zeilen 83–108,
// Fassung f9d9356 vom 2021-07-06), und das nimmt den Schnitt „Italic" der
// Familie. Hier blieb ein Feld mit `style: italic` und ohne `italic name:`
// aufrecht — in der Rasterung, im SVG und beim Messen gleichermaßen.
func TestDerKursiveSchnittKommtAusDerFamilie(t *testing.T) {
	t.Parallel()

	schriften := namensschriften{
		"Probe": 1, "Probe Italic": 2, "Probe Bold Italic": 3, "Eigen Kursiv": 4,
	}

	for _, fall := range []struct {
		name, schrift, kursivname string
		kursiv, kursivFett        float64
	}{
		{name: "mit italic name dieser Name, fett über das Gewicht", schrift: "Probe", kursivname: "Eigen Kursiv", kursiv: 4, kursivFett: 40},
		{name: "ohne italic name der Schnitt der Familie", schrift: "Probe", kursiv: 2, kursivFett: 3},
		{name: "ohne Schnitt in der Familie keiner", schrift: "Andere"},
		{name: "ohne Namen keiner", schrift: ""},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			kursiv, kursivFett := render.Kursivschnitte(schriften, fall.schrift, fall.kursivname)

			if breite := breiteVon(kursiv); breite != fall.kursiv {
				t.Errorf("kursiv %v, erwartet %v", breite, fall.kursiv)
			}

			if breite := breiteVon(kursivFett); breite != fall.kursivFett {
				t.Errorf("kursiv-fett %v, erwartet %v", breite, fall.kursivFett)
			}
		})
	}
}

func breiteVon(schrift render.Schrift) float64 {
	if schrift == nil {
		return 0
	}

	return schrift.Breite('a')
}
