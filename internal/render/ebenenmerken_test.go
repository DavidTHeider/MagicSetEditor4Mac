package render_test

import (
	"errors"
	"image"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Gemerkt wird nur, was wiederkommt** — E06
// (#2278).
//
// Der Ebenenspeicher schlüsselt nach dem **Zeiger** der Quelle (#672): ein
// Bild aus der Bildquelle ist bei jedem Lauf dasselbe. Ein **erzeugtes** —
// das Schachbrett eines leeren Bildfelds, ein Setsymbol, eine Verknüpfung —
// ist bei jedem Lauf ein neues; sein Eintrag wird nie wieder getroffen, hält
// die Quelle fest und füllt den Speicher, bis er ganz geräumt wird. Im
// Dauerbetrieb trug das Schachbrett so 6,4 MB in fünf Minuten.
func TestEinErzeugtesBildWirdNichtGemerkt(t *testing.T) {
	quelle := festeQuelle{"/rahmen.png": image.NewNRGBA(image.Rect(0, 0, 50, 50))}
	plan := &render.Plan{Breite: 100, Hoehe: 100, Ebenen: []render.Ebene{
		{Art: render.ArtBild, Rahmen: render.Rechteck{Breite: 100, Hoehe: 100}, Bild: render.Blatt("/rahmen.png"), Deckung: 1},
		{Art: render.ArtBild, Rahmen: render.Rechteck{Links: 10, Oben: 10, Breite: 60, Hoehe: 40}, Bild: render.Bildbaum{Knotenart: render.Platzhalterknoten}, Deckung: 1},
	}}

	rasterer := render.NeuerRasterer(quelle)

	for range 3 {
		if _, _, err := rasterer.Zeichne(t.Context(), plan, 1, render.GueteMittel); err != nil {
			t.Fatalf("zeichnen: %v", err)
		}
	}

	zahlen := rasterer.Ebenenzahlen()
	if zahlen.Eintraege != 1 || zahlen.Treffer != 2 {
		t.Errorf("Ebenenspeicher %+v — erwartet ein Eintrag (der Rahmen) und zwei Treffer", zahlen)
	}
}

// festeQuelle gibt für denselben Pfad dasselbe Bild — wie
// `msepaket.Bildquelle`, die ihre Bilder vorhält.
type festeQuelle map[string]image.Image

func (q festeQuelle) Bild(pfad string) (image.Image, error) {
	if bild, da := q[pfad]; da {
		return bild, nil
	}

	return nil, errors.New("kein Bild unter " + pfad)
}
