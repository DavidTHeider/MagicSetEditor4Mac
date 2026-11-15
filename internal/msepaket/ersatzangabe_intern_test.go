// **Die Dateiangaben des Ersatzes** —
// #1164.
//
// # Warum sie einzeln geprüft werden
//
// [fs.File] verlangt ein `Stat`, und `fs.FileInfo` verlangt sechs Methoden.
// Fünf davon fragt beim Lesen einer Datei niemand — `fs.ReadFile` braucht die
// Größe und sonst nichts. Sie sind trotzdem die **Antwort dieses Pakets** auf
// eine Frage, die ein Aufrufer stellen darf; ungeprüft stünde dort, was beim
// Schreiben gerade einfiel.
//
// # Was hier nicht geprüft wird — und ein Fehler wäre
//
// **`Dateien()` zeigt eine ersetzte Datei nicht, die es nur im Ersatz gibt.**
// Der Baum darunter kennt sie nicht, und `fs.WalkDir` fragt ihn. Für die
// Vorschau spielt das keine Rolle: sie liest **eine** Datei bei Namen. Wer
// `Dateien()` über einem Ersatz aufzählt, bekommt die Liste der Platte — das
// steht hier, damit es niemand für einen Zufall hält.
package msepaket

import (
	"io/fs"
	"testing"
	"time"
)

func TestDieAngabenDerErsatzdatei(t *testing.T) {
	t.Parallel()

	datei := &ersatzdatei{name: "style", inhalt: []byte("mse version: 2.0.0\n")}

	angabe, err := datei.Stat()
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	if angabe.Name() != "style" {
		t.Errorf("Name %q", angabe.Name())
	}

	if angabe.Size() != int64(len(datei.inhalt)) {
		t.Errorf("Size %d", angabe.Size())
	}

	// **Kein Modus und keine Zeit.** Beide wären erfunden: der Ersatz liegt
	// nirgends, und eine Uhrzeit daran behauptete, er sei irgendwann
	// geschrieben worden.
	if angabe.Mode() != fs.FileMode(0) {
		t.Errorf("Mode %v", angabe.Mode())
	}

	if !angabe.ModTime().Equal(time.Time{}) {
		t.Errorf("ModTime %v", angabe.ModTime())
	}

	if angabe.IsDir() {
		t.Error("die Ersatzdatei gibt sich als Verzeichnis aus")
	}

	if angabe.Sys() != nil {
		t.Errorf("Sys %v", angabe.Sys())
	}
}

// **Gelesen wird bis zum Ende und dann nichts mehr.**
func TestDieErsatzdateiListSichLeer(t *testing.T) {
	t.Parallel()

	datei := &ersatzdatei{name: "style", inhalt: []byte("ab")}

	nach := make([]byte, 1)

	for erwartet := range []byte("ab") {
		gelesen, err := datei.Read(nach)
		if err != nil || gelesen != 1 {
			t.Fatalf("Lesen %d: %d Bytes, %v", erwartet, gelesen, err)
		}
	}

	if _, err := datei.Read(nach); err == nil {
		t.Error("nach dem Ende kam noch etwas")
	}

	if err := datei.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}
