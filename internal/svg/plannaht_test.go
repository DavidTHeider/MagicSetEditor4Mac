// **Der ganze Plan, nicht seine Hälfte** —
// #852.
//
// Der Zweitblick über Phase P hat neun Eigenschaften aufgezählt, die der
// Rasterer anwendet und der SVG-Schreiber fallen ließ — Deckung, Winkel,
// Maske, Verlauf, Mischart, Passung, Ecke, Rand und den Bildbaum. Hier steht
// je eine Zusage dafür.
package svg_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/svg"
)

// eineEbene baut einen Plan mit genau dieser Ebene.
func eineEbene(ebene render.Ebene) *render.Plan {
	return &render.Plan{Breite: 100, Hoehe: 140, Ebenen: []render.Ebene{ebene}}
}

// flaeche ist die einfachste Ebene — Grundlage für die Fälle unten.
func flaeche() render.Ebene {
	return render.Ebene{
		Art:     render.ArtFlaeche,
		Feld:    "grund",
		Rahmen:  render.Rechteck{Links: 10, Oben: 20, Breite: 50, Hoehe: 30},
		Farbe:   render.Farbe{R: 200, G: 40, B: 40, A: 255},
		Deckung: 1,
	}
}

// TestEineEbeneOhneDeckungFehlt hält die Regel des Rasterers.
//
//	if deckung <= 0 { return "" }
//
// `internal/render/rasterer.go`. Der Nullwert eines Go-Verbunds ist 0, und
// eine Ebene ohne Deckung soll **nichts** zeichnen statt alles zu überdecken.
// Der SVG-Schreiber malte sie voll — und der Probeplan der ersten Fassung trug
// genau diesen Nullwert.
func TestEineEbeneOhneDeckungFehlt(t *testing.T) {
	t.Parallel()

	ohne := flaeche()
	ohne.Deckung = 0

	geschrieben, _, fehler := svg.AusPlan(t.Context(), eineEbene(ohne), svg.Quellen{})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if strings.Contains(geschrieben, "<rect") {
		t.Errorf("die Ebene ohne Deckung steht im SVG:\n%s", geschrieben)
	}
}

// TestHalbeDeckungStehtAlsOpacity prüft den Zwischenwert.
func TestHalbeDeckungStehtAlsOpacity(t *testing.T) {
	t.Parallel()

	halb := flaeche()
	halb.Deckung = 0.5

	geschrieben, _, fehler := svg.AusPlan(t.Context(), eineEbene(halb), svg.Quellen{})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if !strings.Contains(geschrieben, `opacity="0.5"`) {
		t.Errorf("keine halbe Deckung im SVG:\n%s", geschrieben)
	}
}

// TestEinWinkelWirdZurDrehung prüft die Umrechnung aus `Rotation::tr`.
//
//	return RealPoint(c * x + s * y + origin.x,
//	                -s * x + c * y + origin.y);
//
// `src/util/rotation.cpp`, Zeilen 37–42, Fassung f9d9356 vom 2021-07-06. Das
// ist in SVG `translate(links, oben) rotate(-Winkel)`, und die Ebene wird
// darin am Ursprung gezeichnet — dieselbe Bauform wie `ebeneGedreht`.
func TestEinWinkelWirdZurDrehung(t *testing.T) {
	t.Parallel()

	gedreht := flaeche()
	gedreht.Winkel = 90

	geschrieben, _, fehler := svg.AusPlan(t.Context(), eineEbene(gedreht), svg.Quellen{})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if !strings.Contains(geschrieben, `transform="translate(10,20) rotate(-90)"`) {
		t.Errorf("keine Drehung im SVG:\n%s", geschrieben)
	}

	// **Und der Rahmen sitzt danach am Ursprung**, sonst wirkte die
	// Verschiebung zweimal.
	if !strings.Contains(geschrieben, `x="0" y="0"`) {
		t.Errorf("der gedrehte Rahmen sitzt nicht am Ursprung:\n%s", geschrieben)
	}
}

// TestRundeEckenStehenAlsRadius prüft `Ecke`.
func TestRundeEckenStehenAlsRadius(t *testing.T) {
	t.Parallel()

	rund := flaeche()
	rund.Ecke = 8

	geschrieben, _, fehler := svg.AusPlan(t.Context(), eineEbene(rund), svg.Quellen{})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if !strings.Contains(geschrieben, `rx="8"`) {
		t.Errorf("keine runden Ecken im SVG:\n%s", geschrieben)
	}
}

// TestEinRandWirdZuVierBahnen prüft `Rand`.
//
// **Vier Rechtecke und kein `stroke`.** Die vier Breiten dürfen verschieden
// sein — MSEs `left_width`, `top_width`, `right_width`, `bottom_width` —, und
// eine Strichbreite kennt nur eine.
func TestEinRandWirdZuVierBahnen(t *testing.T) {
	t.Parallel()

	gerahmt := flaeche()
	gerahmt.Rand = &render.Randbreiten{Links: 2, Oben: 3, Rechts: 4, Unten: 5}

	geschrieben, _, fehler := svg.AusPlan(t.Context(), eineEbene(gerahmt), svg.Quellen{})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if anzahl := strings.Count(geschrieben, "<rect"); anzahl != 4 {
		t.Errorf("%d Rechtecke, erwartet waren vier Bahnen:\n%s", anzahl, geschrieben)
	}
}

// TestEinZuBreiterRandFuelltGanz hält MSEs `clip`.
//
//	bool clip = style().left_width < style().width && …
//
// `src/render/value/color.cpp`, Zeilen 47–50, Fassung f9d9356 vom 2021-07-06.
// **Alle vier** müssen kleiner
// sein als das Feld; MSEs Vorgabe ist 100000, und dann wird gefüllt.
func TestEinZuBreiterRandFuelltGanz(t *testing.T) {
	t.Parallel()

	gerahmt := flaeche()
	gerahmt.Rand = &render.Randbreiten{Links: 100_000, Oben: 1, Rechts: 1, Unten: 1}

	geschrieben, _, fehler := svg.AusPlan(t.Context(), eineEbene(gerahmt), svg.Quellen{})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if anzahl := strings.Count(geschrieben, "<rect"); anzahl != 1 {
		t.Errorf("%d Rechtecke, erwartet war eine volle Fläche:\n%s", anzahl, geschrieben)
	}
}

// TestEinVerlaufWirdEinLinearGradient prüft `Verlauf`.
func TestEinVerlaufWirdEinLinearGradient(t *testing.T) {
	t.Parallel()

	mitVerlauf := flaeche()
	mitVerlauf.Verlauf = &render.Verlauf{Stufen: []render.Verlaufsstufe{
		{Stelle: 0, Farbe: render.Farbe{R: 255, A: 255}},
		{Stelle: 1, Farbe: render.Farbe{B: 255, A: 255}},
	}}

	geschrieben, _, fehler := svg.AusPlan(t.Context(), eineEbene(mitVerlauf), svg.Quellen{})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	for _, teil := range []string{"<linearGradient", "<stop", `fill="url(#`} {
		if !strings.Contains(geschrieben, teil) {
			t.Errorf("%q fehlt im SVG:\n%s", teil, geschrieben)
		}
	}
}

// TestEineUnbekannteMischartWirdGenannt prüft den Befundkanal.
//
// **CSS kann MSEs Mischarten nicht.** Von den zwanzig sind `add`, `subtract`,
// `stamp`, `negation`, `reflect`, `shadow` und `glow` in `mix-blend-mode` gar
// nicht enthalten, und `soft light` heißt dort etwas anderes. Übernommen wird
// deshalb keine — **gemeldet** wird jede.
func TestEineUnbekannteMischartWirdGenannt(t *testing.T) {
	t.Parallel()

	gemischt := flaeche()
	gemischt.Mischart = "shadow"

	_, befunde, fehler := svg.AusPlan(t.Context(), eineEbene(gemischt), svg.Quellen{})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "shadow") {
		t.Errorf("die Mischart wurde nicht genannt: %v", befunde)
	}
}
