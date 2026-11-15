package msepaket

import (
	"fmt"
	"path/filepath"
)

// ModuleEinbinden löst die `include file:`-Verweise dieses Pakets auf.
//
// # Warum das ein eigener Schritt ist und nicht Teil von [Oeffne]
//
// **Weil `msepaket` nicht weiß, wo Pakete liegen.** Das entscheidet der
// Paketverwalter, und `Oeffne` bekommt nur einen Pfad. Wer die Auflösung ins
// Öffnen zöge, machte aus jedem Paketleser einen, der eine Wurzel braucht — auch
// dort, wo niemand ein Modul liest.
//
// # Was es verändert
//
// **Den Kopf**, und das ist der Punkt: [Paket.Stilmodell] und
// [Paket.Spielmodell] lesen danach einen Baum, in dem die Module stehen. Ohne
// diesen Schritt trägt `Stil.Initskript` nur die `include file:`-Zeilen, und
// jedes Feldskript, das eine Modulfunktion ruft, ist unauswertbar
// (#186).
//
// **Die Rohbytes bleiben unberührt.** `KopfRoh` trägt die Datei, wie sie
// dastand ([0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)),
// und die Byte-Zusage aus #48 hängt daran. Eingebunden wird der **Baum**.
//
// Ein Modul, das sich nicht auflösen lässt, ist ein Befund und kein Abbruch
// (AGENTS.md §8): ein Stil, dem ein Modul fehlt, ist mit Vorbehalt lesbar.
func (p *Paket) ModuleEinbinden(quelle Paketquelle) []Befund {
	einbinder := NeuerEinbinder(quelle).ImPaket(paketname(p.Pfad))
	p.Kopf = einbinder.Einbinden(p.Kopf)

	befunde := einbinder.Befunde()
	p.Befunde = append(p.Befunde, befunde...)

	return befunde
}

// Paketleser liest Dateien aus Paketen, die unter einer Wurzel liegen.
//
// **Die Schwester von [Bildquelle]**, und aus demselben Grund hier: `msepaket`
// weiß, wie ein Paket auf der Platte aussieht — ein Verzeichnis oder ein
// Archiv. Wo die Wurzeln sind, sagt der Composition Root.
type Paketleser struct {
	wurzeln []string
}

// NeuerPaketleser baut den Leser über einer Reihe von Paketwurzeln.
//
// **Eine leere Wurzel fällt weg** — aus demselben Grund wie bei
// [NeueBildquelle]: `filepath.Join("", …)` machte daraus einen relativen Pfad.
func NeuerPaketleser(wurzeln ...string) *Paketleser {
	return &Paketleser{wurzeln: echteWurzeln(wurzeln)}
}

// Lies gibt den Inhalt einer Datei aus einem Paket.
func (l *Paketleser) Lies(pfad Paketpfad) ([]byte, error) {
	versuche := make([]error, 0, len(l.wurzeln)+1)
	versuche = append(versuche, ErrKeinPaket)

	for _, wurzel := range l.wurzeln {
		inhalt, err := paketdatei(filepath.Join(wurzel, pfad.Paket), pfad.Datei)
		if err != nil {
			versuche = append(versuche, err)

			continue
		}

		return inhalt, nil
	}

	return nil, fmt.Errorf("/%s/%s: %w", pfad.Paket, pfad.Datei, verbinde(versuche))
}
