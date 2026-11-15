package service

// **Ein Bild nachzeichnen — die Abnahme** —
// #1340,
// Werkstück F von Phase 4.11.
//
// # Die Zusage aus dem Issue
//
//	Ein PNG mit einer klaren Silhouette wird geladen und ergibt eine
//	Zeichnung mit Bézierpunkten, die man weiterbearbeiten kann.
//
// **„Die man weiterbearbeiten kann" ist die Hälfte, die man vergisst.** Ein
// Umriss mit dreihundert Punkten wäre auch eine Zeichnung — und niemand könnte
// ihn anfassen. Deshalb prüft der erste Fall die **Zahl** der Punkte.

import (
	"errors"
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// kreisbild malt einen gefüllten Kreis auf weißen Grund.
func kreisbild(mass int) image.Image {
	gemalt := image.NewRGBA(image.Rect(0, 0, mass, mass))
	mitte := float64(mass) / 2
	halbmesser := float64(mass) * 0.35

	for y := range mass {
		for x := range mass {
			ton := uint8(255)
			if math.Hypot(float64(x)-mitte, float64(y)-mitte) <= halbmesser {
				ton = 0
			}

			gemalt.Set(x, y, color.RGBA{R: ton, G: ton, B: ton, A: 255})
		}
	}

	return gemalt
}

// TestEinKreisWirdEineBearbeitbareZeichnung ist die Abnahme.
func TestEinKreisWirdEineBearbeitbareZeichnung(t *testing.T) {
	t.Parallel()

	teile, err := BildNachzeichnen(kreisbild(120))
	if err != nil {
		t.Fatalf("nachzeichnen: %v", err)
	}

	if len(teile) != 1 {
		t.Fatalf("%d Teile, erwartet 1", len(teile))
	}

	form := teile[0]

	if form.Art != msepaket.TeilForm {
		t.Errorf("Art %q, erwartet %q", form.Art, msepaket.TeilForm)
	}

	// **Weiterbearbeitbar heißt: wenige Punkte.**
	//
	// **Gemessen am 2026-08-31** über die Stufen dieser Kette, für genau
	// dieses Bild: der Rundgang gibt **340** Punkte, `punkteEntfernen` bringt
	// sie auf **36**, und `begradigen` macht aus 16 der 36 Strecken Geraden.
	//
	// **Die 36 sind nicht gewählt, sondern MSEs Schwelle `0.0002`.** Hier
	// stand zuerst „höchstens 24" — geraten, und der Lauf war rot. Die Zahl
	// folgt aus einer Konstante des Altcodes, und eine Erwartung dagegen zu
	// setzen hieße, das Verhalten zu ändern (§15 Satz 4).
	//
	// **Die Schranke steht deshalb dort, wo sie etwas aussagt**: sie fängt
	// eine Stufe ab, die ausfällt. Bliebe `punkteEntfernen` wirkungslos,
	// stünden hier dreihundert.
	if len(form.Punkte) > 60 {
		t.Errorf("%d Punkte — zu viele, um sie zu bearbeiten", len(form.Punkte))
	}

	if len(form.Punkte) < 4 {
		t.Fatalf("%d Punkte — das ist kein Kreis mehr", len(form.Punkte))
	}

	// **Und es sind Bézierpunkte**: mindestens eine Strecke ist eine Kurve,
	// sonst wäre aus dem Kreis ein Vieleck geworden.
	kurven := 0

	for _, punkt := range form.Punkte {
		if punkt.KurveDanach {
			kurven++
		}
	}

	if kurven == 0 {
		t.Error("keine einzige Kurve — ein Kreis aus lauter Geraden")
	}

	// **Die Punkte liegen ungefähr auf dem Kreis.** Das ist die Probe darauf,
	// dass die Vereinfachung die Form behalten hat und nicht nur Punkte
	// weggeworfen hat.
	for i, punkt := range form.Punkte {
		abstand := math.Hypot(punkt.Stelle.X-0.5, punkt.Stelle.Y-0.5)
		if math.Abs(abstand-0.35) > 0.06 {
			t.Errorf("Punkt %d liegt bei %.3f statt bei 0,35 vom Mittelpunkt", i, abstand)
		}
	}
}

// **Ein Rechteck wird zu vier Ecken und vier Geraden.**
//
// Das ist die Gegenprobe zum Kreis: dieselbe Kette, und heraus kommt etwas
// ganz anderes. Käme hier ebenfalls eine runde Form, prüfte der Kreisfall
// nichts.
func TestEinRechteckWirdEckig(t *testing.T) {
	t.Parallel()

	gemalt := image.NewRGBA(image.Rect(0, 0, 80, 80))

	for y := range 80 {
		for x := range 80 {
			ton := uint8(255)
			if x >= 20 && x < 60 && y >= 20 && y < 60 {
				ton = 0
			}

			gemalt.Set(x, y, color.RGBA{R: ton, G: ton, B: ton, A: 255})
		}
	}

	teile, err := BildNachzeichnen(gemalt)
	if err != nil {
		t.Fatalf("nachzeichnen: %v", err)
	}

	if len(teile) != 1 {
		t.Fatalf("%d Teile, erwartet 1", len(teile))
	}

	punkte := teile[0].Punkte
	if len(punkte) > 10 {
		t.Errorf("%d Punkte für ein Rechteck — erwartet höchstens zehn", len(punkte))
	}

	geraden := 0

	for _, punkt := range punkte {
		if !punkt.KurveDanach {
			geraden++
		}
	}

	if geraden < 4 {
		t.Errorf("%d Geraden — ein Rechteck hat vier Kanten", geraden)
	}
}

// **Ein Bild ohne Fläche ist ein fachliches Nein**, kein leeres Ergebnis (§8).
func TestEinLeeresBildLaesstSichNichtNachzeichnen(t *testing.T) {
	t.Parallel()

	gemalt := image.NewRGBA(image.Rect(0, 0, 40, 40))
	for y := range 40 {
		for x := range 40 {
			gemalt.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	_, err := BildNachzeichnen(gemalt)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet ErrValidation", err)
	}
}

// **Ein winziges Bild ebenso** — unter drei Bildpunkten gibt es keinen Umriss.
func TestEinWinzigesBildWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	_, err := BildNachzeichnen(image.NewRGBA(image.Rect(0, 0, 2, 2)))
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet ErrValidation", err)
	}
}

// **Und gar kein Bild.**
func TestOhneBildGibtEsNichtsNachzuzeichnen(t *testing.T) {
	t.Parallel()

	if _, err := BildNachzeichnen(nil); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet ErrValidation", err)
	}
}

// **Ein Ring gibt zwei Formen — und die innere wird abgezogen.**
//
// `read_symbol_shape` entscheidet das über die Zelle zwei links vom Anfang
// (Zeilen 199–202). Ohne diese Zeile wäre ein Ring eine volle Scheibe.
func TestEinRingGibtEinLoch(t *testing.T) {
	t.Parallel()

	const mass = 120

	gemalt := image.NewRGBA(image.Rect(0, 0, mass, mass))
	mitte := float64(mass) / 2

	for y := range mass {
		for x := range mass {
			abstand := math.Hypot(float64(x)-mitte, float64(y)-mitte)

			ton := uint8(255)
			if abstand <= 0.40*mass && abstand >= 0.22*mass {
				ton = 0
			}

			gemalt.Set(x, y, color.RGBA{R: ton, G: ton, B: ton, A: 255})
		}
	}

	teile, err := BildNachzeichnen(gemalt)
	if err != nil {
		t.Fatalf("nachzeichnen: %v", err)
	}

	if len(teile) != 2 {
		t.Fatalf("%d Teile, erwartet 2 — außen und das Loch", len(teile))
	}

	abgezogene := 0

	for _, teil := range teile {
		if teil.Kombination == msepaket.KombinationAbziehen {
			abgezogene++
		}
	}

	if abgezogene != 1 {
		t.Errorf("%d abgezogene Formen, erwartet genau eine", abgezogene)
	}
}

// **Jedes Teil bekommt einen Namen** — sonst stünde die Teilliste voll
// namenloser Zeilen.
func TestJedesNachgezeichneteTeilTraegtEinenNamen(t *testing.T) {
	t.Parallel()

	teile, err := BildNachzeichnen(kreisbild(60))
	if err != nil {
		t.Fatalf("nachzeichnen: %v", err)
	}

	for i, teil := range teile {
		if teil.Name == "" {
			t.Errorf("Teil %d trägt keinen Namen", i)
		}
	}
}
