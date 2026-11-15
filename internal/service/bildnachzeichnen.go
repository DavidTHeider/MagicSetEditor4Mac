// **Ein Bild nachzeichnen** —
// #1340,
// Werkstück F von Phase 4.11.
//
// Portiert aus `import_symbol` und `image_to_symbol`
// (`src/data/format/image_to_symbol.cpp`, Zeilen 207–237, Fassung `f9d9356`
// vom 2021-07-06).
//
// # Was es tut
//
// Aus einem Bild wird eine Symbolzeichnung: Silhouette suchen, umlaufen,
// vereinfachen. Der Editor zeichnete bis heute nur von Hand.

package service

import (
	"fmt"
	"image"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// hoechstmass ist die Kantenlänge, auf die ein großes Bild verkleinert wird.
//
// **MSEs Zahl samt Begründung** (Zeile 229): „100x100 ought to be enough, we
// throw out most afterwards data anyway". Ein größeres Bild gäbe mehr Punkte
// und nach dem Vereinfachen dieselbe Form — bei vielfacher Rechenzeit.
const hoechstmass = 100

// BildNachzeichnen macht aus einem Bild eine Symbolzeichnung.
func BildNachzeichnen(gemalt image.Image) ([]msepaket.Symbolteil, error) {
	if gemalt == nil {
		return nil, fmt.Errorf("%w: es liegt kein Bild vor", domain.ErrValidation)
	}

	kasten := gemalt.Bounds()
	if kasten.Dx() < 3 || kasten.Dy() < 3 {
		return nil, fmt.Errorf(
			"%w: das Bild ist %d × %d groß und damit zu klein für einen Umriss",
			domain.ErrValidation, kasten.Dx(), kasten.Dy())
	}

	// **Ein MSE-1-Symbol trägt seine Zeichnung rechts oben** — 60 × 80 groß,
	// und genommen wird der Ausschnitt (20,0,40,40) (Zeilen 226–228). Der Rest
	// des Bildes ist die farbige Vorschau, die MSE1 danebenlegte.
	if istMSE1Symbol(gemalt) {
		gemalt = render.ZuschnittAus(gemalt, image.Rect(20, 0, 60, 40).Add(kasten.Min), 40, 40)
	} else if kasten.Dx() > hoechstmass || kasten.Dy() > hoechstmass {
		gemalt = render.ZuschnittAus(gemalt, kasten, hoechstmass, hoechstmass)
	}

	feld := schwelle(gemalt)

	teile := make([]msepaket.Symbolteil, 0, 4)

	for {
		teil, gefunden := umrissLesen(feld)
		if !gefunden {
			break
		}

		teile = append(teile, formVereinfachen(teil))
	}

	if len(teile) == 0 {
		return nil, fmt.Errorf(
			"%w: in diesem Bild ist keine Fläche zu erkennen", domain.ErrValidation)
	}

	// **Umgedreht** — `reverse(symbol->parts...)` (Zeile 220). Der Rundgang
	// findet die Formen von links nach rechts; MSE legt die zuletzt gefundene
	// nach vorn, damit ein Loch **über** der Fläche liegt, in der es steckt.
	slices.Reverse(teile)

	for i := range teile {
		teile[i].Name = fmt.Sprintf("Form %d", i+1)
	}

	return teile, nil
}

// istMSE1Symbol erkennt ein Symbol aus MSE 1 an Größe und Farbigkeit.
//
// Portiert aus `is_mse1_symbol` (Zeilen 97–114). **Die Schwelle 5000 ist
// MSEs**, und sie misst, wie weit die rechte Hälfte von Grau abweicht: dort
// steht bei MSE 1 die schwarzweiße Maske.
func istMSE1Symbol(gemalt image.Image) bool {
	kasten := gemalt.Bounds()
	if kasten.Dx() != 60 || kasten.Dy() != 80 {
		return false
	}

	abweichung := 0

	for y := range 80 {
		for x := 20; x < 60; x++ {
			r, g, b, _ := gemalt.At(kasten.Min.X+x, kasten.Min.Y+y).RGBA()
			rr, gg, bb := int(r>>8), int(g>>8), int(b>>8)
			abweichung += abs(rr-bb) + abs(rr-gg) + abs(bb-gg)
		}
	}

	return abweichung <= 5000
}

func abs(a int) int {
	if a < 0 {
		return -a
	}

	return a
}

// **Zuschnitt und Verkleinerung stehen in `internal/render`.**
//
// `render.ZuschnittAus` schneidet eine Fläche heraus **und** rechnet sie auf
// ein Maß um — beides, was hier gebraucht wird. Eine zweite Fassung daneben
// liefe mit jener auseinander (§2), und sie täte es still: eine Kante um einen
// Bildpunkt versetzt sieht man einem Symbol nicht an.
//
// **Eine Abweichung von MSE steht darin**: MSEs `Rescale` glättet, und die
// Abtastung hier tut es auch — das ist dieselbe Art Umrechnung, nicht dieselbe
// Formel. Was danach kommt, ist eine Schwelle auf voll oder leer; ein um einen
// Ton anderer Grauwert kippt sie nur an der Grenze. Vermerkt in
// `docs/PORTIERUNGSKARTE.md`.
