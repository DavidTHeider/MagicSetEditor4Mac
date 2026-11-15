package main

import (
	"strings"
	"testing"
)

// **Die Suche an der Kommandozeile findet keine fremde Karte** — ENG-05
// (#1730).
//
// Nachgestellt am 2026-09-23: im eigenen Profil fand `mse suche Feuer` die
// Karte eines anderen Nutzers — die Abfrage ging ohne Set über den ganzen
// Bestand. Hier an der Fläche, dazu die Gegenprobe: der Eigner findet sie.
func TestDieSucheFindetKeineFremdeKarte(t *testing.T) {
	ort, _, _ := mitKarte(t)

	if code, _, fehler := fahreAn(t, ort, "nutzer-neu", "Bob"); code != CodeGut {
		t.Fatalf("Bob anlegen: %d (%s)", code, fehler)
	}

	if code, _, fehler := fahreAn(t, ort, "nutzer-wechseln", "2"); code != CodeGut {
		t.Fatalf("zu Bob wechseln: %d (%s)", code, fehler)
	}

	if code, aus, fehler := fahreAn(t, ort, "suche", "Feuer"); code != CodeGut ||
		strings.Contains(aus, "Feuerball") {
		t.Errorf("Bob findet die Karte des ersten Nutzers: %d %q (%s)", code, aus, fehler)
	}

	if code, _, fehler := fahreAn(t, ort, "nutzer-wechseln", "1"); code != CodeGut {
		t.Fatalf("zurückwechseln: %d (%s)", code, fehler)
	}

	if code, aus, fehler := fahreAn(t, ort, "suche", "Feuer"); code != CodeGut ||
		!strings.Contains(aus, "Feuerball") {
		t.Errorf("der Eigner findet seine Karte nicht: %d %q (%s)", code, aus, fehler)
	}
}
