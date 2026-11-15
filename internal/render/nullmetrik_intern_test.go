package render

import (
	"encoding/binary"
	"errors"
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// ohneTypotabelle nimmt einer Schriftdatei ihre `OS/2`-Tabelle.
//
// **Damit die Rückfallkette von hinten prüfbar bleibt.** Seit
// #597 steht
// `OS/2` typo **vor** den Rändern; `masseLesen` und `raenderLesen` lassen sich
// austauschen, die rohen Bytes nicht. Wer die Ränder prüfen will, muss die
// Tabelle davor wegnehmen.
//
// **Der Eintrag wird umbenannt, nicht entfernt** — das Verzeichnis behält
// seine Länge, und `sfnt` liest die Datei weiter.
func ohneTypotabelle(t *testing.T, roh []byte) []byte {
	t.Helper()

	kopie := make([]byte, len(roh))
	copy(kopie, roh)

	anzahl := int(binary.BigEndian.Uint16(kopie[4:]))
	for i := range anzahl {
		eintrag := kopie[tabellenkopf+i*tabelleneintrag:]
		if string(eintrag[0:4]) == "OS/2" {
			copy(eintrag[0:4], "XXXX")

			return kopie
		}
	}

	t.Skip("die Testschrift führt keine OS/2-Tabelle")

	return nil
}

// **Eine Schrift ohne `hhea` und ohne `OS/2` bekommt ihre Metrik aus den
// Rändern.**
//
// # Der Befund
//
// `x/image/font/sfnt` gibt für `MATRIX.TTF` aus `Full-Magic-Pack`
// **ascent = 0, descent = 0, height = 0** — und `err = nil`. Dieselbe Datei
// liefert brauchbare Ränder: `bounds.Min.Y = -49856` bei `em = 1000`, also rund
// 0,78 em über der Grundlinie. Betroffen sind `MATRIX.TTF` und
// `MatrixBold.ttf`; `mplantin.ttf`, `beleren-bold` und `ModMatrix.ttf` nennen
// ihre Metrik.
//
// # Was das gekostet hat
//
// Die Grundlinie ist `Zeilenoberkante + Größe × Ascent`. Mit Ascent null fällt
// sie auf die Oberkante, und **jeder Zug in dieser Schrift steht über seinem
// Kasten**. Auf der Archenemy-Karte sitzt der Illustratorname damit sieben
// Bildpunkte zu hoch — gemessen gegen MSEs eigene Ausgabe (y 563–569 statt
// 570–580, #417).
//
// # Warum das eine Anpassung ist und keine Portierung
//
// MSE holt seine Metrik vom Schriftdienst des Betriebssystems (`wxFont`), nicht
// aus der Datei — dort stellt sich die Frage nicht. Die Ersatzrechnung ist
// deshalb **adapted** und in `docs/PORTIERUNGSKARTE.md` so vermerkt
// (#418).
func TestEineSchriftOhneMetrikNimmtIhreRaender(t *testing.T) {
	vorher := masseLesen
	masseLesen = func(*sfnt.Font, *sfnt.Buffer, fixed.Int26_6) (font.Metrics, error) {
		return font.Metrics{}, nil
	}

	t.Cleanup(func() { masseLesen = vorher })

	schrift, err := SchriftAusDatei(ohneTypotabelle(t, schriftbytes(t)))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if schrift.Grundlinie() <= 0 {
		t.Errorf("Grundlinie %g — sie muss aus den Rändern kommen", schrift.Grundlinie())
	}

	if schrift.Zeilenhoehe() <= 0 {
		t.Errorf("Zeilenhöhe %g — sie muss aus den Rändern kommen", schrift.Zeilenhoehe())
	}
}

// **Gibt auch der Rand nichts her, wird die Schrift abgewiesen** — nicht mit
// Zeilenhöhe null durchgereicht, denn dann stünde jeder Zug über seinem Kasten
// und niemand wüsste, warum.
func TestEineSchriftOhneMetrikUndOhneRaenderWirdAbgewiesen(t *testing.T) {
	vorherMasse, vorherRaender := masseLesen, raenderLesen

	masseLesen = func(*sfnt.Font, *sfnt.Buffer, fixed.Int26_6) (font.Metrics, error) {
		return font.Metrics{}, nil
	}

	raenderLesen = func(*sfnt.Font, *sfnt.Buffer, fixed.Int26_6) (
		fixed.Rectangle26_6, error,
	) {
		return fixed.Rectangle26_6{}, nil
	}

	t.Cleanup(func() { masseLesen, raenderLesen = vorherMasse, vorherRaender })

	if _, err := SchriftAusDatei(ohneTypotabelle(t, schriftbytes(t))); !errors.Is(err, errOhneMetrik) {
		t.Errorf("Fehler %v, erwartet errOhneMetrik", err)
	}
}

// **Und ein Fehler beim Lesen der Ränder wird durchgereicht.**
func TestEinFehlerBeimRandlesenWirdGemeldet(t *testing.T) {
	vorherMasse, vorherRaender := masseLesen, raenderLesen

	masseLesen = func(*sfnt.Font, *sfnt.Buffer, fixed.Int26_6) (font.Metrics, error) {
		return font.Metrics{}, nil
	}

	raenderLesen = func(*sfnt.Font, *sfnt.Buffer, fixed.Int26_6) (
		fixed.Rectangle26_6, error,
	) {
		return fixed.Rectangle26_6{}, errNahtGestolpert
	}

	t.Cleanup(func() { masseLesen, raenderLesen = vorherMasse, vorherRaender })

	if _, err := SchriftAusDatei(ohneTypotabelle(t, schriftbytes(t))); !errors.Is(err, errNahtGestolpert) {
		t.Errorf("Fehler %v", err)
	}
}
