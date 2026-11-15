package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Einen Punkt in eine Gerade setzen** —
// #890,
// Werkstück 3 von Phase X.
//
//	new_point->pos = point1.other.pos * (1 - t) + point2.other.pos * t;
//	new_point->lock = LOCK_FREE;
//
// `src/data/action/symbol_part.cpp`, Zeilen 270–274, Fassung f9d9356 vom
// 2021-07-06.
func TestEinPunktInEinerGeradenLiegtDazwischen(t *testing.T) {
	t.Parallel()

	form := msepaket.Symbolteil{Art: msepaket.TeilForm, Punkte: []msepaket.Kontrollpunkt{
		{Stelle: msepaket.Vektor{X: 0, Y: 0}, Sperre: "free"},
		{Stelle: msepaket.Vektor{X: 1, Y: 0}, Sperre: "free"},
	}}

	neu, gut := service.PunktEinfuegen(form, 0, 0.25)
	if !gut {
		t.Fatal("das Einfügen ging nicht")
	}

	if len(neu.Punkte) != 3 {
		t.Fatalf("%d Punkte", len(neu.Punkte))
	}

	// **Hinter dem ersten**, nicht am Ende.
	if hat := neu.Punkte[1].Stelle; !nahe(hat, msepaket.Vektor{X: 0.25, Y: 0}) {
		t.Errorf("der neue Punkt ist %+v", hat)
	}

	if neu.Punkte[1].KurveDanach || neu.Punkte[1].KurveDavor {
		t.Error("aus einer Geraden wurde eine Kurve")
	}

	if neu.Punkte[1].Sperre != "free" {
		t.Errorf("die Sperre ist %q", neu.Punkte[1].Sperre)
	}
}

// **In eine Kurve setzen heißt teilen** — und die Nachbarn ändern sich mit.
//
//	out.pos          = mid23h21 * (1-t) + mid23h34 * t;
//	out.delta_before = mid23h21 - out.pos;
//	out.delta_after  = mid23h34 - out.pos;
//	a21 = half21;
//	a34 = half34;
//
// `src/gfx/bezier.cpp`, Zeilen 49–60, Fassung f9d9356 vom 2021-07-06.
//
// **Die letzten beiden Zeilen sind der Punkt.** Wer nur den neuen Punkt
// ausrechnet und die Nachbarn stehen lässt, bekommt eine Kurve, die woanders
// verläuft als vorher — sichtbar, sobald jemand einen Punkt setzt.
func TestEinPunktInEinerKurveLiegtAufIhr(t *testing.T) {
	t.Parallel()

	// Ein Viertelkreis: von (0,0) nach (1,1), Griffe nach rechts und oben.
	form := msepaket.Symbolteil{Art: msepaket.TeilForm, Punkte: []msepaket.Kontrollpunkt{
		{
			Stelle: msepaket.Vektor{X: 0, Y: 0}, GriffDanach: msepaket.Vektor{X: 0.5, Y: 0},
			KurveDanach: true, Sperre: "size",
		},
		{
			Stelle: msepaket.Vektor{X: 1, Y: 1}, GriffDavor: msepaket.Vektor{X: 0, Y: -0.5},
			KurveDavor: true, Sperre: "size",
		},
	}}

	vorher := aufDerKurve(form.Punkte[0], form.Punkte[1], 0.5)

	neu, gut := service.PunktEinfuegen(form, 0, 0.5)
	if !gut {
		t.Fatal("das Einfügen ging nicht")
	}

	// **Der neue Punkt liegt genau dort, wo die Kurve bei t verlief.**
	if hat := neu.Punkte[1].Stelle; !nahe(hat, vorher) {
		t.Errorf("der neue Punkt ist %+v, die Kurve war bei %+v", hat, vorher)
	}

	// **Und die Nachbargriffe sind gekürzt** — sonst verliefe die Kurve anders.
	if hat := neu.Punkte[0].GriffDanach; !nahe(hat, msepaket.Vektor{X: 0.25, Y: 0}) {
		t.Errorf("der Griff des ersten ist %+v, will die Hälfte", hat)
	}

	if hat := neu.Punkte[2].GriffDavor; !nahe(hat, msepaket.Vektor{X: 0, Y: -0.25}) {
		t.Errorf("der Griff des letzten ist %+v, will die Hälfte", hat)
	}

	// **Und `size` wird zu `direction`**: die beiden Hälften einer geteilten
	// Kurve sind nicht mehr gleich lang (`symbol_part.cpp`, Zeilen 265–267).
	for stelle, punkt := range neu.Punkte {
		if punkt.Sperre != "direction" {
			t.Errorf("Punkt %d hat die Sperre %q, will direction", stelle, punkt.Sperre)
		}
	}
}

// **Und die geteilte Kurve verläuft wie die alte.**
//
// Das ist die Zusage, für die es de Casteljau überhaupt gibt — und sie ist
// prüfbar, ohne die Formel zu wiederholen: an zehn Stellen nachgemessen.
func TestDieGeteilteKurveVerlaeuftWieDieAlte(t *testing.T) {
	t.Parallel()

	form := msepaket.Symbolteil{Art: msepaket.TeilForm, Punkte: []msepaket.Kontrollpunkt{
		{
			Stelle: msepaket.Vektor{X: 0, Y: 0}, GriffDanach: msepaket.Vektor{X: 0.5, Y: 0.2},
			KurveDanach: true, Sperre: "free",
		},
		{
			Stelle: msepaket.Vektor{X: 1, Y: 1}, GriffDavor: msepaket.Vektor{X: -0.3, Y: -0.5},
			KurveDavor: true, Sperre: "free",
		},
	}}

	const geteiltBei = 0.3

	neu, gut := service.PunktEinfuegen(form, 0, geteiltBei)
	if !gut {
		t.Fatal("das Einfügen ging nicht")
	}

	for schritt := range 11 {
		wo := float64(schritt) / 10
		alt := aufDerKurve(form.Punkte[0], form.Punkte[1], wo)

		var hat msepaket.Vektor
		if wo <= geteiltBei {
			hat = aufDerKurve(neu.Punkte[0], neu.Punkte[1], wo/geteiltBei)
		} else {
			hat = aufDerKurve(neu.Punkte[1], neu.Punkte[2], (wo-geteiltBei)/(1-geteiltBei))
		}

		if !nahe(hat, alt) {
			t.Errorf("bei t=%.1f verläuft die geteilte Kurve durch %+v, die alte durch %+v",
				wo, hat, alt)
		}
	}
}

// **Unter drei Punkten wird nicht gelöscht.**
//
//	if (shape->points.size() - to_delete.size() < 2) {
//	  return unique_ptr<Action>(); // no action
//	}
//
// `src/data/action/symbol_part.cpp`, Zeilen 428–431, Fassung f9d9356 vom
// 2021-07-06. **Die Grenze ist zwei, nicht drei** — MSE lässt eine Form mit
// zwei Punkten stehen, obwohl sie nichts umschließt. Nachgelesen, nicht
// gerundet.
func TestEinPunktWirdNichtGeloeschtWennZuWenigBleiben(t *testing.T) {
	t.Parallel()

	form := msepaket.Symbolteil{Art: msepaket.TeilForm, Punkte: []msepaket.Kontrollpunkt{
		{Stelle: msepaket.Vektor{X: 0, Y: 0}},
		{Stelle: msepaket.Vektor{X: 1, Y: 0}},
		{Stelle: msepaket.Vektor{X: 1, Y: 1}},
	}}

	if _, gut := service.PunktLoeschen(form, 0); !gut {
		t.Error("von drei auf zwei ging nicht — MSEs Grenze ist zwei")
	}

	zwei := msepaket.Symbolteil{Art: msepaket.TeilForm, Punkte: form.Punkte[:2]}
	if _, gut := service.PunktLoeschen(zwei, 0); gut {
		t.Error("von zwei auf eins ging — MSE lässt das nicht zu")
	}
}

// aufDerKurve rechnet einen Punkt der Bézierkurve zwischen zwei
// Kontrollpunkten aus.
//
// **Eigens hier und nicht aus dem Quelltext geholt**: eine Prüfung, die
// dieselbe Funktion benutzt wie die Umsetzung, prüft nichts.
func aufDerKurve(a, b msepaket.Kontrollpunkt, t float64) msepaket.Vektor {
	p0 := a.Stelle
	p1 := msepaket.Vektor{X: a.Stelle.X + a.GriffDanach.X, Y: a.Stelle.Y + a.GriffDanach.Y}
	p2 := msepaket.Vektor{X: b.Stelle.X + b.GriffDavor.X, Y: b.Stelle.Y + b.GriffDavor.Y}
	p3 := b.Stelle

	u := 1 - t

	return msepaket.Vektor{
		X: u*u*u*p0.X + 3*u*u*t*p1.X + 3*u*t*t*p2.X + t*t*t*p3.X,
		Y: u*u*u*p0.Y + 3*u*u*t*p1.Y + 3*u*t*t*p2.Y + t*t*t*p3.Y,
	}
}

// **Eine Stelle außerhalb ändert nichts.**
//
// Beide Wege bekommen sie: `hinter` kommt bei MSE aus einem Mausklick auf eine
// Strecke, `stelle` aus einer Auswahl — hier kommen beide über eine Naht, und
// eine Naht bekommt jede Zahl.
func TestEineStelleAusserhalbAendertNichts(t *testing.T) {
	t.Parallel()

	form := msepaket.Symbolteil{Art: msepaket.TeilForm, Punkte: []msepaket.Kontrollpunkt{
		{Stelle: msepaket.Vektor{X: 0, Y: 0}},
		{Stelle: msepaket.Vektor{X: 1, Y: 0}},
		{Stelle: msepaket.Vektor{X: 1, Y: 1}},
	}}

	for _, stelle := range []int{-1, 3, 99} {
		if _, gut := service.PunktEinfuegen(form, stelle, 0.5); gut {
			t.Errorf("das Einfügen hinter %d ging", stelle)
		}

		if _, gut := service.PunktLoeschen(form, stelle); gut {
			t.Errorf("das Löschen an %d ging", stelle)
		}
	}
}

// **Und hinter dem letzten Punkt wird auf die Schlussstrecke gesetzt.**
//
// Ein Symbolzug ist **geschlossen**: die Strecke vom letzten zum ersten Punkt
// gehört dazu. `shape->getPoint(insert_after + 1)` läuft bei MSE über den Rand
// (`mod`, `src/data/symbol.hpp`, Zeilen 191–194, Fassung f9d9356 vom
// 2021-07-06); wer dort abbräche, ließe eine Kante unteilbar.
func TestHinterDemLetztenPunktWirdDieSchlussstreckeGeteilt(t *testing.T) {
	t.Parallel()

	form := msepaket.Symbolteil{Art: msepaket.TeilForm, Punkte: []msepaket.Kontrollpunkt{
		{Stelle: msepaket.Vektor{X: 0, Y: 0}, Sperre: "free"},
		{Stelle: msepaket.Vektor{X: 1, Y: 0}, Sperre: "free"},
		{Stelle: msepaket.Vektor{X: 1, Y: 1}, Sperre: "free"},
	}}

	neu, gut := service.PunktEinfuegen(form, 2, 0.5)
	if !gut {
		t.Fatal("das Einfügen hinter dem letzten ging nicht")
	}

	if len(neu.Punkte) != 4 {
		t.Fatalf("%d Punkte", len(neu.Punkte))
	}

	// **Zwischen (1,1) und (0,0)** — die Schlussstrecke.
	if hat := neu.Punkte[3].Stelle; !nahe(hat, msepaket.Vektor{X: 0.5, Y: 0.5}) {
		t.Errorf("der neue Punkt ist %+v", hat)
	}
}
