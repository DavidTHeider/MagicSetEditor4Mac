package service

// **Was das Zurücknehmen sammelt** —
// #1282,
// Befund 4.
//
// # Warum von innen
//
// Über den ganzen Weg kann `Bildablage.Wegnehmen` nicht scheitern: die Namen
// kommen aus `Bytes` und sind einfache Dateinamen, ein fehlendes Bild ist kein
// Fehler, und ein eben beschriebenes Verzeichnis lässt auch löschen. Der
// Fehlerzweig wäre damit einer, den kein Aufruf betritt (§3).
//
// Hier steht er als das, was er ist: eine Zusage über den Ernstfall, geprüft
// an der Stelle, an der sie steht.

import (
	"errors"
	"strings"
	"testing"
)

func TestDasZuruecknehmenSammeltSeineFehlschlaege(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Platte antwortet nicht")

	fehler := zuruecknehmen(
		map[string]string{"a": "image1", "b": "image2", "c": "image3"},
		map[string]bool{"image2": true},
		func(name string) error {
			if name == "image1" {
				return kaputt
			}

			return nil
		})

	if !errors.Is(fehler, kaputt) {
		t.Fatalf("Fehler = %v, erwartet %v", fehler, kaputt)
	}

	if !strings.Contains(fehler.Error(), "image1") {
		t.Errorf("die Meldung nennt das Bild nicht: %v", fehler)
	}
}

// **Ein benutztes Bild wird nicht angefasst** — das ist die andere Hälfte.
func TestDasZuruecknehmenLaesstBenutzteStehen(t *testing.T) {
	t.Parallel()

	var angefasst []string

	fehler := zuruecknehmen(
		map[string]string{"a": "image1", "b": "image2"},
		map[string]bool{"image1": true},
		func(name string) error {
			angefasst = append(angefasst, name)

			return nil
		})

	if fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if len(angefasst) != 1 || angefasst[0] != "image2" {
		t.Errorf("angefasst wurde %v — erwartet nur image2", angefasst)
	}
}
