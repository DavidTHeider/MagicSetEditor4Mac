package msepaket

import "syscall"

// stempelVon liest den Stempel eines Ortes — dieselbe Rechnung wie auf
// darwin, nur heißen die Felder dort `Mtimespec` und `Ctimespec`.
func stempelVon(pfad string) ortsstempel {
	var stand syscall.Stat_t
	if syscall.Lstat(pfad, &stand) != nil {
		return ortsstempel{}
	}

	return ortsstempel{
		aenderung: stand.Mtim.Nano(),
		status:    stand.Ctim.Nano(),
		modus:     stand.Mode,
		knoten:    stand.Ino,
		da:        true,
	}
}
