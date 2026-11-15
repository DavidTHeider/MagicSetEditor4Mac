package render_test

// **Was eine Glyphe über ihre Zeile hinaus trägt, bleibt stehen** — gefunden
// am 2026-09-24 an der Probekarte „Sonderzeichen ÄÖÜ ß «» ✦ ☙ 漢字 🃏" in
// der WebKit-Probe: gerendert stand „Sonderzeichen AOU ß «»".
//
// # Zwei Fehler, ein Puffer
//
// **Die Punkte eines Ä liegen über der Oberlänge** — bei Helvetica reicht
// Core Text zufolge die Oberlänge bis 0,77 em, das Ä bis 0,89 em. Der
// Textpuffer war genau so hoch wie die Zeile, und was darüber lag, fiel weg.
//
// **Ein ausgewichenes Zeichen hatte keine Breite.** Gemessen wurde ein Stück
// mit der Schrift des Stücks allein, und die meldet für ein Zeichen, das sie
// nicht führt, null — gezeichnet aber wurde aus der Ersatzschrift. Der Puffer
// war so schmal wie das Gemessene, und von der Glyphe blieb ein Streifen.

import (
	"image"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// wieEchte ist eine Schrift, der ein Zeichen fehlt — und die sich dabei
// verhält wie eine echte: ohne Glyphe kein Vorschub.
//
// **Hier fehlte der alte Lauf** (`ersatzschrift_test.go`): seine Probeschrift
// nannte auch für das fehlende Zeichen eine Breite, und der zu schmale Puffer
// fiel nicht auf.
type wieEchte struct {
	luecke
}

func (w wieEchte) Breite(zeichen rune) float64 {
	if zeichen == w.ohne {
		return 0
	}

	return w.luecke.Breite(zeichen)
}

func (w wieEchte) Fuehrt(zeichen rune) bool { return zeichen != w.ohne }

// spalten zählt die Spalten eines Bildes, in denen etwas gemalt ist.
func spalten(bild *image.NRGBA) int {
	gemalt := 0

	for x := range bild.Bounds().Dx() {
		for y := range bild.Bounds().Dy() {
			if bild.NRGBAAt(x, y).A > 0 {
				gemalt++

				break
			}
		}
	}

	return gemalt
}

func TestEinAusgewichenesZeichenStehtInVollerBreite(t *testing.T) {
	const fehlt = '→'

	plan := &render.Plan{Breite: 40, Hoehe: 12, Ebenen: []render.Ebene{{
		Art:        render.ArtText,
		Rahmen:     render.Rechteck{Breite: 40, Hoehe: 12},
		Abschnitte: []render.Abschnitt{{Text: string([]rune{fehlt, fehlt, fehlt, fehlt})}},
		Farbe:      render.Farbe{R: 255, A: 255},
		Satz:       render.Satz{Schrift: "Probe", Groesse: 6, Ausrichtung: render.LinksOben},
		Deckung:    1,
	}}}

	rasterer := render.NeuerRasterer(bildmasse{}).
		MitSchriften(schriften{"Probe": wieEchte{luecke{ohne: fehlt}}}).
		MitErsatzschrift(func(rune) (render.Glyphenschrift, bool) {
			return einheitsschrift{}, true
		})

	bild, _, err := rasterer.Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	// **Vier Zeichen zu je einer halben Schrifthöhe**: 4 · 0,5 · 6 = 12
	// Spalten. Mit der Breite der Hauptschrift waren es drei — so breit wie
	// MSEs Zugabe zum Puffer.
	if gemalt := spalten(bild); gemalt < 10 {
		t.Errorf("gemalt in %d Spalten — vier ausgewichene Zeichen brauchen 12", gemalt)
	}
}

// hoheMarke ist eine Schrift, deren Glyphen über die Oberlänge reichen — wie
// die Punkte eines Ä.
type hoheMarke struct{ einheitsschrift }

// Maske gibt ein Quadrat von einer Schrifthöhe, dessen Oberkante 1,1 Höhen
// über der Grundlinie liegt — die Oberlänge ist 0,8.
func (hoheMarke) Maske(_ rune, groesse float64) (image.Image, image.Point, bool) {
	seite := int(groesse)
	maske := image.NewAlpha(image.Rect(0, 0, seite, seite))

	for i := range maske.Pix {
		maske.Pix[i] = 255
	}

	return maske, image.Point{Y: -int(groesse * 1.1)}, true
}

func TestEineGlypheUeberDerOberlaengeBleibtStehen(t *testing.T) {
	const oben = 20

	plan := &render.Plan{Breite: 40, Hoehe: 60, Ebenen: []render.Ebene{{
		Art:        render.ArtText,
		Rahmen:     render.Rechteck{Oben: oben, Breite: 40, Hoehe: 30},
		Abschnitte: []render.Abschnitt{{Text: "Ä"}},
		Farbe:      render.Farbe{R: 255, A: 255},
		Satz:       render.Satz{Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben},
		Deckung:    1,
	}}}

	rasterer := render.NeuerRasterer(bildmasse{}).
		MitSchriften(schriften{"Probe": hoheMarke{}})

	bild, _, err := rasterer.Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	oberste := -1

	for y := range bild.Bounds().Dy() {
		for x := range bild.Bounds().Dx() {
			if bild.NRGBAAt(x, y).A > 0 {
				oberste = y

				break
			}
		}

		if oberste >= 0 {
			break
		}
	}

	// **Die Zeile beginnt am Rahmen**, die Grundlinie 0,8 Höhen darunter,
	// die Glyphe 1,1 Höhen über ihr — also 0,3 Höhen, drei Punkte, über dem
	// Rahmen. Mit dem alten Puffer war die oberste gemalte Zeile der Rahmen.
	if oberste < 0 || oberste > oben-2 {
		t.Errorf("die oberste gemalte Zeile ist %d — die Glyphe reicht bis %d", oberste, oben-3)
	}
}

// **Eine echte Schrift sagt, was sie führt** — Andale Mono hat ein A und
// kein Zeichen aus der privaten Fläche.
func TestEineEchteSchriftSagtWasSieFuehrt(t *testing.T) {
	schrift := echteSchrift(t)

	if !schrift.Fuehrt('A') {
		t.Error("Andale Mono führt kein A?")
	}

	if schrift.Fuehrt('') {
		t.Error("Andale Mono führt ein Zeichen der privaten Fläche?")
	}
}

// **Findet auch die Ersatzschrift nichts, bleibt es bei der Hauptschrift** —
// das Zeichen fehlt dann, und genau das meldet der Befund.
func TestOhneErsatzBleibtDasZeichenAus(t *testing.T) {
	const fehlt = '→'

	plan := &render.Plan{Breite: 20, Hoehe: 12, Ebenen: []render.Ebene{{
		Art:        render.ArtText,
		Rahmen:     render.Rechteck{Breite: 20, Hoehe: 12},
		Abschnitte: []render.Abschnitt{{Text: string(fehlt)}},
		Farbe:      render.Farbe{R: 255, A: 255},
		Satz:       render.Satz{Schrift: "Probe", Groesse: 6, Ausrichtung: render.LinksOben},
		Deckung:    1,
	}}}

	rasterer := render.NeuerRasterer(bildmasse{}).
		MitSchriften(schriften{"Probe": wieEchte{luecke{ohne: fehlt}}}).
		MitErsatzschrift(func(rune) (render.Glyphenschrift, bool) {
			return nil, false
		})

	bild, befunde, err := rasterer.Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if gemalt := spalten(bild); gemalt != 0 {
		t.Errorf("gemalt in %d Spalten, obwohl keine Schrift das Zeichen führt", gemalt)
	}

	if len(befunde) != 1 {
		t.Errorf("%d Befunde, erwartet einer: %v", len(befunde), befunde)
	}
}

// riesig ist eine Schrift, deren Glyphe vier Höhen über der Grundlinie
// beginnt — weiter, als der Rand des Puffers reicht.
type riesig struct{ einheitsschrift }

func (riesig) Maske(_ rune, groesse float64) (image.Image, image.Point, bool) {
	seite := int(groesse)
	maske := image.NewAlpha(image.Rect(0, 0, seite, seite*5))

	for i := range maske.Pix {
		maske.Pix[i] = 255
	}

	return maske, image.Point{Y: -int(groesse * 4)}, true
}

// **Auch mit Rand hat der Puffer ein Ende** — was weiter ragt als eine halbe
// Zeile, fällt weiterhin weg, und der Lauf stolpert nicht darüber.
func TestWasUeberDenRandRagtFaelltWeg(t *testing.T) {
	const oben = 40

	plan := &render.Plan{Breite: 40, Hoehe: 80, Ebenen: []render.Ebene{{
		Art:        render.ArtText,
		Rahmen:     render.Rechteck{Oben: oben, Breite: 40, Hoehe: 30},
		Abschnitte: []render.Abschnitt{{Text: "X"}},
		Farbe:      render.Farbe{R: 255, A: 255},
		Satz:       render.Satz{Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben},
		Deckung:    1,
	}}}

	rasterer := render.NeuerRasterer(bildmasse{}).
		MitSchriften(schriften{"Probe": riesig{}})

	bild, _, err := rasterer.Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	// **Die Glyphe begänne 40 Punkte über der Grundlinie** (oben 40 + 8), der
	// Rand reicht eine halbe Zeile (6) über den Rahmen — gemalt wird ab 34.
	for y := range oben - 7 {
		for x := range 40 {
			if bild.NRGBAAt(x, y).A > 0 {
				t.Fatalf("gemalt in Zeile %d — über den Rand des Puffers hinaus", y)
			}
		}
	}

	if spalten(bild) == 0 {
		t.Error("gar nichts gemalt")
	}
}
