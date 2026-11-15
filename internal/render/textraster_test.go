package render_test

import (
	"image"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Text steht auf Viertelpunkten, nicht auf ganzen** — QA-03.AK-02
// (#1656).
//
// MSE zeichnet ein Textstück vierfach vergrößert und rechnet es herunter
// (`draw_resampled_text`, `src/gfx/resample_text.cpp`, Zeilen 166–211,
// Fassung f9d9356 vom 2021-07-06). Die Stelle darin ist
// `xsub = int(text_scaling * (pos.x - xi))` — ein Viertelpunkt. Hier stand
// jede Glyphe auf `int(links)`: bis zu einem ganzen Bildpunkt links von ihrer
// Stelle, und das je Glyphe verschieden.
//
// Gemessen am 2026-09-23 an einer Regeltextzeile gegen Core Text auf demselben
// Weg: der Abstand fiel von 4,42 auf 1,92 — und was übrig blieb, war allein
// Quartz' Schriftglättung (`docs/SCHRIFTUNTERSUCHUNG.md`).
func TestEinTextstueckStehtAufViertelpunkten(t *testing.T) {
	plan := &render.Plan{Breite: 30, Hoehe: 12, Ebenen: []render.Ebene{{
		Art:        render.ArtText,
		Rahmen:     render.Rechteck{Links: 10.5, Breite: 19, Hoehe: 12},
		Abschnitte: []render.Abschnitt{{Text: "a"}},
		Farbe:      render.Farbe{A: 255},
		Satz:       render.Satz{Schrift: "Balken", Groesse: 8},
		Deckung:    1,
	}}}

	bild, befunde, err := render.NeuerRasterer(nil).
		MitSchriften(schriften{"Balken": balkenschrift{}}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil || len(befunde) > 0 {
		t.Fatalf("zeichnen: %v, Befunde %v", err, befunde)
	}

	// Der Balken ist vier Bildpunkte breit und beginnt bei 10,5: an den
	// Rändern steht je die Hälfte, dazwischen voll.
	erwartet := []uint8{0, 127, 255, 255, 255, 127, 0}
	zeile := deckungDerZeile(bild, 9, 16)

	for i, soll := range erwartet {
		if zeile[i] != soll {
			t.Errorf("Deckung bei x = %d: %d, erwartet %d — ganze Zeile %v", 9+i, zeile[i], soll, zeile)
		}
	}
}

// balkenschrift zeichnet jedes Zeichen als Balken halber Geviertbreite, der
// eine Geviertgröße hoch auf der Grundlinie steht.
type balkenschrift struct{}

func (balkenschrift) Breite(rune) float64  { return 0.5 }
func (balkenschrift) Zeilenhoehe() float64 { return 1.25 }
func (balkenschrift) Grundlinie() float64  { return 1 }

func (balkenschrift) Maske(_ rune, groesse float64) (image.Image, image.Point, bool) {
	breite, hoehe := max(int(groesse/2), 1), max(int(groesse), 1)

	maske := image.NewAlpha(image.Rect(0, 0, breite, hoehe))
	for i := range maske.Pix {
		maske.Pix[i] = 255
	}

	return maske, image.Point{Y: -hoehe}, true
}

// deckungDerZeile gibt die Deckung der Messzeile von x = von bis bis — der
// Zeile y = 2, die in jeder Schrift hier mitten im Zeichen liegt.
func deckungDerZeile(bild image.Image, von, bis int) []uint8 {
	const y = 2

	aus := make([]uint8, 0, bis-von)

	for x := von; x < bis; x++ {
		_, _, _, a := bild.At(x, y).RGBA()
		aus = append(aus, uint8(a>>8)) //nolint:gosec // 16 Bit auf 8
	}

	return aus
}

// **Ein Kerningpaar verschiebt sein zweites Zeichen, nicht das dritte.**
//
// Der Vorschub eines Zeichens trägt die Korrektur zu seinem Vorgänger vorn
// (`vorschub`, #752) — der Stift rückt erst um sie, dann steht die Glyphe.
// Gezeichnet wurde die Glyphe aber **vor** dem Vorschub: das Paar `AV` rückte
// nicht das `V` an das `A`, sondern das Zeichen danach. Die Breite stimmte,
// und damit fiel es keinem Breitentest auf.
func TestEinKerningpaarVerschiebtSeinZweitesZeichen(t *testing.T) {
	plan := &render.Plan{Breite: 30, Hoehe: 12, Ebenen: []render.Ebene{{
		Art:        render.ArtText,
		Rahmen:     render.Rechteck{Breite: 30, Hoehe: 12},
		Abschnitte: []render.Abschnitt{{Text: "AV"}},
		Farbe:      render.Farbe{A: 255},
		Satz:       render.Satz{Schrift: "Kernbalken", Groesse: 8},
		Deckung:    1,
	}}}

	bild, befunde, err := render.NeuerRasterer(nil).
		MitSchriften(schriften{"Kernbalken": kernbalken{}}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil || len(befunde) > 0 {
		t.Fatalf("zeichnen: %v, Befunde %v", err, befunde)
	}

	// `A` ist unsichtbar und vier Punkte breit, `V` ein Strich von einem
	// Punkt an seinem linken Rand. Das Paar rückt `V` um zwei Punkte heran:
	// der Strich steht bei x = 2, nicht bei x = 4.
	zeile := deckungDerZeile(bild, 0, 8)
	if zeile[2] != 255 || zeile[4] != 0 {
		t.Errorf("Deckung %v — der Strich des V gehört an x = 2", zeile)
	}
}

// kernbalken ist eine Schrift, in der `A` unsichtbar ist und `V` ein Strich,
// und zwischen beiden steht ein Kerningpaar von einem Viertelgeviert.
type kernbalken struct{}

func (kernbalken) Breite(rune) float64  { return 0.5 }
func (kernbalken) Zeilenhoehe() float64 { return 1.25 }
func (kernbalken) Grundlinie() float64  { return 1 }

func (kernbalken) Kernung(vorher, jetzt rune) float64 {
	if vorher == 'A' && jetzt == 'V' {
		return -0.25
	}

	return 0
}

func (kernbalken) Maske(zeichen rune, groesse float64) (image.Image, image.Point, bool) {
	if zeichen != 'V' {
		return image.NewAlpha(image.Rectangle{}), image.Point{}, true
	}

	breite, hoehe := max(int(groesse/8), 1), max(int(groesse), 1)

	maske := image.NewAlpha(image.Rect(0, 0, breite, hoehe))
	for i := range maske.Pix {
		maske.Pix[i] = 255
	}

	return maske, image.Point{Y: -hoehe}, true
}

// **Was über den Puffer hinausragt, schneidet MSE ab — und dieser Rasterer
// auch.** Der Puffer ist `int(breite) + 3` Punkte breit, „some fonts are
// larger then the GetTextExtent tells us" (`src/gfx/resample_text.cpp`,
// Zeile 169, Fassung f9d9356 vom 2021-07-06). Eine Glyphe, die weiter reicht
// als ihr Vorschub und diese drei Punkte, endet an seinem Rand.
func TestEineGlypheEndetAmRandDesPuffers(t *testing.T) {
	plan := &render.Plan{Breite: 30, Hoehe: 12, Ebenen: []render.Ebene{{
		Art:        render.ArtText,
		Rahmen:     render.Rechteck{Breite: 30, Hoehe: 12},
		Abschnitte: []render.Abschnitt{{Text: "W"}},
		Farbe:      render.Farbe{A: 255},
		Satz:       render.Satz{Schrift: "Ueberhang", Groesse: 8},
		Deckung:    1,
	}}}

	bild, befunde, err := render.NeuerRasterer(nil).
		MitSchriften(schriften{"Ueberhang": ueberhang{}}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil || len(befunde) > 0 {
		t.Fatalf("zeichnen: %v, Befunde %v", err, befunde)
	}

	// Vorschub vier Punkte, Glyphe zwanzig: der Puffer endet bei 4 + 3 = 7.
	zeile := deckungDerZeile(bild, 0, 12)
	if zeile[6] != 255 || zeile[7] != 0 {
		t.Errorf("Deckung %v — voll bis x = 6, ab x = 7 abgeschnitten", zeile)
	}
}

// ueberhang zeichnet jedes Zeichen zweieinhalb Geviert breit bei einem
// Vorschub von einem halben.
type ueberhang struct{}

func (ueberhang) Breite(rune) float64  { return 0.5 }
func (ueberhang) Zeilenhoehe() float64 { return 1.25 }
func (ueberhang) Grundlinie() float64  { return 1 }

func (ueberhang) Maske(_ rune, groesse float64) (image.Image, image.Point, bool) {
	breite, hoehe := max(int(groesse*2.5), 1), max(int(groesse), 1)

	maske := image.NewAlpha(image.Rect(0, 0, breite, hoehe))
	for i := range maske.Pix {
		maske.Pix[i] = 255
	}

	return maske, image.Point{Y: -hoehe}, true
}

// **Auch die Glyphen im Stück stehen auf Viertelpunkten**, nicht nur sein
// Anfang: MSE setzt das ganze Stück in den Vierfachpuffer, und dort steht
// jede Glyphe an ihrer Stelle.
//
// Die Gegenprobe hat ihn nachgefordert: ein Vorschub auf ganze Punkte
// abgeschnitten ließ [TestEinTextstueckStehtAufViertelpunkten] grün, denn
// dort steht nur eine Glyphe.
func TestJedeGlypheStehtAufIhrerViertelstelle(t *testing.T) {
	plan := &render.Plan{Breite: 30, Hoehe: 12, Ebenen: []render.Ebene{{
		Art:        render.ArtText,
		Rahmen:     render.Rechteck{Breite: 30, Hoehe: 12},
		Abschnitte: []render.Abschnitt{{Text: "ll"}},
		Farbe:      render.Farbe{A: 255},
		Satz:       render.Satz{Schrift: "Strich", Groesse: 8},
		Deckung:    1,
	}}}

	bild, befunde, err := render.NeuerRasterer(nil).
		MitSchriften(schriften{"Strich": strichschrift{}}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil || len(befunde) > 0 {
		t.Fatalf("zeichnen: %v, Befunde %v", err, befunde)
	}

	// Zwei Striche von einem Punkt, der zweite bei 2,5: er fällt je zur
	// Hälfte auf x = 2 und x = 3.
	erwartet := []uint8{255, 0, 127, 127, 0}
	zeile := deckungDerZeile(bild, 0, 5)

	for i, soll := range erwartet {
		if zeile[i] != soll {
			t.Errorf("Deckung bei x = %d: %d, erwartet %d — ganze Zeile %v", i, zeile[i], soll, zeile)
		}
	}
}

// strichschrift zeichnet jedes Zeichen als Strich von einem Achtelgeviert
// bei einem Vorschub von fünf Sechzehnteln.
type strichschrift struct{}

func (strichschrift) Breite(rune) float64  { return 0.3125 }
func (strichschrift) Zeilenhoehe() float64 { return 1.25 }
func (strichschrift) Grundlinie() float64  { return 1 }

func (strichschrift) Maske(_ rune, groesse float64) (image.Image, image.Point, bool) {
	breite, hoehe := max(int(groesse/8), 1), max(int(groesse), 1)

	maske := image.NewAlpha(image.Rect(0, 0, breite, hoehe))
	for i := range maske.Pix {
		maske.Pix[i] = 255
	}

	return maske, image.Point{Y: -hoehe}, true
}
