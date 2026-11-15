package render_test

import (
	"image"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **`ImageSlice::constrain`, ganz** —
// #576.
//
// [render.Zuschnittflaeche] portiert bisher **einen Aufruf** davon: den aus dem
// Konstruktor des Fensters, mit der Auswahl über dem ganzen Bild und ohne
// bevorzugte Kante. Der Zuschnittdialog ruft dieselbe Regel nach **jeder**
// Bewegung, und dann tragen die anderen Teile:
//
//	void ImageSlice::constrain(PreferedProperty prefer) {
//	  selection.width  = max(1, selection.width);
//	  selection.height = max(1, selection.height);
//	  if (!allow_outside) {
//	    selection.width  = min(selection.width,  source.GetWidth());
//	    selection.height = min(selection.height, source.GetHeight());
//	    selection.x      = max(selection.x, 0);
//	    selection.y      = max(selection.y, 0);
//	    selection.x      = min(selection.x, source.GetWidth()  - selection.width);
//	    selection.y      = min(selection.y, source.GetHeight() - selection.height);
//	  }
//	  if (aspect_fixed) {
//	    int diff = selection.width * target_size.GetHeight()
//	             - selection.height * target_size.GetWidth();
//	    if ((diff > 0 && prefer != PREFER_WIDTH) || prefer == PREFER_HEIGHT) {
//	      selection.width  -= int(diff / target_size.GetHeight());
//	    } else {
//	      selection.height -= int(-diff / target_size.GetWidth());
//	    }
//	  }
//	}
//
// `src/gui/image_slice_window.cpp`, Zeilen 27–52, Fassung f9d9356 vom
// 2021-07-06.
//
// **`allow_outside` ist nicht portiert, und das ist keine Auslassung**: MSE
// setzt es im Konstruktor auf `false` und lässt es dort. Die Ankreuzfläche
// dazu steht im Quelltext **auskommentiert** (Zeile 114: `// allowOutside=
// new CheckBox(…)`). Ein Zweig, den kein Weg betritt, wäre hier ungeprüfter
// Code.
func TestDieZuschnittwahlFolgtMSEsRegel(t *testing.T) {
	t.Parallel()

	quelle := image.Rect(0, 0, 100, 80)

	faelle := map[string]struct {
		wahl      render.Zuschnittwahl
		bevorzugt render.Bevorzugt
		erwartet  image.Rectangle
	}{
		// **Mindestens ein Bildpunkt.** Wer den Rahmen auf nichts zieht,
		// bekommt einen Punkt und kein leeres Bild.
		"eine Auswahl ohne Fläche": {
			wahl: render.Zuschnittwahl{
				Quelle: quelle, Ziel: image.Pt(10, 10),
				Auswahl: image.Rect(5, 5, 5, 5),
			},
			erwartet: image.Rect(5, 5, 6, 6),
		},

		// **Nichts ragt über die Quelle hinaus** — weder in der Größe noch in
		// der Lage.
		"eine Auswahl über den Rand hinaus": {
			wahl: render.Zuschnittwahl{
				Quelle: quelle, Ziel: image.Pt(10, 10),
				Auswahl: image.Rect(90, 70, 130, 110),
			},
			erwartet: image.Rect(60, 40, 100, 80),
		},
		"eine Auswahl links oberhalb": {
			wahl: render.Zuschnittwahl{
				Quelle: quelle, Ziel: image.Pt(10, 10),
				Auswahl: image.Rect(-30, -20, 10, 20),
			},
			erwartet: image.Rect(0, 0, 40, 40),
		},

		// **Größer als die Quelle wird auf die Quelle gestutzt** — und die
		// Lage danach neu bestimmt.
		//
		// 200×180 wird auf 100×80. Damit **muss** die Ecke auf (0,0) fallen:
		// `x = min(10, 100−100)`. Erst danach greift das Seitenverhältnis und
		// nimmt der Breite 20 Punkte. Der erste Anlauf erwartete hier
		// (10,10)-(90,90) — er hatte die Lagegrenze übersehen, die aus der
		// vollen Breite folgt.
		"eine Auswahl größer als die Quelle": {
			wahl: render.Zuschnittwahl{
				Quelle: quelle, Ziel: image.Pt(1, 1),
				Auswahl: image.Rect(10, 10, 210, 190),
			},
			erwartet: image.Rect(0, 0, 80, 80),
		},

		// **Ohne festes Seitenverhältnis bleibt sie, wie sie ist** — das ist
		// MSEs „force to fit".
		"ohne festes Seitenverhältnis": {
			wahl: render.Zuschnittwahl{
				Quelle: quelle, Ziel: image.Pt(10, 20),
				Auswahl: image.Rect(0, 0, 40, 40), Frei: true,
			},
			erwartet: image.Rect(0, 0, 40, 40),
		},

		// **Zu breit: die Breite gibt nach.** 40×40 auf ein Feld 10×20 —
		// `diff = 40*20 − 40*10 = 400`, `400/20 = 20`, Breite 20.
		"zu breit für das Feld": {
			wahl: render.Zuschnittwahl{
				Quelle: quelle, Ziel: image.Pt(10, 20),
				Auswahl: image.Rect(0, 0, 40, 40),
			},
			erwartet: image.Rect(0, 0, 20, 40),
		},

		// **Zu hoch: die Höhe gibt nach.** 40×40 auf ein Feld 20×10 —
		// `diff = 40*10 − 40*20 = −400`, `400/20 = 20`, Höhe 20.
		"zu hoch für das Feld": {
			wahl: render.Zuschnittwahl{
				Quelle: quelle, Ziel: image.Pt(20, 10),
				Auswahl: image.Rect(0, 0, 40, 40),
			},
			erwartet: image.Rect(0, 0, 40, 20),
		},

		// **Wer die Breite zieht, behält sie** — dann wächst die Höhe.
		// Derselbe Fall wie „zu breit", aber mit `PREFER_WIDTH`: der Zweig
		// dreht sich um, `-diff` ist negativ, und `height -= negativ` **wächst**.
		"die Breite ist die gezogene Kante": {
			wahl: render.Zuschnittwahl{
				Quelle: quelle, Ziel: image.Pt(10, 20),
				Auswahl: image.Rect(0, 0, 40, 40),
			},
			bevorzugt: render.BevorzugtBreite,
			erwartet:  image.Rect(0, 0, 40, 80),
		},

		// **Und wer die Höhe zieht, behält sie.** `PREFER_HEIGHT` erzwingt den
		// Breitenzweig, auch wenn `diff` negativ ist.
		"die Höhe ist die gezogene Kante": {
			wahl: render.Zuschnittwahl{
				Quelle: quelle, Ziel: image.Pt(20, 10),
				Auswahl: image.Rect(0, 0, 40, 40),
			},
			bevorzugt: render.BevorzugtHoehe,
			erwartet:  image.Rect(0, 0, 80, 40),
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if gab := fall.wahl.Begrenze(fall.bevorzugt); gab != fall.erwartet {
				t.Errorf("gab %v, erwartet %v", gab, fall.erwartet)
			}
		})
	}
}

// **Und die Vorgabewahl ist derselbe Aufruf.**
//
// [render.Zuschnittflaeche] hat bis #576 eine eigene Rechnung geführt. Sie war
// dieselbe — für **einen** Aufruf. Wer zwei Fassungen einer portierten Regel
// stehen lässt, hat zwei Antworten auf dieselbe Frage, sobald eine davon
// nachgezogen wird.
func TestDieVorgabewahlIstDieselbeRegel(t *testing.T) {
	t.Parallel()

	groessen := []image.Point{
		{X: 100, Y: 80}, {X: 1, Y: 1}, {X: 375, Y: 523}, {X: 4000, Y: 3000}, {X: 3, Y: 7},
	}

	for _, quelle := range groessen {
		for _, ziel := range groessen {
			flaeche := image.Rect(0, 0, quelle.X, quelle.Y)

			wahl := render.Zuschnittwahl{
				Quelle: flaeche, Ziel: ziel, Auswahl: flaeche,
			}

			if gab := wahl.Begrenze(render.BevorzugtKeines); gab != render.Zuschnittflaeche(flaeche, ziel.X, ziel.Y) {
				t.Errorf("Quelle %v, Ziel %v: %v gegen %v",
					quelle, ziel, gab, render.Zuschnittflaeche(flaeche, ziel.X, ziel.Y))
			}
		}
	}
}

// **Das Ziehen ist portierte Regel und keine Fensterarbeit** —
// #576.
//
//	void ImageSliceSelector::onMotion(wxMouseEvent& ev) {
//	  double deltaX = (ev.GetX() - mouseX) / scaleX;
//	  double deltaY = (ev.GetY() - mouseY) / scaleY;
//	  if (dragX == 0 && dragY == 0) {
//	    slice.selection.x = int(start_selection.x + deltaX);
//	    slice.selection.y = int(start_selection.y + deltaY);
//	  } else {
//	    if (slice.aspect_fixed) {
//	      double d = (dragX * deltaX * slice.target_size.GetHeight()
//	                + dragY * deltaY * slice.target_size.GetWidth()) / 2;
//	      deltaX = dragX * d / slice.target_size.GetHeight();
//	      deltaY = dragY * d / slice.target_size.GetWidth();
//	    }
//	    if (dragX) {
//	      slice.selection.x     = int(start_selection.x     + deltaX * (1 - dragX) / 2);
//	      slice.selection.width = int(start_selection.width + deltaX * dragX);
//	    }
//	    if (dragY) { … dasselbe senkrecht … }
//	  }
//	}
//
// `src/gui/image_slice_window.cpp`, Zeilen 609–643, Fassung f9d9356 vom
// 2021-07-06.
//
// **Warum das nicht ins Frontend gehört**: es ist dieselbe Frage wie
// [Zuschnittwahl.Begrenze] — zwei Fassungen einer portierten Regel geben zwei
// Antworten, sobald jemand eine davon nachzieht (§15). Das Fenster rechnet
// Bildschirmpunkte in Quellpunkte um und reicht sie herein.
func TestDasZiehenFolgtMSEsRegel(t *testing.T) {
	t.Parallel()

	quelle := image.Rect(0, 0, 400, 300)

	faelle := map[string]struct {
		wahl           render.Zuschnittwahl
		x, y           float64
		kanteX, kanteY int
		erwartet       image.Rectangle
	}{
		// **Ohne Ecke wird verschoben** — die Größe bleibt.
		"das Ganze verschieben": {
			wahl: render.Zuschnittwahl{
				Quelle: quelle, Ziel: image.Pt(100, 100),
				Auswahl: image.Rect(0, 0, 200, 200),
			},
			x: 50, y: 20,
			erwartet: image.Rect(50, 20, 250, 220),
		},

		// **Und das Verschieben endet am Rand** — [Zuschnittwahl.Begrenze]
		// läuft danach, wie in MSE über `EVENT_SLICE_CHANGED`.
		"über den Rand hinaus verschieben": {
			wahl: render.Zuschnittwahl{
				Quelle: quelle, Ziel: image.Pt(100, 100),
				Auswahl: image.Rect(0, 0, 200, 200),
			},
			x: 500, y: 500,
			erwartet: image.Rect(200, 100, 400, 300),
		},

		// **Die untere rechte Ecke.** `d = (1·60·100 + 1·40·100)/2 = 5000`,
		// also `deltaX = deltaY = 50` — das Mittel beider Wege, weil das
		// Seitenverhältnis 1:1 ist.
		"die untere rechte Ecke ziehen": {
			wahl: render.Zuschnittwahl{
				Quelle: quelle, Ziel: image.Pt(100, 100),
				Auswahl: image.Rect(0, 0, 100, 100),
			},
			x: 60, y: 40, kanteX: 1, kanteY: 1,
			erwartet: image.Rect(0, 0, 150, 150),
		},

		// **Die obere linke Ecke wandert mit.** `dragX = −1`, also
		// `x += deltaX·(1−(−1))/2 = deltaX`, und die Breite nimmt um `deltaX`
		// ab.
		"die obere linke Ecke ziehen": {
			wahl: render.Zuschnittwahl{
				Quelle: quelle, Ziel: image.Pt(100, 100),
				Auswahl: image.Rect(50, 50, 250, 250),
			},
			x: 40, y: 40, kanteX: -1, kanteY: -1,
			erwartet: image.Rect(90, 90, 250, 250),
		},

		// **Eine Wahl ohne Ziel gibt nichts** — dieselbe Wache wie in
		// [Zuschnittwahl.Begrenze]: ein Feld ohne Fläche liegt auf dieser
		// Karte nicht, und ein Zug darauf hat kein Ergebnis.
		"eine Wahl ohne Zielfläche": {
			wahl: render.Zuschnittwahl{
				Quelle: quelle, Ziel: image.Pt(0, 0),
				Auswahl: image.Rect(0, 0, 100, 100),
			},
			x: 10, y: 10,
			erwartet: image.Rectangle{},
		},

		// **Ohne festes Seitenverhältnis zieht jede Kante für sich** — das ist
		// MSEs „force to fit", und nur dort gibt es überhaupt Kantengriffe:
		// *„only drag corners if aspect fixed"* (Zeile 596).
		"eine Kante ohne festes Seitenverhältnis": {
			wahl: render.Zuschnittwahl{
				Quelle: quelle, Ziel: image.Pt(100, 100),
				Auswahl: image.Rect(0, 0, 100, 100), Frei: true,
			},
			x: 60, y: 0, kanteX: 1,
			erwartet: image.Rect(0, 0, 160, 100),
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if gab := fall.wahl.Ziehe(fall.x, fall.y, fall.kanteX, fall.kanteY); gab != fall.erwartet {
				t.Errorf("gab %v, erwartet %v", gab, fall.erwartet)
			}
		})
	}
}
