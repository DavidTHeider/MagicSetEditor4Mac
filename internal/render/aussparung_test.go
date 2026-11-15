package render_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Ein Vergleich, der eine Fläche aussparen kann** — die Zusage aus
// [0021](../../docs/entscheidungen/0021-das-referenzbild-kommt-vom-nutzenden.md).
//
// # Warum es das braucht
//
// Gegen eine echte MSE2-Ausgabe fehlt uns die **Illustration**: das Referenzbild
// ist ein fertiges JPG, die Quelldatei der Kunst liegt nicht bei. Ein Vergleich
// über die ganze Karte misst dann zu einem Drittel etwas, das nie stimmen kann,
// und jede Verbesserung an Rahmen und Text verschwindet in diesem Rauschen.
//
// **Ausgespart wird trotzdem nicht weggeschaut**: der Lauf meldet beide Zahlen.
// Diese Datei prüft nur, dass die Aussparung wirklich aussetzt — und dass sie
// die verglichenen Bildpunkte mitzählt, denn ohne Nenner ist der Anteil eine
// Zahl ohne Maßstab.
func flaeche(breite, hoehe int, farbe color.RGBA) *image.RGBA {
	bild := image.NewRGBA(image.Rect(0, 0, breite, hoehe))

	for y := range hoehe {
		for x := range breite {
			bild.SetRGBA(x, y, farbe)
		}
	}

	return bild
}

func TestEineAussparungBleibtAusserBetracht(t *testing.T) {
	weiss := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	gold := flaeche(10, 10, weiss)
	neu := flaeche(10, 10, weiss)

	// **Ein schwarzer Fleck, ganz innerhalb der Aussparung.** Ohne sie risse er
	// den Vergleich auf; mit ihr darf er nicht auffallen.
	for y := 2; y < 5; y++ {
		for x := 2; x < 5; x++ {
			neu.SetRGBA(x, y, color.RGBA{A: 255})
		}
	}

	ohne, err := render.VergleicheAusserhalb(gold, neu, image.Rect(2, 2, 5, 5))
	if err != nil {
		t.Fatalf("vergleichen: %v", err)
	}

	if !ohne.Innerhalb() {
		t.Errorf("die Aussparung greift nicht: %s", ohne)
	}

	if ohne.Punkte != 100-9 {
		t.Errorf("%d verglichene Bildpunkte, erwartet %d", ohne.Punkte, 100-9)
	}

	// **Die Gegenprobe im selben Test.** Ohne sie sagte der Lauf nur, dass
	// *irgendetwas* grün ist — er sagte nicht, dass der Fleck ohne Aussparung
	// wirklich auffiele. Ein Wächter, der seine eigene Wirkung nicht zeigt,
	// hält die halbe Zusage (§10).
	ganz, err := render.Vergleiche(gold, neu)
	if err != nil {
		t.Fatalf("vergleichen: %v", err)
	}

	if ganz.Innerhalb() {
		t.Errorf("ohne Aussparung müsste derselbe Fleck auffallen: %s", ganz)
	}
}

// Eine leere Aussparung ist der gewöhnliche Vergleich.
func TestEineLeereAussparungSpartNichtsAus(t *testing.T) {
	gold := flaeche(4, 4, color.RGBA{R: 255, A: 255})
	neu := flaeche(4, 4, color.RGBA{R: 255, A: 255})

	ohne, err := render.VergleicheAusserhalb(gold, neu, image.Rectangle{})
	if err != nil {
		t.Fatalf("vergleichen: %v", err)
	}

	if ohne.Punkte != 16 {
		t.Errorf("%d verglichene Bildpunkte, erwartet 16", ohne.Punkte)
	}
}

// **Eine Aussparung über das ganze Bild lässt nichts zu vergleichen übrig.**
//
// Das ist kein Erfolg, sondern ein Fehler: ein Vergleich ohne Bildpunkte hätte
// den Abstand null und wäre damit immer grün — die denkbar stillste Art, eine
// Messung wegzudefinieren.
func TestEineAussparungUeberAllesIstEinFehler(t *testing.T) {
	gold := flaeche(4, 4, color.RGBA{A: 255})
	neu := flaeche(4, 4, color.RGBA{A: 255})

	if _, err := render.VergleicheAusserhalb(gold, neu, image.Rect(0, 0, 4, 4)); err == nil {
		t.Fatal("eine Aussparung über das ganze Bild muss ein Fehler sein")
	}
}

// **Die Aussparung liegt relativ zur linken oberen Ecke, nicht absolut.**
//
// Der Unterschied fällt erst bei einem Bild mit versetztem Ursprung auf. Der
// erste Anlauf von [render.VergleicheAusserhalb] verglich die Aussparung gegen
// absolute Koordinaten und las die Bildpunkte relativ — bei Bounds ab `(10,10)`
// sparte er die falsche Fläche aus. Angestrichen vom Zweitblick über Phase A.
func TestEineAussparungGiltRelativZurEcke(t *testing.T) {
	weiss := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	// **Ein Ausschnitt ab (10,10)**, damit Bounds.Min nicht null ist.
	gross := flaeche(20, 20, weiss)
	gold := gross.SubImage(image.Rect(10, 10, 20, 20))

	grossNeu := flaeche(20, 20, weiss)
	for y := 12; y < 15; y++ {
		for x := 12; x < 15; x++ {
			grossNeu.SetRGBA(x, y, color.RGBA{A: 255})
		}
	}

	neu := grossNeu.SubImage(image.Rect(10, 10, 20, 20))

	// Der Fleck liegt bei (2,2)…(5,5) **im Ausschnitt**.
	ohne, err := render.VergleicheAusserhalb(gold, neu, image.Rect(2, 2, 5, 5))
	if err != nil {
		t.Fatalf("vergleichen: %v", err)
	}

	if !ohne.Innerhalb() {
		t.Errorf("die Aussparung greift bei versetztem Ursprung nicht: %s", ohne)
	}
}
