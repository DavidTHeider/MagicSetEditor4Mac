package fenster

// **Die Maus der Leitung: ziehen, loslassen, rechts klicken** — A04
// (#2218).
//
// Eine eigene Datei, seit `abnahme.go` mit ihr über 400 Zeilen ging (§2).
// **Ohne Build-Tag**: die Abfolge eines Zugs wird ohne AppKit geprüft; nur
// das Schicken (`abnahme_hand_darwin.go`) braucht den Abnahmebau.
//
// **Was sie belegt und was nicht**: die Ereignisse gehen ab der
// Ereignisschlange von AppKit denselben Weg wie die einer echten Maus —
// Trefferprüfung, Mausverfolgung der Webansicht, WebKits Zeigerereignisse
// mit `isTrusted`. Der Fensterserver liefert sie nicht; ein Zug aus einer
// anderen Anwendung (dem Finder) oder über den Bildschirmrand hinaus ist
// damit nicht belegt.

import (
	"errors"
	"fmt"
)

// Ort ist eine Stelle der Seite, in Punkten.
type Ort struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Die Arten aus `NSEvent.h` (`NSEventType`), nachgelesen im SDK.
const (
	artMausRunter   uint64 = 1
	artMausHoch     uint64 = 2
	artRechtsRunter uint64 = 3
	artRechtsHoch   uint64 = 4
	artMausBewegt   uint64 = 5
	artMausGezogen  uint64 = 6
	artTasteRunter  uint64 = 10
	artTasteHoch    uint64 = 11
)

// mausschritt ist ein Ereignis eines Zugs: welche Art, an welcher Stelle —
// und für einen Klick, der wievielte in Folge er ist und mit welchem
// Umschalter (P4, #2372). Null Klicks schickt AppKit als einen.
type mausschritt struct {
	art      uint64
	ort      Ort
	klicks   int
	schalter uint64
}

var (
	errZugZuKurz     = errors.New("ein Zug braucht mindestens zwei Punkte")
	errKlickzahl     = errors.New("ein Klick zählt einen bis drei")
	errKlickschalter = errors.New("unbekannter Umschalter an einem Klick")
)

// klickschritte gibt einen bis drei Klicks an derselben Stelle — **mit der
// Klickzahl**, wie AppKit sie bei einer echten Maus zählt: WebKit liest
// daraus `detail`, und beim zweiten feuert `dblclick`. Null heißt einer.
func klickschritte(ort Ort, anzahl int, mit []string) ([]mausschritt, error) {
	if anzahl == 0 {
		anzahl = 1
	}

	if anzahl < 1 || anzahl > 3 {
		return nil, fmt.Errorf("%w: %d", errKlickzahl, anzahl)
	}

	var schalter uint64

	for _, name := range mit {
		wert, da := umschalter[name]
		if !da {
			return nil, fmt.Errorf("%w: %q", errKlickschalter, name)
		}

		schalter |= wert
	}

	schritte := make([]mausschritt, 0, 2*anzahl)
	for klick := 1; klick <= anzahl; klick++ {
		schritte = append(schritte,
			mausschritt{art: artMausRunter, ort: ort, klicks: klick, schalter: schalter},
			mausschritt{art: artMausHoch, ort: ort, klicks: klick, schalter: schalter})
	}

	return schritte, nil
}

// zugschritte gibt die Abfolge eines Zugs, wie AppKit sie bei einem echten
// Ziehen liefert: runter am ersten Punkt, gezogen durch die übrigen, hoch am
// letzten — außer `halten` lässt die Taste unten.
func zugschritte(weg []Ort, halten bool) ([]mausschritt, error) {
	if len(weg) < 2 {
		return nil, errZugZuKurz
	}

	schritte := []mausschritt{{art: artMausRunter, ort: weg[0]}}
	for _, ort := range weg[1:] {
		schritte = append(schritte, mausschritt{art: artMausGezogen, ort: ort})
	}

	if !halten {
		schritte = append(schritte, mausschritt{art: artMausHoch, ort: weg[len(weg)-1]})
	}

	return schritte, nil
}

// rechtsklickschritte gibt einen Klick mit der rechten Taste.
func rechtsklickschritte(ort Ort) []mausschritt {
	return []mausschritt{{art: artRechtsRunter, ort: ort}, {art: artRechtsHoch, ort: ort}}
}

// bewegeschritte gibt eine Bewegung des Zeigers **ohne** gedrückte Taste.
func bewegeschritte(ort Ort) []mausschritt {
	return []mausschritt{{art: artMausBewegt, ort: ort}}
}
