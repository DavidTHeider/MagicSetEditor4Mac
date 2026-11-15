package render_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Der Rasterer gibt seinen gesetzten Zug heraus** —
// #1029.
//
// Ohne diesen Weg steht der Umbruch nirgends: `Ebene` trägt die **Eingabe**,
// und das Ergebnis lebte und starb im Zeichnen. Ein Messlauf, der wissen will,
// **wo** eine Zeile steht, kam nicht daran.
func TestDerRastererGibtSeinenGesetztenZugHeraus(t *testing.T) {
	t.Parallel()

	ebene := render.Ebene{
		Art:        render.ArtText,
		Rahmen:     render.Rechteck{Breite: 40, Hoehe: 60},
		Abschnitte: []render.Abschnitt{{Text: "eins zwei drei vier"}},
		Satz:       render.Satz{Schrift: "Aufrecht", Groesse: 10},
		Deckung:    1,
	}

	gesetzt, err := render.NeuerRasterer(nil).
		MitSchriften(schriften{"Aufrecht": einheitsschrift{}}).
		GesetzterZug(ebene)
	if err != nil {
		t.Fatalf("GesetzterZug: %v", err)
	}

	// Bei 5 Punkten je Zeichen und 40 Punkten Breite bricht der Zug um.
	if len(gesetzt.Zeilen) < 2 {
		t.Fatalf("%d Zeilen — erwartet war ein Umbruch: %+v",
			len(gesetzt.Zeilen), gesetzt.Zeilen)
	}

	// **Die Oberkanten sind ungerundet** — das ist der Punkt. Eine Zeile
	// steht nicht auf einem ganzen Punkt, und genau diese Nachkommastelle
	// entscheidet, ob der gezeichnete Abstand 16 oder 17 wird.
	if gesetzt.Zeilen[1].Oben <= gesetzt.Zeilen[0].Oben {
		t.Errorf("die zweite Zeile steht bei %.4f, die erste bei %.4f",
			gesetzt.Zeilen[1].Oben, gesetzt.Zeilen[0].Oben)
	}

	if gesetzt.Hoehe <= 0 {
		t.Errorf("die Höhe ist %.4f", gesetzt.Hoehe)
	}
}

// **Und er weist ab, was er nicht setzen kann.**
//
// Jeder dieser Fälle gäbe sonst einen leeren Zug, der wie ein Ergebnis
// aussieht — ein Messlauf läse null Zeilen und hielte es für den Satz.
func TestDerGesetzteZugWeistUnbrauchbaresAb(t *testing.T) {
	t.Parallel()

	text := render.Ebene{
		Art:        render.ArtText,
		Rahmen:     render.Rechteck{Breite: 40, Hoehe: 60},
		Abschnitte: []render.Abschnitt{{Text: "eins"}},
		Satz:       render.Satz{Schrift: "Aufrecht", Groesse: 10},
	}

	ohneNamen := text
	ohneNamen.Satz.Schrift = "  "

	fremde := text
	fremde.Satz.Schrift = "Gibtesnicht"

	for name, fall := range map[string]struct {
		rasterer *render.Rasterer
		ebene    render.Ebene
		erwartet string
	}{
		"keine Textebene": {
			render.NeuerRasterer(nil).MitSchriften(schriften{}),
			render.Ebene{Art: render.ArtBild},
			"malt Bild",
		},
		"ohne Schriftsuche": {
			render.NeuerRasterer(nil), text,
			"keine Schriftsuche",
		},
		"ohne Schriftnamen": {
			render.NeuerRasterer(nil).MitSchriften(schriften{}), ohneNamen,
			"ohne Schriftnamen",
		},
		"unbekannte Schrift": {
			render.NeuerRasterer(nil).MitSchriften(schriften{}), fremde,
			"nicht zu finden",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := fall.rasterer.GesetzterZug(fall.ebene)
			if err == nil {
				t.Fatal("angenommen, obwohl der Zug nicht zu setzen ist")
			}

			if !strings.Contains(err.Error(), fall.erwartet) {
				t.Errorf("Meldung %q nennt nicht %q", err, fall.erwartet)
			}
		})
	}
}
