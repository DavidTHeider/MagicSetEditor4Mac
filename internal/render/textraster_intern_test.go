package render

import (
	"image"
	"slices"
	"testing"
)

// **Das Herunterrechnen, wie MSE es rechnet** — `downsample_to_alpha`,
// `src/gfx/resample_text.cpp`, Zeilen 26–140, Fassung f9d9356 vom 2021-07-06.
//
// Die erwarteten Zahlen sind von Hand aus MSEs Rechnung gewonnen, nicht aus
// dieser — mit ganzzahliger Teilung an jeder Stelle, an der MSE teilt.
func TestDasHerunterrechnenRechnetWieMSE(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name                         string
		breite, hoehe, ausgabebreite int
		vierfach                     []uint8
		erwartet                     []uint8
	}{
		{
			name:   "ein voller Block bleibt voll",
			breite: 1, hoehe: 1, ausgabebreite: 1,
			vierfach: wiederholt([]uint8{255, 255, 255, 255}, 4),
			erwartet: []uint8{255},
		},
		{
			name:   "ein halber Block wird halb, abgeschnitten",
			breite: 1, hoehe: 1, ausgabebreite: 1,
			vierfach: wiederholt([]uint8{255, 255, 0, 0}, 4),
			erwartet: []uint8{127},
		},
		{
			// Zeilen 3/4, 3/4, 3/4 und 7/4 geben 0, 0, 0, 1 — und 1/4 gibt 0.
			// Einmal durch sechzehn gäbe 16/16 = 1.
			name:   "zweimal geteilt ist nicht einmal geteilt",
			breite: 1, hoehe: 1, ausgabebreite: 1,
			vierfach: slices.Concat(
				[]uint8{3, 0, 0, 0}, []uint8{3, 0, 0, 0}, []uint8{3, 0, 0, 0}, []uint8{7, 0, 0, 0}),
			erwartet: []uint8{0},
		},
		{
			// out_fact = 4096/8 = 512, mul = 128 + min(256, 128·8/4) = 384:
			// (128 · 384) >> 8 = 192 — gestaucht wird fetter.
			name:   "gestaucht auf die Hälfte wird fetter",
			breite: 2, hoehe: 1, ausgabebreite: 1,
			vierfach: wiederholt(wiederholt([]uint8{128}, 8), 4),
			erwartet: []uint8{192},
		},
		{
			// (255 · 384) >> 8 = 382, und `top` klemmt auf 255.
			name:   "und nicht über voll",
			breite: 2, hoehe: 1, ausgabebreite: 1,
			vierfach: wiederholt(wiederholt([]uint8{255}, 8), 4),
			erwartet: []uint8{255},
		},
		{
			// out_fact = 12288/8 = 1536, mul = 128 + 1024/12 = 213; der erste
			// Ausgabepunkt nimmt 1536 Anteile von 255, dann 1536 und 1024
			// Anteile von null: 391680 >> 12 = 95, (95 · 213) >> 8 = 79.
			name:   "gedehnt teilt ein Eingabepunkt sich auf zwei",
			breite: 2, hoehe: 1, ausgabebreite: 3,
			vierfach: wiederholt([]uint8{255, 0, 0, 0, 0, 0, 0, 0}, 4),
			erwartet: []uint8{79, 0, 0},
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			got := herunterrechnen(fall.vierfach, fall.breite, fall.hoehe, fall.ausgabebreite)
			if !slices.Equal(got, fall.erwartet) {
				t.Errorf("%v, erwartet %v", got, fall.erwartet)
			}
		})
	}
}

// wiederholt hängt ein Stück n-mal aneinander.
func wiederholt(stueck []uint8, n int) []uint8 {
	aus := make([]uint8, 0, len(stueck)*n)
	for range n {
		aus = append(aus, stueck...)
	}

	return aus
}

// **Auf null gestaucht wird nichts gezeichnet — und nichts stürzt ab.**
//
// Erreichbar ist das: ein Symboltext mit `max_stretch: 0` in einem Feld der
// Breite null bekommt die Dehnung null ([eingepasst]). MSE teilt dann in
// `downsample_to_alpha` durch die Ausgabebreite null
// (`mul = 128 + min(256, 128*w1/(text_scaling*w2))`, Zeile 60, Fassung
// f9d9356 vom 2021-07-06); hier bleibt das Bild, wie es war.
func TestAufNullGestauchtWirdNichtsGezeichnet(t *testing.T) {
	t.Parallel()

	lauf := neuerVierfachlauf(2, 0, 4, 4, 3, 0)
	lauf.deckung[0] = 255

	ziel := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	(&Rasterer{}).lege(ziel, Ebene{Farbe: Farbe{A: 255}, Deckung: 1}, nil, ziel.Bounds(), lauf)

	for _, wert := range ziel.Pix {
		if wert != 0 {
			t.Fatalf("gezeichnet, obwohl die Ausgabe keine Breite hat: %v", ziel.Pix)
		}
	}
}
