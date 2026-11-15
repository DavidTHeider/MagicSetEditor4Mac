package render_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **`shrink` staucht waagerecht — die Schriftgröße bleibt.**
//
// # Der Befund
//
//	double TextStyle::getStretch() const {
//	  if (layout->width > 0 && (alignment() & ALIGN_STRETCH)) {
//	    double factor = (width - padding_left - padding_right) / layout->width;
//	    if (!(alignment() & ALIGN_IF_OVERFLOW) || factor < 1.0) return factor;
//	  }
//	  return 1.0;
//	}
//
// `src/data/field/text.cpp`, Zeilen 73–81, Fassung f9d9356 vom 2021-07-06. Der
// Faktor geht als **waagerechte Zoomstufe** in die Zeichenfläche
// (`RotatedDC`), nicht in die Schriftgröße.
//
// Bis zum 2026-08-17 verkleinerte dieses Programm die **Größe** — und damit
// auch Höhe, Grundlinie und Zeilenhöhe. Das ist ein anderes Modell, kein
// Näherungsfehler: bei MSE steht ein langer Kartenname in voller Höhe und
// schmal, bei uns kleiner und proportional. Die Angabe steht **676-mal** in den
// 190 Stilen von `M15-Magic-Pack`
// (#319).
func TestEinZuBreiterZugWirdGestauchtUndNichtVerkleinert(t *testing.T) {
	// Zehn Zeichen à 5 Punkte sind 50 breit, der Kasten ist 25: Faktor 0,5.
	gesetzt := render.SetzeAbschnitte(
		render.Satz{
			Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben,
			Stauchen: true, Einzeilig: true, NurBeiUeberlauf: true,
		},
		render.Rechteck{Breite: 25, Hoehe: 30},
		[]render.Abschnitt{{Text: "abcdefghij"}},
		render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}, nil)

	if gesetzt.Stauchung != 0.5 {
		t.Errorf("die Stauchung ist %g, erwartet 0,5", gesetzt.Stauchung)
	}

	// **Die Größe bleibt.** Sie war der Fehler: eine verkleinerte Schrift macht
	// auch die Zeile niedriger, und die Zeile trägt die Grundlinie.
	if gesetzt.Groesse != 10 {
		t.Errorf("die Schriftgröße ist %g, erwartet 10", gesetzt.Groesse)
	}

	// **Und die Breite bleibt die ungestauchte.** `content_width` fragt nach
	// dem gesetzten Zug, und aus ihr rechnet MSE den Faktor erst aus — eine
	// schon gestauchte Breite ergäbe beim nächsten Mal Faktor 1.
	if breite := gesetzt.Zeilen[0].Breite; breite != 50 {
		t.Errorf("die Zeilenbreite ist %g, erwartet 50", breite)
	}
}

// **Mit `overflow` wird nur gestaucht, nicht gedehnt.**
//
// `factor < 1.0` — ein Zug, der hineinpasst, bleibt, wie er ist. Ohne
// `overflow` würde er auf die Kastenbreite **gedehnt**; im Pack steht diese
// Form zweimal.
func TestGedehntWirdNurOhneUeberlaufbedingung(t *testing.T) {
	faelle := []struct {
		name            string
		nurBeiUeberlauf bool
		erwartet        float64
	}{
		{"mit overflow bleibt es bei eins", true, 1},
		{"ohne overflow wird gedehnt", false, 2},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			// Vier Zeichen à 5 sind 20 breit, der Kasten ist 40.
			gesetzt := render.SetzeAbschnitte(
				render.Satz{
					Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben,
					Stauchen: true, Einzeilig: true, NurBeiUeberlauf: fall.nurBeiUeberlauf,
				},
				render.Rechteck{Breite: 40, Hoehe: 30},
				[]render.Abschnitt{{Text: "abcd"}},
				render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}, nil)

			if gesetzt.Stauchung != fall.erwartet {
				t.Errorf("die Stauchung ist %g, erwartet %g",
					gesetzt.Stauchung, fall.erwartet)
			}
		})
	}
}

// **Ein gestauchter Zug wird nicht mehr ausgerichtet.**
//
// MSE lässt im Stretch-Zweig die Stellen stehen, wie sie sind — kein
// `align_delta_x` (`src/render/text/viewer.cpp`, Zeilen 863–865). Er beginnt
// also an der linken Kante des Kastens. Passt er dagegen hinein, greift die
// gewöhnliche Ausrichtung: 180 Feldstile im Pack tragen
// `center middle shrink-overflow`, und die stehen mittig, solange sie nicht
// überlaufen.
//
// Fassung f9d9356 vom 2021-07-06.
func TestEinGestauchterZugStehtLinks(t *testing.T) {
	satz := render.Satz{
		Schrift: "Probe", Groesse: 10, Ausrichtung: render.MitteOben,
		Stauchen: true, Einzeilig: true, NurBeiUeberlauf: true,
	}

	eng := render.SetzeAbschnitte(satz, render.Rechteck{Links: 100, Breite: 25, Hoehe: 30},
		[]render.Abschnitt{{Text: "abcdefghij"}},
		render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}, nil)

	if links := eng.Zeilen[0].Links; links != 100 {
		t.Errorf("der gestauchte Zug beginnt bei %g, erwartet 100", links)
	}

	weit := render.SetzeAbschnitte(satz, render.Rechteck{Links: 100, Breite: 100, Hoehe: 30},
		[]render.Abschnitt{{Text: "abcd"}},
		render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}, nil)

	// 20 breit in einem Kasten von 100, mittig: 100 + (100-20)/2 = 140.
	if links := weit.Zeilen[0].Links; links != 140 {
		t.Errorf("der ungestauchte Zug beginnt bei %g, erwartet 140", links)
	}
}

// **Die senkrechte Ausrichtung überlebt die Stauchung.**
//
// Gestaucht wird waagerecht; oben, mittig und unten bleiben, wie sie sind.
// Ohne diese Probe bestünde eine Umsetzung, die jede Ausrichtung auf „links
// oben" wirft.
func TestDieSenkrechteAusrichtungUeberlebtDieStauchung(t *testing.T) {
	faelle := []struct {
		name        string
		ausrichtung render.Ausrichtung
		oben        float64
	}{
		// Die Zeile ist 12 hoch (Größe 10 × Zeilenhöhe 1,2), der Kasten 30.
		{"oben", render.RechtsOben, 0},
		{"mittig", render.MitteMitte, 9},
		{"unten", render.LinksUnten, 18},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			gesetzt := render.SetzeAbschnitte(
				render.Satz{
					Schrift: "Probe", Groesse: 10, Ausrichtung: fall.ausrichtung,
					Stauchen: true, Einzeilig: true, NurBeiUeberlauf: true,
				},
				render.Rechteck{Breite: 25, Hoehe: 30},
				[]render.Abschnitt{{Text: "abcdefghij"}},
				render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

			if gesetzt.Zeilen[0].Links != 0 {
				t.Errorf("die Zeile beginnt bei %g, erwartet 0 — gestaucht heißt links",
					gesetzt.Zeilen[0].Links)
			}

			if oben := gesetzt.Zeilen[0].Oben; oben != fall.oben {
				t.Errorf("die Zeile sitzt bei %g, erwartet %g", oben, fall.oben)
			}
		})
	}
}
