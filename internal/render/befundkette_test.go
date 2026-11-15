package render_test

// **Ein Befund aus einem Kind geht nicht verloren**
// (#473).
//
// Gefunden vom Zweitblick durch Codex (2026-08-18) als **konstruierte**
// Gegenprobe: sie folgt aus dem Code, war aber nie ausgeführt. Deshalb steht
// sie hier zuerst als roter Test.
//
// # Der Befund
//
// Jede Einbild-Umformung holt ihr Kind so:
//
//	unten, tiefer := r.ausBaumTief(baum.Kinder[0].Baum, tiefe+1)
//	if unten == nil {
//	    return nil, tiefer
//	}
//
// `tiefer` wird **nur im Fehlschlag** weitergereicht. Ein Kind, das ein Bild
// *und* einen Befund liefert — `recolor_image` ohne Farbe gibt sein
// Ausgangsbild zurück und meldet den Mangel —, verliert seinen Befund an
// jedem Elternteil, das gelingt.
//
// # Warum das materiell ist
//
// §8: fremde Daten sind der Normalfall. Ein halb geschriebenes Paket soll
// benannte Befunde liefern, nicht ein plausibles Teilbild ohne Warnung.

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// TestEinBefundAusEinemKindUeberlebtDasElternteil ist genau die Verschachtelung
// aus dem Zweitblick: `saturate(recolor_image("a.png"), amount: 0.5)`.
func TestEinBefundAusEinemKindUeberlebtDasElternteil(t *testing.T) {
	quelle := maskenquelle{bilder: map[string]image.Image{
		"a.png": maskenflaeche(4, 4, color.NRGBA{R: 200, G: 100, B: 50, A: 255}),
	}}

	// **`recolor_image` ohne `red:` und Geschwister** gibt sein Ausgangsbild
	// zurück und meldet „recolor_image braucht …".
	ohneFarben := render.Bildbaum{
		Knotenart: "recolor_image",
		Kinder:    []render.Bildkind{{Name: "input", Baum: render.Blatt("a.png")}},
	}

	faelle := []struct {
		name string
		baum render.Bildbaum
	}{
		{
			"saturate darüber",
			render.Bildbaum{
				Knotenart: "saturate",
				Zahlen:    map[string]float64{"amount": 0.5},
				Kinder:    []render.Bildkind{{Name: "input", Baum: ohneFarben}},
			},
		},
		{
			"invert_image darüber",
			render.Bildbaum{
				Knotenart: "invert_image",
				Kinder:    []render.Bildkind{{Name: "input", Baum: ohneFarben}},
			},
		},
		{
			"set_alpha darüber",
			render.Bildbaum{
				Knotenart: "set_alpha",
				Zahlen:    map[string]float64{"alpha": 0.5},
				Kinder:    []render.Bildkind{{Name: "input", Baum: ohneFarben}},
			},
		},
		{
			"enlarge darüber",
			render.Bildbaum{
				Knotenart: "enlarge",
				Zahlen:    map[string]float64{"border_size": 0.25},
				Kinder:    []render.Bildkind{{Name: "input", Baum: ohneFarben}},
			},
		},
		{
			"zwei Ebenen darüber",
			render.Bildbaum{
				Knotenart: "flip_horizontal",
				Kinder: []render.Bildkind{{Name: "input", Baum: render.Bildbaum{
					Knotenart: "saturate",
					Zahlen:    map[string]float64{"amount": 0.5},
					Kinder:    []render.Bildkind{{Name: "input", Baum: ohneFarben}},
				}}},
			},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			plan := &render.Plan{Breite: 4, Hoehe: 4, Ebenen: []render.Ebene{{
				Art:     render.ArtBild,
				Rahmen:  render.Rechteck{Breite: 4, Hoehe: 4},
				Bild:    fall.baum,
				Deckung: 1,
			}}}

			bild, befunde, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
			if err != nil {
				t.Fatalf("zeichnen: %v", err)
			}

			if len(befunde) == 0 {
				t.Fatal("kein Befund — der des Kindes ist verlorengegangen")
			}

			if got := befunde[0].String(); !strings.Contains(got, "recolor_image") {
				t.Errorf("der Befund lautet %q, erwartet einer über recolor_image", got)
			}

			// **Und gezeichnet ist trotzdem worden.** Ein Befund ersetzt das
			// Bild nicht — sichtbar falsch schlägt unsichtbar fehlend (§8).
			if bild.NRGBAAt(1, 1).A == 0 {
				t.Error("nichts gezeichnet — der Befund hat das Bild verdrängt")
			}
		})
	}
}
