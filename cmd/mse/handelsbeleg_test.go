package main

// **`mse handelsbeleg`** — der Beleg eines Handels, ohne einen Schritt zu
// gehen (Hüllenprobe `nat04handel`, #1672). Der geglückte Weg steht neben
// dem Dienst in `internal/flaechen/handelsbeleg_test.go`; hier die Absagen.

import "testing"

func TestHandelsbelegVerlangtGenauEinenVorgang(t *testing.T) {
	ort := eigenesZuhause(t)

	for _, argumente := range [][]string{
		{"handelsbeleg"},
		{"handelsbeleg", "v-1", "v-2"},
	} {
		if code, _, _ := fahreAn(t, ort, argumente...); code != CodeBedienung {
			t.Errorf("%v gab %d", argumente, code)
		}
	}
}

func TestEinUnbekannterHandelHatKeinenBelegAnDerKommandozeile(t *testing.T) {
	ort := eigenesZuhause(t)

	if code, _, fehler := fahreAn(t, ort, "handelsbeleg", "v-gibt-es-nicht"); code != CodeFachlich {
		t.Errorf("es kam %d (%s)", code, fehler)
	}
}
