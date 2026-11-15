package domain

// **Was ein Handel von einem Exemplar nimmt — je Variante** —
// #2492.
//
// # Der Fund
//
// **Unter „Zusagen" ließ sich jede Besitzzeile des Drucks festhalten**, auch
// eine in anderem Zustand, anderer Sprache oder Ausführung als vereinbart.
// Geprüft wurden Druck und Gesamtmenge. Verbucht ging dann das gewählte
// Exemplar weg, und drüben kam die **abgesprochene** Zeile an — die Belege
// passten nicht zusammen, und die Gegenzeichnung fiel still aus. Gemessen
// mit der Hüllenprobe `nat04handel` am 2026-10-19.
//
// # Die Regel
//
// **Eine Zusage gilt einer Zeile, die sie erfüllt** — Druck, Zustand, Sprache
// und Ausführung, mit denselben Gleichheiten wie überall:
// [Druckschluessel] für den Druck, [Exemplar.Ausfuehrung] für „leer heißt
// nonfoil". **Und die Grenze ist die der Variante**: zwei Zeilen desselben
// Drucks in zwei Zuständen sind zwei Grenzen und keine gemeinsame.

// Zusagevertrag sind die Zeilen, die diese Partei weggibt.
type Zusagevertrag []Handelsposition

// GrenzeFuer zählt, wie viel der Vertrag von der Variante dieses Exemplars
// nimmt — null, wenn keine Zeile passt.
func (v Zusagevertrag) GrenzeFuer(exemplar Exemplar) int {
	grenze := 0

	for _, zeile := range v {
		if zeile.alsExemplar().GleicheVariante(exemplar) {
			grenze += zeile.Menge
		}
	}

	return grenze
}

// GleicheVariante sagt, ob zwei Exemplare dieselbe Vereinbarung erfüllen.
func (e Exemplar) GleicheVariante(anderes Exemplar) bool {
	return e.Druckschluessel() == anderes.Druckschluessel() &&
		e.Zustand == anderes.Zustand && e.Sprache == anderes.Sprache &&
		e.Ausfuehrung() == anderes.Ausfuehrung()
}

// alsExemplar ist die Variante, die eine Zeile verlangt.
func (p Handelsposition) alsExemplar() Exemplar {
	return Exemplar{
		Printing: p.Printing, DruckID: p.DruckID,
		Zustand: Zustand(p.Guete), Sprache: p.Sprache, Finish: Finish(p.Finish),
	}
}
