// Der **Winkel** eines Feldstils — `angle:`
// (#486).
//
// Eigene Datei, weil `stilmasse_test.go` mit ihm über 400 Zeilen ginge (§2).
package msepaket_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **`angle:` wird gelesen** — die Drehung eines Feldes in Grad.
//
//	Rotation ValueViewer::getRotation() const {
//	  return Rotation(deg_to_rad(getStyle()->angle), bounding_box, 1.0, getStretch());
//	}
//
// `src/render/value/viewer.cpp`, Zeilen 44–46, Fassung f9d9356 vom
// 2021-07-06. Über die Stilpakete von `M15-Magic-Pack` steht die Angabe
// 75-mal in 19 von 190 Paketen
// (#486).
//
// **Roh gelesen, weil sie gerechnet sein darf**: 19 der 75 stehen als
// `{if …}`.
func TestAngleWirdGelesen(t *testing.T) {
	faelle := []struct{ name, quelle, erwartet string }{
		{"eine Zahl", "angle: 270", "270"},
		{
			"gerechnet", `angle: { if card.name == "x" then 90 else 0 }`,
			`{ if card.name == "x" then 90 else 0 }`,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			pfad := probestilpaket(t, `mse version: 2.0.0
game: probe
card width: 100
card height: 100
card style:
	text:
		left: 1
		`+fall.quelle+`
`)

			paket, err := msepaket.Oeffne(pfad)
			if err != nil {
				t.Fatalf("öffnen: %v", err)
			}

			t.Cleanup(func() {
				if fehler := paket.Schliessen(); fehler != nil {
					t.Errorf("schließen: %v", fehler)
				}
			})

			stil, _ := paket.Stilmodell()
			if len(stil.Felder) != 1 {
				t.Fatalf("%d Feldstile", len(stil.Felder))
			}

			if haben := strings.TrimSpace(stil.Felder[0].Winkel); haben != fall.erwartet {
				t.Errorf("der Winkel ist %q, erwartet %q", haben, fall.erwartet)
			}
		})
	}
}

// **`scale down to:` wird gelesen** — die kleinste Schriftgröße, auf die ein
// zu langer Text verkleinert werden darf
// (#498).
//
//	REFLECT(scale_down_to);
//
// `src/data/font.cpp`, Zeile 127, Fassung f9d9356 vom 2021-07-06. **571
// Angaben in 174 der 190 Stilpakete** von `M15-Magic-Pack`.
func TestScaleDownToWirdGelesen(t *testing.T) {
	pfad := probestilpaket(t, `mse version: 2.0.0
game: probe
card width: 100
card height: 100
card style:
	text:
		left: 1
		font:
			name: Probe
			size: 13
			scale down to: 6
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, _ := paket.Stilmodell()
	if len(stil.Felder) != 1 {
		t.Fatalf("%d Feldstile", len(stil.Felder))
	}

	if haben := strings.TrimSpace(stil.Felder[0].Schrift.Kleinstgroesse); haben != "6" {
		t.Errorf("die Kleinstgröße ist %q, erwartet 6", haben)
	}
}
