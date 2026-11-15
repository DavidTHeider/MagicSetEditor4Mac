// Das Zufallspack
// (#1072).
//
// # Warum das hier steht
//
// `msepaket` liest, **was in der Datei steht**. Wie daraus ein Booster wird —
// die Tiefen, die Gewichte, das Ziehen — ist eine fachliche Regel über
// mehrere Einheiten (§2).
//
// # Der Filter wird eingesetzt, nicht gebaut
//
// Eine Packart wählt ihre Karten über einen MSE-Script-Ausdruck über `card`.
// Ihn hier auszuwerten hieße, den Planbauer und die ganze Umgebung in den
// Zieher zu holen — und ein Test bräuchte dann ein Spielpaket. Der Aufrufer
// gibt stattdessen die **schon gefilterten** Karten je Packart.

package service

import (
	"fmt"
	"math/rand/v2"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Packzieher zieht Karten nach den Packarten eines Spiels.
type Packzieher struct {
	arten map[string]msepaket.Pack

	// reihenfolge ist die Reihenfolge der Datei. **Sie zählt**: MSE erzeugt je
	// Tiefe „in game file order" (`src/data/pack.cpp`, Zeile 405, Fassung (Fassung f9d9356 vom 2021-07-06)
	// f9d9356 vom 2021-07-06), und eine Abbildung hat keine Reihenfolge.
	reihenfolge []string

	// karten sind die Karten je Packart — vom Aufrufer gefiltert.
	karten map[string][]*domain.CardRow
}

// NeuerPackzieher baut ihn.
func NeuerPackzieher(
	arten []msepaket.Pack, karten map[string][]*domain.CardRow,
) *Packzieher {
	zieher := &Packzieher{
		arten:       make(map[string]msepaket.Pack, len(arten)),
		reihenfolge: make([]string, 0, len(arten)),
		karten:      karten,
	}

	for _, art := range arten {
		if _, doppelt := zieher.arten[art.Name]; doppelt {
			// **Die erste gewinnt.** MSE sucht linear und nimmt den ersten
			// Treffer (`PackGenerator::get`, Zeilen 440–456).
			continue
		}

		zieher.arten[art.Name] = art
		zieher.reihenfolge = append(zieher.reihenfolge, art.Name)
	}

	return zieher
}

// packstand ist eine Packart im Lauf.
type packstand struct {
	art     msepaket.Pack
	karten  []*domain.CardRow
	gewicht float64
	tiefe   int

	// verlangt sind die Züge, die noch offen sind.
	verlangt int
}

// Ziehen zieht aus einer Packart und gibt die Karten.
//
// # Der Ablauf
//
// **Von der größten Tiefe zur kleinsten** (`PackGenerator::generate`, Zeilen
// 399–420): eine Packart darf nur tiefere nennen, und beim Erzeugen ändert
// sie deren Zahl. Wer von oben nach unten ginge, zöge aus einem Posten,
// bevor jemand ihn verlangt hat.
//
// **Der Same macht den Lauf wiederholbar.** Ohne ihn wäre kein Test möglich,
// und das steht so in #889.
func (z *Packzieher) Ziehen(name string, same uint64) ([]*domain.CardRow, error) {
	staende, fehler := z.staende(name)
	if fehler != nil {
		return nil, fehler
	}

	staende[name].verlangt = 1

	// **Ein aussäbarer Würfel und ausdrücklich kein kryptografischer.**
	// Gefragt ist Wiederholbarkeit: derselbe Same gibt dieselbe Ziehung, und
	// ohne diese Zusage ist kein Test möglich (#889). `crypto/rand` hat
	// keinen Samen und wäre hier die falsche Wahl, nicht die sicherere.
	//
	// **Der zweite Wert ist der goldene Schnitt in 64 Bit.** PCG braucht zwei
	// Wörter; zweimal denselben zu geben verkürzte die Folge.
	// #nosec G404 -- gefragt ist Wiederholbarkeit, nicht Unvorhersagbarkeit
	wuerfel := rand.New(rand.NewPCG(same, same^0x9E3779B97F4A7C15))

	tiefste := 0
	for _, stand := range staende {
		tiefste = max(tiefste, stand.tiefe)
	}

	gezogen := make([]*domain.CardRow, 0, 16)

	for tiefe := tiefste; tiefe >= 0; tiefe-- {
		for _, art := range z.reihenfolge {
			stand := staende[art]
			if stand == nil || stand.tiefe != tiefe {
				continue
			}

			gezogen = z.erzeuge(staende, stand, wuerfel, gezogen)
		}
	}

	return gezogen, nil
}

// staende baut die Stände aller erreichbaren Packarten.
//
// **Ein Kreis wird abgewiesen, nicht durchlaufen.** MSE hätte dort eine
// endlose Rekursion; hier ist es ein Befund mit dem Weg darin — dieselbe
// Haltung wie beim Einbinder (`msepaket.Einbinder`).
func (z *Packzieher) staende(name string) (map[string]*packstand, error) {
	staende := map[string]*packstand{}

	var bauen func(string, []string) error

	bauen = func(name string, imLauf []string) error {
		if stelle := indexVon(imLauf, name); stelle >= 0 {
			return fmt.Errorf("%w: die Packarten nennen sich im Kreis: %v",
				domain.ErrValidation, append(imLauf[stelle:], name))
		}

		if _, fertig := staende[name]; fertig {
			return nil
		}

		art, bekannt := z.arten[name]
		if !bekannt {
			return fmt.Errorf("%w: die Packart %q gibt es nicht", domain.ErrNotFound, name)
		}

		for _, posten := range art.Posten {
			if fehler := bauen(posten.Name, append(imLauf, name)); fehler != nil {
				return fehler
			}
		}

		stand := &packstand{art: art, karten: z.karten[name]}
		stand.tiefe = tiefeAus(art, staende)
		stand.gewicht = gewichtAus(art, stand.karten, staende)

		staende[name] = stand

		return nil
	}

	return staende, bauen(name, nil)
}

// tiefeAus ist `depth` — eins über der tiefsten genannten Art.
//
// `src/data/pack.cpp`, Zeilen 140–145, Fassung f9d9356 vom 2021-07-06.
func tiefeAus(art msepaket.Pack, staende map[string]*packstand) int {
	tiefe := 0

	for _, posten := range art.Posten {
		tiefe = max(tiefe, 1+staende[posten.Name].tiefe)
	}

	return tiefe
}

// gewichtAus ist `total_weight` (`src/data/pack.cpp`, Zeilen 119–139). (Fassung f9d9356 vom 2021-07-06)
//
// **Je Auswahlart anders**, und das ist keine Feinheit: `proportional` wiegt
// einen Posten mit dem Gewicht **seiner** Karten, `nonempty` zählt ihn nur,
// wenn er welche hat, und `first` nimmt den ersten, der überhaupt etwas hat.
func gewichtAus(
	art msepaket.Pack, karten []*domain.CardRow, staende map[string]*packstand,
) float64 {
	if art.Auswahl == "first" {
		gewicht := 0.0
		if len(karten) > 0 {
			gewicht = 1
		}

		for _, posten := range art.Posten {
			if gewicht > 0 {
				break
			}

			gewicht = posten.Gewicht
		}

		return gewicht
	}

	gewicht := float64(len(karten))

	// **Ohne Wache auf `nil`.** Die Stände der Posten sind vorher gebaut —
	// `staende` geht in die Tiefe, bevor es den eigenen anlegt.
	for _, posten := range art.Posten {
		unten := staende[posten.Name]

		switch art.Auswahl {
		case "proportional", "equal proportional":
			gewicht += posten.Gewicht * unten.gewicht

		case "nonempty", "equal nonempty":
			if unten.gewicht > 0 {
				gewicht += posten.Gewicht
			}

		default:
			gewicht += posten.Gewicht
		}
	}

	return gewicht
}

// indexVon gibt die Stelle eines Namens — oder −1.
func indexVon(namen []string, gesucht string) int {
	for stelle, name := range namen {
		if name == gesucht {
			return stelle
		}
	}

	return -1
}
