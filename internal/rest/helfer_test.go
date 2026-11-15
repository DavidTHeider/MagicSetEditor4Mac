package rest_test

import (
	"os"
	"testing"
)

// probemarke ist eine Marke von der Länge einer echten.
//
// **Fest und nicht erzeugt.** Ein Test, der seine Marke selbst würfelt, prüft
// bei jedem Lauf einen anderen Fall; hier soll die Prüfung selbst geprüft
// werden, nicht das Würfeln. Dass eine erzeugte Marke lang und einmalig ist,
// steht in einem eigenen Fall.
const probemarke = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

// lies gibt eine Quelldatei dieses Pakets.
//
// **Für Zusagen über den Quelltext**, wo eine Messung am Verhalten nicht
// verlässlich wäre — und daneben steht dann, was sie deshalb nicht sagt.
func lies(t *testing.T, name string) string {
	t.Helper()

	roh, err := os.ReadFile(name) // #nosec G304 -- eine Quelldatei dieses Pakets
	if err != nil {
		t.Fatalf("%s lesen: %v", name, err)
	}

	return string(roh)
}
