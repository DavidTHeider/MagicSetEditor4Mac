package render_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Beim Verkleinern rechnet MSE bilinear, beim Vergrößern das Flächenmittel**
// (#532).
//
// # Woher das kommt
//
// MSE bringt ein Bild mit `image.Rescale(w, h)` auf die Feldgröße
// (`src/render/value/image.cpp`, Zeilen 33–40, Fassung f9d9356 vom
// 2021-07-06). Das ist `wxImage::Scale` mit dem Vorgabewert
// `wxIMAGE_QUALITY_NORMAL` (`wx/image.h`, Zeile 356).
//
// **Was dieser Wert rechnet, steht in keiner Kopfdatei** — sie sagt nur
// „default quality, suitable for most icons". Gemessen am 2026-08-19 mit
// vierzig Zeilen C++ gegen **wx 3.3.3**, die Fassung aus
// `mse2-quelle/build/CMakeCache.txt`, an einer echten Illustration
// (626 × 457):
//
//	Verkleinern auf 316 × 231:   NORMAL == BILINEAR      RMS 0,0000
//	Vergrößern  auf 1200 × 900:  NORMAL == BOX_AVERAGE   RMS 0,0000
//
// **Wir rechneten es genau vertauscht.** Der Grund stand im Kommentar und war
// selbst gemessen (#113):
// bilinear lässt feine Linien beim starken Verkleinern flimmern. Das bleibt
// richtig — und wird trotzdem aufgegeben: es ist eine Aussage über
// Bildqualität, nicht über Parität. MSE flimmert ebenso.
//
// **Was die Angleichung bringt, ist ein Stück und nicht alles.** Die
// Vorhersage aus #532
// lautete RMS 0 für das Kunstfenster von `A-Cloister Gargoyle`; gemessen
// wurden danach **2,362** gegen vorher 3,123.
//
// Der Rest ist **wx' bilineare Arithmetik**: dieselbe Filterwahl, andere
// Rechnung. Gegen wx' eigenes Ergebnis liegt unsere Abtastung bei 2,305 —
// also fast der ganze verbliebene Abstand. Eine eckenbündige Abbildung der
// Abtaststellen ist messbar schlechter (4,55) und deshalb nicht die Erklärung.
//
// Wer die Zahl weiter senken will, bildet wx' Rundung nach, nicht seine
// Filterwahl.

// impulsbild malt einen einzelnen hellen Punkt — daran unterscheiden sich die
// Filter am deutlichsten.
//
// **Eine Kante taugt dafür nicht.** Bei ihr liefern Flächenmittel und bilinear
// dieselben Werte, sobald die Abtaststellen zwischen zwei gleichen Nachbarn
// liegen; erst ein Impuls trennt sie.
func impulsbild(breite, hoehe, stelle int) *image.NRGBA {
	bild := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))

	for y := range hoehe {
		for x := range breite {
			wert := uint8(0)
			if x == stelle {
				wert = 255
			}

			bild.SetNRGBA(x, y, color.NRGBA{R: wert, G: wert, B: wert, A: 255})
		}
	}

	return bild
}

// gezeichnetesBild rastert ein Bild in ein Rechteck und gibt das Ergebnis.
func gezeichnetesBild(t *testing.T, quelle image.Image, breite, hoehe float64) *image.NRGBA {
	t.Helper()

	plan := &render.Plan{
		Breite: breite, Hoehe: hoehe,
		Ebenen: []render.Ebene{{
			Art:     render.ArtBild,
			Rahmen:  render.Rechteck{Breite: breite, Hoehe: hoehe},
			Bild:    render.Blatt("/quelle.png"),
			Deckung: 1,
		}},
	}

	bild, befunde, err := render.NeuerRasterer(bilder{"/quelle.png": quelle}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("Zeichne: %v", err)
	}

	for _, befund := range befunde {
		t.Fatalf("Befund: %v", befund)
	}

	return bild
}

// TestBeimVerkleinernWirdBilinearGerechnet ist die Kernzusage.
//
// # Die Rechnung, von Hand
//
// Quelle 8 Punkte breit, ein Impuls bei 3; Ziel 6 Punkte, Faktor 1,333.
// Die Abtaststelle des dritten Zielpunkts ist
//
//	(2 + 0,5) / 6 · 8 − 0,5 = 2,8333
//
// **Bilinear** mischt `src[2]` (0) und `src[3]` (255) im Verhältnis 0,8333:
// **212,5 → 213**.
//
// **Das Flächenmittel** nimmt den ganzen Bereich [2,167 … 3,5], also die
// Punkte 2 und 3 zu gleichen Teilen: **127,5 → 128**.
func TestBeimVerkleinernWirdBilinearGerechnet(t *testing.T) {
	t.Parallel()

	klein := gezeichnetesBild(t, impulsbild(8, 2, 3), 6, 2)

	if wert := klein.NRGBAAt(2, 1).R; wert != 213 {
		t.Errorf("der dritte Zielpunkt ist %d, bilinear wären 213 (das "+
			"Flächenmittel gäbe 128)", wert)
	}
}

// TestBeimVergroessernWirdGemittelt ist die andere Richtung.
//
// **Beim Vergrößern nimmt wx das Flächenmittel**, und das ist bei ganzzahliger
// Vergrößerung dasselbe wie nächster Nachbar: jeder Zielpunkt liegt ganz in
// **einem** Quellpunkt. Bilinear ergäbe stattdessen einen weichen Verlauf.
func TestBeimVergroessernWirdGemittelt(t *testing.T) {
	t.Parallel()

	gross := gezeichnetesBild(t, impulsbild(4, 2, 1), 16, 8)

	var zwischenwerte int

	for x := range 16 {
		if wert := gross.NRGBAAt(x, 4).R; wert != 0 && wert != 255 {
			zwischenwerte++
		}
	}

	if zwischenwerte != 0 {
		t.Errorf("%d Zwischenwerte, erwartet 0 — bilinear statt gemittelt "+
			"vergrößert", zwischenwerte)
	}
}

// TestBeiGemischterRichtungWirdGemittelt hält die dritte gemessene Regel fest.
//
// **wx entscheidet für den ganzen Aufruf, nicht je Achse.** Am 2026-08-19
// gegen wx 3.3.3 gemessen: 626 × 457 auf 300 × 900 (waagerecht kleiner,
// senkrecht größer) und auf 900 × 300 (umgekehrt) — beide Male
// `NORMAL == BOX_AVERAGE` bei RMS **0,0000**.
func TestBeiGemischterRichtungWirdGemittelt(t *testing.T) {
	t.Parallel()

	// Waagerecht verkleinert (8 → 6), senkrecht vergrößert (2 → 8).
	gemischt := gezeichnetesBild(t, impulsbild(8, 2, 3), 6, 8)

	if wert := gemischt.NRGBAAt(2, 4).R; wert != 128 {
		t.Errorf("der dritte Zielpunkt ist %d, das Flächenmittel gäbe 128 "+
			"(bilinear gäbe 213)", wert)
	}
}

// TestDieAbtastungHaeltSichAmRandFest prüft das Klemmen an allen vier Kanten.
//
// **Festhalten statt Wiederholen**: am Rand eines Kartenrahmens sähe eine
// Wiederholung wie ein schmaler Streifen der Gegenseite aus. Die Kanten werden
// erreicht, sobald ein Zielbildpunkt über die Quelle hinausgreift — beim
// bilinearen Lesen ist das der äußerste Punkt jeder Seite.
func TestDieAbtastungHaeltSichAmRandFest(t *testing.T) {
	t.Parallel()

	// Eine Quelle mit vier verschiedenen Ecken; verkleinert wird knapp, damit
	// die bilineare Abtastung an allen Rändern hinausgreift.
	quelle := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := range 4 {
		for x := range 4 {
			quelle.SetNRGBA(x, y, color.NRGBA{
				R: uint8(60 + 40*x), G: uint8(60 + 40*y), B: 128, A: 255,
			})
		}
	}

	klein := gezeichnetesBild(t, quelle, 3, 3)

	// **Die Ecken bleiben die Ecken.** Ohne Festhalten läse die Abtastung
	// außerhalb und bekäme dort den Nullwert — also durchsichtiges Schwarz.
	for _, ecke := range []image.Point{{}, {X: 2}, {Y: 2}, {X: 2, Y: 2}} {
		if farbe := klein.NRGBAAt(ecke.X, ecke.Y); farbe.A != 255 || farbe.B != 128 {
			t.Errorf("die Ecke %v ist %v — außerhalb gelesen statt festgehalten",
				ecke, farbe)
		}
	}
}
