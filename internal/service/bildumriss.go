// **Ein Bild nachzeichnen: der Rundgang** —
// #1340,
// Werkstück F von Phase 4.11.
//
// Portiert aus `find_symbol_shape_start` und `read_symbol_shape`
// (`src/data/format/image_to_symbol.cpp`, Zeilen 129–204, Fassung `f9d9356`
// vom 2021-07-06).
//
// # Wie der Rundgang geht
//
// Der Zeiger steht **zwischen** vier Zellen, nicht auf einer:
//
//	a b
//	 ·
//	c d
//
// Aus den vier Wahrheitswerten wird eine Zahl `0xabcd`, und eine Tabelle sagt,
// wohin der nächste Schritt geht. Das sind vierzehn Fälle; die beiden fehlenden
// — alle vier voll und alle vier leer — kann der Rundgang nicht erreichen, denn
// dann stünde er in der Luft oder im Boden. MSE wirft dort einen Fehler, und
// dieser Satz ist der Grund, warum hier auch einer steht.
//
// **Zwei Fälle sind zweideutig** (`0x0110` und `0x1001`): eine Diagonale lässt
// sich von zwei Seiten betreten. MSE entscheidet sie über die **letzte**
// Bewegung, und das ist keine Feinheit — ohne sie liefe der Rundgang an einer
// Sanduhrform in die falsche Hälfte.

package service

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// umrissAnfang sucht die erste volle Zelle, über der eine leere liegt.
//
// **Spaltenweise und nicht zeilenweise** — MSE läuft `x` außen (Zeilen
// 130–131). Die Reihenfolge bestimmt, welche Form zuerst gefunden wird, und
// damit die Reihenfolge der Teile in der fertigen Zeichnung.
func umrissAnfang(feld *schwellenfeld) (x, y int, gefunden bool) {
	for spalte := range feld.breite {
		for zeile := range feld.hoehe {
			// **Die Zelle darüber muss leer sein**: sonst fänge der Rundgang
			// mitten in einer Fläche an. Und sie darf nicht schon markiert
			// sein, sonst fände er dieselbe Form wieder.
			if feld.voll(spalte, zeile) && !feld.voll(spalte, zeile-1) &&
				!feld.gemerkt(spalte, zeile) {
				return spalte, zeile, true
			}
		}
	}

	return 0, 0, false
}

// umrissLesen läuft **einmal** um eine Form herum.
func umrissLesen(feld *schwellenfeld) (msepaket.Symbolteil, bool) {
	// **Die Schranke.** MSE hat keine; sie steht hier, weil ein Feld mit einer
	// Zelle, die der Tabelle widerspricht, sonst niemals zurückkäme — und ein
	// hängendes Fenster ist die schlechteste aller Antworten.
	//
	// Acht Schritte je Zelle sind großzügig: ein Rundgang berührt jede Kante
	// höchstens zweimal.
	return umrissLesenMitSchranke(feld, 8*(feld.breite+2)*(feld.hoehe+2))
}

// umrissLesenMitSchranke ist dasselbe mit einer gesetzten Obergrenze.
//
// **Sie steht getrennt, damit ein Lauf sie prüfen kann.** Über ein echtes Bild
// ist die Schranke unerreichbar; ein Fehlerweg, den niemand betritt, ist aber
// trotzdem einer (§3).
func umrissLesenMitSchranke(feld *schwellenfeld, schranke int) (msepaket.Symbolteil, bool) {
	xs, ys, gefunden := umrissAnfang(feld)
	if !gefunden {
		return msepaket.Symbolteil{}, false
	}

	feld.merken(xs, ys)

	// **Einen nach rechts** — sonst deutete die erste Bewegung so, als käme
	// der Zeiger von oben (Zeile 152).
	xs++

	x, y := xs, ys
	altX, altY := x, y
	letzte := 1

	punkte := make([]msepaket.Kontrollpunkt, 0, 64)

	for schritt := 0; ; schritt++ {
		if schritt > schranke {
			return msepaket.Symbolteil{}, false
		}

		x, y = naechsterSchritt(feld, x, y, letzte)

		punkte = append(punkte, msepaket.Kontrollpunkt{
			Stelle: msepaket.Vektor{
				X: float64(x) / float64(feld.breite),
				Y: float64(y) / float64(feld.hoehe),
			},
		})

		// **Markiert wird nur beim Schritt nach rechts** — damit trägt jede
		// Form genau ihre Oberkante als Marke, und der nächste Anfang findet
		// sie nicht wieder (Zeile 191).
		if x > altX {
			feld.merken(altX, y)
		}

		letzte = (x + y) - (altX + altY)
		altX, altY = x, y

		if x == xs && y == ys {
			break
		}
	}

	// **Innen oder außen?** Zwei Zellen links vom Anfang: ist dort Fläche,
	// läuft der Zug um ein **Loch**, und MSE zieht es ab (Zeilen 199–202).
	kombination := msepaket.KombinationVereinen
	if feld.voll(x-2, y-1) {
		kombination = msepaket.KombinationAbziehen
	}

	return msepaket.Symbolteil{
		Art:         msepaket.TeilForm,
		Kombination: kombination,
		Punkte:      punkte,
	}, true
}

// naechsterSchritt ist MSEs Sprungtabelle über `0xabcd`.
//
// **Ein widersprüchlicher Stand bleibt stehen.** MSE wirft dort
// `InternalError("in the ground/air")` (Zeile 181); hier gibt es die Stelle
// unverändert zurück, und der Zug läuft in die Schranke.
//
// **Das ist ein Ausgang statt zweier**, und das ist der Grund: ein eigener
// Rückgabepfad für diesen Fall wäre über ein Feld nicht erreichbar — der
// Rundgang steht immer auf einer Kante, und eine Kante hat nie vier gleiche
// Nachbarn. Ein Zweig, den kein Lauf betreten kann, ist kein Fehlerweg,
// sondern eine Behauptung (§3). So bleibt **eine** Antwort auf einen
// unmöglichen Stand, und sie ist geprüft.
func naechsterSchritt(feld *schwellenfeld, x, y, letzte int) (neuX, neuY int) {
	zahl := 0
	for _, ecke := range [4]struct{ dx, dy, wert int }{
		{-1, -1, 0x1000},
		{0, -1, 0x0100},
		{-1, 0, 0x0010},
		{0, 0, 0x0001},
	} {
		if feld.voll(x+ecke.dx, y+ecke.dy) {
			zahl += ecke.wert
		}
	}

	switch zahl {
	case 0x0001, 0x0011, 0x1011:
		return x + 1, y
	case 0x0010, 0x1010, 0x1110:
		return x, y + 1
	case 0x0100, 0x0101, 0x0111:
		return x, y - 1
	case 0x1000, 0x1100, 0x1101:
		return x - 1, y
	case 0x0110:
		// **Zweideutig**: die Diagonale lässt sich von links wie von rechts
		// betreten, und die letzte Bewegung sagt, an welcher der beiden Zellen
		// der Zug hängt.
		return x, y - letzte
	case 0x1001:
		return x + letzte, y
	default:
		// **In der Luft oder im Boden** — alle vier gleich.
		return x, y
	}
}
