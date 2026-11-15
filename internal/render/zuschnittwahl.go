package render

import "image"

// Bevorzugt sagt, welche Kante gerade gezogen wurde und deshalb stehen bleibt.
//
//	enum PreferedProperty { PREFER_NONE, PREFER_WIDTH, PREFER_HEIGHT };
//
// `src/gui/image_slice_window.hpp`, Zeilen 20–24, Fassung f9d9356 vom
// 2021-07-06.
//
// **Der Kommentar dort sagt, wozu**: *„Which option is just changed, and
// therefore more important?"* Wer die Breite zieht, will die Breite — die Höhe
// gibt nach, auch wenn sie dabei **wächst**.
type Bevorzugt int

const (
	// BevorzugtKeines ist der Fall aller Bewegungen, die keine Kante nennen —
	// das Verschieben, das Öffnen des Fensters.
	BevorzugtKeines Bevorzugt = iota
	BevorzugtBreite
	BevorzugtHoehe
)

// Zuschnittwahl ist der Ausschnitt, aus dem ein Bild geschnitten wird.
//
//	class ImageSlice {
//	  Image  source;
//	  wxSize target_size;
//	  wxRect selection;
//	  bool   allow_outside;
//	  bool   aspect_fixed;
//	};
//
// `src/gui/image_slice_window.hpp`, Zeilen 27–48, Fassung f9d9356 vom
// 2021-07-06.
//
// **Was hier fehlt und warum:**
//
//   - `allow_outside` — MSE setzt es im Konstruktor auf `false` und lässt es
//     dort; die Ankreuzfläche dazu steht **auskommentiert** im Quelltext
//     (Zeile 114). Ein Zweig, den kein Weg betritt, wäre ungeprüfter Code.
//   - `background` — er färbt, was außerhalb der Quelle liegt, und das gibt es
//     ohne `allow_outside` nicht.
//   - `sharpen` samt `sharpen_amount` — ein Filter, kein Ausschnitt. MSE hat
//     ihn in der Vorgabe **aus** (#576
//     nennt ihn nicht).
//
// **`Frei` statt `aspect_fixed`, und die Bedeutung ist umgedreht.** Der
// Nullwert einer Struktur in Go ist `false`, und MSEs Vorgabe ist
// `aspect_fixed = true` — ein Feld `Fest bool` läge damit in seiner Vorgabe
// falsch, und jede vergessene Zuweisung ergäbe stillschweigend „force to fit".
type Zuschnittwahl struct {
	// Quelle sind die Grenzen des Ursprungsbildes.
	Quelle image.Rectangle

	// Ziel ist die Größe des Feldes, auf das zugeschnitten wird.
	Ziel image.Point

	// Auswahl ist der gewählte Ausschnitt — in den Punkten der Quelle.
	Auswahl image.Rectangle

	// Frei hebt die Bindung an das Seitenverhältnis des Feldes auf; MSE nennt
	// das „force to fit".
	Frei bool
}

// Begrenze gibt die Auswahl, nachdem MSEs Regeln auf sie angewandt sind.
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
// **Die Reihenfolge ist die Regel.** Erst der Mindestpunkt, dann die Grenzen
// der Quelle, **dann** das Seitenverhältnis — und der letzte Schritt darf
// wieder herausführen, sofern er verkleinert. Wer die drei Schritte tauschte,
// bekäme bei einer Auswahl am Rand ein anderes Bild.
//
// **Ganzzahlig geteilt, und zwar zur Null hin.** C++ und Go schneiden gleich
// ab; bei `-diff` unter null **wächst** die Höhe dadurch, und genau das ist
// der Sinn von [BevorzugtBreite].
func (w Zuschnittwahl) Begrenze(bevorzugt Bevorzugt) image.Rectangle {
	if w.Quelle.Empty() || w.Ziel.X <= 0 || w.Ziel.Y <= 0 {
		return image.Rectangle{}
	}

	breite := max(w.Auswahl.Dx(), 1)
	hoehe := max(w.Auswahl.Dy(), 1)

	breite = min(breite, w.Quelle.Dx())
	hoehe = min(hoehe, w.Quelle.Dy())

	// **In den Punkten der Quelle gerechnet.** MSE misst gegen null, weil ein
	// `wxImage` dort beginnt; ein [image.Rectangle] muss das nicht.
	x := min(max(w.Auswahl.Min.X, w.Quelle.Min.X), w.Quelle.Max.X-breite)
	y := min(max(w.Auswahl.Min.Y, w.Quelle.Min.Y), w.Quelle.Max.Y-hoehe)

	if !w.Frei {
		unterschied := breite*w.Ziel.Y - hoehe*w.Ziel.X

		if (unterschied > 0 && bevorzugt != BevorzugtBreite) || bevorzugt == BevorzugtHoehe {
			breite -= unterschied / w.Ziel.Y
		} else {
			hoehe -= -unterschied / w.Ziel.X
		}
	}

	return image.Rect(x, y, x+breite, y+hoehe)
}

// Ziehe gibt die Auswahl, nachdem der Rahmen um eine Strecke gezogen wurde.
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
//	    if (dragY) {
//	      slice.selection.y      = int(start_selection.y      + deltaY * (1 - dragY) / 2);
//	      slice.selection.height = int(start_selection.height + deltaY * dragY);
//	    }
//	  }
//	  ProcessEvent(EVENT_SLICE_CHANGED); // → constrain()
//	}
//
// `src/gui/image_slice_window.cpp`, Zeilen 609–646, Fassung f9d9356 vom
// 2021-07-06.
//
// **Die Strecke kommt in Quellpunkten herein, nicht in Bildschirmpunkten.**
// MSE teilt sie durch `scaleX`, den Maßstab seines Fensters; hier tut das
// Fenster es, denn nur es kennt seine Größe. Was hier bleibt, ist die Regel.
//
// **kanteX und kanteY sind MSEs `dragX`/`dragY`** — je −1, 0 oder 1. Beide
// null heißt: der ganze Rahmen wird verschoben.
//
// **Bei festem Seitenverhältnis zieht nur eine Ecke.** MSE setzt eine
// Kantenmitte dann auf „kein Griff" zurück (Zeile 596: *„only drag corners if
// aspect fixed"*), und damit wird aus dem Zug ein Verschieben. Dieselbe Zeile
// steht hier, weil sonst ein Zug an der Kante die Fläche kippte.
//
// **Ganzzahlig zur Null hin**, wie C++ `int(…)`: Go schneidet gleich ab.
func (w Zuschnittwahl) Ziehe(x, y float64, kanteX, kanteY int) image.Rectangle {
	if w.Quelle.Empty() || w.Ziel.X <= 0 || w.Ziel.Y <= 0 {
		return image.Rectangle{}
	}

	if !w.Frei && (kanteX == 0 || kanteY == 0) {
		kanteX, kanteY = 0, 0
	}

	gezogen := w.Auswahl

	if kanteX == 0 && kanteY == 0 {
		gezogen = gezogen.Add(image.Pt(int(x), int(y)))
	} else {
		if !w.Frei {
			mass := (float64(kanteX)*x*float64(w.Ziel.Y) + float64(kanteY)*y*float64(w.Ziel.X)) / 2
			x = float64(kanteX) * mass / float64(w.Ziel.Y)
			y = float64(kanteY) * mass / float64(w.Ziel.X)
		}

		links, oben := gezogen.Min.X, gezogen.Min.Y
		breite, hoehe := gezogen.Dx(), gezogen.Dy()

		if kanteX != 0 {
			links = int(float64(links) + x*float64(1-kanteX)/2)
			breite = int(float64(breite) + x*float64(kanteX))
		}

		if kanteY != 0 {
			oben = int(float64(oben) + y*float64(1-kanteY)/2)
			hoehe = int(float64(hoehe) + y*float64(kanteY))
		}

		gezogen = image.Rect(links, oben, links+breite, oben+hoehe)
	}

	// **Und danach die Grenzen** — MSE meldet `EVENT_SLICE_CHANGED`, und das
	// Fenster ruft `constrain()`. Ein Zug ohne sie ließe den Rahmen aus dem
	// Bild laufen.
	return Zuschnittwahl{
		Quelle: w.Quelle, Ziel: w.Ziel, Auswahl: gezogen, Frei: w.Frei,
	}.Begrenze(BevorzugtKeines)
}
