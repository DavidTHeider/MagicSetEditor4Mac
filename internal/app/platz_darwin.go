//go:build darwin

package app

// **Wie viel neben dem Bestand frei ist** — Werkstück O6 von Phase 5.2
// (#1430).
//
// **`Statfs` und keine Schätzung.** Das Volumen des Bestands ist nicht
// unbedingt das Systemvolumen; wer den Platz auf `/` zählte, gäbe für einen
// Bestand auf einer externen Platte die falsche Zahl.

import (
	"math"
	"syscall"
)

// freierPlatz gibt die freien Bytes des Volumens, auf dem `ort` liegt.
func freierPlatz(ort string) (int64, error) {
	var angaben syscall.Statfs_t

	if fehler := syscall.Statfs(ort, &angaben); fehler != nil {
		return 0, fehler //nolint:wrapcheck // der Aufrufer sagt, worum es ging
	}

	// **`Bavail` und nicht `Bfree`**: der Teil, den ein gewöhnlicher Prozess
	// wirklich benutzen darf. Der Unterschied ist die Reserve, die macOS für
	// sich behält — sie mitzuzählen hieße, den Tausch mit Speicher zu
	// planen, den es nicht gibt.
	return freieBytes(angaben.Bavail, uint64(angaben.Bsize)), nil
}

// platzAuf gibt die freien und die gesamten Bytes des Volumens, auf dem
// `ort` liegt — für das Betriebsband (OPS02, #2348), das auch den Anteil
// braucht.
func platzAuf(ort string) (frei, gesamt uint64, fehler error) {
	var angaben syscall.Statfs_t

	if fehler := syscall.Statfs(ort, &angaben); fehler != nil {
		return 0, 0, fehler //nolint:wrapcheck // der Aufrufer sagt, worum es ging
	}

	groesse := uint64(angaben.Bsize)

	// #nosec G115 -- freieBytes deckelt bei MaxInt64 und ist nie negativ
	return uint64(freieBytes(angaben.Bavail, groesse)), uint64(freieBytes(angaben.Blocks, groesse)), nil
}

// freieBytes rechnet Blöcke mal Blockgröße — und **deckelt, statt zu kippen**.
//
// # Warum das eine eigene Funktion ist
//
// Hier stand `int64(Bavail) * int64(Bsize)` mit einem `//nolint:gosec` daneben
// und der Begründung „Bavail und Bsize sind nie negativ". Das ist wahr und
// beantwortet die Frage nicht: gerügt war das **Produkt**, und ein Produkt
// zweier nichtnegativer Zahlen kann sehr wohl über `math.MaxInt64` gehen —
// dann ist die Zahl negativ, und die Vorprüfung meldete auf einem sehr großen
// Volumen „zu wenig Platz". Eine Ausnahme, deren Grund am gerügten Fall
// vorbeigeht, ist keine Begründung (§10).
//
// **Und die Schranke ist prüfbar, weil sie hier steht**: `Statfs` lässt sich
// in einem Lauf nicht dazu bringen, eine Zahl dieser Größe zu melden.
func freieBytes(bloecke, groesse uint64) int64 {
	if groesse != 0 && bloecke > uint64(math.MaxInt64)/groesse {
		return math.MaxInt64
	}

	// **Nach der Schranke passt das Produkt.** `groesse == 0` kommt hier an
	// und ergibt 0 — ein Volumen ohne Blockgröße hat keinen zählbaren Platz.
	return int64(bloecke * groesse) // #nosec G115 -- die Schranke darüber hält das Produkt unter MaxInt64
}
