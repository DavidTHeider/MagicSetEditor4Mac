package render_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Ein Foil ist eine Bildebene über der fertigen Karte — mehr nicht.
//
// **Keine Mischart.** In `yugioh-series10` ist `foils:` ein Block mit `left`,
// `top`, `width` und `height`, jedes aus einem Skript; ein Wort für „multiplizieren"
// oder „überlagern" kommt in keiner gesichteten Datei vor. `overlay` erscheint
// dort ausschließlich als **Feldname** (`atkoverlay`, `defoverlay`, und in
// `magic-m15` ein Styling-Feld „Should there be an overlay applied, such as
// foil?").
//
// Eine Mischart zu bauen, die kein Paket nennt, wäre dieselbe Erfindung wie
// eine Funktion, die keine Vorlage ruft.
func TestEinFoilIstEineBildebeneUeberDerKarte(t *testing.T) {
	// Eine Karte aus zwei Ebenen, darüber ein halbdurchsichtiges Foil.
	foil := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	foil.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 255, B: 255, A: 128})

	plan := &render.Plan{Breite: 1, Hoehe: 1, Ebenen: []render.Ebene{
		{
			Art: render.ArtFlaeche, Rahmen: render.Rechteck{Breite: 1, Hoehe: 1},
			Farbe: render.Farbe{A: 255}, Deckung: 1,
		},
		{
			Art: render.ArtBild, Bild: render.Blatt("/foil.png"),
			Rahmen: render.Rechteck{Breite: 1, Hoehe: 1}, Deckung: 1,
		},
	}}

	bild, befunde, err := render.NeuerRasterer(bilder{"/foil.png": foil}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil || len(befunde) != 0 {
		t.Fatalf("%v / %v", befunde, err)
	}

	// Schwarz unten, halbdeckendes Weiß oben: 255·(128/255) = 128.
	if got := farbeBei(t, bild, 0, 0); got != (render.Farbe{R: 128, G: 128, B: 128, A: 255}) {
		t.Errorf("%s, erwartet war rgba(128,128,128,255)", got)
	}
}

// Ein Foil mit Maske wirkt nur, wo die Maske es zulässt.
//
// So sieht ein Rahmenfoil aus: das Glanzbild über der ganzen Karte, die Maske
// schneidet die Bildfläche aus.
func TestEinFoilMitMaskeWirktNurWoDieMaskeEsZulaesst(t *testing.T) {
	foil := einfarbig(2, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	maske := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	maske.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	maske.SetNRGBA(1, 0, color.NRGBA{A: 255})

	plan := &render.Plan{Breite: 2, Hoehe: 1, Ebenen: []render.Ebene{
		{
			Art: render.ArtFlaeche, Rahmen: render.Rechteck{Breite: 2, Hoehe: 1},
			Farbe: render.Farbe{R: 40, G: 40, B: 40, A: 255}, Deckung: 1,
		},
		{
			Art: render.ArtBild, Bild: render.Blatt("/foil.png"), Maske: render.Blatt("/rahmen.png"),
			Rahmen: render.Rechteck{Breite: 2, Hoehe: 1}, Deckung: 1,
		},
	}}

	bild, befunde, err := render.NeuerRasterer(bilder{"/foil.png": foil, "/rahmen.png": maske}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil || len(befunde) != 0 {
		t.Fatalf("%v / %v", befunde, err)
	}

	if links := farbeBei(t, bild, 0, 0); links.R != 255 {
		t.Errorf("links %s — das Foil kam nicht durch", links)
	}

	if rechts := farbeBei(t, bild, 1, 0); rechts.R != 40 {
		t.Errorf("rechts %s — das Foil kam durch, wo die Maske schwarz ist", rechts)
	}
}

// Ein Foil mit halber Deckung mischt sich anteilig ein.
func TestEinFoilMitHalberDeckung(t *testing.T) {
	foil := einfarbig(1, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	plan := &render.Plan{Breite: 1, Hoehe: 1, Ebenen: []render.Ebene{
		{
			Art: render.ArtFlaeche, Rahmen: render.Rechteck{Breite: 1, Hoehe: 1},
			Farbe: render.Farbe{A: 255}, Deckung: 1,
		},
		{
			Art: render.ArtBild, Bild: render.Blatt("/foil.png"),
			Rahmen: render.Rechteck{Breite: 1, Hoehe: 1}, Deckung: 0.5,
		},
	}}

	bild, _, err := render.NeuerRasterer(bilder{"/foil.png": foil}).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	// 255·0,5 = 127,5 → 128 als Alpha des Foils; darüber Weiß auf Schwarz.
	if got := farbeBei(t, bild, 0, 0); got.R < 120 || got.R > 135 {
		t.Errorf("%s — halbe Deckung ergab nicht ungefähr halbes Weiß", got)
	}
}

// **Bildschirm und Druck kommen aus demselben Plan** (#34).
//
// Eine Vorschau und ein Druckbild, die auseinanderlaufen, tun das an genau
// einer Stelle, wenn es zwei Pläne gibt. Deshalb gibt es einen — und die
// Auflösung ist ein Zahlenwert des Rasterlaufs, keine Eigenschaft des Plans.
func TestBildschirmUndDruckKommenAusDemselbenPlan(t *testing.T) {
	quelle := einfarbig(4, 4, color.NRGBA{R: 200, G: 100, B: 50, A: 255})

	plan := &render.Plan{Breite: 10, Hoehe: 10, Ebenen: []render.Ebene{
		{
			Art: render.ArtBild, Bild: render.Blatt("/rahmen.png"),
			Rahmen: render.Rechteck{Breite: 10, Hoehe: 10}, Deckung: 1,
		},
		{
			Art: render.ArtFlaeche, Farbe: render.Farbe{B: 255, A: 255},
			Rahmen: render.Rechteck{Links: 2, Oben: 2, Breite: 6, Hoehe: 6}, Deckung: 1,
		},
	}}

	rasterer := render.NeuerRasterer(bilder{"/rahmen.png": quelle})

	bildschirm, _, err := rasterer.Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("Bildschirm: %v", err)
	}

	druck, _, err := rasterer.Zeichne(t.Context(), plan, 4, render.GueteMittel)
	if err != nil {
		t.Fatalf("Druck: %v", err)
	}

	if druck.Bounds().Dx() != 4*bildschirm.Bounds().Dx() {
		t.Fatalf("Druck %d breit, Bildschirm %d — nicht das Vierfache",
			druck.Bounds().Dx(), bildschirm.Bounds().Dx())
	}

	// **Dieselben Stellen tragen dieselben Farben.** Geprüft an drei Punkten:
	// im Rahmen, in der Fläche und an ihrer Grenze. Der Plan misst in
	// Vorlagenpunkten; ein Punkt bei (1,1) ist auf dem Bildschirm der Bildpunkt
	// (1,1) und im Druck der Bildpunkt (4,4).
	for _, punkt := range []struct{ x, y int }{{1, 1}, {5, 5}, {8, 8}} {
		auf := farbeBei(t, bildschirm, punkt.x, punkt.y)
		imDruck := farbeBei(t, druck, punkt.x*4, punkt.y*4)

		if auf != imDruck {
			t.Errorf("bei (%d,%d): Bildschirm %s, Druck %s", punkt.x, punkt.y, auf, imDruck)
		}
	}
}
