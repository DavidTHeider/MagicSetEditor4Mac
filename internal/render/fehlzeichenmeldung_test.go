package render_test

// **Ein Befund muss zeigen, was er gefunden hat**
// (#476).
//
// Beim Bauen des systematischen Demokartensatzes stand da:
//
//	Rasterbefund: in der Schrift MPlantin fehlen Zeichen:
//
// **Ohne ein einziges Zeichen dahinter.** Die Meldung ist entstanden, weil ein
// Zeichen keine Maske hergab — aber es hat keine Gestalt, und roh in die
// Meldung geschrieben ist es unsichtbar. Wer das liest, kann nichts tun.

import (
	"image"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// maskenlos gibt für ein bestimmtes Zeichen keine Maske her.
type maskenlos struct {
	ohne rune
}

func (maskenlos) Breite(rune) float64  { return 0.5 }
func (maskenlos) Zeilenhoehe() float64 { return 1.2 }
func (maskenlos) Grundlinie() float64  { return 0.8 }

func (m maskenlos) Suche(string, bool) (render.Schrift, error) { return m, nil }

func (m maskenlos) Maske(zeichen rune, groesse float64) (image.Image, image.Point, bool) {
	if zeichen == m.ohne {
		return nil, image.Point{}, false
	}

	return einheitsschrift{}.Maske(zeichen, groesse)
}

func TestEinFehlendesZeichenIstImBefundZuErkennen(t *testing.T) {
	faelle := []struct {
		name    string
		zeichen rune
		text    string
		nennt   string
	}{
		// **Ein sichtbares Zeichen** steht wie bisher da.
		{"ein Buchstabe", 'q', "aqb", `"q"`},

		// **Ein unsichtbares** braucht seine Nummer, sonst steht in der
		// Meldung nichts.
		{"ein geschütztes Leerzeichen", ' ', "a b", "U+00A0"},
		{"ein schmales Leerzeichen", ' ', "a b", "U+2009"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			plan := &render.Plan{Breite: 60, Hoehe: 20, Ebenen: []render.Ebene{{
				Art:        render.ArtText,
				Rahmen:     render.Rechteck{Breite: 60, Hoehe: 20},
				Abschnitte: []render.Abschnitt{{Text: fall.text}},
				Satz: render.Satz{
					Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben,
				},
				Deckung: 1,
			}}}

			_, befunde, err := render.NeuerRasterer(bildmasse{}).
				MitSchriften(schriften{"Probe": maskenlos{ohne: fall.zeichen}}).
				Zeichne(t.Context(), plan, 1, render.GueteMittel)
			if err != nil {
				t.Fatalf("zeichnen: %v", err)
			}

			if len(befunde) != 1 {
				t.Fatalf("%d Befunde, erwartet einen: %v", len(befunde), befunde)
			}

			if !strings.Contains(befunde[0].Meldung, fall.nennt) {
				t.Errorf("die Meldung %q nennt %s nicht",
					befunde[0].Meldung, fall.nennt)
			}
		})
	}
}

// **Ein Zeilenumbruch wird nie gezeichnet**
// (#476).
//
// MSE2 entfernt den Umbruch, der die Zeile gebrochen hat — `// remove the '\n'
// or ' ' that caused the break`, `src/render/text/viewer.cpp`, Zeilen 664–666,
// Fassung f9d9356 vom 2021-07-06.
//
// **Ein gestauchter Zug ging hier daran vorbei.** `SetzeAbschnitte` ersetzt
// für ihn die Absatzbildung durch die rohen Stücke (`abschnitte.go`, Zeilen
// 33–40), und darin blieb der Umbruch als Zeichen stehen. Gefunden am
// Spielstein des Demokartensatzes: ein leeres Feld, dessen einziger Abschnitt
// `"\n"` lautet, meldete `in der Schrift MPlantin fehlen Zeichen: U+000A`.
func TestEinGestauchterZugZeichnetKeinenUmbruch(t *testing.T) {
	faelle := []struct{ name, text string }{
		{"allein", "\n"},
		{"am Ende", "2/2\n"},
		{"in der Mitte", "2\n2"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			gesetzt := render.SetzeAbschnitte(
				render.Satz{
					Schrift: "Probe", Groesse: 10,
					Ausrichtung: render.LinksOben, Stauchen: true, Einzeilig: true,
				},
				render.Rechteck{Breite: 60, Hoehe: 20},
				[]render.Abschnitt{{Text: fall.text, Blockende: true}},
				render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

			for _, zeile := range gesetzt.Zeilen {
				for _, stueck := range zeile.Stuecke {
					if strings.ContainsRune(stueck.Text, '\n') {
						t.Errorf("das Stück %q trägt einen Umbruch", stueck.Text)
					}
				}
			}
		})
	}
}
