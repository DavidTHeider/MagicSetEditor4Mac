package render_test

import (
	"math"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Das Seitenverhältnis einer Symbolzeichnung** —
// #735.
//
//	double margin_x = min(0.4999, max(0., min(bounds.min.x, 1-bounds.max.x)));
//	double margin_y = min(0.4999, max(0., min(bounds.min.y, 1-bounds.max.y)));
//	double delta = 2 * (margin_y - margin_x);
//	if (delta > 0) return 1 / (1 - delta); else return 1 + delta;
//
// `Symbol::aspectRatio`, `src/data/symbol.cpp`, Zeilen 342–358, Fassung
// f9d9356 vom 2021-07-06.
//
// **Der Rand ist der kleinere der beiden Abstände zum Bildrand**, je Richtung.
// Ist der senkrechte größer als der waagerechte, liegt das Symbol flach — und
// `render_symbol` klemmt das Bild darauf, statt es zu dehnen.
func TestDasSeitenverhaeltnisEinerZeichnung(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		teile    []render.Symbolteil
		erwartet float64
	}{
		{
			// **Ohne Teile sind die Grenzen unendlich**, und beide Ränder
			// fallen auf die Klemme 0,4999 — die Differenz ist null.
			name:     "keine Teile",
			erwartet: 1,
		},
		{
			name:     "die ganze Fläche",
			teile:    []render.Symbolteil{viereck(0, 0, 1, 1)},
			erwartet: 1,
		},
		{
			// margin_x = 0, margin_y = 0,25 → delta = 0,5 → 1/(1−0,5)
			name:     "flach",
			teile:    []render.Symbolteil{viereck(0, 0.25, 1, 0.75)},
			erwartet: 2,
		},
		{
			// margin_x = 0,25, margin_y = 0 → delta = −0,5
			name:     "hoch",
			teile:    []render.Symbolteil{viereck(0.25, 0, 0.75, 1)},
			erwartet: 0.5,
		},
		{
			// **Ein Strich in der Mitte klemmt bei 0,4999**, sonst teilte die
			// Rechnung durch null. Er ist so hoch, wie ein Symbol nur sein
			// kann: 1 + 2·(0 − 0,4999).
			name:     "ein Strich in der Mitte",
			teile:    []render.Symbolteil{viereck(0.5, 0, 0.5, 1)},
			erwartet: 1 - 2*0.4999,
		},
		{
			// **Ein Symbol darf über den Rand hinausragen** — `max(0., …)`
			// fängt den negativen Abstand ab.
			name:     "über den Rand hinaus",
			teile:    []render.Symbolteil{viereck(-0.5, 0, 1.5, 1)},
			erwartet: 1,
		},
		{
			// **Die Grenze folgt der Kurve, nicht ihren Griffen.** Beide
			// Griffe ziehen auf y = 0,8; die Kurve erreicht 0,725.
			//
			// Mit den Griffen als Grenze käme margin_y = 0,2 heraus und damit
			// ein Verhältnis von 1 — der Fall unterscheidet die beiden.
			name:     "eine Kurve bleibt unter ihren Griffen",
			teile:    []render.Symbolteil{bogen()},
			erwartet: 1 / (1 - 2*(0.275-0.2)),
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if got := render.Seitenverhaeltnis(fall.teile); math.Abs(got-fall.erwartet) > 1e-9 {
				t.Errorf("Seitenverhältnis = %.6f, erwartet %.6f", got, fall.erwartet)
			}
		})
	}
}

// **Eine Symmetrie zählt jede Kopie mit.**
//
//	FOR_EACH_CONST(p, parts) { for (int i = 0 ; i < copies ; ++i) { … } }
//
// `SymbolSymmetry::calculateBounds`, `src/data/symbol.cpp`, Zeilen 252–277,
// Fassung f9d9356 vom 2021-07-06.
//
// **Ein Balken rechts der Mitte**, viermal um sie gedreht, ergibt ein Kreuz —
// und ein Kreuz ist quadratisch. Wer nur die erste Kopie misst, bekommt 5.
func TestEineSymmetrieZaehltJedeKopieMit(t *testing.T) {
	t.Parallel()

	balken := viereck(0.5, 0.45, 0.95, 0.55)

	// Allein: x ∈ [0,50; 0,95], y ∈ [0,45; 0,55] → 0,05 gegen 0,45.
	if got := render.Seitenverhaeltnis([]render.Symbolteil{balken}); math.Abs(got-5) > 1e-9 {
		t.Fatalf("der Balken allein = %.6f, erwartet 5", got)
	}

	gedreht := render.Symbolteil{
		Art:        render.TeilSymmetrie,
		Spiegelart: render.SpiegelungDrehung,
		Kopien:     4,
		Mitte:      render.Punkt{X: 0.5, Y: 0.5},
		Griff:      render.Punkt{X: 0.5},
		Teile:      []render.Symbolteil{balken},
	}

	// Vier Kopien belegen x wie y ∈ [0,05; 0,95].
	if got := render.Seitenverhaeltnis([]render.Symbolteil{gedreht}); math.Abs(got-1) > 1e-9 {
		t.Errorf("mit vier Kopien = %.6f, erwartet 1", got)
	}
}

// **Und eine Achsenspiegelung nimmt den zweiten Zweig.**
//
//	Matrix2D rot(cos(a+b),sin(a+b), sin(a+b),-cos(a+b));
//
// Dieselbe Stelle, Zeile 268. Der Winkel `b` ist das Doppelte des Griffs;
// mit dem Griff (1,1) spiegelt die zweite Kopie an der Gegendiagonale.
//
// **Eine Spiegelung um die Mitte allein änderte nichts** — der Rand nimmt
// ohnehin die kleinere der beiden Seiten. Erst eine Achse, die nicht
// waagerecht oder senkrecht liegt, verschiebt beide Ränder.
func TestEineAchsenspiegelungZaehltIhreKopieMit(t *testing.T) {
	t.Parallel()

	balken := viereck(0.5, 0.45, 0.95, 0.6)

	if got := render.Seitenverhaeltnis([]render.Symbolteil{balken}); math.Abs(got-1/0.3) > 1e-9 {
		t.Fatalf("der Balken allein = %.6f, erwartet %.6f", got, 1/0.3)
	}

	gespiegelt := render.Symbolteil{
		Art:        render.TeilSymmetrie,
		Spiegelart: render.SpiegelungAchse,
		Kopien:     2,
		Mitte:      render.Punkt{X: 0.5, Y: 0.5},
		Griff:      render.Punkt{X: 1, Y: 1},
		Teile:      []render.Symbolteil{balken},
	}

	// Die gespiegelte Kopie liegt bei x ∈ [0,40; 0,55], y ∈ [0,05; 0,50].
	if got := render.Seitenverhaeltnis([]render.Symbolteil{gespiegelt}); math.Abs(got-1) > 1e-9 {
		t.Errorf("mit der Spiegelung = %.6f, erwartet 1", got)
	}
}

// **Eine Symmetrie ohne Kopien zeichnet nichts.**
//
//	int copies = kind == SYMMETRY_REFLECTION ? this->copies & ~1 : this->copies;
//
// Dieselbe Stelle, Zeile 256: eine Achsenspiegelung rundet auf eine gerade
// Zahl ab, und aus einer einzigen Kopie wird damit keine.
func TestEineSymmetrieOhneKopienBleibtLeer(t *testing.T) {
	t.Parallel()

	ohne := render.Symbolteil{
		Art:        render.TeilSymmetrie,
		Spiegelart: render.SpiegelungAchse,
		Kopien:     1,
		Mitte:      render.Punkt{X: 0.5, Y: 0.5},
		Griff:      render.Punkt{X: 1},
		Teile:      []render.Symbolteil{viereck(0, 0.25, 1, 0.75)},
	}

	if got := render.Seitenverhaeltnis([]render.Symbolteil{ohne}); math.Abs(got-1) > 1e-9 {
		t.Errorf("Seitenverhältnis = %.6f, erwartet 1", got)
	}
}

// **Eine Gruppe reicht ihre Kinder durch.**
func TestEineGruppeReichtIhreKinderDurch(t *testing.T) {
	t.Parallel()

	gruppe := render.Symbolteil{
		Art:   render.TeilGruppe,
		Teile: []render.Symbolteil{viereck(0, 0.25, 1, 0.75)},
	}

	if got := render.Seitenverhaeltnis([]render.Symbolteil{gruppe}); math.Abs(got-2) > 1e-9 {
		t.Errorf("Seitenverhältnis = %.6f, erwartet 2", got)
	}
}

// bogen ist ein Zug aus einer Kurve und einer Geraden zurück.
func bogen() render.Symbolteil {
	return render.Symbolteil{
		Art: render.TeilForm,
		Punkte: []render.Symbolpunkt{
			{
				Stelle:      render.Punkt{X: 0.2, Y: 0.5},
				GriffDanach: render.Punkt{Y: 0.3},
				KurveDanach: true,
			},
			{
				Stelle:     render.Punkt{X: 0.8, Y: 0.5},
				GriffDavor: render.Punkt{Y: 0.3},
				KurveDavor: true,
			},
		},
	}
}

// **Eine Symmetrie in einer Symmetrie** —
// #751.
//
// # Was hier geprüft wird
//
// Die Tests daneben führen **eine** Drehung, **eine** Spiegelung und eine
// flache Gruppe. Dass die Verkettung auch verschachtelt trägt, war bis hierher
// ein Quelltextbeweis und kein Wächter — angestrichen vom Zweitblick durch
// Codex am 2026-08-21, der die Richtung gegen `symbol.cpp` Zeile 252 und
// `viewer.cpp` Zeile 150 gelesen und **kein** Auseinanderlaufen gefunden hat.
//
// # Und warum die Erwartung nicht ausgerechnet ist
//
// **Weil eine von Hand gerechnete Zahl dieselbe Rechnung zweimal prüft.** Die
// vier Rechtecke stehen deshalb einzeln da, jedes an der Stelle, an die die
// Drehung es bringt — und der Vergleich läuft gegen sie, nicht gegen ein
// Ergebnis, das aus derselben Kette stammt. Beide Seiten geben 5,0 und nicht
// die 1, die bei einer leeren Rechnung herauskäme.
//
// # Was er fängt, und was **niemand** fangen kann
//
// **Er fängt den Ursprung.** Wird `m.angewandt(…)` bei der Versetzung der
// Kopien weggelassen, wird dieser Test rot — und **nur** dieser: bei einer
// flachen Symmetrie ist die äußere Matrix die Einheit, und der Fehler ist
// unsichtbar. Gegenprobe gefahren.
//
// **Die Reihenfolge der Verkettung fängt er nicht, und das ist keine Lücke.**
// `drehung.verkettet(m)` gegen `m.verkettet(drehung)` ist bei **Drehungen**
// dasselbe: zwei Drehungen der Ebene vertauschen. Sichtbar würde der
// Unterschied erst mit einer Spiegelung im Spiel, und deren Zweig steht
// nebenan. Der Zweitblick hat die Richtung ohnehin gegen `symbol.cpp` Zeile 252
// und `viewer.cpp` Zeile 150 gelesen; hier steht, warum kein Wächter sie
// prüfen kann.
func TestEineSymmetrieInEinerSymmetrie(t *testing.T) {
	t.Parallel()

	// Ein Balken rechts außen: x ∈ [0,80; 0,95], y ∈ [0,45; 0,55].
	balken := viereck(0.8, 0.45, 0.95, 0.55)

	// **Innen**: zwei Kopien um (0,75 | 0,5) — die zweite spiegelt an diesem
	// Punkt, also x ↦ 1,5 − x.
	innen := render.Symbolteil{
		Art: render.TeilSymmetrie, Spiegelart: render.SpiegelungDrehung,
		Kopien: 2, Mitte: render.Punkt{X: 0.75, Y: 0.5},
		Griff: render.Punkt{X: 0.5},
		Teile: []render.Symbolteil{balken},
	}

	// **Außen**: zwei Kopien um die Bildmitte — x ↦ 1 − x.
	//
	// **Mit einer Gruppe dazwischen**, damit beide Verschachtelungen in einem
	// Fall stehen: Symmetrie in Gruppe **und** Gruppe in Symmetrie.
	aussen := render.Symbolteil{
		Art: render.TeilSymmetrie, Spiegelart: render.SpiegelungDrehung,
		Kopien: 2, Mitte: render.Punkt{X: 0.5, Y: 0.5},
		Griff: render.Punkt{X: 0.5},
		Teile: []render.Symbolteil{{
			Art:   render.TeilGruppe,
			Teile: []render.Symbolteil{innen},
		}},
	}

	// Dieselben vier Balken, jeder an seiner Stelle hingeschrieben.
	einzeln := []render.Symbolteil{
		viereck(0.8, 0.45, 0.95, 0.55),
		viereck(0.55, 0.45, 0.7, 0.55),
		viereck(0.05, 0.45, 0.2, 0.55),
		viereck(0.3, 0.45, 0.45, 0.55),
	}

	verschachtelt := render.Seitenverhaeltnis([]render.Symbolteil{aussen})
	nebeneinander := render.Seitenverhaeltnis(einzeln)

	if math.Abs(verschachtelt-nebeneinander) > 1e-9 {
		t.Errorf("verschachtelt = %.6f, nebeneinander = %.6f — die Verkettung "+
			"bringt die Kopien woandershin", verschachtelt, nebeneinander)
	}
}
