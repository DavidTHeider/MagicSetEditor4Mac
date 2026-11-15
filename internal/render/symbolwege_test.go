package render_test

import (
	"image"
	"image/color"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Jeder produktive Weg gibt die Größe mit, in der sein Bild gebraucht wird** —
// #751.
//
// # Der Befund
//
// `Symbolflaeche` wählt die Auflösung eines `symbol_variation` aus der
// **gefragten** Größe (#735). Dass sie richtig rechnet, prüfte
// `symbolflaeche_test.go`; dass die produktiven Wege sie überhaupt fragen,
// prüfte **nur der Testeingang** — also der Weg, den kein Betrieb nimmt.
// Angestrichen vom Zweitblick durch Codex am 2026-08-21.
//
// Es sind fünf, und sie stehen an vier Stellen:
//
//  1. die Maske einer Ebene (`bild.go`),
//  2. das Bild einer Ebene (ebenda),
//  3. ein Symbol in einem Symbolzug (`symbolzug.go`),
//  4. ein Bildwert für `write_image_file` (`bildausbaum.go`),
//  5. ein Symbol als eigene Datei für `symbols_to_html` (`symbolalsbild.go`).
//
// # Warum es am Ergebnis nicht abzulesen ist
//
// **Weil das Bild danach ohnehin auf den Rahmen gebracht wird.** Ein Weg, der
// versehentlich `image.Point{}` gibt, liefert dasselbe Bild — nur unschärfer.
// Der Unterschied steht in der Auflösung **vor** der Verkleinerung, und die
// sieht von außen niemand. Deshalb meldet der Rasterer sie auf Verlangen
// ([render.Rasterer.MeldeGroesseFuerTest]).
func TestJederWegGibtSeineGroesseMit(t *testing.T) {
	t.Parallel()

	t.Run("das Bild einer Ebene", func(t *testing.T) {
		t.Parallel()

		gemeldet := mitschreiber()

		plan := &render.Plan{Breite: 40, Hoehe: 30, Ebenen: []render.Ebene{{
			Art:     render.ArtBild,
			Deckung: 1,
			Rahmen:  render.Rechteck{Breite: 30, Hoehe: 20},
			Bild:    render.Blatt("/teil.png"),
		}}}

		if _, _, err := rastererMit(gemeldet).Zeichne(t.Context(), plan, 2, render.GueteMittel); err != nil {
			t.Fatalf("zeichnen: %v", err)
		}

		// **Mal der Vergrößerung**: die Ebene wird bei 2 gezeichnet, und das
		// Bild wird in dieser Größe gebraucht — nicht in der des Plans.
		gemeldet.erwarte(t, image.Point{X: 60, Y: 40})
	})

	t.Run("die Maske einer Ebene", func(t *testing.T) {
		t.Parallel()

		gemeldet := mitschreiber()

		plan := &render.Plan{Breite: 40, Hoehe: 30, Ebenen: []render.Ebene{{
			Art:     render.ArtBild,
			Deckung: 1,
			Rahmen:  render.Rechteck{Breite: 30, Hoehe: 20},
			Bild:    render.Blatt("/teil.png"),
			Maske:   render.Blatt("/maske.png"),
		}}}

		if _, _, err := rastererMit(gemeldet).Zeichne(t.Context(), plan, 1, render.GueteMittel); err != nil {
			t.Fatalf("zeichnen: %v", err)
		}

		// Maske **und** Bild fragen dieselbe Größe — zwei Meldungen.
		gemeldet.erwarte(t, image.Point{X: 30, Y: 20}, image.Point{X: 30, Y: 20})
	})

	t.Run("ein Bildwert für write_image_file", func(t *testing.T) {
		t.Parallel()

		gemeldet := mitschreiber()

		if bild, _ := rastererMit(gemeldet).
			BildAusBaum(render.Blatt("/teil.png"), 24, 18); bild == nil {
			t.Fatal("kein Bild")
		}

		gemeldet.erwarte(t, image.Point{X: 24, Y: 18})
	})

	t.Run("ein Symbol als eigene Datei", func(t *testing.T) {
		t.Parallel()

		gemeldet := mitschreiber()

		bild, befund := rastererMit(gemeldet).SymbolAlsBild(
			render.Symbol{Bild: render.Blatt("/teil.png")},
			render.Symbolmasse{Groesse: 20, Bildgroesse: 40})
		if befund != "" || bild == nil {
			t.Fatalf("Befund: %s", befund)
		}

		// **Der Kasten des Symbols**, nicht die Größe der Schrift: das Bild
		// ist 8 × 6, und `Groesse/Bildgroesse` ist ein Halb.
		gemeldet.erwarte(t, image.Point{X: 4, Y: 3})
	})

	t.Run("ein Symbol in einem Symbolzug", func(t *testing.T) {
		t.Parallel()

		gemeldet := mitschreiber()

		plan := &render.Plan{Breite: 60, Hoehe: 60, Ebenen: []render.Ebene{{
			Art:     render.ArtSymbolzug,
			Rahmen:  render.Rechteck{Breite: 60, Hoehe: 60},
			Deckung: 1,
			Satz:    render.Satz{Groesse: 40},
			Abschnitte: []render.Abschnitt{{
				Symbolmasse: render.Symbolmasse{Groesse: 40, Bildgroesse: 40},
				Symbole:     []render.Symbol{{Bild: render.Blatt("/teil.png")}},
			}},
		}}}

		if _, _, err := rastererMit(gemeldet).Zeichne(t.Context(), plan, 1, render.GueteMittel); err != nil {
			t.Fatalf("zeichnen: %v", err)
		}

		// Derselbe Kasten wie oben, hier ohne Verkleinerung.
		gemeldet.erwarte(t, image.Point{X: 8, Y: 6})
	})
}

// rastererMit baut einen Rasterer, der seine gefragten Größen meldet.
func rastererMit(gemeldet *groessenmitschrift) *render.Rasterer {
	return render.NeuerRasterer(bilder{
		"/teil.png":  einfarbig(8, 6, color.NRGBA{R: 200, A: 255}),
		"/maske.png": einfarbig(8, 6, color.NRGBA{R: 200, A: 255}),
	}).MeldeGroesseFuerTest(gemeldet.nimm)
}

// groessenmitschrift hält fest, welche Größen gefragt wurden.
//
// **Mit Sperre**: ein Rasterer läuft heute je Bildbaum auf einer Kopie, und ein
// Test darf ihn nebenläufig benutzen (#748).
type groessenmitschrift struct {
	sperre  sync.Mutex
	gefragt []image.Point
}

func mitschreiber() *groessenmitschrift { return &groessenmitschrift{} }

func (g *groessenmitschrift) nimm(punkt image.Point) {
	g.sperre.Lock()
	defer g.sperre.Unlock()

	g.gefragt = append(g.gefragt, punkt)
}

func (g *groessenmitschrift) erwarte(t *testing.T, punkte ...image.Point) {
	t.Helper()

	g.sperre.Lock()
	defer g.sperre.Unlock()

	if len(g.gefragt) != len(punkte) {
		t.Fatalf("%d Läufe gemeldet, erwartet %d: %v", len(g.gefragt), len(punkte), g.gefragt)
	}

	for i, punkt := range punkte {
		if g.gefragt[i] != punkt {
			t.Errorf("%d. Lauf fragte %v, erwartet %v", i+1, g.gefragt[i], punkt)
		}
	}
}
