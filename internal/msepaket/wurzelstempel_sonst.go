//go:build !darwin && !linux

package msepaket

import "os"

// stempelVon liest Änderungszeit und Rechte — eine Statuszeit gibt es hier
// nicht (siehe [ortsstempel]).
func stempelVon(pfad string) ortsstempel {
	info, fehler := os.Lstat(pfad)
	if fehler != nil {
		return ortsstempel{}
	}

	// **Statuszeit und Knoten ausdrücklich null** — es gibt sie hier nicht,
	// und ein Stempel ohne sie vergleicht nur Zeit und Rechte.
	return ortsstempel{aenderung: info.ModTime().UnixNano(), status: 0, modus: uint32(info.Mode()), knoten: 0, da: true}
}
