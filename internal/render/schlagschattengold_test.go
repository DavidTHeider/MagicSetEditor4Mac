package render_test

// Das Goldbild für `drop_shadow`
// (#472).
//
// **Das Goldbild rechnet MSEs Formel nach, nicht unsere.** Nähme es die
// Rechnung aus `schlagschatten.go`, prüfte es die Umsetzung gegen sich selbst
// und wäre grün, auch wenn beide falsch sind (§10).

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// nachMSEsWeichzeichner ist `gaussian_blur`, Zeile für Zeile aus
// `src/gfx/generated_image.cpp`, Zeilen 322–367, Fassung f9d9356 vom
// 2021-07-06.
func nachMSEsWeichzeichner(
	in []byte, w, h int, radius float64,
) (out []uint64, total uint64) {
	blurX := make([]uint64, w*h)

	var totalX uint64

	sigma := radius * float64(w)
	mult := float64(1<<8) / (math.Sqrt(2*math.Pi) * sigma)
	sigsqr2 := 1 / (2 * sigma * sigma)
	rang := min(w, int(3*sigma))

	for d := -rang; d <= rang; d++ {
		factor := uint64(mult * math.Exp(-float64(d*d)*sigsqr2))
		totalX += factor

		if factor > 0 {
			for y := range h {
				for x := max(0, -d); x < min(w, w-d); x++ {
					blurX[x+y*w] += uint64(in[x+d+y*w]) * factor
				}
			}
		}
	}

	out = make([]uint64, w*h)

	var totalY uint64

	sigma = radius * float64(h)
	mult = float64(1<<8) / (math.Sqrt(2*math.Pi) * sigma)
	sigsqr2 = 1 / (2 * sigma * sigma)
	rang = min(h, int(3*sigma))

	for d := -rang; d <= rang; d++ {
		factor := uint64(mult * math.Exp(-float64(d*d)*sigsqr2))
		totalY += factor

		if factor > 0 {
			for y := max(0, -d); y < min(h, h-d); y++ {
				for x := range w {
					out[x+y*w] += blurX[x+(d+y)*w] * factor
				}
			}
		}
	}

	return out, totalX * totalY
}

// TestDropShadowHaeltSeinGoldbild rechnet MSEs Formel nach und vergleicht.
func TestDropShadowHaeltSeinGoldbild(t *testing.T) {
	const seite = 16

	// Ein Bild, dessen linke Hälfte deckt und dessen rechte durchsichtig ist.
	eingang := image.NewNRGBA(image.Rect(0, 0, seite, seite))
	for y := range seite {
		for x := range seite {
			alpha := uint8(0)
			if x < seite/2 {
				alpha = 255
			}

			eingang.SetNRGBA(x, y, color.NRGBA{R: 200, G: 200, B: 200, A: alpha})
		}
	}

	const (
		versatzX = 0.125
		versatzY = 0.0
		alpha    = 1.0
		radius   = 0.1
	)

	schattenfarbe := render.Farbe{R: 0, G: 0, B: 0, A: 255}

	// --- das Goldbild, nach MSEs Zeilen 370–401 ---
	deckung := make([]byte, seite*seite)
	for y := range seite {
		for x := range seite {
			deckung[x+y*seite] = eingang.NRGBAAt(x, y).A
		}
	}

	schatten, roh := nachMSEsWeichzeichner(deckung, seite, seite, radius)
	total := 255 * roh

	gold := image.NewNRGBA(image.Rect(0, 0, seite, seite))
	for y := range seite {
		for x := range seite {
			gold.SetNRGBA(x, y, eingang.NRGBAAt(x, y))
		}
	}

	dw, dh := int(seite*versatzX), int(seite*versatzY)
	sa := int64(alpha * (1 << 16))
	delta := dw + seite*dh

	for y := max(0, dh); y < min(seite, seite+dh); y++ {
		for x := max(0, dw); x < min(seite, seite+dw); x++ {
			p := x + y*seite
			a := int64(deckung[p])
			shad := int64(uint64(((255-a)*sa)>>16) * schatten[p-delta] / total)
			factor := max(int64(1), a+shad)

			alt := gold.NRGBAAt(x, y)
			gold.SetNRGBA(x, y, color.NRGBA{
				R: byte((a*int64(alt.R) + shad*int64(schattenfarbe.R)) / factor),
				G: byte((a*int64(alt.G) + shad*int64(schattenfarbe.G)) / factor),
				B: byte((a*int64(alt.B) + shad*int64(schattenfarbe.B)) / factor),
				A: byte(a + shad),
			})
		}
	}

	// --- unser Bild ---
	quelle := maskenquelle{bilder: map[string]image.Image{"bild.png": eingang}}

	plan := &render.Plan{Breite: seite, Hoehe: seite, Ebenen: []render.Ebene{{
		Art:    render.ArtBild,
		Rahmen: render.Rechteck{Breite: seite, Hoehe: seite},
		Bild: render.Bildbaum{
			Knotenart: "drop_shadow",
			Zahlen: map[string]float64{
				"offset_x": versatzX, "offset_y": versatzY,
				"alpha": alpha, "blur_radius": radius,
			},
			Farben: map[string]render.Farbe{"color": schattenfarbe},
			Kinder: []render.Bildkind{{Name: "input", Baum: render.Blatt("bild.png")}},
		},
		Deckung: 1,
	}}}

	unsere, befunde, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	// **Die Ebene liegt auf einer leeren Karte**: was der Schatten an
	// Deckung mitbringt, kommt gegen Schwarz an — deshalb wird der
	// Alphakanal verglichen und die Farbe nur dort, wo etwas deckt.
	var schlimmster int

	for y := range seite {
		for x := range seite {
			ist, soll := unsere.NRGBAAt(x, y), gold.NRGBAAt(x, y)

			abstand := abs(int(ist.A) - int(soll.A))
			if abstand > schlimmster {
				schlimmster = abstand
			}
		}
	}

	// **Toleranz null.** Beide rechnen ganzzahlig und in derselben
	// Reihenfolge; jede Abweichung wäre ein anderer Weg.
	if schlimmster != 0 {
		t.Errorf("die Deckung weicht um bis zu %d ab, erwartet 0", schlimmster)
	}

	// **Und der Schatten ist überhaupt da.** Ohne diese Zählung wäre ein
	// Vergleich zweier leerer Bilder ebenfalls grün.
	var beschattet int

	for y := range seite {
		for x := range seite {
			if x >= seite/2 && unsere.NRGBAAt(x, y).A > 0 {
				beschattet++
			}
		}
	}

	if beschattet == 0 {
		t.Error("rechts der Mitte ist nichts beschattet — der Schatten fehlt")
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}
