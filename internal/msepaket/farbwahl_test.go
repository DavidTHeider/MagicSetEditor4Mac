package msepaket_test

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Eine Farbstiloption erlaubt eine eigene Farbe, solange nichts anderes
// dasteht** (#634).
//
// **ported.** `ColorField::ColorField() : allow_custom(true)` —
// `src/data/field/color.cpp`, Zeilen 15–18, Fassung f9d9356 vom 2021-07-06.
//
// **Und über das ganze Pack ist es nirgends abgeschaltet**: alle 105
// Farbstiloptionen lassen die eigene Farbe zu. Der Farbwähler ist damit nicht
// die Ausnahme, sondern der Normalfall
// (#630).
func TestEineFarboptionErlaubtOhneAngabeDieEigeneFarbe(t *testing.T) {
	optionen := stiloptionenDesPruefstuecks(t)

	for name, erwartet := range map[string]bool{
		"randfarbe":  true,
		"eigenfarbe": true,
		"festfarbe":  false,
	} {
		option, da := optionen[name]
		if !da {
			t.Fatalf("die Stiloption %q fehlt", name)
		}

		if option.Kind != domain.KindColor {
			t.Errorf("%s: Art %q, erwartet %q", name, option.Kind, domain.KindColor)
		}

		if option.EigeneFarbe != erwartet {
			t.Errorf("%s: EigeneFarbe %v, erwartet %v", name, option.EigeneFarbe, erwartet)
		}
	}
}

// **Die Wahlen tragen Name und Farbe nebeneinander** — der Name steht in
// `Choices`, die Farbe an derselben Stelle in `Wahlfarben`.
func TestEineFarboptionTraegtNamenUndFarben(t *testing.T) {
	rand := stiloptionenDesPruefstuecks(t)["randfarbe"]

	if len(rand.Choices) != 2 || len(rand.Wahlfarben) != 2 {
		t.Fatalf("%d Namen, %d Farben — erwartet je zwei", len(rand.Choices), len(rand.Wahlfarben))
	}

	if rand.Choices[0] != "schwarz" || rand.Wahlfarben[0] != "rgb(0,0,0)" {
		t.Errorf("erste Wahl %q = %q", rand.Choices[0], rand.Wahlfarben[0])
	}
}

// stiloptionenDesPruefstuecks gibt die Stiloptionen nach Namen.
func stiloptionenDesPruefstuecks(t *testing.T) map[string]domain.StyleField {
	t.Helper()

	paket, err := msepaket.Oeffne(filepath.Join("..", "..", "testdaten", "probe-stil.mse-style"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = paket.Schliessen() }) //nolint:errcheck // gelesen ist gelesen

	stil, befunde := paket.Stilmodell()
	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	nachNamen := map[string]domain.StyleField{}
	for _, option := range stil.Stiloptionen {
		nachNamen[option.Name] = option
	}

	return nachNamen
}
