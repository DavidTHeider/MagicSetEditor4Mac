package render_test

import (
	"context"
	"image/color"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Ein Rasterer bedient mehrere Wege gleichzeitig** —
// #748.
//
// # Der Befund
//
// `internal/app` baut **einen** `render.Rasterer` und reicht ihn an Vorschau,
// Symbolpalette, Bildausfuhr, Druckbogen und Vorlagenausfuhr. Die Oberfläche
// lässt zwei Ausfuhren **verschiedener Sets** ausdrücklich parallel laufen —
// `useVorgang` sperrt über `ausfuehren-${satz.id}`, und
// #741 hat
// genau diese Parallelität wiederhergestellt.
//
// Der Rasterer trug bis dahin die Gegenzusage im Klartext: „Ein Rasterer
// zeichnet nacheinander, nicht nebenläufig … **wer das ändert, ändert es hier
// mit.**" Geändert hat es #741, mitgeändert wurde hier nichts.
//
// # Was dieser Wächter prüft
//
// **Zwei echte Rasterwege**, nicht zwei Attrappen: eine Karte über
// [render.Rasterer.Zeichne] und ein Bildwert über
// [render.Rasterer.BildAusBaum], beide auf **demselben** Rasterer und beide
// mehrfach. Er läuft mit `-race`, und ohne die Behebung meldet der Melder den
// ungesicherten Zugriff auf `gefragteGroesse`, `baumbefunde` und den
// Ebenenspeicher.
//
// **Was er nicht zuverlässig erreicht, ist der Ebenenspeicher.** Ob eine Ebene
// gemerkt wird, hängt am Bildtyp, an der Größe und daran, ob der Eintrag schon
// dasteht — beim Bauen dieses Tests lief `merkeEbenenbild` dreimal
// hintereinander **null**-mal, und die Gegenprobe ohne Sperre blieb deshalb
// grün. Die Ablage hat ihren eigenen Wächter
// (`ebenenspeicherlauf_intern_test.go`); dieser hier steht für die Lauffelder.
//
// **Der Unterschied zu `TestZweiAusfuhrenKreuzenSichNicht`** (`internal/app`):
// der prüft die Naht darüber und benutzt eine Attrappe, die den Rasterer nicht
// berührt. Ein Wächter, der einen Weg nicht betritt, sagt über ihn nichts (§10).
func TestEinRastererTraegtZweiLaeufeNebeneinander(t *testing.T) {
	t.Parallel()

	rasterer := render.NeuerRasterer(bilder{"/rahmen.png": einfarbig(8, 6, color.NRGBA{R: 200, G: 30, B: 30, A: 255})})

	// **Ein Plan mit einer Bildebene**, damit der Ebenenspeicher betreten wird —
	// er ist der Zustand, dessen Kreuzung nicht nur ein falsches Bild, sondern
	// einen Abbruch ergäbe.
	plan := &render.Plan{
		Breite: 30, Hoehe: 20,
		Ebenen: []render.Ebene{{
			Art:     render.ArtBild,
			Deckung: 1,
			Rahmen:  render.Rechteck{Breite: 30, Hoehe: 20},
			Bild:    render.Bildbaum{Knotenart: "datei", Texte: map[string]string{"pfad": "/rahmen.png"}},
		}},
	}

	// **Ein leeres Bild** liest `gefragteGroesse` — das zweite geteilte Feld
	// (`bildwerkzeuge.go`, `blank_image`).
	leer := render.Bildbaum{Knotenart: "blank_image"}

	var warte sync.WaitGroup

	warte.Add(2)

	go func() {
		defer warte.Done()

		// **Auch hier wechselt die Vergrößerung**, in einer anderen Folge als
		// im zweiten Lauf.
		for stelle := range 40 {
			if _, _, err := rasterer.Zeichne(context.Background(), plan,
				1+float64(stelle)/100, render.GueteMittel); err != nil {
				t.Errorf("zeichnen: %v", err)

				return
			}
		}
	}()

	go func() {
		defer warte.Done()

		for stelle := range 40 {
			// **Ein Bildwert liest `gefragteGroesse`** — das eine geteilte Feld.
			bild, _ := rasterer.BildAusBaum(leer, 3+stelle, 5+stelle)
			if bild == nil {
				t.Error("kein Bild aus dem Baum")

				return
			}

			// **Und eine zweite Karte in wechselnder Vergrößerung.**
			if _, _, err := rasterer.Zeichne(context.Background(), plan,
				1+float64(stelle)/40, render.GueteMittel); err != nil {
				t.Errorf("zweite Karte: %v", err)

				return
			}
		}
	}()

	warte.Wait()
}
