// **Passung, Auflösung und die beiden Wächter am Anfang jeder Ebene** —
// #852.
package svg_test

import (
	"image"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/svg"
)

// TestEinePassungWirdZuPreserveAspectRatio prüft alle neun Ausrichtungen.
func TestEinePassungWirdZuPreserveAspectRatio(t *testing.T) {
	t.Parallel()

	faelle := map[render.Ausrichtung]string{
		render.LinksOben:   "xMinYMin meet",
		render.MitteOben:   "xMidYMin meet",
		render.RechtsOben:  "xMaxYMin meet",
		render.LinksMitte:  "xMinYMid meet",
		render.MitteMitte:  "xMidYMid meet",
		render.RechtsMitte: "xMaxYMid meet",
		render.LinksUnten:  "xMinYMax meet",
		render.MitteUnten:  "xMidYMax meet",
		render.RechtsUnten: "xMaxYMax meet",
	}

	for ausrichtung, erwartet := range faelle {
		t.Run(string(ausrichtung), func(t *testing.T) {
			t.Parallel()

			bildebene := render.Ebene{
				Art:     render.ArtBild,
				Feld:    "art",
				Rahmen:  render.Rechteck{Links: 5, Oben: 5, Breite: 40, Hoehe: 30},
				Bild:    render.Blatt("/probe.mse-style/art.png"),
				Passung: &render.Bildpassung{Ausrichtung: ausrichtung},
				Deckung: 1,
			}

			geschrieben, _, fehler := svg.AusPlan(t.Context(), eineEbene(bildebene),
				svg.Quellen{Baeume: baumprobe{}})
			if fehler != nil {
				t.Fatalf("schreiben: %v", fehler)
			}

			if !strings.Contains(geschrieben, `preserveAspectRatio="`+erwartet+`"`) {
				t.Errorf("%q fehlt im SVG:\n%s", erwartet, geschrieben)
			}
		})
	}
}

// TestOhnePassungWirdGestreckt hält MSEs `ASPECT_STRETCH`.
func TestOhnePassungWirdGestreckt(t *testing.T) {
	t.Parallel()

	bildebene := render.Ebene{
		Art:     render.ArtBild,
		Feld:    "art",
		Rahmen:  render.Rechteck{Links: 5, Oben: 5, Breite: 40, Hoehe: 30},
		Bild:    render.Blatt("/probe.mse-style/art.png"),
		Deckung: 1,
	}

	geschrieben, _, fehler := svg.AusPlan(t.Context(), eineEbene(bildebene),
		svg.Quellen{Baeume: baumprobe{}})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if !strings.Contains(geschrieben, `preserveAspectRatio="none"`) {
		t.Errorf("ohne Passung wird nicht gestreckt:\n%s", geschrieben)
	}
}

// gemerkteGroesse merkt sich, in welcher Größe ein Bild gefragt wurde.
type gemerkteGroesse struct{ breite, hoehe int }

func (g *gemerkteGroesse) BildAusBaum(_ render.Bildbaum, breite, hoehe int) (
	bild *image.NRGBA, befund string,
) {
	g.breite, g.hoehe = breite, hoehe

	return image.NewNRGBA(image.Rect(0, 0, 2, 2)), ""
}

// TestDieVergroesserungGiltFuerDieEingebettetenBilder prüft die Auflösung.
//
// **Sie gilt nur für sie.** Flächen und Text bleiben vektoriell; ein
// Kartenrahmen ist ein Bild, und wer eine SVG in Druckgröße will, braucht ihn
// in Druckauflösung.
func TestDieVergroesserungGiltFuerDieEingebettetenBilder(t *testing.T) {
	t.Parallel()

	gemerkt := &gemerkteGroesse{}

	bildebene := render.Ebene{
		Art:     render.ArtBild,
		Feld:    "art",
		Rahmen:  render.Rechteck{Breite: 40, Hoehe: 30},
		Bild:    render.Blatt("/probe.mse-style/art.png"),
		Deckung: 1,
	}

	if _, _, fehler := svg.AusPlan(t.Context(), eineEbene(bildebene),
		svg.Quellen{Baeume: gemerkt, Vergroesserung: 4}); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if gemerkt.breite != 160 || gemerkt.hoehe != 120 {
		t.Errorf("gefragt wurde %d×%d, erwartet waren 160×120",
			gemerkt.breite, gemerkt.hoehe)
	}

	// **Und der Rahmen im SVG bleibt in Vorlagenpunkten** — sonst säße das
	// Bild viermal zu groß auf der Karte.
	gemerkt = &gemerkteGroesse{}
	if _, _, fehler := svg.AusPlan(t.Context(), eineEbene(bildebene),
		svg.Quellen{Baeume: gemerkt}); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if gemerkt.breite != 40 || gemerkt.hoehe != 30 {
		t.Errorf("ohne Vergrößerung wurde %d×%d gefragt, erwartet waren 40×30",
			gemerkt.breite, gemerkt.hoehe)
	}
}

// TestEinLeererPlanGibtEinLeeresSVG hält den Randfall.
func TestEinLeererPlanGibtEinLeeresSVG(t *testing.T) {
	t.Parallel()

	geschrieben, befunde, fehler := svg.AusPlan(t.Context(),
		&render.Plan{Breite: 10, Hoehe: 10}, svg.Quellen{})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("unerwartete Befunde: %v", befunde)
	}

	if !strings.HasSuffix(geschrieben, "</svg>") || strings.Contains(geschrieben, "<defs>") {
		t.Errorf("ein leerer Plan gibt kein leeres SVG:\n%s", geschrieben)
	}
}

// TestEineUnsichtbareEbeneFehlt hält `render.Sichtbar`.
func TestEineUnsichtbareEbeneFehlt(t *testing.T) {
	t.Parallel()

	ohneFlaeche := flaeche()
	ohneFlaeche.Rahmen.Breite = 0

	geschrieben, _, fehler := svg.AusPlan(t.Context(), eineEbene(ohneFlaeche), svg.Quellen{})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if strings.Contains(geschrieben, "<rect") {
		t.Errorf("die unsichtbare Ebene steht im SVG:\n%s", geschrieben)
	}
}

// TestEineArtOhneZeichnung faellt nicht um.
//
// **Der Zweig gehört dazu**: `render.Ebenenart` ist eine Zeichenkette, und ein
// Plan aus einer künftigen Fassung kann eine Art tragen, die es hier noch
// nicht gibt.
func TestEineUnbekannteArtFaelltNichtUm(t *testing.T) {
	t.Parallel()

	fremd := flaeche()
	fremd.Art = render.Ebenenart("Hologramm")

	geschrieben, befunde, fehler := svg.AusPlan(t.Context(), eineEbene(fremd), svg.Quellen{})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(befunde) != 0 || strings.Contains(geschrieben, "<rect") {
		t.Errorf("eine unbekannte Art wurde gezeichnet: %v\n%s", befunde, geschrieben)
	}
}

// TestEinSchraegerWinkelWirdGenannt prüft die Abweichung zum Rasterer.
func TestEinSchraegerWinkelWirdGenannt(t *testing.T) {
	t.Parallel()

	schraeg := flaeche()
	schraeg.Winkel = 30

	geschrieben, befunde, fehler := svg.AusPlan(t.Context(), eineEbene(schraeg), svg.Quellen{})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "nicht gerade") {
		t.Errorf("der schräge Winkel wurde nicht genannt: %v", befunde)
	}

	// **Gedreht wird er trotzdem** — SVG kann es, MSE tut es.
	if !strings.Contains(geschrieben, "rotate(-30)") {
		t.Errorf("der schräge Winkel wurde nicht gedreht:\n%s", geschrieben)
	}
}

// TestEinGeraderWinkelWirdNichtGenannt ist die Gegenprobe dazu.
func TestEinGeraderWinkelWirdNichtGenannt(t *testing.T) {
	t.Parallel()

	for _, grad := range []float64{90, 180, 270, 360, -90, 270.0000001} {
		gerade := flaeche()
		gerade.Winkel = grad

		_, befunde, fehler := svg.AusPlan(t.Context(), eineEbene(gerade), svg.Quellen{})
		if fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}

		if len(befunde) != 0 {
			t.Errorf("%g Grad gilt als schräg: %v", grad, befunde)
		}
	}
}

// TestDieZielgroesseStehtAmSVGUndDerPlanImViewBox trennt beide Maße.
//
// **`width` ist die bestellte Größe, `viewBox` die des Plans.** Damit skalieren
// Flächen und Text mit, ohne dass irgendeine Stelle im Dokument umgerechnet
// werden müsste — und die Rechnung steht an einer Stelle statt an vierzig.
func TestDieZielgroesseStehtAmSVGUndDerPlanImViewBox(t *testing.T) {
	t.Parallel()

	geschrieben, _, fehler := svg.AusPlan(t.Context(),
		&render.Plan{Breite: 100, Hoehe: 140}, svg.Quellen{Vergroesserung: 7.5})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if !strings.Contains(geschrieben, `width="750" height="1050"`) {
		t.Errorf("die Zielgröße fehlt:\n%s", geschrieben)
	}

	if !strings.Contains(geschrieben, `viewBox="0 0 100 140"`) {
		t.Errorf("der viewBox trägt nicht den Plan:\n%s", geschrieben)
	}
}
