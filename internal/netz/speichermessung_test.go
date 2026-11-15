package netz_test

// **Kalt und warm, gemessen** — OPS-08.AK-04
// (#1725):
// kalt heißt hier ein Gang über das Netz, **nachgestellt mit 50 ms** — die
// Größenordnung einer Scryfall-Antwort aus Europa; warm das Lesen von der
// Platte.

import (
	"strconv"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

func BenchmarkSpeicherKaltUndWarm(b *testing.B) {
	antwort := make([]byte, 8<<10)
	holen := func() ([]byte, error) { //nolint:unparam // die Gestalt verlangt Speicher.Hole
		time.Sleep(50 * time.Millisecond)

		return antwort, nil
	}

	b.Run("kalt", func(b *testing.B) {
		speicher := netz.NeuerSpeicher(b.TempDir(), 24*time.Hour, time.Now)
		stelle := 0

		for b.Loop() {
			stelle++
			if _, _, fehler := speicher.Hole(b.Context(), "karte/"+strconv.Itoa(stelle), holen); fehler != nil {
				b.Fatal(fehler)
			}
		}
	})

	b.Run("warm", func(b *testing.B) {
		speicher := netz.NeuerSpeicher(b.TempDir(), 24*time.Hour, time.Now)
		if _, _, fehler := speicher.Hole(b.Context(), "karte", holen); fehler != nil {
			b.Fatal(fehler)
		}

		for b.Loop() {
			if _, _, fehler := speicher.Hole(b.Context(), "karte", holen); fehler != nil {
				b.Fatal(fehler)
			}
		}
	})
}
