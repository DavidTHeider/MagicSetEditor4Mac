//go:build fremdpakete

// **Die geschichtete Ziehung nimmt aus jeder Schicht**
// (#610).
//
// Der Grund, warum es sie gibt, steht in `schichten_test.go`. Hier steht, was
// sie zusagt — und die Zusage, die zählt, ist die **seltene** Schicht: eine
// verteilte Ziehung trifft die häufige ohnehin.
package app

import "testing"

// probenkarten baut n Karten mit fortlaufender Kennung.
func probenkarten(n int) []KarteDTO {
	karten := make([]KarteDTO, n)
	for i := range karten {
		karten[i] = KarteDTO{ID: string(rune('a' + i%26)), Nummer: i}
	}

	return karten
}

// TestDieZiehungTrifftJedeSchicht ist die Gegenprobe zum Befund aus #610.
//
// **Der erste Fall ist der gemessene**: `Colorless` trägt 3.216 Karten, davon
// 25 mit `conspiracy` in zwei Schichten von 13 und 12. Eine verteilte Ziehung
// von 68 fand **null** davon; die geschichtete findet sie.
func TestDieZiehungTrifftJedeSchicht(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		groessen []int
		anzahl   int
	}{
		{"Colorless: 3.191 gewöhnliche, 13 und 12 seltene", []int{3191, 13, 12}, 68},
		{"eine Schicht", []int{100}, 10},
		{"mehr Schichten als Plätze — die seltenen gewinnen", []int{50, 40, 3, 2, 1}, 3},
		{"jede Schicht eine Karte", []int{1, 1, 1, 1}, 4},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			karten, schichten := aufbauen(fall.groessen)
			gezogen := geschichteteStichprobe(karten, schichten, fall.anzahl)

			pruefeZiehung(t, gezogen, karten, schichten, fall.groessen, fall.anzahl)
		})
	}
}

// aufbauen baut Karten und ihre Schichten aus den Schichtgrößen.
func aufbauen(groessen []int) (karten []KarteDTO, schichten []string) {
	for nummer, groesse := range groessen {
		for range groesse {
			schichten = append(schichten, string(rune('A'+nummer)))
		}
	}

	return probenkarten(len(schichten)), schichten
}

// pruefeZiehung hält die drei Zusagen der Ziehung gegen ein Ergebnis.
func pruefeZiehung(
	t *testing.T,
	gezogen, karten []KarteDTO, schichten []string, groessen []int, anzahl int,
) {
	t.Helper()

	if len(gezogen) != min(anzahl, len(karten)) {
		t.Errorf("gezogen = %d, erwartet %d", len(gezogen), min(anzahl, len(karten)))
	}

	getroffen := map[string]bool{}
	for _, karte := range gezogen {
		getroffen[schichtVon(karte, karten, schichten)] = true
	}

	// **Die seltensten zuerst.** Reicht der Platz nicht für alle Schichten,
	// müssen die kleinen darunter sein — sie sind der Grund für die Schichtung.
	erwartet := min(anzahl, len(groessen))
	if len(getroffen) < erwartet {
		t.Errorf("getroffene Schichten = %d, erwartet mindestens %d (%v)",
			len(getroffen), erwartet, getroffen)
	}

	// **Keine Karte zweimal.**
	gesehen := map[int]bool{}
	for _, karte := range gezogen {
		if gesehen[karte.Nummer] {
			t.Errorf("Karte %d steht zweimal in der Ziehung", karte.Nummer)
		}

		gesehen[karte.Nummer] = true
	}
}

// schichtVon sucht die Schicht, aus der eine gezogene Karte stammt.
func schichtVon(gezogen KarteDTO, karten []KarteDTO, schichten []string) string {
	for stelle := range karten {
		if karten[stelle].Nummer == gezogen.Nummer {
			return schichten[stelle]
		}
	}

	return ""
}
