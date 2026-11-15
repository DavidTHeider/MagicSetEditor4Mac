package service

// **Ein Zwischenspeicher auf der Platte** — OPS-08.AK-03
// (#1725).
//
// # Nur, was das Muster als Eintrag erkennt
//
// **Ein Ordner kann mehr enthalten als den Speicher** — eine Sperrdatei, eine
// halb geschriebene Nachbardatei, eine Datei, die jemand hineingelegt hat.
// Gezählt, beschnitten und geleert wird nur, was `Eintrag` erkennt; alles
// andere bleibt, wie es ist.
//
// # Beschnitten wird nach der Uhr der Platte
//
// **Die ältesten zuerst**, gemessen an der letzten Änderung. Ein Eintrag,
// der gerade neu geholt wurde, ist jung; einer, den seit Monaten niemand
// brauchte, alt. Das ist nicht „zuletzt gelesen" — ein Lesen ändert die Uhr
// nicht —, aber es kostet keinen Schreibvorgang je Treffer.

import (
	"os"
	"path/filepath"
	"sort"
)

// Plattenspeicher ist ein Ordner, dessen Einträge ein Muster erkennt.
type Plattenspeicher struct {
	Ordner  string
	Eintrag func(name string) bool
}

// Plattenbelegung sagt, wie viele Einträge wie viel Platz belegen.
type Plattenbelegung struct {
	Eintraege int
	Bytes     int64
}

type platteneintrag struct {
	pfad  string
	bytes int64
	zeit  int64
}

// eintraege liest die Einträge — ein Ordner, den es nicht gibt, hat keine.
func (p Plattenspeicher) eintraege() []platteneintrag {
	gelesen, fehler := os.ReadDir(p.Ordner)
	if fehler != nil {
		return nil
	}

	eintraege := make([]platteneintrag, 0, len(gelesen))

	for _, eintrag := range gelesen {
		if eintrag.IsDir() || !p.Eintrag(eintrag.Name()) {
			continue
		}

		// **Eine Datei, die zwischen Lesen und Fragen verschwand, zählt nicht**
		// — sie ist schon fort, und das ist, was ein Leeren wollte.
		if info, infofehler := eintrag.Info(); infofehler == nil {
			eintraege = append(eintraege, platteneintrag{
				pfad: filepath.Join(p.Ordner, eintrag.Name()), bytes: info.Size(), zeit: info.ModTime().UnixNano(),
			})
		}
	}

	return eintraege
}

// Belegung zählt die Einträge und ihre Bytes.
func (p Plattenspeicher) Belegung() Plattenbelegung {
	var belegung Plattenbelegung

	for _, eintrag := range p.eintraege() {
		belegung.Eintraege++
		belegung.Bytes += eintrag.bytes
	}

	return belegung
}

// Beschneiden nimmt die ältesten Einträge weg, bis die Grenze in Bytes
// eingehalten ist, und gibt ihre Zahl.
func (p Plattenspeicher) Beschneiden(grenze int64) int {
	eintraege := p.eintraege()
	sort.Slice(eintraege, func(links, rechts int) bool { return eintraege[links].zeit < eintraege[rechts].zeit })

	var summe int64
	for _, eintrag := range eintraege {
		summe += eintrag.bytes
	}

	weg := 0

	for _, eintrag := range eintraege {
		if summe <= grenze {
			break
		}

		if os.Remove(eintrag.pfad) == nil {
			summe -= eintrag.bytes
			weg++
		}
	}

	return weg
}

// Leeren nimmt alle Einträge weg und gibt ihre Zahl.
func (p Plattenspeicher) Leeren() int {
	weg := 0

	for _, eintrag := range p.eintraege() {
		if os.Remove(eintrag.pfad) == nil {
			weg++
		}
	}

	return weg
}
