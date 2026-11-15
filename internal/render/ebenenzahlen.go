package render

// **Der Ebenenspeicher zeigt sich und lässt sich leeren** — OPS-08
// (#1725).

// Speicherzahlen sind Treffer, Fehlgriffe und Belegung eines Speichers.
type Speicherzahlen struct {
	Treffer, Fehlgriffe int64
	Eintraege           int
	Bytes, GrenzeBytes  int64
}

// Ebenenzahlen gibt den Stand des Ebenenspeichers.
func (r *Rasterer) Ebenenzahlen() Speicherzahlen {
	r.ebenen.sperre.Lock()
	defer r.ebenen.sperre.Unlock()

	return Speicherzahlen{
		Treffer: r.ebenen.treffer.Load(), Fehlgriffe: r.ebenen.fehlgriffe.Load(),
		Eintraege: len(r.ebenen.bilder), Bytes: int64(r.ebenen.platz), GrenzeBytes: GrenzeEbenenspeicher,
	}
}

// EbenenLeeren wirft die gemerkten Ebenen weg — die Zähler bleiben.
func (r *Rasterer) EbenenLeeren() {
	r.ebenen.sperre.Lock()
	defer r.ebenen.sperre.Unlock()

	r.ebenen.bilder = nil
	r.ebenen.platz = 0
}
