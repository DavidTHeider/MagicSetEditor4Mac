package render_test

import (
	"errors"
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

var errKeinBild = errors.New("das Bild ist in keinem Paket zu finden")

// bildquelle gibt zu einem Pfad ein einfarbiges Bild.
type bildquelle struct{ da map[string]color.RGBA }

func (b bildquelle) Bild(pfad string) (image.Image, error) {
	farbe, gibtEs := b.da[pfad]
	if !gibtEs {
		return nil, errKeinBild
	}

	// **Acht Punkte breit, nicht zwei.** Ein Verlauf über zwei Punkte hat keine
	// Mitte: bei MSEs Formel (`x`, nicht `x + 0,5`) liegt der rechte Punkt
	// genau dort, und jede Aussage über „links gegen rechts" prüfte dann die
	// Breite des Prüfbilds statt der Regel.
	const seite = 8

	bild := image.NewRGBA(image.Rect(0, 0, seite, seite))
	for y := range seite {
		for x := range seite {
			bild.SetRGBA(x, y, farbe)
		}
	}

	return bild, nil
}

// **Ein Bildbaum wird ausgeführt, und was noch nicht geht, wird benannt.**
//
// # Die Zusage, um die es geht
//
// Vor dem Bildbaum trug der Plan einen **Pfad**, und `recolor_image` hatte
// nichts zurückzugeben: 965 der 2.729 Befunde über `M15-Magic-Pack` lauteten
// auf diesen einen Namen. Mit dem Baum sind es **null** — die Bildebenen sind
// von 795 auf 1.760 gestiegen
// (#270).
//
// **`recolor_image` ist seither gebaut**, und dieser Test steht deshalb auf
// `rotate`: er prüft die Zusage, nicht die eine Verknüpfung. Wer den
// letzten ungebauten Namen einbaut, sucht sich den nächsten aus
// `generated_image.hpp` — oder löscht diesen Test, weil es nichts mehr gibt,
// wofür er gälte.
//
// **Er stand bis zum 2026-08-18 auf `drop_shadow`**, und der ist jetzt gebaut
// (#472).
// `rotate` ist der nächste: im vollen Pack steht **kein bestätigter Aufruf**,
// und deshalb ist er benannt abgelehnt statt gebaut.
//
// # Was hier geprüft wird, und was nicht
//
// Dass eine **nicht gebaute** Verknüpfung ihr Ausgangsbild zeichnet und den
// Befund mitschickt. **Nicht**, dass sie richtig einfärbt — das ist die Arbeit
// danach, je Verknüpfung mit eigenem Goldbildvergleich (§3).
func TestEineNichtGebauteVerknuepfungZeichnetIhrAusgangsbild(t *testing.T) {
	rot := color.RGBA{R: 255, A: 255}

	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{"/a.png": rot}})

	plan := &render.Plan{Breite: 2, Hoehe: 2, Ebenen: []render.Ebene{{
		Art:    render.ArtBild,
		Rahmen: render.Rechteck{Breite: 2, Hoehe: 2},
		Bild: render.Bildbaum{
			// Eine erfundene Art — der Grund steht bei [nochNichtGebaut].
			Knotenart: nochNichtGebaut,
			Kinder:    []render.Bildkind{{Name: "input", Baum: render.Blatt("/a.png")}},
			Zahlen:    map[string]float64{"blur_radius": 0.5},
		},
		Deckung: 1,
	}}}

	bild, befunde, err := rasterer.Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	// **Gezeichnet ist das Ausgangsbild** — voll deckend: die Deckung zu
	// setzen kann dieses Programm noch nicht.
	r, g, b, _ := bild.At(0, 0).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("der Bildpunkt ist (%d,%d,%d), erwartet das rote Ausgangsbild",
			r>>8, g>>8, b>>8)
	}

	// **Und der Befund reist mit.** Wer die Zahl der Befunde als Maß nimmt,
	// soll nicht dadurch besser dastehen, dass etwas halb getan wurde.
	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet einer: %v", len(befunde), befunde)
	}

	if got := befunde[0].String(); !strings.Contains(got, nochNichtGebaut) || !strings.Contains(got, "/a.png") {
		t.Errorf("der Befund lautet %q und nennt weder Verknüpfung noch Ausgangsbild", got)
	}
}

// **Ein Baum ohne Kind ist ein Fehler, kein leeres Bild.**
func TestEineVerknuepfungOhneAusgangsbildWirdBenannt(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{})

	plan := &render.Plan{Breite: 2, Hoehe: 2, Ebenen: []render.Ebene{{
		Art:     render.ArtBild,
		Rahmen:  render.Rechteck{Breite: 2, Hoehe: 2},
		Bild:    render.Bildbaum{Knotenart: "masked_blend"},
		Deckung: 1,
	}}}

	_, befunde, err := rasterer.Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].String(), "kein Ausgangsbild") {
		t.Errorf("die Befunde lauten %v, erwartet einer über das fehlende Ausgangsbild", befunde)
	}
}

// **Ein unlesbares Blatt tief im Baum nennt seinen Pfad.**
func TestEinUnlesbaresBlattImBaumNenntSeinenPfad(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{})

	plan := &render.Plan{Breite: 2, Hoehe: 2, Ebenen: []render.Ebene{{
		Art:    render.ArtBild,
		Rahmen: render.Rechteck{Breite: 2, Hoehe: 2},
		Bild: render.Bildbaum{
			Knotenart: "recolor_image",
			Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/fehlt.png")}},
		},
		Deckung: 1,
	}}}

	_, befunde, err := rasterer.Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].String(), "/fehlt.png") {
		t.Errorf("die Befunde lauten %v, erwartet einer über /fehlt.png", befunde)
	}
}

// **Die Beschreibung nennt das tiefste Blatt** — sie steht in jedem Befund.
func TestDieBeschreibungEinesBaumsNenntSeinBlatt(t *testing.T) {
	faelle := map[string]render.Bildbaum{
		"kein Bild":             {},
		"/a.png":                render.Blatt("/a.png"),
		"masked_blend":          {Knotenart: "masked_blend"},
		"recolor_image(/a.png)": {Knotenart: "recolor_image", Kinder: []render.Bildkind{{Name: "image", Baum: render.Blatt("/a.png")}}},
	}

	for erwartet, baum := range faelle {
		if got := baum.Beschreibung(); got != erwartet {
			t.Errorf("%+v beschreibt sich als %q, erwartet %q", baum, got, erwartet)
		}
	}
}

// **Eine Ebene ohne Bild wird benannt.**
func TestEineEbeneOhneBildWirdBenannt(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{})

	// **Unmittelbar und nicht über `Zeichne`**: die Planprüfung fängt den Fall
	// schon vorher ab, und dann bliebe dieser Zweig ungeprüft — ein Weg, den
	// kein Aufruf betritt, ist entweder tot oder eine Lücke (§3). Hier ist er
	// die zweite Verteidigungslinie.
	if _, befund := rasterer.AusBaumFuerTest(render.Bildbaum{}); befund == "" {
		t.Error("eine Ebene ohne Bild ergibt keinen Befund")
	}
}

// **Eine nicht gebaute Verknüpfung ohne lesbares Ausgangsbild zeichnet nichts.**
//
// Der Rückfall ist das Ausgangsbild — gibt es auch das nicht, bleibt der
// Befund von unten stehen und nennt die Datei. Ein leeres Bild mit der Meldung
// „noch nicht gebaut" verschwiege den eigentlichen Mangel (§8).
func TestEineNichtGebauteVerknuepfungOhneAusgangsbild(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{}})

	bild, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: nochNichtGebaut,
		Kinder:    []render.Bildkind{{Name: "input", Baum: render.Blatt("/fehlt.png")}},
	})

	if bild != nil || !strings.Contains(befund, "/fehlt.png") {
		t.Errorf("Bild %v, Befund %q", bild != nil, befund)
	}
}
