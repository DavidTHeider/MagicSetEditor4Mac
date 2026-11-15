package vorschauweg

// **Beide Fächer zeigen sich, und das der Vorschauen lässt sich leeren** —
// OPS-08 (#1725).
// Das der Spielbilder leert [Texturfach.Raeume] schon.

// Zahlen sind Treffer, Fehlgriffe und Belegung eines Fachs.
type Zahlen struct {
	Treffer, Fehlgriffe int64
	Eintraege           int
	Bytes               int64
	GrenzeEintraege     int
	GrenzeBytes         int64
}

// Zahlen gibt den Stand des Fachs der Vorschauen.
func (f *Fach) Zahlen() Zahlen {
	f.schranke.Lock()
	defer f.schranke.Unlock()

	var bytes int64
	for _, bild := range f.bilder {
		bytes += int64(len(bild))
	}

	return Zahlen{
		Treffer: f.treffer, Fehlgriffe: f.fehlgriffe, Eintraege: len(f.bilder), Bytes: bytes,
		GrenzeEintraege: Faecher,
	}
}

// Leere wirft die Vorschauen weg — die Zähler bleiben.
func (f *Fach) Leere() {
	f.schranke.Lock()
	defer f.schranke.Unlock()

	f.bilder = nil
	f.reihe = nil
}

// Zahlen gibt den Stand des Fachs der Spielbilder.
func (t *Texturfach) Zahlen() Zahlen {
	t.schranke.Lock()
	defer t.schranke.Unlock()

	return Zahlen{
		Treffer: t.treffer, Fehlgriffe: t.fehlgriffe, Eintraege: len(t.bilder),
		Bytes: int64(t.belegt), GrenzeBytes: int64(t.budget),
	}
}
