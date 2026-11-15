package msepaket

import "syscall"

// stempelVon liest den Stempel eines Ortes — eines Verweises selbst, nicht
// seines Ziels —, und den leeren, wenn es ihn nicht gibt.
//
// **Über `syscall.Lstat` und nicht `os.Lstat`**: so kommt `Stat_t` ohne
// Typzusicherung, und die Statuszeit steht darin.
func stempelVon(pfad string) ortsstempel {
	var stand syscall.Stat_t
	if syscall.Lstat(pfad, &stand) != nil {
		return ortsstempel{}
	}

	return ortsstempel{
		aenderung: stand.Mtimespec.Nano(),
		status:    stand.Ctimespec.Nano(),
		modus:     uint32(stand.Mode),
		knoten:    stand.Ino,
		da:        true,
	}
}
