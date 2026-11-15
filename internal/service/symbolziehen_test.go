package service_test

import (
	"math"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Einen Punkt ziehen** —
// #890,
// Werkstück 3 von Phase X.
//
// **Einen Griff ziehen — und was die Sperre daraus macht.**
//
//	if (lock == LOCK_DIR) {
//	  getOther(wh) = -getHandle(wh) * getOther(wh).length() / getHandle(wh).length();
//	} else if (lock == LOCK_SIZE) {
//	  getOther(wh) = -getHandle(wh);
//	}
//
// `src/data/symbol.cpp`, Zeilen 55–62, Fassung f9d9356 vom 2021-07-06.
func TestEinGezogenerGriffZiehtDenAnderenNachSeinerSperre(t *testing.T) {
	t.Parallel()

	grund := msepaket.Kontrollpunkt{
		Stelle:      msepaket.Vektor{X: 0.5, Y: 0.5},
		GriffDavor:  msepaket.Vektor{X: -0.2, Y: 0},
		GriffDanach: msepaket.Vektor{X: 0.1, Y: 0},
	}

	t.Run("free lässt den anderen stehen", func(t *testing.T) {
		t.Parallel()

		punkt := grund
		punkt.Sperre = "free"

		gezogen := service.GriffZiehen(punkt, true, msepaket.Vektor{X: 0, Y: 0.3})

		if !nahe(gezogen.GriffDanach, grund.GriffDanach) {
			t.Errorf("der andere Griff ist %+v", gezogen.GriffDanach)
		}
	})

	t.Run("size spiegelt ihn", func(t *testing.T) {
		t.Parallel()

		punkt := grund
		punkt.Sperre = "size"

		gezogen := service.GriffZiehen(punkt, true, msepaket.Vektor{X: 0, Y: 0.3})

		// Der gezogene Griff ist (-0.2, 0.3); der andere ist sein Gegenteil.
		if !nahe(gezogen.GriffDanach, msepaket.Vektor{X: 0.2, Y: -0.3}) {
			t.Errorf("der andere Griff ist %+v", gezogen.GriffDanach)
		}
	})

	t.Run("direction dreht ihn und behält seine Länge", func(t *testing.T) {
		t.Parallel()

		punkt := grund
		punkt.Sperre = "direction"

		gezogen := service.GriffZiehen(punkt, true, msepaket.Vektor{X: 0, Y: 0.3})

		// **Die Länge bleibt** — 0.1, wie vorher.
		if hat := laenge(gezogen.GriffDanach); math.Abs(hat-0.1) > 1e-9 {
			t.Errorf("die Länge ist %v, will 0.1", hat)
		}

		// **Und die Richtung ist die entgegengesetzte.**
		if !entgegengesetzt(gezogen.GriffDavor, gezogen.GriffDanach) {
			t.Errorf("die Griffe zeigen nicht entgegengesetzt: %+v und %+v",
				gezogen.GriffDavor, gezogen.GriffDanach)
		}
	})
}

// **Eine Sperre zu setzen mittelt, was dasteht.**
//
//	if (lock == LOCK_SIZE) {
//	  delta_before = (delta_before - delta_after) * 0.5;
//	  delta_after  = -delta_before;
//	}
//
// `src/data/symbol.cpp`, Zeilen 63–75, Fassung f9d9356 vom 2021-07-06.
//
// **Sie schnappt nicht auf einen der beiden.** Wer beim Sperren den einen Griff
// auf den anderen setzte, verschöbe die Kurve — MSE nimmt die Mitte.
func TestEineNeueSperreMitteltDieGriffe(t *testing.T) {
	t.Parallel()

	punkt := msepaket.Kontrollpunkt{
		GriffDavor:  msepaket.Vektor{X: -0.2, Y: 0},
		GriffDanach: msepaket.Vektor{X: 0.4, Y: 0},
		Sperre:      "free",
	}

	t.Run("size mittelt die Länge", func(t *testing.T) {
		t.Parallel()

		gesperrt := service.SperreSetzen(punkt, "size")

		// Die Mitte von minus 0,2 und 0,4 ist minus 0,3.
		if !nahe(gesperrt.GriffDavor, msepaket.Vektor{X: -0.3, Y: 0}) {
			t.Errorf("der Griff davor ist %+v", gesperrt.GriffDavor)
		}

		if !nahe(gesperrt.GriffDanach, msepaket.Vektor{X: 0.3, Y: 0}) {
			t.Errorf("der Griff danach ist %+v", gesperrt.GriffDanach)
		}
	})

	t.Run("direction behält beide Längen", func(t *testing.T) {
		t.Parallel()

		schief := msepaket.Kontrollpunkt{
			GriffDavor:  msepaket.Vektor{X: -0.2, Y: 0.1},
			GriffDanach: msepaket.Vektor{X: 0.4, Y: 0.3},
			Sperre:      "free",
		}

		gesperrt := service.SperreSetzen(schief, "direction")

		for name, paar := range map[string][2]msepaket.Vektor{
			"davor":  {gesperrt.GriffDavor, schief.GriffDavor},
			"danach": {gesperrt.GriffDanach, schief.GriffDanach},
		} {
			if hat, will := laenge(paar[0]), laenge(paar[1]); math.Abs(hat-will) > 1e-9 {
				t.Errorf("die Länge %s ist %v, war %v", name, hat, will)
			}
		}

		if !entgegengesetzt(gesperrt.GriffDavor, gesperrt.GriffDanach) {
			t.Errorf("die Griffe zeigen nicht entgegengesetzt: %+v und %+v",
				gesperrt.GriffDavor, gesperrt.GriffDanach)
		}
	})

	t.Run("free lässt beide stehen", func(t *testing.T) {
		t.Parallel()

		gesperrt := service.SperreSetzen(punkt, "free")

		if !nahe(gesperrt.GriffDavor, punkt.GriffDavor) ||
			!nahe(gesperrt.GriffDanach, punkt.GriffDanach) {
			t.Errorf("die Griffe sind %+v und %+v", gesperrt.GriffDavor, gesperrt.GriffDanach)
		}
	})
}

func laenge(v msepaket.Vektor) float64 { return math.Hypot(v.X, v.Y) }

// entgegengesetzt sagt, ob zwei Vektoren in genau entgegengesetzte Richtungen
// zeigen.
func entgegengesetzt(a, b msepaket.Vektor) bool {
	la, lb := laenge(a), laenge(b)
	if la == 0 || lb == 0 {
		return false
	}

	// Das Kreuzprodukt ist null (gleiche Gerade) und das Skalarprodukt negativ.
	return math.Abs(a.X*b.Y-a.Y*b.X) < 1e-9 && a.X*b.X+a.Y*b.Y < 0
}

// **Ein Griff der Länge null bricht nichts.**
//
// MSE teilt an beiden Stellen ohne Prüfung — `getHandle(wh).length()` in
// `onUpdateHandle` und `normalized()` in `onUpdateLock`
// (`src/data/symbol.cpp`, Zeilen 58 und 67, Fassung f9d9356 vom 2021-07-06).
//
// **Der Fall ist erreichbar**: jemand zieht den Griff genau auf seinen Punkt.
// Bei MSE entstünde dort ein `NaN`, und ein `NaN` in einer Symboldatei fällt
// erst auf, wenn jemand sie öffnet. Hier bleibt der andere Griff stehen — eine
// **benannte Abweichung** (§15, Regel 7), und die bessere Antwort: sie ändert
// nichts, statt etwas Unbrauchbares zu schreiben.
func TestEinGriffDerLaengeNullBrichtNichts(t *testing.T) {
	t.Parallel()

	t.Run("beim Ziehen", func(t *testing.T) {
		t.Parallel()

		punkt := msepaket.Kontrollpunkt{
			GriffDavor:  msepaket.Vektor{X: 0.1, Y: 0.1},
			GriffDanach: msepaket.Vektor{X: 0.2, Y: 0},
			Sperre:      "direction",
		}

		// Der Griff davor wird genau auf null gezogen.
		gezogen := service.GriffZiehen(punkt, true, msepaket.Vektor{X: -0.1, Y: -0.1})

		if !nahe(gezogen.GriffDanach, punkt.GriffDanach) {
			t.Errorf("der andere Griff ist %+v", gezogen.GriffDanach)
		}

		if math.IsNaN(gezogen.GriffDanach.X) || math.IsNaN(gezogen.GriffDanach.Y) {
			t.Error("im anderen Griff steht NaN")
		}
	})

	t.Run("beim Sperren", func(t *testing.T) {
		t.Parallel()

		// **Beide Griffe gleich** — ihre Differenz ist der Nullvektor, und
		// `normalized()` teilte dort durch null.
		punkt := msepaket.Kontrollpunkt{
			GriffDavor:  msepaket.Vektor{X: 0.2, Y: 0.1},
			GriffDanach: msepaket.Vektor{X: 0.2, Y: 0.1},
			Sperre:      "free",
		}

		gesperrt := service.SperreSetzen(punkt, "direction")

		for _, griff := range []msepaket.Vektor{gesperrt.GriffDavor, gesperrt.GriffDanach} {
			if math.IsNaN(griff.X) || math.IsNaN(griff.Y) {
				t.Errorf("in einem Griff steht NaN: %+v", griff)
			}
		}
	})
}
