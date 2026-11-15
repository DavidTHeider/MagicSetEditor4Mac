package render_test

import (
	"context"
	"errors"
	"image"
	"image/color"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// bilder ist eine Bildquelle aus dem Arbeitsspeicher.
type bilder map[string]image.Image

func (b bilder) Bild(pfad string) (image.Image, error) {
	bild, da := b[pfad]
	if !da {
		return nil, errors.New("nicht vorhanden")
	}

	return bild, nil
}

// einfarbig baut ein Bild einer Farbe.
func einfarbig(breite, hoehe int, farbe color.NRGBA) image.Image {
	bild := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))
	for y := range hoehe {
		for x := range breite {
			bild.SetNRGBA(x, y, farbe)
		}
	}

	return bild
}

func farbeBei(t *testing.T, bild *image.NRGBA, x, y int) render.Farbe {
	t.Helper()

	stelle := bild.PixOffset(x, y)

	return render.Farbe{
		R: bild.Pix[stelle], G: bild.Pix[stelle+1],
		B: bild.Pix[stelle+2], A: bild.Pix[stelle+3],
	}
}

// Alphaverrechnung an Beispielen mit bekanntem Ergebnis.
//
// **Von Hand nachgerechnet, nicht aus dem Lauf übernommen.** Ein Sollwert, den
// jemand aus der ersten Ausgabe abschreibt, prüft, dass sich nichts ändert —
// nicht, dass es stimmt.
func TestAlphaWirdNachDerFormelVerrechnet(t *testing.T) {
	faelle := []struct {
		name     string
		unten    render.Farbe
		oben     render.Farbe
		erwartet render.Farbe
	}{
		{
			name:  "undurchsichtig über undurchsichtig",
			unten: render.Farbe{R: 255, A: 255}, oben: render.Farbe{B: 255, A: 255},
			erwartet: render.Farbe{B: 255, A: 255},
		},
		{
			// **128 ist nicht die Hälfte von 255**, und der erste Anlauf dieses
			// Falls rechnete so: 128/255 = 0,50196, nicht 0,5.
			//
			//	aus = 0,50196 + 1·0,49804 = 1
			//	rot = (0·0,50196 + 255·1·0,49804)/1 = 255 − 128 = 127
			//
			// Der Code gab 127, der Sollwert sagte 128, und richtig lag der
			// Code. Genau dafür wird von Hand nachgerechnet statt aus dem Lauf
			// abgeschrieben — ein übernommener Sollwert hätte den Fehler in der
			// Rechnung nie gezeigt, weil es keinen gab.
			name:  "halb über undurchsichtig",
			unten: render.Farbe{R: 255, A: 255}, oben: render.Farbe{A: 128},
			erwartet: render.Farbe{R: 127, A: 255},
		},
		{
			name:  "durchsichtig ändert nichts",
			unten: render.Farbe{R: 255, G: 128, A: 255}, oben: render.Farbe{B: 255, A: 0},
			erwartet: render.Farbe{R: 255, G: 128, A: 255},
		},
		{
			// Auf Leerem: aus = 0,5, und die Farbe bleibt die obere.
			name:  "halb über nichts",
			unten: render.Farbe{}, oben: render.Farbe{R: 200, A: 128},
			erwartet: render.Farbe{R: 200, A: 128},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			plan := &render.Plan{Breite: 1, Hoehe: 1, Ebenen: []render.Ebene{
				{
					Art: render.ArtFlaeche, Rahmen: render.Rechteck{Breite: 1, Hoehe: 1},
					Farbe: fall.unten, Deckung: 1,
				},
				{
					Art: render.ArtFlaeche, Rahmen: render.Rechteck{Breite: 1, Hoehe: 1},
					Farbe: fall.oben, Deckung: 1,
				},
			}}

			bild, befunde, err := render.NeuerRasterer(bilder{}).Zeichne(t.Context(), plan, 1, render.GueteMittel)
			if err != nil || len(befunde) != 0 {
				t.Fatalf("%v / %v", befunde, err)
			}

			if got := farbeBei(t, bild, 0, 0); got != fall.erwartet {
				t.Errorf("%s, erwartet war %s", got, fall.erwartet)
			}
		})
	}
}

// Die Maske wirkt **vor** dem Verlauf, nicht danach.
//
// Andersherum radierte sie weg, was darunter schon stand — der Verlauf
// überdeckte, was die Maske freigelegt hat. Auf einer einzelnen Karte sieht das
// fast richtig aus, und genau deshalb steht der Fall hier.
func TestDieMaskeWirktVorDemVerlaufUndNichtDanach(t *testing.T) {
	// Ein Untergrund aus Rot, darüber eine blaue Fläche mit einer Maske, die
	// links ganz und rechts gar nicht durchlässt.
	maske := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	maske.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	maske.SetNRGBA(1, 0, color.NRGBA{A: 255})

	plan := &render.Plan{Breite: 2, Hoehe: 1, Ebenen: []render.Ebene{
		{
			Art: render.ArtFlaeche, Rahmen: render.Rechteck{Breite: 2, Hoehe: 1},
			Farbe: render.Farbe{R: 255, A: 255}, Deckung: 1,
		},
		{
			// **Mit Verlauf**, denn die Zusage lautet „Maske vor dem Verlauf".
			// Der erste Anlauf dieses Tests nahm eine einfarbige Fläche und
			// prüfte damit nur die Hälfte der Behauptung. Gefunden vom Nachlauf
			// des Zweitblicks über Phase 3.
			Art: render.ArtFlaeche, Rahmen: render.Rechteck{Breite: 2, Hoehe: 1},
			Verlauf: &render.Verlauf{Stufen: []render.Verlaufsstufe{
				{Stelle: 0, Farbe: render.Farbe{B: 255, A: 255}},
				{Stelle: 1, Farbe: render.Farbe{B: 255, A: 255}},
			}},
			Maske: render.Blatt("/m.png"), Deckung: 1,
		},
	}}

	bild, befunde, err := render.NeuerRasterer(bilder{"/m.png": maske}).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil || len(befunde) != 0 {
		t.Fatalf("%v / %v", befunde, err)
	}

	if links := farbeBei(t, bild, 0, 0); links != (render.Farbe{B: 255, A: 255}) {
		t.Errorf("links %s — die Maske ließ nicht durch, wo sie weiß ist", links)
	}

	// **Rechts steht der Untergrund, nicht Leere.** Wirkte die Maske erst auf
	// das fertige Bild, hätte sie hier auch das Rot ausradiert.
	if rechts := farbeBei(t, bild, 1, 0); rechts != (render.Farbe{R: 255, A: 255}) {
		t.Errorf("rechts %s — die Maske hat den Untergrund mitgenommen", rechts)
	}
}

// Deckung 0 zeichnet nichts — der Nullwert eines Go-Verbunds ist 0.
//
// Eine Ebene, die jemand ohne Deckung hinschreibt, soll nichts zeichnen statt
// alles zu überdecken. Das ist eine Falle, die man einmal baut und dreimal
// sucht.
func TestDeckungNullZeichnetNichts(t *testing.T) {
	plan := &render.Plan{Breite: 1, Hoehe: 1, Ebenen: []render.Ebene{
		{
			Art: render.ArtFlaeche, Rahmen: render.Rechteck{Breite: 1, Hoehe: 1},
			Farbe: render.Farbe{R: 255, A: 255}, Deckung: 1,
		},
		{
			Art: render.ArtFlaeche, Rahmen: render.Rechteck{Breite: 1, Hoehe: 1},
			Farbe: render.Farbe{B: 255, A: 255},
		},
	}}

	bild, _, err := render.NeuerRasterer(bilder{}).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	if got := farbeBei(t, bild, 0, 0); got != (render.Farbe{R: 255, A: 255}) {
		t.Errorf("%s — eine Ebene ohne Deckung hat gezeichnet", got)
	}
}

func TestZeichneWeistUnbrauchbareVergroesserungenAb(t *testing.T) {
	plan := &render.Plan{Breite: 10, Hoehe: 10}

	for _, vergroesserung := range []float64{0, -1, 0.001} {
		if _, _, err := render.NeuerRasterer(bilder{}).
			Zeichne(t.Context(), plan, vergroesserung, render.GueteMittel); !errors.Is(err, render.ErrPlan) {
			t.Errorf("Vergrößerung %g: Fehler %v, erwartet war ErrPlan", vergroesserung, err)
		}
	}

	if _, _, err := render.NeuerRasterer(bilder{}).
		Zeichne(t.Context(), nil, 1, render.GueteMittel); !errors.Is(err, render.ErrPlan) {
		t.Errorf("Fehler %v, erwartet war ErrPlan", err)
	}
}

// Ein laufender Rasterlauf lässt sich abbrechen (AGENTS.md §7).
//
// **Für dieses Projekt eigens genannt:** eine Kartenvorschau wird abgebrochen,
// nicht abgewartet. Wer während des Tippens vier Läufe anstößt, zeigt sonst den
// drittletzten Stand.
func TestEinRasterlaufLaesstSichAbbrechen(t *testing.T) {
	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	plan := &render.Plan{Breite: 10, Hoehe: 10, Ebenen: []render.Ebene{
		{Art: render.ArtFlaeche, Rahmen: render.Rechteck{Breite: 10, Hoehe: 10}, Deckung: 1},
	}}

	if _, _, err := render.NeuerRasterer(bilder{}).Zeichne(ctx, plan, 1, render.GueteMittel); !errors.Is(err, context.Canceled) {
		t.Errorf("Fehler %v, erwartet war der Abbruch", err)
	}
}
