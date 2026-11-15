package render_test

// **Das Maß eines erzeugten Bilds, wie es eine Wahl meldet**
// (#2303).
//
// `prepare_choice_viewer` erzeugt das Stilbild und übernimmt die Größe, die
// `generateCached` meldet (`src/render/value/choice.cpp`, Zeilen 35–58):
// das Bild, wie der Erzeuger es für die gefragte Größe liefert, gebracht
// durch `conform_image` (`src/gfx/generated_image.cpp`, Zeilen 30–66). Ein
// leerer Baum ist `BlankImage` — so groß, wie gefragt, mindestens 1 × 1
// (Zeilen 83–89). Beides Fassung f9d9356 vom 2021-07-06.
//
// **Gefragt am 2026-10-11 über einen Orakelstil** (`rarity` in 44 × 22,
// `alignment: middle right`): eine Datei von 30 × 10 meldet 44 × 14, eine
// von 10 × 30 meldet 7 × 22, ein breites Setsymbol 44 × 22, ein
// quadratisches 22 × 22, ein leerer Wert 44 × 22; mit `render style: both`
// 30 × 10, 10 × 30 und für den leeren Wert 1 × 1.

import (
	"image"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestDasErzeugteMassFolgtMSE2(t *testing.T) {
	quelle := bildmasse{da: map[string]image.Point{
		"/breit.png": {X: 30, Y: 10}, "/hoch.png": {X: 10, Y: 30},
	}}
	kasten := image.Point{X: 44, Y: 22}
	symbol := func(teil render.Symbolteil) render.Bildbaum {
		return render.Bildbaum{
			Knotenart: "symbol_variation", Symbol: []render.Symbolteil{teil},
			Farben: map[string]render.Farbe{"fill_color": {A: 255}, "border_color": {A: 255}},
		}
	}

	for _, fall := range []struct {
		name      string
		baum      render.Bildbaum
		gefragt   image.Point
		einpassen bool
		erwartet  image.Point
	}{
		{name: "breite Datei, eingepasst", baum: render.Blatt("/breit.png"), gefragt: kasten, einpassen: true, erwartet: image.Point{X: 44, Y: 14}},
		{name: "hohe Datei, eingepasst", baum: render.Blatt("/hoch.png"), gefragt: kasten, einpassen: true, erwartet: image.Point{X: 7, Y: 22}},
		{name: "breites Symbol", baum: symbol(viereck(0.1, 0.35, 0.9, 0.65)), gefragt: kasten, einpassen: true, erwartet: image.Point{X: 44, Y: 22}},
		{name: "quadratisches Symbol", baum: symbol(viereck(0.2, 0.2, 0.8, 0.8)), gefragt: kasten, einpassen: true, erwartet: image.Point{X: 22, Y: 22}},
		{name: "kein Bild", baum: render.Bildbaum{}, gefragt: kasten, einpassen: true, erwartet: kasten},
		{name: "Datei in Originalgröße", baum: render.Blatt("/breit.png"), erwartet: image.Point{X: 30, Y: 10}},
		{name: "kein Bild in Originalgröße", baum: render.Bildbaum{}, erwartet: image.Point{X: 1, Y: 1}},
		// **Gestreckt** heißt: die gefragte Größe, wie sie ist —
		// `ASPECT_STRETCH`, die Vorgabe einer Wahl ohne `alignment:`.
		{name: "gestreckt", baum: render.Blatt("/breit.png"), gefragt: kasten, erwartet: kasten},
	} {
		t.Run(fall.name, func(t *testing.T) {
			mass, da := render.NeuerRasterer(quelle).ErzeugtesMass(fall.baum, fall.gefragt, fall.einpassen)
			if !da || mass != fall.erwartet {
				t.Errorf("Maß %v (%v), erwartet %v", mass, da, fall.erwartet)
			}
		})
	}
}

// **Was sich nicht erzeugen lässt, hat kein Maß** — bei MSE2 wirft der
// Erzeuger, und das Maß bleibt, wo es war.
func TestEinBildDasFehltHatKeinMass(t *testing.T) {
	_, da := render.NeuerRasterer(bildmasse{}).ErzeugtesMass(
		render.Blatt("/fehlt.png"), image.Point{X: 44, Y: 22}, true)
	if da {
		t.Error("ein fehlendes Bild hat ein Maß")
	}
}
