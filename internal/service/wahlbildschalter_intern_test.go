package service

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Der Schalter, an dem die Oberfläche entscheidet, ob sie überhaupt fragt**
// (#1225).
//
// **Ein interner Test**, weil die Frage eine interne ist: `Optionen` reicht
// das Ergebnis nur durch, und ein Test über den ganzen Wähler bräuchte für
// jede der vier Antworten ein eigenes Stilpaket.
//
// **Zwei Wege, wie in [Planbauer.Optionsbilder]** — ein `choice images:`-Block
// oder ein **gerechnetes** `image:` (`src/gui/value/choice.cpp`, Zeilen
// 147–160, Fassung f9d9356 vom 2021-07-06).
func TestDerWahlbildschalter(t *testing.T) {
	for _, fall := range []struct {
		name    string
		sicht   []msepaket.Feldstil
		gesucht string
		will    bool
	}{
		{
			name:    "ohne Sichtseite gar nichts",
			gesucht: "hintergrund",
		},
		{
			name:    "ein Wahlbildblock zählt",
			sicht:   []msepaket.Feldstil{{Feld: "hintergrund", Wahlbilder: []msepaket.Wahlbild{{Wahl: "hell"}}}},
			gesucht: "hintergrund",
			will:    true,
		},
		{
			name:    "ein gerechnetes Bild zählt",
			sicht:   []msepaket.Feldstil{{Feld: "hintergrund", Bild: `{ input + ".png" }`}},
			gesucht: "hintergrund",
			will:    true,
		},
		{
			// **#1934**: `isScripted()` ist auch für den Zeichenkettenmodus
			// wahr (`src/script/image.cpp` Zeile 73, Fassung f9d9356 vom
			// 2021-07-06).
			name:    "ein Bildname mit Ausdruck zählt",
			sicht:   []msepaket.Feldstil{{Feld: "hintergrund", Bild: "grund-{input}.png"}},
			gesucht: "hintergrund",
			will:    true,
		},
		{
			name:    "ein festes Bild zählt nicht",
			sicht:   []msepaket.Feldstil{{Feld: "hintergrund", Bild: "hell.png"}},
			gesucht: "hintergrund",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			stil := msepaket.Stil{Optionsstile: fall.sicht}

			if hat := hatWahlbilder(stil, fall.gesucht); hat != fall.will {
				t.Errorf("hatWahlbilder = %v, erwartet %v", hat, fall.will)
			}
		})
	}
}
