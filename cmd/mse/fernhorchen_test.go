package main

// **Über fremde Netze erreichbar sein, ohne Fenster** — D04
// (#2235).
//
// **Der Weg selbst läuft gegen das echte Netz** und steht deshalb hinter
// `netzlauf` (`internal/flaechen/fernhandel_netzlauf_test.go`). Hier steht,
// was ohne Netz zu sagen ist: der Befehl nimmt kein Argument.

import (
	"strings"
	"testing"
)

func TestFernhorchenNimmtKeinArgument(t *testing.T) {
	ort, _ := eingefahren(t)

	code, _, klage := fahreAn(t, ort, "fernhorchen", "127.0.0.1:0")

	// **Am Satz und nicht am Rückgabewert allein**: ein unbekannter Befehl
	// gäbe denselben Wert, und der Lauf wäre grün, ohne dass es den Befehl
	// gibt — so stand er im ersten Anlauf.
	if code != CodeBedienung || !strings.Contains(klage, "fernhorchen nimmt kein Argument") {
		t.Errorf("Rückgabewert %d, erwartet %d mit der Absage des Befehls: %s", code, CodeBedienung, klage)
	}
}
