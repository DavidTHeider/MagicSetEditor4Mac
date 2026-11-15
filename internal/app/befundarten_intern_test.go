//go:build fremdpakete

// Die **Befundarten** des Kanonlaufs — deckeln
// (#513).
//
// Eigene Datei, weil `kanonlauf_intern_test.go` mit ihnen über 400 Zeilen
// ginge (§2). Das **Zählen** steht seit
// #497 in
// `kanongruende_intern_test.go`, wo auch die Karte dazu gemerkt wird.
package app

import "testing"

// pruefeRate hält eine Befundart gegen ihre Rate je 1.000 Karten.
func pruefeRate(t *testing.T, art string, befunde, karten int, deckel float64) {
	t.Helper()

	if karten == 0 {
		return
	}

	rate := float64(befunde) * 1000 / float64(karten)
	if rate > deckel {
		t.Errorf("%d Befunde der Art %s bei %d Karten — %.2f je 1.000, Deckel %.2f",
			befunde, art, karten, rate, deckel)
	}
}
