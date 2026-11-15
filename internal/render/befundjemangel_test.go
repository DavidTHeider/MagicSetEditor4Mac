package render_test

// **Ein Mangel, ein Befund — und jeder mit seiner eigenen Art** — QA-03
// (#1656).
// Zwei Wege, auf denen eine Folge der Umgebung als Portierung zählte oder
// eine Portierung unter der Umgebung verschwand:
//
//  1. **Die Schrift für den Text auf einem Symbol fehlte** — die Ziffer im
//     Manakreis —, und die Meldung hieß »die Symbolbilder fehlen: MPlantin«.
//     Kein Bild fehlte; `ArtAus` las »Symbolbilder« und zählte Portierung.
//     Ohne `MSE_SCHRIFTWURZEL` waren das 426 der 679 Portierungsbefunde aus
//     #1656.
//  2. **Zwei Mängel einer Textebene standen in einer Meldung**, und `ArtAus`
//     ordnete sie nach dem ersten Treffer: fehlende Zeichen (Umgebung)
//     verdeckten ein Symbolbild, das in keinem Paket liegt (Vorlage).

import (
	"image"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestEineFehlendeSchriftAufEinemSymbolIstEinUmgebungsbefund(t *testing.T) {
	_, befunde, err := render.NeuerRasterer(bilder{"/probe.mse-symbol-font/kreis.png": weissesFeld()}).
		MitSchriften(schriften{}).
		Zeichne(t.Context(), symbolplan("1"), 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet einen: %+v", len(befunde), befunde)
	}

	befund := befunde[0]

	if strings.Contains(befund.Meldung, "Symbolbild") ||
		!strings.Contains(befund.Meldung, "Andale") ||
		!strings.Contains(befund.Meldung, "ist nicht zu finden") {
		t.Errorf("die Meldung nennt nicht die Schrift: %q", befund.Meldung)
	}

	if befund.Art != render.ArtUmgebung {
		t.Errorf("Art %s, erwartet %s — eine fehlende Schriftdatei ist die Umgebung", befund.Art, render.ArtUmgebung)
	}
}

func TestZweiMaengelEinerEbeneSindZweiBefunde(t *testing.T) {
	plan := &render.Plan{Breite: 80, Hoehe: 20, Ebenen: []render.Ebene{{
		Art:    render.ArtText,
		Rahmen: render.Rechteck{Breite: 80, Hoehe: 20},
		Abschnitte: []render.Abschnitt{
			{Text: "aq"},
			{
				Text:        "R",
				Symbole:     []render.Symbol{{Bild: render.Blatt("/fehlt.png")}},
				Symbolmasse: render.Symbolmasse{Groesse: 10, Bildgroesse: 10},
			},
		},
		Satz:    render.Satz{Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben},
		Deckung: 1,
	}}}

	_, befunde, err := render.NeuerRasterer(bildmasse{}).
		MitSchriften(schriften{"Probe": maskenlos{ohne: 'q'}}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	arten := map[render.Befundart]string{}
	for _, befund := range befunde {
		arten[befund.Art] = befund.Meldung
	}

	if len(befunde) != 2 ||
		!strings.Contains(arten[render.ArtUmgebung], "fehlen Zeichen") ||
		!strings.Contains(arten[render.ArtVorlage], "Symbolbilder fehlen") {
		t.Errorf("erwartet zwei Befunde, Zeichen als Umgebung und Bild als Vorlage: %+v", befunde)
	}
}

// **Fehlende Symbolbilder je Art, dieselbe Schrift einmal.** Ein Bild, das in
// keinem Paket liegt (Vorlage), und eines, dessen Verknüpfung hier fehlt
// (Portierung), stehen in einem Zug — und zwei Symbole vermissen dieselbe
// Schrift. Drei Befunde, jeder mit seiner Art, keiner doppelt.
func TestEinSymbolzugTrenntSeineMaengelNachArt(t *testing.T) {
	nichtGebaut := render.Bildbaum{
		Knotenart: nochNichtGebaut,
		Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/da.png")}},
	}
	mitZiffer := func(ziffer string) render.Symbol {
		return render.Symbol{
			Bild: render.Blatt("/da.png"), Text: ziffer,
			Textsatz: render.Symboltextsatz{Schrift: "Fehlt", Groesse: 0.6},
		}
	}

	plan := &render.Plan{Breite: 80, Hoehe: 20, Ebenen: []render.Ebene{{
		Art:     render.ArtSymbolzug,
		Rahmen:  render.Rechteck{Breite: 80, Hoehe: 20},
		Deckung: 1,
		Satz:    render.Satz{Groesse: 10},
		Abschnitte: []render.Abschnitt{{
			Symbolmasse: render.Symbolmasse{Groesse: 10, Bildgroesse: 10},
			Symbole: []render.Symbol{
				{Bild: render.Blatt("/fehlt.png")},
				{Bild: nichtGebaut},
				mitZiffer("1"), mitZiffer("2"),
			},
		}},
	}}}

	_, befunde, err := render.NeuerRasterer(bildmasse{da: map[string]image.Point{"/da.png": {X: 10, Y: 10}}}).
		MitSchriften(schriften{}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	arten := map[render.Befundart][]string{}
	for _, befund := range befunde {
		arten[befund.Art] = append(arten[befund.Art], befund.Meldung)
	}

	if len(befunde) != 3 || len(arten[render.ArtVorlage]) != 1 || len(arten[render.ArtPortierung]) != 1 ||
		len(arten[render.ArtUmgebung]) != 1 ||
		!strings.Contains(arten[render.ArtVorlage][0], "/fehlt.png") ||
		!strings.Contains(arten[render.ArtPortierung][0], nochNichtGebaut) ||
		!strings.Contains(arten[render.ArtUmgebung][0], "die Schrift Fehlt") {
		t.Errorf("erwartet je einen Befund der drei Arten: %v", arten)
	}
}
