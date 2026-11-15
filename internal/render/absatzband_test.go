package render_test

// **Ein fester Absatz bekommt sein eigenes Band**
// (#479).
//
// `paragraph height: 49` steht in `magic-m15-planeswalker.mse-style` und ist
// das, was die drei Fähigkeiten einer Planeswalkerkarte **neben** ihre
// Loyalitätssymbole setzt. MSE richtet dann jeden Absatz für sich aus, in
// einem Rechteck fester Höhe:
//
//	alignParagraph(start, last + 1, chars, style,
//	               RealRect(0, style.padding_top + n*style.paragraph_height,
//	                        s.width, style.paragraph_height));
//
// `src/render/text/viewer.cpp`, Zeilen 749–764, Fassung f9d9356 vom
// 2021-07-06. Die Absatzgrenze ist dort `break_after != SOFT` (Zeile 758).

import (
	"math"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestEinFesterAbsatzBekommtSeinBand(t *testing.T) {
	const bandhoehe = 49

	faelle := []struct {
		name        string
		ausrichtung render.Ausrichtung
		erwartet    []float64
	}{
		{
			// **Oben ausgerichtet**: jeder Absatz beginnt an seinem Band.
			name: "oben", ausrichtung: render.LinksOben,
			erwartet: []float64{0, bandhoehe, 2 * bandhoehe},
		},
		{
			// **In der Mitte**: eine Zeile von 12 Punkten sitzt in einem Band
			// von 49 um (49−12)/2 = 18,5 tiefer.
			name: "mittig", ausrichtung: render.LinksMitte,
			erwartet: []float64{18.5, bandhoehe + 18.5, 2*bandhoehe + 18.5},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			gesetzt := render.SetzeAbschnitte(
				render.Satz{
					Schrift: "Probe", Groesse: 10,
					Ausrichtung: fall.ausrichtung, Absatzhoehe: bandhoehe,
				},
				render.Rechteck{Breite: 200, Hoehe: 3 * bandhoehe},
				[]render.Abschnitt{{Text: "eins\nzwei\ndrei"}},
				render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

			if len(gesetzt.Zeilen) != 3 {
				t.Fatalf("%d Zeilen, erwartet drei: %v", len(gesetzt.Zeilen), gesetzt.Zeilen)
			}

			for i, zeile := range gesetzt.Zeilen {
				if math.Abs(zeile.Oben-fall.erwartet[i]) > 0.01 {
					t.Errorf("Zeile %d beginnt bei %.2f, erwartet %.2f",
						i+1, zeile.Oben, fall.erwartet[i])
				}
			}
		})
	}
}

// **Ohne Angabe ändert sich nichts** — die Gegenprobe.
//
// `paragraph_height <= 0` heißt bei MSE: der ganze Kasten wird als eines
// ausgerichtet (`viewer.cpp:749`). Ohne diesen Fall wäre nicht belegt, dass
// die Bandrechnung nur dort greift, wo sie hingehört.
func TestOhneAbsatzhoeheAendertSichNichts(t *testing.T) {
	setze := func(hoehe float64) []render.Zeile {
		return render.SetzeAbschnitte(
			render.Satz{
				Schrift: "Probe", Groesse: 10,
				Ausrichtung: render.LinksOben, Absatzhoehe: hoehe,
			},
			render.Rechteck{Breite: 200, Hoehe: 147},
			[]render.Abschnitt{{Text: "eins\nzwei\ndrei"}},
			render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil).Zeilen
	}

	ohne, null := setze(0), setze(-1)

	if len(ohne) != len(null) {
		t.Fatalf("%d gegen %d Zeilen", len(ohne), len(null))
	}

	for i := range ohne {
		if ohne[i].Oben != null[i].Oben {
			t.Errorf("Zeile %d: %.2f ohne Angabe, %.2f bei −1",
				i+1, ohne[i].Oben, null[i].Oben)
		}
	}

	// Die Zeilen stehen dicht, nicht in Bändern von 49.
	if len(ohne) > 1 && ohne[1].Oben > 30 {
		t.Errorf("die zweite Zeile beginnt bei %.2f — das sieht nach einem "+
			"Band aus, und ohne Angabe darf es keines geben", ohne[1].Oben)
	}
}

// **Ein umgebrochener Absatz bleibt ein Absatz.**
//
// MSE schließt einen Absatz an `break_after != SOFT`; ein weicher Umbruch —
// die Zeile war zu lang — schließt keinen (`viewer.cpp`, Zeile 758). Beide
// Zeilen stehen deshalb im **selben** Band, und erst der harte Umbruch
// beginnt das nächste.
func TestEinUmgebrochenerAbsatzBleibtEinAbsatz(t *testing.T) {
	const bandhoehe = 49

	// Die Probeschrift ist 0,5 breit je Zeichen, die Größe 10 — also 5 Punkte
	// je Zeichen. In 200 Punkte passen 40 Zeichen.
	//
	// **Mit Leerzeichen**, denn ohne Trennstelle bricht nichts um: MSE
	// schneidet am Leerzeichen, nicht mitten im Wort.
	lang := strings.Repeat("aa ", 20)

	gesetzt := render.SetzeAbschnitte(
		render.Satz{
			Schrift: "Probe", Groesse: 10,
			Ausrichtung: render.LinksOben, Absatzhoehe: bandhoehe,
		},
		render.Rechteck{Breite: 200, Hoehe: 2 * bandhoehe},
		[]render.Abschnitt{{Text: lang + "\nkurz"}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

	if len(gesetzt.Zeilen) != 3 {
		t.Fatalf("%d Zeilen, erwartet drei (zwei umgebrochene und eine)",
			len(gesetzt.Zeilen))
	}

	// Die ersten beiden im ersten Band, dicht beieinander.
	if abstand := gesetzt.Zeilen[1].Oben - gesetzt.Zeilen[0].Oben; abstand >= bandhoehe {
		t.Errorf("die zweite Zeile steht %.2f unter der ersten — sie gehört in "+
			"dasselbe Band", abstand)
	}

	// Die dritte beginnt das zweite Band.
	if math.Abs(gesetzt.Zeilen[2].Oben-bandhoehe) > 0.01 {
		t.Errorf("die dritte Zeile beginnt bei %.2f, erwartet %d",
			gesetzt.Zeilen[2].Oben, bandhoehe)
	}
}
