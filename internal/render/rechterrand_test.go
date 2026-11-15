package render_test

// **Der rechte Absatzrand verkürzt den Umbruch nicht** (Befund 2 des
// Nachlaufs über E/F).
//
// MSE prüft beim Umbruch gegen `lineRight` — den rechten Rand der Maske
// abzüglich `padding_right`, **ohne** `margin_right`
// (`TextViewer::prepareLinesAtScale`, `src/render/text/viewer.cpp`,
// Zeilen 614–616, und `lineRight`, Zeilen 739–741). Der rechte Absatzrand
// wirkt beim Abrücken einer Zeile (`fitLineWidth`, Zeile 549) und beim
// Ausrichten (`Line::alignHorizontal`, Zeile 835). Fassung f9d9356 vom
// 2021-07-06.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestDerRechteAbsatzrandVerkuerztDenUmbruchNicht(t *testing.T) {
	for _, fall := range []struct {
		name   string
		satz   render.Satz
		quelle render.Bildquelle
		rahmen render.Rechteck
		text   string
	}{
		// „ab cd" ist 25 breit: es passt in 40, nicht aber in 40 - 20.
		{name: "ohne Maske", rahmen: render.Rechteck{Breite: 40, Hoehe: 100}, text: "ab cd"},
		// Die Maske ist links bis Spalte 49 weiß; „aaaa aaaa" ist 45 breit.
		{
			name: "mit Maske", satz: render.Satz{Zeilenmaske: "maske.png"},
			quelle: zeilenmaskenquelle{links, links, links, links, links, links},
			rahmen: render.Rechteck{Breite: 100, Hoehe: 60}, text: "aaaa aaaa",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			satz := fall.satz
			satz.Schrift, satz.Groesse, satz.Ausrichtung = "Probe", 10, render.LinksOben

			gesetzt := render.SetzeAbschnitte(satz, fall.rahmen,
				[]render.Abschnitt{{Text: fall.text, Margen: &render.Margen{Rechts: 20}}},
				render.Schriftsatz{Aufrecht: einheitsschrift{}}, fall.quelle)

			zeilen := make([]string, 0, len(gesetzt.Zeilen))
			for _, zeile := range gesetzt.Zeilen {
				zeilen = append(zeilen, zeile.Text)
			}

			if len(zeilen) != 1 {
				t.Errorf("%d Zeilen %q, erwartet eine", len(zeilen), strings.Join(zeilen, "|"))
			}
		})
	}
}
