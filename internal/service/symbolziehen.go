// **Ziehen und Sperren im Symboleditor** —
// #890,
// Werkstück 3 von Phase X.
//
// Portiert aus `ControlPoint::onUpdateHandle` und `ControlPoint::onUpdateLock`
// (`src/data/symbol.cpp`, Zeilen 55–75, Fassung f9d9356 vom 2021-07-06).
//
// # Was eine Sperre ist
//
// Sie bindet die beiden Griffe eines Punktes aneinander, **während jemand
// zeichnet**. In der Datei steht sie nur, damit sie nicht verloren geht; für
// das Rastern ist sie ohne Belang. Genau deshalb steht sie hier und nicht in
// `msepaket`: sie ist eine Regel der Bedienung.

package service

import (
	"math"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// GriffZiehen verschiebt einen Griff um `um` und zieht den anderen nach, wie
// die Sperre es verlangt.
//
//	if (lock == LOCK_DIR) {
//	  getOther(wh) = -getHandle(wh) * getOther(wh).length() / getHandle(wh).length();
//	} else if (lock == LOCK_SIZE) {
//	  getOther(wh) = -getHandle(wh);
//	}
//
// `src/data/symbol.cpp`, Zeilen 55–62, Fassung f9d9356 vom 2021-07-06.
//
// **`direction` behält die Länge des anderen** und dreht nur seine Richtung;
// `size` macht ihn zum genauen Gegenteil. Der Unterschied ist der zwischen
// „glatt" und „glatt und symmetrisch", und er ist am Symbol zu sehen.
func GriffZiehen(
	punkt msepaket.Kontrollpunkt, davor bool, um msepaket.Vektor,
) msepaket.Kontrollpunkt {
	gezogen, anderer := &punkt.GriffDanach, &punkt.GriffDavor
	if davor {
		gezogen, anderer = anderer, gezogen
	}

	*gezogen = plus(*gezogen, um)

	switch punkt.Sperre {
	case "size":
		*anderer = mal(*gezogen, -1)
	case "direction":
		// **Bei Länge null bleibt der andere stehen.** MSE teilt hier durch
		// `getHandle(wh).length()`; ein Griff der Länge null gäbe dort eine
		// Division durch null. Der Fall ist erreichbar — jemand zieht den
		// Griff genau auf den Punkt.
		if lang := laengeVon(*gezogen); lang != 0 {
			*anderer = mal(*gezogen, -laengeVon(*anderer)/lang)
		}
	}

	return punkt
}

// SperreSetzen setzt die Sperre und mittelt, was dasteht.
//
//	if (lock == LOCK_DIR) {
//	  Vector2D dir = (delta_before - delta_after).normalized();
//	  delta_before = dir * delta_before.length();
//	  delta_after  = dir * -delta_after.length();
//	} else if (lock == LOCK_SIZE) {
//	  delta_before = (delta_before - delta_after) * 0.5;
//	  delta_after  = -delta_before;
//	}
//
// `src/data/symbol.cpp`, Zeilen 63–75, Fassung f9d9356 vom 2021-07-06.
//
// **Sie schnappt nicht auf einen der beiden.** Wer beim Sperren den einen Griff
// auf den anderen setzte, verschöbe die Kurve sichtbar; MSE nimmt die Mitte
// und behält bei `direction` sogar beide Längen.
func SperreSetzen(punkt msepaket.Kontrollpunkt, sperre string) msepaket.Kontrollpunkt {
	punkt.Sperre = sperre

	switch sperre {
	case "size":
		punkt.GriffDavor = mal(minus(punkt.GriffDavor, punkt.GriffDanach), 0.5)
		punkt.GriffDanach = mal(punkt.GriffDavor, -1)
	case "direction":
		richtung := einheit(minus(punkt.GriffDavor, punkt.GriffDanach))
		punkt.GriffDavor = mal(richtung, laengeVon(punkt.GriffDavor))
		punkt.GriffDanach = mal(richtung, -laengeVon(punkt.GriffDanach))
	}

	return punkt
}

// laengeVon gibt die Länge eines Vektors.
func laengeVon(v msepaket.Vektor) float64 { return math.Hypot(v.X, v.Y) }

// einheit gibt den Vektor der Länge eins in derselben Richtung.
//
// **Der Nullvektor bleibt der Nullvektor.** `Vector2D::normalized` teilt bei
// MSE ohne Prüfung; hier wäre das eine stille Ansammlung von `NaN` in einer
// Datei, die jemand später öffnet.
func einheit(v msepaket.Vektor) msepaket.Vektor {
	lang := laengeVon(v)
	if lang == 0 {
		return v
	}

	return mal(v, 1/lang)
}
