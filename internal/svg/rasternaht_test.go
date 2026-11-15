// **Was SVG nicht ausdrückt, geht durch den Rasterer** —
// #852.
//
// Zwei Ebenenarten: der Symbolzug und ein Textzug mit Symbolen. Wohin ein
// `<sym>` mitten im Regeltext kommt, weiß erst der Umbruch; die Symbolmaschine
// des Rasterers ein zweites Mal zu bauen wäre die zweite Rechnung über
// dieselbe Sache.
package svg_test

import (
	"context"
	"errors"
	"image"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/svg"
)

// rasterprobe zeichnet einen Plan — oder scheitert.
type rasterprobe struct {
	fehler  error
	meldung string

	// ohneFlaeche gibt ein Bild ohne Ausdehnung — der einzige Weg, `png.Encode`
	// scheitern zu lassen.
	ohneFlaeche bool

	// gesehen hält den Plan fest, mit dem gezeichnet wurde.
	gesehen *render.Plan
}

func (r *rasterprobe) Zeichne(_ context.Context, plan *render.Plan, _ float64, _ render.Guete) (
	*image.NRGBA, []render.Befund, error,
) {
	r.gesehen = plan

	if r.fehler != nil {
		return nil, nil, r.fehler
	}

	var befunde []render.Befund
	if r.meldung != "" {
		befunde = []render.Befund{{Ebene: 0, Meldung: r.meldung}}
	}

	if r.ohneFlaeche {
		return image.NewNRGBA(image.Rect(0, 0, 0, 0)), befunde, nil
	}

	return image.NewNRGBA(image.Rect(0, 0, 4, 4)), befunde, nil
}

// symbolzugplan gibt einen Plan mit einer Symbolzugebene.
func symbolzugplan() *render.Plan {
	return &render.Plan{
		Breite: 100, Hoehe: 140,
		Ebenen: []render.Ebene{{
			Art:     render.ArtSymbolzug,
			Feld:    "casting cost",
			Rahmen:  render.Rechteck{Links: 10, Oben: 10, Breite: 60, Hoehe: 14},
			Satz:    render.Satz{Schrift: "Helvetica", Groesse: 12},
			Deckung: 1,
		}},
	}
}

func TestEinSymbolzugWirdGerastert(t *testing.T) {
	t.Parallel()

	raster := &rasterprobe{}

	geschrieben, befunde, fehler := svg.AusPlan(t.Context(), symbolzugplan(),
		svg.Quellen{Raster: raster, Schriften: suchprobe{}})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("unerwartete Befunde: %v", befunde)
	}

	if !strings.Contains(geschrieben, `href="data:image/png;base64,`) {
		t.Errorf("der Symbolzug steht nicht als Bild im SVG:\n%s", geschrieben)
	}

	// **Ein Plan mit einer Ebene, in Kartengröße.** Das Ergebnis ist außerhalb
	// der Ebene durchsichtig und wird deshalb an `0,0` gelegt — wo genau die
	// Symbole sitzen, hat der Rasterer schon entschieden.
	if raster.gesehen == nil || len(raster.gesehen.Ebenen) != 1 {
		t.Fatalf("der Teilplan trägt nicht genau eine Ebene: %v", raster.gesehen)
	}

	if raster.gesehen.Breite != 100 || raster.gesehen.Hoehe != 140 {
		t.Errorf("der Teilplan misst %g×%g statt der Kartengröße",
			raster.gesehen.Breite, raster.gesehen.Hoehe)
	}

	if !strings.Contains(geschrieben, `x="0" y="0" width="100" height="140"`) {
		t.Errorf("das Bild sitzt nicht auf der ganzen Karte:\n%s", geschrieben)
	}
}

// **Drehung, Deckung und Maske gehen nicht zweimal ein.**
//
// Sie hängen an der Hülle um die Ebene; die Ebene selbst geht ungedreht und
// voll deckend in den Rasterer. Dieselbe Überlegung wie in
// `render.ebeneGedreht` — was die Ebene an Deckung mitbringt, steht schon im
// Puffer.
func TestDieGerasterteEbeneTraegtIhreHuelleNichtZweimal(t *testing.T) {
	t.Parallel()

	plan := symbolzugplan()
	plan.Ebenen[0].Deckung = 0.5
	plan.Ebenen[0].Winkel = 90
	plan.Ebenen[0].Maske = render.Blatt("/probe.mse-style/maske.png")

	raster := &rasterprobe{}

	geschrieben, _, fehler := svg.AusPlan(t.Context(), plan,
		svg.Quellen{Raster: raster, Baeume: baumprobe{}, Schriften: suchprobe{}})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	innen := raster.gesehen.Ebenen[0]
	if innen.Deckung != 1 || innen.Winkel != 0 || !innen.Maske.Leer() {
		t.Errorf("die Ebene trägt ihre Hülle noch: Deckung %g, Winkel %g, Maske %v",
			innen.Deckung, innen.Winkel, innen.Maske)
	}

	// **Und die Hülle trägt sie einmal.**
	for _, teil := range []string{`opacity="0.5"`, "rotate(-90)", `mask="url(#`} {
		if !strings.Contains(geschrieben, teil) {
			t.Errorf("%q fehlt an der Hülle:\n%s", teil, geschrieben)
		}
	}
}

// **Ohne Rasterer wird es genannt.**
func TestOhneRastererBleibtEinBefund(t *testing.T) {
	t.Parallel()

	_, befunde, fehler := svg.AusPlan(t.Context(), symbolzugplan(),
		svg.Quellen{Schriften: suchprobe{}})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "kein Rasterer") {
		t.Errorf("kein Befund über den fehlenden Rasterer: %v", befunde)
	}
}

// **Ein Fehler beim Zeichnen wird genannt, nicht durchgereicht.**
//
// Eine Karte, von der eine Ebene fehlt, ist brauchbar; eine leere Datei ohne
// Meldung ist es nicht (§8).
func TestEinFehlerBeimZeichnenWirdGenannt(t *testing.T) {
	t.Parallel()

	_, befunde, fehler := svg.AusPlan(t.Context(), symbolzugplan(),
		svg.Quellen{
			Raster:    &rasterprobe{fehler: errors.New("kein Speicher")},
			Schriften: suchprobe{},
		})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "kein Speicher") {
		t.Errorf("der Zeichenfehler wurde verschluckt: %v", befunde)
	}
}

// **Die Befunde des Teilplans bekommen die richtige Ebenennummer.**
//
// Im Teilplan ist die Ebene die nullte; auf der Karte ist sie es meist nicht.
// Ohne diesen Schritt zeigte jeder Befund aus diesem Weg auf die erste Ebene.
func TestEinBefundDesTeilplansZeigtAufDieRichtigeEbene(t *testing.T) {
	t.Parallel()

	plan := symbolzugplan()
	plan.Ebenen = append([]render.Ebene{{
		Art:     render.ArtFlaeche,
		Rahmen:  render.Rechteck{Breite: 100, Hoehe: 140},
		Deckung: 1,
	}}, plan.Ebenen...)

	_, befunde, fehler := svg.AusPlan(t.Context(), plan,
		svg.Quellen{
			Raster:    &rasterprobe{meldung: "das Symbolbild fehlt"},
			Schriften: suchprobe{},
		})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet war einer: %v", len(befunde), befunde)
	}

	if befunde[0].Ebene != 1 || befunde[0].Feld != "casting cost" {
		t.Errorf("der Befund zeigt auf Ebene %d (%q) statt auf die zweite",
			befunde[0].Ebene, befunde[0].Feld)
	}
}

// **Ein Textzug mit Symbolen geht denselben Weg.**
func TestEinTextzugMitSymbolenWirdGerastert(t *testing.T) {
	t.Parallel()

	plan := &render.Plan{
		Breite: 100, Hoehe: 140,
		Ebenen: []render.Ebene{{
			Art:    render.ArtText,
			Feld:   "rule text",
			Rahmen: render.Rechteck{Links: 8, Oben: 60, Breite: 84, Hoehe: 40},
			Abschnitte: []render.Abschnitt{
				{Text: "Tippe: "},
				{Symbole: []render.Symbol{{Bild: render.Blatt("/mana/t.png")}}},
			},
			Satz:    render.Satz{Schrift: "Helvetica", Groesse: 10},
			Deckung: 1,
		}},
	}

	raster := &rasterprobe{}

	geschrieben, _, fehler := svg.AusPlan(t.Context(), plan,
		svg.Quellen{Raster: raster, Schriften: suchprobe{}})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if strings.Contains(geschrieben, "<text") {
		t.Errorf("der Zug steht als Text, obwohl er Symbole trägt:\n%s", geschrieben)
	}

	if raster.gesehen == nil {
		t.Error("der Rasterer wurde nicht gefragt")
	}
}

// **Auch der gerasterte Weg meldet einen Kodiererfehler.**
func TestEinKodiererfehlerImSymbolzugWirdGenannt(t *testing.T) {
	t.Parallel()

	_, befunde, fehler := svg.AusPlan(t.Context(), symbolzugplan(),
		svg.Quellen{Raster: &rasterprobe{ohneFlaeche: true}, Schriften: suchprobe{}})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "einbetten") {
		t.Errorf("der Kodiererfehler wurde verschluckt: %v", befunde)
	}
}
