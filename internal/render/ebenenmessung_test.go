package render

// **Kalt und warm, gemessen** — OPS-08.AK-04
// (#1725).
// Die Zahlen stehen in `docs/ZWISCHENSPEICHER.md`, mit dem Befehl, der sie
// erhebt; in der Pipeline läuft keine Messung.

import (
	"image"
	"testing"
)

func BenchmarkEbeneKaltUndWarm(b *testing.B) {
	quelle := image.NewNRGBA(image.Rect(0, 0, 750, 1050))
	rahmen := image.Rect(0, 0, 375, 523)

	b.Run("kalt", func(b *testing.B) {
		for b.Loop() {
			NeuerRasterer(nil).abgetasteteEbene(quelle, rahmen, GueteHoch, true)
		}
	})

	b.Run("warm", func(b *testing.B) {
		r := NeuerRasterer(nil)
		r.abgetasteteEbene(quelle, rahmen, GueteHoch, true)

		for b.Loop() {
			r.abgetasteteEbene(quelle, rahmen, GueteHoch, true)
		}
	})
}
