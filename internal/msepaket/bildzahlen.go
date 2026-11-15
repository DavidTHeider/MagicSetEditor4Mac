package msepaket

// **Die Bildquelle zeigt sich und lässt sich leeren** — OPS-08
// (#1725).

import "image"

// Speicherzahlen sind Treffer, Fehlgriffe und Belegung eines Speichers.
type Speicherzahlen struct {
	Treffer, Fehlgriffe        int64
	Eintraege, GrenzeEintraege int
	Bytes, GrenzeBytes         int64
}

// Zahlen gibt den Stand der gemerkten Bilder.
func (b *Bildquelle) Zahlen() Speicherzahlen {
	b.sperre.RLock()
	defer b.sperre.RUnlock()

	return Speicherzahlen{
		Treffer: b.treffer.Load(), Fehlgriffe: b.fehlgriffe.Load(),
		Eintraege: len(b.gemerkt), GrenzeEintraege: GrenzeBildspeicher,
		Bytes: b.belegt, GrenzeBytes: b.grenzeBytes,
	}
}

// Leeren wirft die gemerkten Bilder weg — die Zähler bleiben.
func (b *Bildquelle) Leeren() {
	b.sperre.Lock()
	defer b.sperre.Unlock()

	b.gemerkt = make(map[string]image.Image, GrenzeBildspeicher)
	b.belegt = 0
}
