package render

// **Der Ebenenspeicher zählt** — OPS-08.AK-01
// (#1725):
// Treffer, Fehlgriffe und Belegung, damit seine Trefferquote beobachtbar ist
// statt behauptet — und er lässt sich leeren.

import (
	"image"
	"testing"
)

func TestDerEbenenspeicherZaehltUndLeert(t *testing.T) {
	t.Parallel()

	r := NeuerRasterer(nil)
	quelle := ebenenquelle(20)
	rahmen := image.Rect(0, 0, 8, 8)

	r.abgetasteteEbene(quelle, rahmen, GueteMittel, true)
	r.abgetasteteEbene(quelle, rahmen, GueteMittel, true)
	r.abgetasteteEbene(quelle, image.Rect(0, 0, 9, 8), GueteMittel, true)

	zahlen := r.Ebenenzahlen()
	if zahlen.Treffer != 1 || zahlen.Fehlgriffe != 2 || zahlen.Eintraege != 2 ||
		zahlen.Bytes != 8*8*4+9*8*4 || zahlen.GrenzeBytes != GrenzeEbenenspeicher {
		t.Errorf("%+v", zahlen)
	}

	r.EbenenLeeren()

	if nachher := r.Ebenenzahlen(); nachher.Eintraege != 0 || nachher.Bytes != 0 || nachher.Treffer != 1 {
		t.Errorf("nach dem Leeren: %+v", nachher)
	}
}
