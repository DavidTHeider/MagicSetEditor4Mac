package service_test

// **Eine Vorlage im anderen Seitenverhältnis füllt genau die Karte** — I04
// (#2262),
// Befund des Zweitblicks J05 (#2269).
//
// Die Zusage des Bogens steht an [service.Bogenleger.Lege]: das Bild wird auf
// das Kartenmaß gezogen, **gleich welches Seitenverhältnis der Plan hat**.
// Bis dahin rasterte der Weg proportional zur Kartenbreite und ließ
// `mitAnschnitt` die Höhe auffüllen — bei einer quadratischen Vorlage für eine
// Karte von 63 × 88 mm und 3 mm Anschnitt standen oben und unten je 15,5 mm
// **wiederholter Randpunkt** statt Karteninhalt, und ohne Anschnitt streckte
// erst der PDF-Leser ein Bild, dessen Bildpunkte nicht mehr zur Auflösung
// passten.
//
// Gemessen wird am geschriebenen PDF: die Bildpunkte werden aus dem Strom
// gelesen, und Inhaltsmarken in der Vorlage — vier Farbbänder, je ein
// Zwanzigstel der Vorlage breit — müssen nach dem Legen **an der Schnittkante**
// beginnen und nicht daneben.
//
// # Was er nicht belegt
//
// **Die Güte der Streckung.** Dass ein gestrecktes Bild unscharf wird, ist
// der Preis der Zusage und steht als Befund daneben; hier steht nur, wo der
// Inhalt liegt.

import (
	"bytes"
	"compress/zlib"
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// anteilBand ist die Dicke eines Farbbands: ein Zwanzigstel der Vorlage.
const anteilBand = 0.05

// Die fünf Farben der Vorlage — weit genug auseinander, dass ein Mischpunkt an
// einer Kante nie für eine dritte Farbe gehalten wird.
var (
	farbeOben   = render.Farbe{R: 220, G: 30, B: 30, A: 255}
	farbeUnten  = render.Farbe{R: 30, G: 30, B: 220, A: 255}
	farbeLinks  = render.Farbe{R: 30, G: 200, B: 30, A: 255}
	farbeRechts = render.Farbe{R: 230, G: 220, B: 30, A: 255}
	farbeGrund  = render.Farbe{R: 128, G: 128, B: 128, A: 255}
)

// markenPlan ist eine Vorlage mit Farbbändern an allen vier Rändern.
//
// **Oben und unten über die ganze Breite, links und rechts dazwischen**: die
// Bänder überlappen sich nicht, und jede Kante hat genau eine Farbe.
type markenPlan struct{ breite, hoehe float64 }

func (p markenPlan) Kartenplan(context.Context, *domain.CardRow) (*render.Plan, []domain.Befund, error) {
	band := func(links, oben, breite, hoehe float64, farbe render.Farbe) render.Ebene {
		return render.Ebene{
			Art:     render.ArtFlaeche,
			Rahmen:  render.Rechteck{Links: links, Oben: oben, Breite: breite, Hoehe: hoehe},
			Farbe:   farbe,
			Deckung: 1,
		}
	}

	dick, hoch := p.breite*anteilBand, p.hoehe*anteilBand

	return &render.Plan{
		Breite: p.breite, Hoehe: p.hoehe,
		Ebenen: []render.Ebene{
			band(0, 0, p.breite, p.hoehe, farbeGrund),
			band(0, 0, p.breite, hoch, farbeOben),
			band(0, p.hoehe-hoch, p.breite, hoch, farbeUnten),
			band(0, hoch, dick, p.hoehe-2*hoch, farbeLinks),
			band(p.breite-dick, hoch, dick, p.hoehe-2*hoch, farbeRechts),
		},
	}, nil, nil
}

// rohbild ist ein Bild aus dem PDF: Bildpunkte als RGB, zeilenweise von oben.
type rohbild struct {
	breit, hoch int
	rgb         []byte
}

var bildstrom = regexp.MustCompile(
	`<< /Type /XObject /Subtype /Image /Width (\d+) /Height (\d+)` +
		` /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /FlateDecode /Length (\d+) >>\nstream\n`)

// bildpunkte liest die Bilder eines Bogens in der Reihenfolge, in der sie im
// Dokument stehen — Seite für Seite, Karte für Karte.
func bildpunkte(roh []byte) ([]rohbild, error) {
	var bilder []rohbild

	for _, stelle := range bildstrom.FindAllSubmatchIndex(roh, -1) {
		zahl := func(nummer int) int {
			wert, _ := strconv.Atoi(string(roh[stelle[2*nummer]:stelle[2*nummer+1]])) //nolint:errcheck // das Muster lässt nur Ziffern zu

			return wert
		}

		breit, hoch, laenge := zahl(1), zahl(2), zahl(3)
		ab := stelle[1]

		leser, fehler := zlib.NewReader(bytes.NewReader(roh[ab : ab+laenge]))
		if fehler != nil {
			return nil, fmt.Errorf("Bild %d: %w", len(bilder)+1, fehler)
		}

		entpackt, fehler := io.ReadAll(leser)
		if fehler != nil {
			return nil, fmt.Errorf("Bild %d: %w", len(bilder)+1, fehler)
		}

		if len(entpackt) != breit*hoch*3 {
			return nil, fmt.Errorf("Bild %d: %d Bytes für %d × %d", len(bilder)+1, len(entpackt), breit, hoch)
		}

		bilder = append(bilder, rohbild{breit: breit, hoch: hoch, rgb: entpackt})
	}

	return bilder, nil
}

// farbe gibt einen Bildpunkt.
func (b rohbild) farbe(x, y int) render.Farbe {
	ab := (y*b.breit + x) * 3

	return render.Farbe{R: b.rgb[ab], G: b.rgb[ab+1], B: b.rgb[ab+2], A: 255}
}

// aehnlich sagt, ob zwei Farben dieselbe meinen — mit Spielraum für einen
// Mischpunkt an der Kante.
func aehnlich(a, b render.Farbe) bool {
	abstand := func(x, y uint8) int { return int(math.Abs(float64(x) - float64(y))) }

	return abstand(a.R, b.R)+abstand(a.G, b.G)+abstand(a.B, b.B) < 90
}

// wechsel gibt den ersten Bildpunkt entlang einer Linie, dessen Farbe nicht
// mehr die des Anfangs ist.
func (b rohbild) wechsel(vonX, vonY, schrittX, schrittY, anzahl int) (int, error) {
	anfang := b.farbe(vonX, vonY)

	for nummer := range anzahl {
		if !aehnlich(b.farbe(vonX+nummer*schrittX, vonY+nummer*schrittY), anfang) {
			return nummer, nil
		}
	}

	return 0, fmt.Errorf("von (%d, %d) ändert sich die Farbe %v nicht", vonX, vonY, anfang)
}

// bogenMitVorlage legt eine Karte einer Vorlage auf einen Bogen und gibt den
// Inhalt des PDFs.
func bogenMitVorlage(t *testing.T, masse service.Bogenmasse, plan markenPlan) []byte {
	t.Helper()

	leger := service.NeuerBogenleger(kartenDouble{leserDouble: karten("Probe")}, plan,
		render.NeuerRasterer(nil))
	ziel := filepath.Join(t.TempDir(), "bogen.pdf")

	if _, fehler := leger.Lege(t.Context(), 7, ziel, masse, domain.Kartenauswahl{}); fehler != nil {
		t.Fatalf("legen: %v", fehler)
	}

	roh, fehler := os.ReadFile(ziel) // #nosec G304 -- aus t.TempDir()
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	return roh
}

// **Der Inhalt liegt in der Schnittkante, der Anschnitt nur außerhalb.**
//
// Vier Wechsel entlang der Mittellinien des Bildes: oben und unten (senkrecht
// durch die Mitte) und links und rechts (waagerecht durch die Mitte). Jeder
// muss um die Dicke seines Bandes — ein Zwanzigstel der **Karte** — hinter der
// Schnittkante liegen. Eine Vorlage im falschen Verhältnis schöbe sie um
// Zentimeter, weil der Anschnitt dann aus Karteninhalt bestünde.
func TestEineVorlageImAndernVerhaeltnisFuelltGenauDieKarte(t *testing.T) {
	t.Parallel()

	vorlagen := []struct {
		name          string
		breite, hoehe float64
	}{
		{"quadratisch", 630, 630},
		{"zu hoch", 630, 1000},
		{"zu breit", 880, 630},
	}

	for _, vorlage := range vorlagen {
		for _, anschnitt := range []float64{0, 3} {
			t.Run(fmt.Sprintf("%s, %g mm Anschnitt", vorlage.name, anschnitt), func(t *testing.T) {
				t.Parallel()

				masse := service.StandardBogen()
				masse.Anschnitt = anschnitt

				roh := bogenMitVorlage(t, masse, markenPlan{vorlage.breite, vorlage.hoehe})

				seiten, fehler := vermiss(roh)
				if fehler != nil {
					t.Fatalf("vermessen: %v", fehler)
				}

				for _, befund := range massbefunde(seiten, masse, 1) {
					t.Error(befund)
				}

				bilder, fehler := bildpunkte(roh)
				if fehler != nil || len(bilder) != 1 {
					t.Fatalf("%d Bilder (%v), erwartet eines", len(bilder), fehler)
				}

				bild := bilder[0]
				// **Zwei Bildpunkte Spielraum**: die Kante eines Bandes liegt
				// in der Rasterung auf einem Bildpunkt, nicht zwischen zweien.
				spielraum := 2 * 25.4 / float64(masse.DPI)
				mmJeX := seiten[0].bilder[0].breite / float64(bild.breit)
				mmJeY := seiten[0].bilder[0].hoehe / float64(bild.hoch)

				erwartet := []struct {
					was                string
					vonX, vonY         int
					schrittX, schrittY int
					anzahl             int
					mm, mmJe           float64
				}{
					{
						"Oberkante des oberen Bands", bild.breit / 2, 0, 0, 1, bild.hoch,
						anschnitt + anteilBand*masse.Kartenhoehe, mmJeY,
					},
					{
						"Oberkante des unteren Bands", bild.breit / 2, bild.hoch - 1, 0, -1, bild.hoch,
						anschnitt + anteilBand*masse.Kartenhoehe, mmJeY,
					},
					{
						"Innenkante des linken Bands", 0, bild.hoch / 2, 1, 0, bild.breit,
						anschnitt + anteilBand*masse.Kartenbreite, mmJeX,
					},
					{
						"Innenkante des rechten Bands", bild.breit - 1, bild.hoch / 2, -1, 0, bild.breit,
						anschnitt + anteilBand*masse.Kartenbreite, mmJeX,
					},
				}

				for _, wechsel := range erwartet {
					punkte, fehler := bild.wechsel(wechsel.vonX, wechsel.vonY, wechsel.schrittX, wechsel.schrittY, wechsel.anzahl)
					if fehler != nil {
						t.Errorf("%s: %v", wechsel.was, fehler)

						continue
					}

					if ist := float64(punkte) * wechsel.mmJe; math.Abs(ist-wechsel.mm) > spielraum {
						t.Errorf("%s liegt %.2f mm von der Bildkante, erwartet %.2f mm (Spielraum %.2f)",
							wechsel.was, ist, wechsel.mm, spielraum)
					}
				}
			})
		}
	}
}
