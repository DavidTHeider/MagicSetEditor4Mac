// **Ein Schreibfehler gibt den belegten Namen wieder frei** — der zweite
// Zweitblick über die Freigabe 0.3.0, Befund 4
// (#1170).
//
// # Der Befund
//
// `namenBelegen` reservierte den Namen mit `O_EXCL`, **schloss den Griff** und
// öffnete denselben Pfad danach ein zweites Mal mit `os.WriteFile`. Scheitert
// dieser zweite Zug, bleibt die reservierte Datei liegen — leer oder halb —,
// und der nächste Lauf überspringt ihren Namen, weil er belegt ist.
//
// Übrig bleibt ein Bild, das keines ist, unter einem Namen, den niemand mehr
// vergibt. **Und es geht in die nächste Sicherung mit.**
//
// # Warum der Schreiber hereinkommt
//
// Ein Fehlschlag genau zwischen Reservierung und Schreiben lässt sich nicht
// bestellen (§3) — dieselbe Lage wie bei `austauschen`, wo `umbenennen` aus
// demselben Grund ein Parameter ist. Der Schreiber ist die Naht, an der der
// Fehler eingespeist wird.
package service

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEinSchreibfehlerGibtDenBelegtenNamenWiederFrei(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	kaputt := errors.New("die Platte ist voll")

	_, err := namenBelegen(ort, "kunst", []byte("bild"), 8,
		func(io.WriteCloser, []byte) error { return kaputt })
	if !errors.Is(err, kaputt) {
		t.Fatalf("Fehler: %v, erwartet %v", err, kaputt)
	}

	eintraege, lesefehler := os.ReadDir(ort)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	for _, eintrag := range eintraege {
		t.Errorf("%q ist liegengeblieben — der Name bleibt damit für immer "+
			"belegt, und die Datei geht in die nächste Sicherung mit",
			eintrag.Name())
	}
}

// **Und der zweite Lauf bekommt denselben Namen wieder.** Das ist die Folge,
// um die es geht: eine Reservierung, die nach einem Fehlschlag stehen bleibt,
// verbraucht den Namen endgültig.
func TestNachEinemFehlschlagIstDerersteNameWiederFrei(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()

	if _, err := namenBelegen(ort, "kunst", []byte("bild"), 8,
		func(io.WriteCloser, []byte) error { return errors.New("kaputt") }); err == nil {
		t.Fatal("kein Fehler")
	}

	name, err := namenBelegen(ort, "kunst", []byte("bild"), 8, fuelleUndSchliesse)
	if err != nil {
		t.Fatalf("zweiter Lauf: %v", err)
	}

	if name != "kunst1" {
		t.Errorf("Name %q, erwartet kunst1 — der erste Lauf hat ihn behalten", name)
	}

	inhalt, lesefehler := os.ReadFile(filepath.Join(ort, name)) // #nosec G304 -- t.TempDir()
	if lesefehler != nil || string(inhalt) != "bild" {
		t.Errorf("Inhalt %q (%v), erwartet \"bild\"", strings.TrimSpace(string(inhalt)), lesefehler)
	}
}
