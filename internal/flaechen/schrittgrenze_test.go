package flaechen_test

import (
	"testing"
	"time"
)

// **Vor einer Schrittgrenze wird gewartet** — gefunden in der Pipeline am
// 2026-10-12: `TestDerGanzeFaktorwegAnDerKommandozeile` bestätigte mit dem
// Code des vorigen Schritts, und die Kommandozeile prüfte ihn unter Last
// erst nach der nächsten Grenze — zwei Schritte alt, abgewiesen.
func TestVorEinerSchrittgrenzeWirdGewartet(t *testing.T) {
	t.Parallel()

	schritt := time.Unix(1_760_000_010, 0) // Beginn eines Schritts (durch 30 teilbar)

	for _, fall := range []struct {
		name   string
		jetzt  time.Time
		warten time.Duration
	}{
		{"am Anfang des Schritts", schritt, 0},
		{"zwanzig Sekunden hinein", schritt.Add(20 * time.Second), 0},
		{"fünf Sekunden vor der Grenze", schritt.Add(25 * time.Second), 5 * time.Second},
		{"eine halbe Sekunde vor der Grenze", schritt.Add(29500 * time.Millisecond), 500 * time.Millisecond},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if got := wartenBisZurGrenze(fall.jetzt); got != fall.warten {
				t.Errorf("gewartet wird %v, erwartet %v", got, fall.warten)
			}
		})
	}
}
