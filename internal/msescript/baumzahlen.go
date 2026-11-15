package msescript

// **Der Baumspeicher zeigt sich und lässt sich leeren** — OPS-08
// (#1725).

import "sync/atomic"

// baumtreffer und baumfehlgriffe zählen die Abfragen.
var baumtreffer, baumfehlgriffe atomic.Int64

// Speicherzahlen sind Treffer, Fehlgriffe und Belegung eines Speichers.
type Speicherzahlen struct {
	Treffer, Fehlgriffe        int64
	Eintraege, GrenzeEintraege int
	Bytes, GrenzeBytes         int64
}

// Baumzahlen gibt den Stand des Baumspeichers.
func Baumzahlen() Speicherzahlen {
	baumsperre.RLock()
	defer baumsperre.RUnlock()

	return Speicherzahlen{
		Treffer: baumtreffer.Load(), Fehlgriffe: baumfehlgriffe.Load(),
		Eintraege: len(gemerkteBaeume), GrenzeEintraege: GrenzeBaumspeicher,
		Bytes: baumbelegung, GrenzeBytes: baumgrenzeBytes,
	}
}

// BaumspeicherLeeren wirft die gemerkten Bäume weg — die Zähler bleiben.
func BaumspeicherLeeren() {
	baumsperre.Lock()
	defer baumsperre.Unlock()

	gemerkteBaeume = map[string]*Knoten{}
	baumbelegung = 0
}
