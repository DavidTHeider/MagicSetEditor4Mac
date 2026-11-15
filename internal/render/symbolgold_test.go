package render_test

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// probesymbol ist eine **eigene** Zeichnung — ein Ring mit einem Balken darin.
//
// **Nichts Fremdes** (`docs/entscheidungen/0005`): ein Setsymbol aus einem
// bezogenen Paket wäre ein fremdes Bild im Repository. Diese Form nutzt jeden
// Weg des Rasterers — Kurven, ein Loch über `subtract`, eine Drehsymmetrie.
func probesymbol() []render.Symbolteil {
	loch := kreis(0.5, 0.5, 0.22)
	loch.Kombination = render.KombinationAbziehen

	speichen := render.Symbolteil{
		Art:        render.TeilSymmetrie,
		Spiegelart: render.SpiegelungDrehung,
		Kopien:     3,
		Mitte:      render.Punkt{X: 0.5, Y: 0.5},
		Griff:      render.Punkt{X: 0.5, Y: 0},
		Teile:      []render.Symbolteil{viereck(0.47, 0.1, 0.53, 0.5)},
	}

	return []render.Symbolteil{speichen, loch, kreis(0.5, 0.5, 0.45)}
}

// kreis baut eine runde Form aus vier Bézierstücken.
//
// **0,5523 ist der Griff eines Viertelkreises** — die übliche Näherung, und
// dieselbe, die ein Zeichenprogramm schreibt.
func kreis(mitteX, mitteY, halbmesser float64) render.Symbolteil {
	griff := halbmesser * 0.5523

	stellen := []render.Punkt{
		{X: mitteX, Y: mitteY - halbmesser},
		{X: mitteX + halbmesser, Y: mitteY},
		{X: mitteX, Y: mitteY + halbmesser},
		{X: mitteX - halbmesser, Y: mitteY},
	}

	griffe := []render.Punkt{
		{X: griff, Y: 0}, {X: 0, Y: griff}, {X: -griff, Y: 0}, {X: 0, Y: -griff},
	}

	punkte := make([]render.Symbolpunkt, 0, len(stellen))

	// **Der Griff davor zeigt zurück.** Er ist die Gegenrichtung des Griffs
	// danach — ein Kreis hat an jeder Stelle eine Tangente, und die beiden
	// Griffe liegen auf ihr.
	for stelle := range stellen {
		punkte = append(punkte, render.Symbolpunkt{
			Stelle:      stellen[stelle],
			GriffDavor:  render.Punkt{X: -griffe[stelle].X, Y: -griffe[stelle].Y},
			GriffDanach: griffe[stelle],
			KurveDanach: true,
			KurveDavor:  true,
		})
	}

	return render.Symbolteil{Art: render.TeilForm, Punkte: punkte}
}

func symbolgold(t *testing.T) image.Image {
	t.Helper()

	return render.FaerbeSymbol(
		render.RastereSymbol(probesymbol(), 120, 120, 0.05, false),
		render.Vollfilter{
			Fuellfarbe: color.NRGBA{R: 210, G: 180, B: 60, A: 255},
			Randfarbe:  color.NRGBA{R: 20, G: 20, B: 20, A: 255},
		})
}

func symbolgoldpfad(t *testing.T) string {
	t.Helper()

	wurzel, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Wurzel: %v", err)
	}

	return filepath.Join(wurzel, "testdaten", "goldbilder", "probesymbol.png")
}

// **Das Goldbild hält den Symbolrasterer fest.**
//
// # Woran es *nicht* gemessen ist
//
// An MSE2 nicht. Ein Vergleichsbild von dort setzte voraus, dass MSE2 hier
// läuft und ein Symbol vorliegt; im bezogenen Pack liegt **keine einzige**
// `.mse-symbol` — das Symbol kommt aus dem Set, und für die Referenzkarten
// dieses Projekts gibt es keins
// (#323).
//
// Was dieses Bild also zusagt: **der Lauf bleibt, wie er ist.** Dass er MSE
// trifft, sagen die Einzeltests daneben, und jeder von ihnen nennt seine
// Fundstelle im C++-Quelltext. Das ist weniger, als ein Goldbild sonst hier
// wert ist, und es steht deshalb dabei.
func TestDerSymbolrasterlaufBleibtWieErIst(t *testing.T) {
	pfad := symbolgoldpfad(t)
	neu := symbolgold(t)

	datei, err := os.Open(pfad) //nolint:gosec // fester Pfad im Repository
	if errors.Is(err, os.ErrNotExist) {
		schreibe(t, pfad, neu)
		t.Fatalf("das Goldbild fehlte und wurde angelegt: %s — ansehen und den Lauf wiederholen", pfad)
	}

	if err != nil {
		t.Fatalf("Goldbild nicht lesbar: %v", err)
	}

	defer func() {
		if fehler := datei.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	}()

	gold, err := png.Decode(datei)
	if err != nil {
		t.Fatalf("Goldbild nicht lesbar: %v", err)
	}

	vergleich, err := render.Vergleiche(gold, neu)
	if err != nil {
		t.Fatalf("vergleichen: %v", err)
	}

	if !vergleich.Innerhalb() {
		ablegen(t, neu, gold)
		t.Errorf("%s — über der Toleranz von %g. Differenzbild in testdaten/abweichungen/",
			vergleich, render.Toleranz)
	}
}

// **Die Gegenprobe**: ein Symbol ohne seinen Rand ist ein anderes Bild.
//
// Ohne sie wäre der Vergleich oben eine Zusage ohne Deckung — eine Toleranz,
// die nie greift, sieht aus wie eine, die hält.
func TestEinSymbolOhneRandLiegtUeberDerToleranz(t *testing.T) {
	mitRand := symbolgold(t)

	ohneRand := render.FaerbeSymbol(
		render.RastereSymbol(probesymbol(), 120, 120, 0, false),
		render.Vollfilter{
			Fuellfarbe: color.NRGBA{R: 210, G: 180, B: 60, A: 255},
			Randfarbe:  color.NRGBA{R: 20, G: 20, B: 20, A: 255},
		})

	vergleich, err := render.Vergleiche(mitRand, ohneRand)
	if err != nil {
		t.Fatalf("vergleichen: %v", err)
	}

	if vergleich.Innerhalb() {
		t.Errorf("%s — ein Symbol ohne Rand blieb innerhalb der Toleranz. "+
			"Dann prüft der Goldbildvergleich nichts", vergleich)
	}
}
