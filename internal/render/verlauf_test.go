package render_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// verlaufsplan baut eine Fläche mit einem Verlauf von Schwarz nach Weiß.
func verlaufsplan(winkel, breite, hoehe float64) *render.Plan {
	return &render.Plan{Breite: breite, Hoehe: hoehe, Ebenen: []render.Ebene{
		{
			Art:    render.ArtFlaeche,
			Rahmen: render.Rechteck{Breite: breite, Hoehe: hoehe},
			Verlauf: &render.Verlauf{Winkel: winkel, Stufen: []render.Verlaufsstufe{
				{Stelle: 0, Farbe: render.Farbe{A: 255}},
				{Stelle: 1, Farbe: render.Farbe{R: 255, G: 255, B: 255, A: 255}},
			}},
			Deckung: 1,
		},
	}}
}

// Ein Verlauf läuft in die Richtung, die sein Winkel nennt.
//
// Geprüft wird die **Richtung**, nicht der genaue Wert: dass links dunkler ist
// als rechts, ist die Zusage; dass die Mitte auf zwei Stellen genau bei 127
// liegt, wäre eine Zusage über die Rundung.
func TestEinVerlaufLaeuftInSeineRichtung(t *testing.T) {
	faelle := []struct {
		name               string
		winkel             float64
		dunklerX, dunklerY int
		hellerX, hellerY   int
	}{
		{name: "nach rechts", winkel: 0, dunklerX: 0, dunklerY: 1, hellerX: 3, hellerY: 1},
		{name: "nach unten", winkel: 90, dunklerX: 1, dunklerY: 0, hellerX: 1, hellerY: 3},
		{name: "nach links", winkel: 180, dunklerX: 3, dunklerY: 1, hellerX: 0, hellerY: 1},
		{name: "nach oben", winkel: 270, dunklerX: 1, dunklerY: 3, hellerX: 1, hellerY: 0},
		{name: "schräg", winkel: 45, dunklerX: 0, dunklerY: 0, hellerX: 3, hellerY: 3},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			bild, _, err := render.NeuerRasterer(bilder{}).
				Zeichne(t.Context(), verlaufsplan(fall.winkel, 4, 4), 1, render.GueteMittel)
			if err != nil {
				t.Fatalf("abgewiesen: %v", err)
			}

			dunkler := farbeBei(t, bild, fall.dunklerX, fall.dunklerY)
			heller := farbeBei(t, bild, fall.hellerX, fall.hellerY)

			if dunkler.R >= heller.R {
				t.Errorf("dunkel %s, hell %s — der Verlauf läuft falsch herum", dunkler, heller)
			}
		})
	}
}

// Die Enden des Verlaufs sind seine Randfarben.
func TestDieEndenEinesVerlaufsSindSeineRandfarben(t *testing.T) {
	bild, _, err := render.NeuerRasterer(bilder{}).Zeichne(t.Context(), verlaufsplan(0, 8, 1), 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	if links := farbeBei(t, bild, 0, 0); links.R > 32 {
		t.Errorf("links %s — das dunkle Ende ist zu hell", links)
	}

	if rechts := farbeBei(t, bild, 7, 0); rechts.R < 223 {
		t.Errorf("rechts %s — das helle Ende ist zu dunkel", rechts)
	}
}

// Zwei Stufen an derselben Stelle sind ein harter Übergang, kein Absturz.
func TestZweiStufenAnDerselbenStelle(t *testing.T) {
	plan := &render.Plan{Breite: 4, Hoehe: 1, Ebenen: []render.Ebene{
		{
			Art:    render.ArtFlaeche,
			Rahmen: render.Rechteck{Breite: 4, Hoehe: 1},
			Verlauf: &render.Verlauf{Stufen: []render.Verlaufsstufe{
				{Stelle: 0.5, Farbe: render.Farbe{A: 255}},
				{Stelle: 0.5, Farbe: render.Farbe{R: 255, A: 255}},
			}},
			Deckung: 1,
		},
	}}

	bild, befunde, err := render.NeuerRasterer(bilder{}).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil || len(befunde) != 0 {
		t.Fatalf("%v / %v", befunde, err)
	}

	// Vor der Stufe die erste Farbe, dahinter die zweite.
	if vorne := farbeBei(t, bild, 0, 0); vorne.R != 0 {
		t.Errorf("vorne %s", vorne)
	}

	if hinten := farbeBei(t, bild, 3, 0); hinten.R != 255 {
		t.Errorf("hinten %s", hinten)
	}
}

// Eine Ebene, die über den Rand hinausragt, wird abgeschnitten statt zu stürzen.
func TestEineEbeneUeberDemRandWirdAbgeschnitten(t *testing.T) {
	plan := &render.Plan{Breite: 2, Hoehe: 2, Ebenen: []render.Ebene{
		{
			Art: render.ArtFlaeche, Farbe: render.Farbe{R: 255, A: 255},
			Rahmen: render.Rechteck{Links: -5, Oben: -5, Breite: 20, Hoehe: 20}, Deckung: 1,
		},
		{
			Art: render.ArtBild, Bild: render.Blatt("/b.png"),
			Rahmen: render.Rechteck{Links: 1, Oben: 1, Breite: 20, Hoehe: 20}, Deckung: 1,
		},
	}}

	quelle := einfarbig(2, 2, color.NRGBA{B: 255, A: 255})

	bild, befunde, err := render.NeuerRasterer(bilder{"/b.png": quelle}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil || len(befunde) != 0 {
		t.Fatalf("%v / %v", befunde, err)
	}

	if got := farbeBei(t, bild, 0, 0); got != (render.Farbe{R: 255, A: 255}) {
		t.Errorf("%s", got)
	}

	if got := farbeBei(t, bild, 1, 1); got.B != 255 {
		t.Errorf("%s — das Bild hat den Ausschnitt nicht getroffen", got)
	}
}

// Ein völlig durchsichtiges Bild ändert nichts.
func TestEinDurchsichtigesBildAendertNichts(t *testing.T) {
	quelle := einfarbig(2, 2, color.NRGBA{R: 255, A: 0})

	plan := &render.Plan{Breite: 2, Hoehe: 2, Ebenen: []render.Ebene{
		{
			Art: render.ArtFlaeche, Farbe: render.Farbe{G: 255, A: 255},
			Rahmen: render.Rechteck{Breite: 2, Hoehe: 2}, Deckung: 1,
		},
		{
			Art: render.ArtBild, Bild: render.Blatt("/durchsichtig.png"),
			Rahmen: render.Rechteck{Breite: 2, Hoehe: 2}, Deckung: 1,
		},
	}}

	bild, _, err := render.NeuerRasterer(bilder{"/durchsichtig.png": quelle}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	if got := farbeBei(t, bild, 0, 0); got != (render.Farbe{G: 255, A: 255}) {
		t.Errorf("%s", got)
	}
}

// Ein Bild ohne Fläche wird übergangen, nicht durch null geteilt.
func TestEinBildOhneFlaecheWirdUebergangen(t *testing.T) {
	leer := image.NewNRGBA(image.Rect(0, 0, 0, 0))

	plan := &render.Plan{Breite: 2, Hoehe: 2, Ebenen: []render.Ebene{
		{
			Art: render.ArtBild, Bild: render.Blatt("/leer.png"),
			Rahmen: render.Rechteck{Breite: 2, Hoehe: 2}, Deckung: 1,
		},
	}}

	bild, _, err := render.NeuerRasterer(bilder{"/leer.png": leer}).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	if got := farbeBei(t, bild, 0, 0); got != (render.Farbe{}) {
		t.Errorf("%s", got)
	}
}

// Ein Verlauf ohne Ausdehnung ergibt seine erste Farbe.
func TestEinVerlaufOhneAusdehnung(t *testing.T) {
	plan := &render.Plan{Breite: 2, Hoehe: 2, Ebenen: []render.Ebene{
		{
			Art:    render.ArtFlaeche,
			Rahmen: render.Rechteck{Breite: 0, Hoehe: 0},
			Verlauf: &render.Verlauf{Stufen: []render.Verlaufsstufe{
				{Stelle: 0, Farbe: render.Farbe{A: 255}},
				{Stelle: 1, Farbe: render.Farbe{R: 255, A: 255}},
			}},
			Deckung: 1,
		},
	}}

	if _, befunde, err := render.NeuerRasterer(bilder{}).Zeichne(t.Context(), plan, 1, render.GueteMittel); err != nil ||
		len(befunde) != 0 {
		t.Errorf("%v / %v", befunde, err)
	}
}

// Zwei durchsichtige Farben übereinander bleiben durchsichtig.
func TestZweiDurchsichtigeFarbenBleibenDurchsichtig(t *testing.T) {
	plan := &render.Plan{Breite: 1, Hoehe: 1, Ebenen: []render.Ebene{
		{
			Art: render.ArtFlaeche, Farbe: render.Farbe{R: 255, A: 0},
			Rahmen: render.Rechteck{Breite: 1, Hoehe: 1}, Deckung: 1,
		},
	}}

	bild, _, err := render.NeuerRasterer(bilder{}).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	if got := farbeBei(t, bild, 0, 0); got.A != 0 {
		t.Errorf("%s", got)
	}
}
