package render_test

// **Eine Ebene ohne Fläche wird gar nicht erst betreten**
// (#442).
//
//	bool ValueViewer::isVisible() const {
//	  return getStyle()->visible
//	    && bounding_box.width > 0
//	    && bounding_box.height > 0
//	    && fabs(bounding_box.x) < 100000
//	    && fabs(bounding_box.y) < 100000;
//	}
//
// `src/render/value/viewer.cpp`, Zeilen 36–42, Fassung f9d9356 vom 2021-07-06.
// MSE prüft das **vor** dem Zeichnen (`viewer.cpp`, Zeile 62) — und damit vor
// jedem Bild, das die Ebene bräuchte.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestEineEbeneOhneFlaecheSuchtKeineMaske(t *testing.T) {
	// **So verbirgt eine Vorlage ein Feld**: `magic-m15-mainframe-dfc`
	// schreibt für `art mirror` `height: {… card.card_stamp != "none" … else 0}`
	// und dazu `mask: {"image_masks/mythic_stamp_" + stamp_shape() + ".png"}`.
	// Bei `none` gibt es die Datei nicht — und MSE fragt nie danach.
	plan := &render.Plan{Breite: 40, Hoehe: 20, Ebenen: []render.Ebene{{
		Art:     render.ArtBild,
		Rahmen:  render.Rechteck{Breite: 40, Hoehe: 0},
		Bild:    render.Blatt("/da.png"),
		Maske:   render.Blatt("/gibtsnicht.png"),
		Deckung: 1,
	}}}

	_, befunde, err := render.NeuerRasterer(bildmasse{}).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) > 0 {
		t.Errorf("Befunde: %+v", befunde)
	}
}

// **Und mit Fläche wird sie gesucht** — die Gegenprobe.
func TestEineEbeneMitFlaecheSuchtIhreMaske(t *testing.T) {
	plan := &render.Plan{Breite: 40, Hoehe: 20, Ebenen: []render.Ebene{{
		Art:     render.ArtBild,
		Rahmen:  render.Rechteck{Breite: 40, Hoehe: 20},
		Bild:    render.Blatt("/da.png"),
		Maske:   render.Blatt("/gibtsnicht.png"),
		Deckung: 1,
	}}}

	_, befunde, err := render.NeuerRasterer(bildmasse{}).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "gibtsnicht.png") {
		t.Errorf("Befunde: %+v", befunde)
	}
}

// **Weit außerhalb der Karte zählt ebenso als unsichtbar.**
//
// `fabs(bounding_box.x) < 100000` — die Grenze steht in MSEs eigener
// Bedingung; sie fängt eine Rechnung ab, die ins Unermessliche gelaufen ist.
func TestEineEbeneWeitDraussenWirdUebergangen(t *testing.T) {
	plan := &render.Plan{Breite: 40, Hoehe: 20, Ebenen: []render.Ebene{{
		Art:     render.ArtBild,
		Rahmen:  render.Rechteck{Links: 100000, Breite: 40, Hoehe: 20},
		Bild:    render.Blatt("/da.png"),
		Maske:   render.Blatt("/gibtsnicht.png"),
		Deckung: 1,
	}}}

	_, befunde, err := render.NeuerRasterer(bildmasse{}).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) > 0 {
		t.Errorf("Befunde: %+v", befunde)
	}
}
