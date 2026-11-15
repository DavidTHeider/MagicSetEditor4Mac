package render_test

import (
	"errors"
	"image"
	"image/color"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// gestreift malt senkrechte Streifen — ein Muster, das eine Verschiebung
// **sichtbar** macht.
//
// **Nicht eine Fläche**: über einer einfarbigen Fläche ist jeder Versatz gleich
// gut, und die Probe belegte nichts.
func gestreift(breite, hoehe, versatz int) *image.NRGBA {
	bild := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))

	for y := range hoehe {
		for x := range breite {
			wert := uint8(0)
			if ((x+versatz)/2)%2 == 0 {
				wert = 255
			}

			bild.SetNRGBA(x, y, color.NRGBA{R: wert, G: wert, B: wert, A: 255})
		}
	}

	return bild
}

// TestProbeVersatzFindetDieVerschiebung ist die Kernzusage.
func TestProbeVersatzFindetDieVerschiebung(t *testing.T) {
	t.Parallel()

	// Unser Bild steht um einen Bildpunkt nach links versetzt.
	unser := gestreift(12, 6, 1)
	referenz := gestreift(12, 6, 0)

	versatz, err := render.ProbeVersatz(unser, referenz, 1)
	if err != nil {
		t.Fatalf("ProbeVersatz: %v", err)
	}

	// **(−1,0):** was die Referenz bei x zeigt, steht bei uns an x−1 — unser
	// Inhalt sitzt einen Bildpunkt weiter links.
	if versatz.X != -1 || versatz.Y != 0 {
		t.Errorf("bester Versatz (%d,%d), erwartet (-1,0)", versatz.X, versatz.Y)
	}

	if versatz.Bester != 0 {
		t.Errorf("Abstand beim besten Versatz %.3f, erwartet 0", versatz.Bester)
	}

	if versatz.Ohne <= versatz.Bester {
		t.Errorf("ohne Versatz %.3f, mit %.3f — die Probe hat nichts gefunden",
			versatz.Ohne, versatz.Bester)
	}
}

// TestProbeVersatzOhneVerschiebungBleibtBeiNull prüft den Normalfall.
func TestProbeVersatzOhneVerschiebungBleibtBeiNull(t *testing.T) {
	t.Parallel()

	versatz, err := render.ProbeVersatz(gestreift(12, 6, 0), gestreift(12, 6, 0), 1)
	if err != nil {
		t.Fatalf("ProbeVersatz: %v", err)
	}

	if versatz.X != 0 || versatz.Y != 0 {
		t.Errorf("bester Versatz (%d,%d), erwartet (0,0)", versatz.X, versatz.Y)
	}

	if versatz.Gewinn() != 0 {
		t.Errorf("Gewinn %.3f, erwartet 0", versatz.Gewinn())
	}
}

// TestProbeVersatzMisstJedenVersatzUeberDieselbeFlaeche ist die Zusage, ohne
// die die Zahlen nicht vergleichbar wären.
//
// **Ein Versatz verkleinert die Überlappung.** Wer jeden Kandidaten über seine
// eigene Überlappung misst, vergleicht Zahlen über verschiedene Mengen — genau
// der Fehler aus §10. Gemessen wird deshalb für **alle** über dasselbe innere
// Rechteck.
func TestProbeVersatzMisstJedenVersatzUeberDieselbeFlaeche(t *testing.T) {
	t.Parallel()

	versatz, err := render.ProbeVersatz(gestreift(12, 6, 0), gestreift(12, 6, 0), 2)
	if err != nil {
		t.Fatalf("ProbeVersatz: %v", err)
	}

	// 12 × 6 um 2 auf jeder Seite eingerückt sind 8 × 2.
	if versatz.Punkte != 16 {
		t.Errorf("%d Bildpunkte verglichen, erwartet 16", versatz.Punkte)
	}
}

// TestProbeVersatzSagtWennDerBereichZuKleinIst prüft die Reichweitenangabe.
//
// **Liegt das Beste am Rand des Suchbereichs, ist der Bereich zu klein** — dann
// sagt die Zahl nicht, wie groß der Versatz wirklich ist, sondern nur, dass er
// mindestens so groß ist.
func TestProbeVersatzSagtWennDerBereichZuKleinIst(t *testing.T) {
	t.Parallel()

	// Um zwei versetzt, aber nur bis eins gesucht.
	versatz, err := render.ProbeVersatz(gestreift(16, 10, 2), gestreift(16, 10, 0), 1)
	if err != nil {
		t.Fatalf("ProbeVersatz: %v", err)
	}

	if !versatz.AmRand() {
		t.Errorf("bester Versatz (%d,%d) bei Reichweite %d — AmRand sagt nein",
			versatz.X, versatz.Y, versatz.Reichweite)
	}

	weiter, err := render.ProbeVersatz(gestreift(16, 10, 2), gestreift(16, 10, 0), 3)
	if err != nil {
		t.Fatalf("ProbeVersatz: %v", err)
	}

	if weiter.X != -2 || weiter.AmRand() {
		t.Errorf("mit Reichweite 3: Versatz (%d,%d), AmRand %v — erwartet (-2,0) und nein",
			weiter.X, weiter.Y, weiter.AmRand())
	}
}

// TestProbeVersatzFindetAuchSenkrecht prüft die zweite Achse.
func TestProbeVersatzFindetAuchSenkrecht(t *testing.T) {
	t.Parallel()

	unser := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	referenz := image.NewNRGBA(image.Rect(0, 0, 8, 8))

	for y := range 8 {
		for x := range 8 {
			hell := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			dunkel := color.NRGBA{A: 255}

			// Waagerechte Streifen; unser Bild steht um eins nach oben.
			if (y/2)%2 == 0 {
				referenz.SetNRGBA(x, y, hell)
			} else {
				referenz.SetNRGBA(x, y, dunkel)
			}

			if ((y+1)/2)%2 == 0 {
				unser.SetNRGBA(x, y, hell)
			} else {
				unser.SetNRGBA(x, y, dunkel)
			}
		}
	}

	versatz, err := render.ProbeVersatz(unser, referenz, 1)
	if err != nil {
		t.Fatalf("ProbeVersatz: %v", err)
	}

	if versatz.X != 0 || versatz.Y != -1 {
		t.Errorf("bester Versatz (%d,%d), erwartet (0,-1)", versatz.X, versatz.Y)
	}
}

// TestProbeVersatzWeistUnbrauchbareEingabenAb prüft die Fehlerwege.
func TestProbeVersatzWeistUnbrauchbareEingabenAb(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name           string
		unser, referen image.Image
		reichweite     int
	}{
		{"kein Bild", nil, grau(8, 8, 0), 1},
		{"keine Referenz", grau(8, 8, 0), nil, 1},
		{"verschiedene Größen", grau(8, 8, 0), grau(9, 8, 0), 1},
		{"negative Reichweite", grau(8, 8, 0), grau(8, 8, 0), -1},
		{"Bild zu klein für die Reichweite", grau(4, 4, 0), grau(4, 4, 0), 2},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, err := render.ProbeVersatz(fall.unser, fall.referen, fall.reichweite)
			if !errors.Is(err, render.ErrPlan) {
				t.Errorf("Fehler %v, erwartet ErrPlan", err)
			}
		})
	}
}

// TestProbeVersatzOhneReichweiteMisstNurDieLage prüft die Reichweite null.
//
// **Sie ist kein Sonderfall, sondern der Grenzfall**: ohne Suchbereich ist der
// beste Versatz (0,0), und `Ohne` und `Bester` sind dieselbe Zahl.
func TestProbeVersatzOhneReichweiteMisstNurDieLage(t *testing.T) {
	t.Parallel()

	versatz, err := render.ProbeVersatz(gestreift(12, 6, 1), gestreift(12, 6, 0), 0)
	if err != nil {
		t.Fatalf("ProbeVersatz: %v", err)
	}

	if versatz.X != 0 || versatz.Y != 0 || versatz.Ohne != versatz.Bester {
		t.Errorf("(%d,%d) ohne %.3f, bester %.3f — erwartet (0,0) und gleich",
			versatz.X, versatz.Y, versatz.Ohne, versatz.Bester)
	}

	if versatz.AmRand() {
		t.Error("AmRand bei Reichweite 0 — es gibt keinen Rand zu treffen")
	}
}
