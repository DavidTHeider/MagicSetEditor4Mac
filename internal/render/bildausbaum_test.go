package render_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Ein Bildbaum als eigenes Bild** —
// #660.
//
// Die Größenregel ist MSEs `conform_image` (`src/gfx/generated_image.cpp`,
// Zeilen 30–66, Fassung f9d9356 vom 2021-07-06), und ihre vier Fälle stehen
// hier als Tabelle.
func TestBildAusBaum(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name          string
		breite, hoehe int
		zielB, zielH  int
	}{
		{"ohne Angabe bleibt die Quellgröße", 0, 0, 40, 20},
		{"nur die Breite — die Höhe folgt dem Verhältnis", 80, 0, 80, 40},
		{"nur die Höhe — die Breite folgt", 0, 40, 80, 40},
		{"beides — gedehnt, denn ASPECT_STRETCH ist die Vorgabe", 30, 30, 30, 30},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			bild, befund := render.NeuerRasterer(bilder{"q.png": vollflaeche(40, 20)}).
				BildAusBaum(render.Blatt("q.png"), fall.breite, fall.hoehe)
			if befund != "" {
				t.Fatalf("Befund: %s", befund)
			}

			if bild.Bounds().Dx() != fall.zielB || bild.Bounds().Dy() != fall.zielH {
				t.Errorf("%dx%d, erwartet %dx%d",
					bild.Bounds().Dx(), bild.Bounds().Dy(), fall.zielB, fall.zielH)
			}

			if gedeckteFlaeche(bild) == 0 {
				t.Error("das Bild ist leer")
			}
		})
	}
}

// **Was nicht ausführbar ist, wird benannt** — dieselbe Regel wie beim
// Symbolbild (#658): die Datei entstünde sonst leer, und die Vorlage verwiese
// darauf.
func TestEinUnausfuehrbarerBaumWirdBenannt(t *testing.T) {
	t.Parallel()

	bild, befund := render.NeuerRasterer(bilder{}).
		BildAusBaum(render.Blatt("fehlt.png"), 10, 10)

	if befund == "" || bild != nil {
		t.Errorf("Befund %q, Bild %v", befund, bild != nil)
	}
}

// **Ein Bild ohne Fläche gibt keine Division durch null.**
//
// MSE fängt es über `image.Ok()` ab (`conform_image`, `generated_image.cpp`,
// Zeile 34 und `SymbolInFont::getImage`, Zeile 126, Fassung f9d9356 vom
// 2021-07-06). Hier steht die Wache in der Rechnung, denn ohne sie teilte
// `breite*quellhoehe/quellbreite` durch null.
func TestEinBildOhneFlaeche(t *testing.T) {
	t.Parallel()

	bild, befund := render.NeuerRasterer(bilder{"leer.png": vollflaeche(0, 0)}).
		BildAusBaum(render.Blatt("leer.png"), 12, 0)
	if befund != "" {
		t.Fatalf("Befund: %s", befund)
	}

	if bild.Bounds().Dx() != 12 || bild.Bounds().Dy() != 1 {
		t.Errorf("%dx%d, erwartet 12x1", bild.Bounds().Dx(), bild.Bounds().Dy())
	}
}

// **Ein Teilbefund geht nicht verloren** — Befund des Zweitblicks durch Codex
// vom 2026-08-21.
//
// [Rasterer.ausBaum] gibt **Bild und Befund**: eine noch nicht gebaute
// Verknüpfung zeichnet das Ausgangsbild und meldet es dazu
// (`bildausfuehrung.go`, Zeilen 207–215). Beide neuen Einstiege — dieser und
// [Rasterer.SymbolAlsBild] — haben den Befund verworfen, sobald ein Bild kam.
//
// **Was das kostet**: eine Beigabe, die halb falsch aussieht, wird geschrieben,
// und die Ausfuhr meldet nichts. Dieselbe Klasse an zwei Stellen — der
// Zweitblick hat sie beide genannt.
func TestEinTeilbefundGehtNichtVerloren(t *testing.T) {
	t.Parallel()

	baum := render.Bildbaum{
		Knotenart: "rotate",
		Kinder: []render.Bildkind{
			{Name: "input", Baum: render.Blatt("q.png")},
		},
	}

	bild, befund := render.NeuerRasterer(bilder{"q.png": vollflaeche(40, 20)}).
		BildAusBaum(baum, 0, 0)

	if bild == nil {
		t.Fatal("kein Bild — die Verknüpfung soll das Ausgangsbild zeichnen")
	}

	if befund == "" {
		t.Error("der Befund über die ungebaute Verknüpfung fehlt")
	}
}
