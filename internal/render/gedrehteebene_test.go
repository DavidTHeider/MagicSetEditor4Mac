package render_test

// **Ein gedrehtes Feld wird gedreht gezeichnet**
// (#486).
//
// MSE gibt jedem Feld einen eigenen gedrehten Zeichenkontext:
//
//	Rotation ValueViewer::getRotation() const {
//	  return Rotation(deg_to_rad(getStyle()->angle), bounding_box, 1.0, getStretch());
//	}
//
// `src/render/value/viewer.cpp`, Zeilen 44–46; `bounding_box` ist dabei
// `RealRect(left, top, width, height)` des Feldstils (`data/field.hpp:124`).
// Die Umrechnung steht in `src/util/rotation.cpp`, Zeilen 37–42:
//
//	RealPoint Rotation::tr(const RealPoint& p) const {
//	  double s = sin(angle), c = cos(angle);
//	  double x = p.x * zoomX, y = p.y * zoomY;
//	  return RealPoint(c * x + s * y + origin.x,
//	                  -s * x + c * y + origin.y);
//	}
//
// Fassung f9d9356 vom 2021-07-06.
//
// **Was das für die geraden Winkel heißt** — `origin` ist (`links`, `oben`),
// (`bx`, `by`) ein Punkt im ungedrehten Feld:
//
//	  0°  → (links + bx, oben + by)
//	 90°  → (links + by, oben − bx)
//	180°  → (links − bx, oben − by)
//	270°  → (links − by, oben + bx)
//
// **Das Feld verlässt dabei seinen eigenen Kasten**, und das ist so gewollt:
// `name 2` von `magic-m15-aftermath` steht bei links 342, oben 295, breit 181,
// hoch 21 und `angle: 270` — gezeichnet wird damit die Spalte x 321…342,
// y 295…476, also senkrecht am rechten Rand der unteren Hälfte.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestEineGedrehteEbeneWirdGedrehtGezeichnet(t *testing.T) {
	// Eine Fläche von 3 × 1 an (5, 5). Ungedreht deckt sie (5,5)…(7,5).
	const links, oben, breite, hoehe = 5, 5, 3, 1

	faelle := []struct {
		name   string
		winkel float64
		punkte [][2]int
	}{
		{"ohne Drehung", 0, [][2]int{{5, 5}, {6, 5}, {7, 5}}},
		{"90 Grad", 90, [][2]int{{5, 5}, {5, 4}, {5, 3}}},
		{"180 Grad", 180, [][2]int{{5, 5}, {4, 5}, {3, 5}}},
		{"270 Grad", 270, [][2]int{{5, 5}, {5, 6}, {5, 7}}},

		// **Ein voller Umlauf ist keine Drehung** — und geht trotzdem durch
		// den gedrehten Weg, denn `Winkel != 0`. Ein gerechneter Winkel darf
		// 360 heißen.
		{"360 Grad", 360, [][2]int{{5, 5}, {6, 5}, {7, 5}}},

		// **Und ein negativer zählt rückwärts**: −90 ist 270.
		{"minus 90 Grad", -90, [][2]int{{5, 5}, {5, 6}, {5, 7}}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			plan := &render.Plan{Breite: 12, Hoehe: 12, Ebenen: []render.Ebene{{
				Art:    render.ArtFlaeche,
				Rahmen: render.Rechteck{Links: links, Oben: oben, Breite: breite, Hoehe: hoehe},
				Farbe:  render.Farbe{R: 255, A: 255},
				Winkel: fall.winkel, Deckung: 1,
			}}}

			bild, befunde, err := render.NeuerRasterer(bildmasse{}).
				Zeichne(t.Context(), plan, 1, render.GueteMittel)
			if err != nil {
				t.Fatalf("zeichnen: %v", err)
			}

			if len(befunde) != 0 {
				t.Fatalf("Befunde: %v", befunde)
			}

			for _, punkt := range fall.punkte {
				if farbe := bild.NRGBAAt(punkt[0], punkt[1]); farbe.R != 255 || farbe.A != 255 {
					t.Errorf("(%d,%d) ist %v, erwartet rot und deckend",
						punkt[0], punkt[1], farbe)
				}
			}

			// **Und sonst nirgends.** Ohne diese Zählung wäre eine Ebene, die
			// alles einfärbt, ebenfalls grün.
			var gemalt int

			for y := range 12 {
				for x := range 12 {
					if bild.NRGBAAt(x, y).A > 0 {
						gemalt++
					}
				}
			}

			if gemalt != len(fall.punkte) {
				t.Errorf("%d Bildpunkte gemalt, erwartet %d", gemalt, len(fall.punkte))
			}
		})
	}
}

// **Ein schräger Winkel wird benannt und ungedreht gezeichnet.**
//
// MSE unterscheidet selbst zwischen geraden und schrägen Winkeln
// (`is_straight(angle)`, `src/util/rotation.cpp`, Fassung f9d9356 vom
// 2021-07-06). Die schrägen sind hier
// nicht gebaut; im vollen Pack stehen sie viermal (`41`, `3.5`, `0.65`) gegen
// 142 gerade.
func TestEinSchraegerWinkelWirdBenannt(t *testing.T) {
	plan := &render.Plan{Breite: 12, Hoehe: 12, Ebenen: []render.Ebene{{
		Art:    render.ArtFlaeche,
		Rahmen: render.Rechteck{Links: 5, Oben: 5, Breite: 3, Hoehe: 1},
		Farbe:  render.Farbe{R: 255, A: 255},
		Winkel: 41, Deckung: 1,
	}}}

	bild, befunde, err := render.NeuerRasterer(bildmasse{}).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet einer: %v", len(befunde), befunde)
	}

	// **Und gezeichnet ist trotzdem worden** — ungedreht (§8).
	if farbe := bild.NRGBAAt(6, 5); farbe.R != 255 {
		t.Errorf("(6,5) ist %v, erwartet rot — die Ebene fehlt ganz", farbe)
	}
}

// **Was über den Rand hinausragt, fällt weg.**
//
// Eine gedrehte Ebene verlässt ihren eigenen Kasten; sie darf dabei auch die
// Karte verlassen. Was draußen liegt, wird nicht gezeichnet — und nicht
// gemeldet, denn es ist kein Mangel, sondern Absicht der Vorlage.
func TestEineGedrehteEbeneRagtNichtUeberDenRand(t *testing.T) {
	// Bei 180 Grad läuft die Ebene von (1,1) nach links und oben — also
	// größtenteils hinaus.
	plan := &render.Plan{Breite: 12, Hoehe: 12, Ebenen: []render.Ebene{{
		Art:    render.ArtFlaeche,
		Rahmen: render.Rechteck{Links: 1, Oben: 1, Breite: 5, Hoehe: 1},
		Farbe:  render.Farbe{R: 255, A: 255},
		Winkel: 180, Deckung: 1,
	}}}

	bild, befunde, err := render.NeuerRasterer(bildmasse{}).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	var gemalt int

	for y := range 12 {
		for x := range 12 {
			if bild.NRGBAAt(x, y).A > 0 {
				gemalt++
			}
		}
	}

	// (1,1), (0,1) — der Rest liegt bei x < 0.
	if gemalt != 2 {
		t.Errorf("%d Bildpunkte gemalt, erwartet 2", gemalt)
	}
}

// **Eine Ebene, die auf null Bildpunkte zusammenfällt, zeichnet nichts.**
//
// Sie ist nach [sichtbar] noch sichtbar — ihre Maße sind größer als null —,
// aber bei dieser Vergrößerung bleibt kein ganzer Bildpunkt übrig.
func TestEineGedrehteEbeneOhneFlaecheZeichnetNichts(t *testing.T) {
	plan := &render.Plan{Breite: 12, Hoehe: 12, Ebenen: []render.Ebene{{
		Art:    render.ArtFlaeche,
		Rahmen: render.Rechteck{Links: 5, Oben: 5, Breite: 0.4, Hoehe: 0.4},
		Farbe:  render.Farbe{R: 255, A: 255},
		Winkel: 90, Deckung: 1,
	}}}

	bild, befunde, err := render.NeuerRasterer(bildmasse{}).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	for y := range 12 {
		for x := range 12 {
			if bild.NRGBAAt(x, y).A > 0 {
				t.Fatalf("(%d,%d) ist gemalt", x, y)
			}
		}
	}
}

// **Ein gedrehter Textzug läuft senkrecht.**
//
// Das ist der Fall, um den es bei `magic-m15-aftermath` geht: Name, Typzeile
// und Regeltext der unteren Hälfte stehen dort gedreht. Ein Textzug lässt den
// größten Teil seiner Fläche **durchsichtig** — und die bleibt es auch nach
// der Drehung.
func TestEinGedrehterTextzugLaeuftSenkrecht(t *testing.T) {
	// Das Feld ist 6 breit und 2 hoch, gedreht um 270 Grad an (8, 2).
	// Gezeichnet wird damit die Spalte x 7…8, y 2…7.
	plan := &render.Plan{Breite: 12, Hoehe: 12, Ebenen: []render.Ebene{{
		Art:        render.ArtText,
		Rahmen:     render.Rechteck{Links: 8, Oben: 2, Breite: 6, Hoehe: 2},
		Abschnitte: []render.Abschnitt{{Text: "ab"}},
		Satz: render.Satz{
			Schrift: "Probe", Groesse: 2, Ausrichtung: render.LinksOben,
		},
		Farbe:  render.Farbe{R: 255, A: 255},
		Winkel: 270, Deckung: 1,
	}}}

	bild, befunde, err := render.NeuerRasterer(bildmasse{}).
		MitSchriften(schriften{"Probe": einheitsschrift{}}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	var links, rechts, oben, unten int

	erster := true

	for y := range 12 {
		for x := range 12 {
			if bild.NRGBAAt(x, y).A == 0 {
				continue
			}

			if erster {
				links, rechts, oben, unten = x, x, y, y
				erster = false
			}

			links, rechts = min(links, x), max(rechts, x)
			oben, unten = min(oben, y), max(unten, y)
		}
	}

	if erster {
		t.Fatal("nichts gezeichnet")
	}

	// **Senkrecht heißt: höher als breit.** Ungedreht wäre der Zug breiter
	// als hoch, denn er läuft über die Zeile.
	if rechts-links >= unten-oben {
		t.Errorf("der Zug ist %d breit und %d hoch — er läuft nicht senkrecht",
			rechts-links+1, unten-oben+1)
	}

	// **Und er steht links von seinem Ursprung**, wie `tr` es für 270 Grad
	// sagt: (links − by, oben + bx).
	if rechts > 8 {
		t.Errorf("der Zug reicht bis x=%d, erwartet höchstens 8", rechts)
	}
}
