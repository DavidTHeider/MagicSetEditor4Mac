package service_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die vier Grundformen des Symboleditors** —
// #890,
// Werkstück 3 von Phase X.
//
// Portiert aus `SymbolBasicShapeEditor::makeCenteredShape`
// (`src/gui/symbol/basic_shape_editor.cpp`, Zeilen 188–297, Fassung f9d9356
// vom 2021-07-06).
//
// **Die Zahlen sind nachgelesen, nicht hergeleitet.** Der Kreis benutzt eine
// Konstante, die MSE selbst „magic" nennt, und das Vieleck eine Herleitung
// über drei Gleichungen. Wer das nachrechnet statt nachzulesen, bekommt eine
// Form, die fast stimmt.
func TestEinRechteckHatVierEcken(t *testing.T) {
	t.Parallel()

	form := service.Grundform(service.FormRechteck,
		msepaket.Vektor{X: 0.5, Y: 0.5}, msepaket.Vektor{X: 0.25, Y: 0.1}, 0, false)

	if len(form.Punkte) != 4 {
		t.Fatalf("%d Punkte: %+v", len(form.Punkte), form.Punkte)
	}

	// **Gegen den Uhrzeigersinn ab links oben** — so legt MSE sie.
	will := []msepaket.Vektor{
		{X: 0.25, Y: 0.4},
		{X: 0.75, Y: 0.4},
		{X: 0.75, Y: 0.6},
		{X: 0.25, Y: 0.6},
	}

	for stelle, punkt := range form.Punkte {
		if !nahe(punkt.Stelle, will[stelle]) {
			t.Errorf("Punkt %d ist %+v, will %+v", stelle, punkt.Stelle, will[stelle])
		}

		if punkt.KurveDanach {
			t.Errorf("Punkt %d ist eine Kurve — ein Rechteck hat gerade Seiten", stelle)
		}
	}
}

// **Ein Kreis hat vier Punkte, und seine Griffe tragen die Konstante.**
//
//	const double kr = 0.5522847498f; // = 4/3 * (sqrt(2) - 1)
//
// Ebd., Zeile 200. Vier Bézierbögen nähern einen Kreis genau dann gut, wenn
// die Griffe diese Länge haben; eine gerade Zahl daneben gäbe eine Form, die
// aussieht wie ein Kreis und keiner ist.
func TestEinKreisTraegtDieMagischeKonstante(t *testing.T) {
	t.Parallel()

	form := service.Grundform(service.FormKreis,
		msepaket.Vektor{X: 0.5, Y: 0.5}, msepaket.Vektor{X: 0.4, Y: 0.4}, 0, false)

	if len(form.Punkte) != 4 {
		t.Fatalf("%d Punkte", len(form.Punkte))
	}

	const kr = 0.5522847498

	if hat := form.Punkte[0].GriffDavor.Y; math.Abs(hat-kr*0.4) > 1e-9 {
		t.Errorf("der Griff ist %v, will %v", hat, kr*0.4)
	}

	for stelle, punkt := range form.Punkte {
		if !punkt.KurveDanach {
			t.Errorf("Punkt %d ist eine Gerade — ein Kreis besteht aus Bögen", stelle)
		}

		// **Die Sperre ist `size`**, ebd.: die beiden Griffe eines
		// Kreispunktes zeigen in entgegengesetzte Richtungen und sind gleich
		// lang. Wer sie freigäbe, verlöre die Rundung beim ersten Ziehen.
		if punkt.Sperre != "size" {
			t.Errorf("Punkt %d hat die Sperre %q", stelle, punkt.Sperre)
		}
	}
}

// **Ein Vieleck liegt auf einem Kreis um seine eigene Mitte — und ist genau
// so breit, wie es soll.**
//
// Beide Eigenschaften sind prüfbar, ohne die Herleitung zu wiederholen. Die
// zweite ist die wichtigere: der Umkreis allein sagt nichts über die Größe.
func TestEinViereckLiegtAufSeinemUmkreis(t *testing.T) {
	t.Parallel()

	for _, ecken := range []int{3, 5, 6, 7} {
		t.Run(eckenname(ecken), func(t *testing.T) {
			t.Parallel()

			form := service.Grundform(service.FormVieleck,
				msepaket.Vektor{X: 0.5, Y: 0.5}, msepaket.Vektor{X: 0.4, Y: 0.4},
				ecken, false)

			if len(form.Punkte) != ecken {
				t.Fatalf("%d Punkte, will %d", len(form.Punkte), ecken)
			}

			mitte := schwerpunkt(form.Punkte)
			erster := punktabstand(form.Punkte[0].Stelle, mitte)

			for stelle, punkt := range form.Punkte {
				if hat := punktabstand(punkt.Stelle, mitte); math.Abs(hat-erster) > 1e-9 {
					t.Errorf("Punkt %d hat den Abstand %v, der erste %v", stelle, hat, erster)
				}
			}

			// **Und die Breite ist genau `2 · halb.X`.**
			//
			// Das ist die Eigenschaft, die MSEs Herleitung überhaupt
			// herstellt: „define: width = 2 = |fc|" (ebd., Zeile 246). Der
			// Umkreis allein sagt nichts über die Größe — mit einem falschen
			// `beta` liegen die Punkte immer noch auf einem Kreis, nur auf
			// einem anderen. **Genau daran ist der erste Anlauf dieses Tests
			// gescheitert**: die Ganzzahldivision `(n+1)/4` durch eine
			// Gleitkommadivision zu ersetzen ließ ihn grün.
			if breite := spanne(form.Punkte); math.Abs(breite-0.8) > 1e-9 {
				t.Errorf("die Breite ist %v, will 0.8", breite)
			}
		})
	}
}

// **Ein Stern hat doppelt so viele Punkte, und sie wechseln sich ab.**
//
// **Und die Näherung wird mitportiert.** MSE schreibt selbst dazu: „This is
// only correct for n<=6, but gives acceptable results for higher n"
// (ebd., Zeile 277). Sie stillschweigend zu verbessern wäre eine
// Verhaltensänderung (§15, Regel 4) — der Stern sähe dann anders aus als in
// MSE, und niemand wüsste, warum.
func TestEinSternWechseltAussenUndInnen(t *testing.T) {
	t.Parallel()

	form := service.Grundform(service.FormStern,
		msepaket.Vektor{X: 0.5, Y: 0.5}, msepaket.Vektor{X: 0.4, Y: 0.4}, 5, false)

	if len(form.Punkte) != 10 {
		t.Fatalf("%d Punkte, will 10", len(form.Punkte))
	}

	// **Auch der Stern ist an seiner Breite ausgerichtet** — er benutzt
	// dieselben `alpha` und `ra` wie das Vieleck.
	if breite := spanne(form.Punkte); math.Abs(breite-0.8) > 1e-9 {
		t.Errorf("die Breite ist %v, will 0.8", breite)
	}

	mitte := schwerpunkt(form.Punkte)

	aussen := punktabstand(form.Punkte[0].Stelle, mitte)
	innen := punktabstand(form.Punkte[1].Stelle, mitte)

	if innen >= aussen {
		t.Fatalf("innen %v ist nicht näher als außen %v", innen, aussen)
	}

	for stelle, punkt := range form.Punkte {
		will := aussen
		if stelle%2 == 1 {
			will = innen
		}

		if hat := punktabstand(punkt.Stelle, mitte); math.Abs(hat-will) > 1e-9 {
			t.Errorf("Punkt %d hat den Abstand %v, will %v", stelle, hat, will)
		}
	}
}

// **Der Name sagt, was es ist** — und bei drei bis sechs Ecken ist das kein
// „Vieleck".
func TestEineGrundformTraegtIhrenNamen(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		art   service.Formart
		ecken int
		eng   bool
		will  string
	}{
		{service.FormKreis, 0, false, "Ellipse"},
		{service.FormKreis, 0, true, "Kreis"},
		{service.FormRechteck, 0, false, "Rechteck"},
		{service.FormRechteck, 0, true, "Quadrat"},
		{service.FormVieleck, 3, false, "Dreieck"},
		{service.FormVieleck, 4, false, "Raute"},
		{service.FormVieleck, 5, false, "Fünfeck"},
		{service.FormVieleck, 6, false, "Sechseck"},
		{service.FormVieleck, 7, false, "Vieleck"},
		{service.FormStern, 5, false, "Stern"},
	} {
		t.Run(fall.will, func(t *testing.T) {
			t.Parallel()

			form := service.Grundform(fall.art,
				msepaket.Vektor{X: 0.5, Y: 0.5}, msepaket.Vektor{X: 0.4, Y: 0.4},
				fall.ecken, fall.eng)

			if form.Name != fall.will {
				t.Errorf("die Form heißt %q", form.Name)
			}
		})
	}
}

func nahe(a, b msepaket.Vektor) bool {
	return math.Abs(a.X-b.X) < 1e-9 && math.Abs(a.Y-b.Y) < 1e-9
}

func punktabstand(a, b msepaket.Vektor) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

// spanne gibt die waagerechte Ausdehnung.
func spanne(punkte []msepaket.Kontrollpunkt) float64 {
	links, rechts := math.Inf(1), math.Inf(-1)
	for _, punkt := range punkte {
		links, rechts = math.Min(links, punkt.Stelle.X), math.Max(rechts, punkt.Stelle.X)
	}

	return rechts - links
}

func schwerpunkt(punkte []msepaket.Kontrollpunkt) msepaket.Vektor {
	var mitte msepaket.Vektor
	for _, punkt := range punkte {
		mitte.X += punkt.Stelle.X
		mitte.Y += punkt.Stelle.Y
	}

	mitte.X /= float64(len(punkte))
	mitte.Y /= float64(len(punkte))

	return mitte
}

func eckenname(ecken int) string {
	return strconv.Itoa(ecken) + " Ecken"
}

// **Weniger als drei Ecken gibt keine Form.**
//
// MSEs Editor kann das nicht: sein Zähler beginnt bei drei
// (`src/gui/symbol/basic_shape_editor.cpp`, `sides`, Zeile 55, Fassung
// f9d9356 vom 2021-07-06). Hier ist die
// Zahl ein Parameter, und ein Aufrufer kann jede hineinschreiben — eine Form
// mit zwei Punkten wäre eine Strecke, eine mit null eine leere Datei mit einem
// `part:`-Block darin.
//
// **Kein Fehler, sondern keine Punkte**: der Aufrufer sieht die leere Form und
// entscheidet; ein Fehler hier nähme ihm die Wahl und stünde an einer Stelle,
// die von Geometrie handelt und nicht von Bedienung.
func TestEineFormMitZuWenigEckenBleibtLeer(t *testing.T) {
	t.Parallel()

	for _, art := range []service.Formart{service.FormVieleck, service.FormStern} {
		t.Run(string(art), func(t *testing.T) {
			t.Parallel()

			for _, ecken := range []int{0, 1, 2} {
				form := service.Grundform(art,
					msepaket.Vektor{X: 0.5, Y: 0.5}, msepaket.Vektor{X: 0.4, Y: 0.4},
					ecken, false)

				if len(form.Punkte) != 0 {
					t.Errorf("%d Ecken gaben %d Punkte", ecken, len(form.Punkte))
				}
			}
		})
	}
}

// **Eine unbekannte Art gibt eine leere Form** — sie hat weder Namen noch
// Punkte, und der Aufrufer sieht es.
func TestEineUnbekannteFormartBleibtLeer(t *testing.T) {
	t.Parallel()

	form := service.Grundform(service.Formart("Kegelschnitt"),
		msepaket.Vektor{X: 0.5, Y: 0.5}, msepaket.Vektor{X: 0.4, Y: 0.4}, 5, false)

	if form.Name != "" || len(form.Punkte) != 0 {
		t.Errorf("die Form ist %+v", form)
	}
}
