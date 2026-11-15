// **Punkte setzen und löschen im Symboleditor** —
// #890,
// Werkstück 3 von Phase X.
//
// Portiert aus `ControlPointAddAction` und `control_point_remove_action`
// (`src/data/action/symbol_part.cpp`, Zeilen 253–291 und 427–435) sowie
// `deCasteljau` (`src/gfx/bezier.cpp`, Zeilen 49–60), Fassung f9d9356 vom
// 2021-07-06.

package service

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"

// PunktEinfuegen setzt einen Punkt hinter `hinter` auf die Strecke zum
// nächsten, an der Stelle `t` zwischen 0 und 1.
//
// Gibt `false`, wenn `hinter` außerhalb liegt.
//
// # Warum die Nachbarn sich ändern
//
// Bei einer Kurve wird nicht ein Punkt **hinzugefügt**, sondern die Kurve
// **geteilt**. Die beiden Hälften einer Bézierkurve haben andere Griffe als das
// Ganze — wer nur den neuen Punkt ausrechnet und die Nachbarn stehen lässt,
// bekommt eine Kurve, die woanders verläuft als vorher.
//
// **Und deshalb wird `size` zu `direction`**: die beiden Hälften sind nicht
// mehr gleich lang, und eine Sperre, die Gleichheit erzwingt, verschöbe die
// Kurve beim nächsten Anfassen (`symbol_part.cpp`, Zeilen 265–267).
func PunktEinfuegen(form msepaket.Symbolteil, hinter int, t float64) (msepaket.Symbolteil, bool) {
	if hinter < 0 || hinter >= len(form.Punkte) {
		return form, false
	}

	punkte := make([]msepaket.Kontrollpunkt, len(form.Punkte))
	copy(punkte, form.Punkte)

	danach := (hinter + 1) % len(punkte)
	erster, zweiter := punkte[hinter], punkte[danach]

	var neuer msepaket.Kontrollpunkt

	if erster.KurveDanach {
		neuer = geteilt(&erster, &zweiter, t)
	} else {
		neuer = msepaket.Kontrollpunkt{
			Stelle: msepaket.Vektor{
				X: erster.Stelle.X*(1-t) + zweiter.Stelle.X*t,
				Y: erster.Stelle.Y*(1-t) + zweiter.Stelle.Y*t,
			},
			Sperre: "free",
		}
	}

	punkte[hinter], punkte[danach] = erster, zweiter

	// **Eine eigene Scheibe und kein `append` in die alte.** Die Nachbarn sind
	// gerade geändert worden; ein `append`, das die vorhandene Scheibe
	// weiterschreibt, überschriebe sie, sobald sie Platz hätte.
	mitNeuem := make([]msepaket.Kontrollpunkt, 0, len(punkte)+1)
	mitNeuem = append(mitNeuem, punkte[:hinter+1]...)
	mitNeuem = append(mitNeuem, neuer)
	mitNeuem = append(mitNeuem, punkte[hinter+1:]...)

	form.Punkte = mitNeuem

	return form, true
}

// geteilt rechnet die Teilung einer Bézierkurve aus und kürzt die Nachbarn.
//
//	Vector2D half21   = a21 * t;
//	Vector2D half34   = a34 * (1-t);
//	Vector2D mid23    = (a1 + a21) * (1-t) + (a34 + a4) * t;
//	Vector2D mid23h21 = (a1 + half21) * (1-t) + mid23 * t;
//	Vector2D mid23h34 = (a4 + half34) * t     + mid23 * (1-t);
//	out.pos          = mid23h21 * (1-t) + mid23h34 * t;
//	out.delta_before = mid23h21 - out.pos;
//	out.delta_after  = mid23h34 - out.pos;
//	a21 = half21;
//	a34 = half34;
//
// `src/gfx/bezier.cpp`, Zeilen 49–60, Fassung f9d9356 vom 2021-07-06. `a21`
// ist der Griff **nach** dem ersten Punkt, `a34` der **vor** dem zweiten; die
// letzten beiden Zeilen schreiben sie zurück.
func geteilt(erster, zweiter *msepaket.Kontrollpunkt, t float64) msepaket.Kontrollpunkt {
	a1, a4 := erster.Stelle, zweiter.Stelle
	a21, a34 := erster.GriffDanach, zweiter.GriffDavor

	halb21 := mal(a21, t)
	halb34 := mal(a34, 1-t)

	mitte23 := plus(mal(plus(a1, a21), 1-t), mal(plus(a34, a4), t))
	mitte21 := plus(mal(plus(a1, halb21), 1-t), mal(mitte23, t))
	mitte34 := plus(mal(plus(a4, halb34), t), mal(mitte23, 1-t))

	stelle := plus(mal(mitte21, 1-t), mal(mitte34, t))

	erster.GriffDanach, zweiter.GriffDavor = halb21, halb34
	entsperren(erster)
	entsperren(zweiter)

	return msepaket.Kontrollpunkt{
		Stelle:      stelle,
		GriffDavor:  minus(mitte21, stelle),
		GriffDanach: minus(mitte34, stelle),
		KurveDavor:  true,
		KurveDanach: true,
		Sperre:      "direction",
	}
}

// entsperren löst eine Größensperre.
//
//	if (point1.other.lock == LOCK_SIZE) point1.other.lock = LOCK_DIR;
//
// Ebd., Zeile 265. Eine andere Sperre bleibt, wie sie ist.
func entsperren(punkt *msepaket.Kontrollpunkt) {
	if punkt.Sperre == "size" {
		punkt.Sperre = "direction"
	}
}

// PunktLoeschen entfernt den Punkt an `stelle`.
//
// Gibt `false`, wenn die Stelle außerhalb liegt oder danach **weniger als
// zwei** Punkte blieben.
//
//	if (shape->points.size() - to_delete.size() < 2) {
//	  return unique_ptr<Action>(); // no action
//	}
//
// `src/data/action/symbol_part.cpp`, Zeilen 428–431, Fassung f9d9356 vom
// 2021-07-06. **Die Grenze ist zwei, nicht drei** — MSE lässt eine Form mit
// zwei Punkten stehen, obwohl sie nichts umschließt. Nachgelesen, nicht
// gerundet; der Kommentar daneben („TODO : remove part?") sagt, dass MSE es
// selbst für ungeklärt hält.
func PunktLoeschen(form msepaket.Symbolteil, stelle int) (msepaket.Symbolteil, bool) {
	if stelle < 0 || stelle >= len(form.Punkte) || len(form.Punkte)-1 < 2 {
		return form, false
	}

	punkte := make([]msepaket.Kontrollpunkt, 0, len(form.Punkte)-1)
	punkte = append(punkte, form.Punkte[:stelle]...)
	punkte = append(punkte, form.Punkte[stelle+1:]...)

	form.Punkte = punkte

	return form, true
}

func plus(a, b msepaket.Vektor) msepaket.Vektor {
	return msepaket.Vektor{X: a.X + b.X, Y: a.Y + b.Y}
}

func minus(a, b msepaket.Vektor) msepaket.Vektor {
	return msepaket.Vektor{X: a.X - b.X, Y: a.Y - b.Y}
}

func mal(a msepaket.Vektor, faktor float64) msepaket.Vektor {
	return msepaket.Vektor{X: a.X * faktor, Y: a.Y * faktor}
}
