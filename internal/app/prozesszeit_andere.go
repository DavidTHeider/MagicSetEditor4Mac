//go:build !windows

package app

import "syscall"

// prozesszeitMs ist die Rechenzeit dieses Prozesses, Nutzer und System
// zusammen — aus `getrusage`, das macOS und Linux gleich kennen.
//
// **Ohne Fehlerzweig**: `getrusage` scheitert nur an einem ungültigen Zeiger
// oder einem unbekannten `who` (EFAULT, EINVAL — `man 2 getrusage`), und
// beides steht hier fest. Scheiterte es doch, bliebe die Zahl null.
func prozesszeitMs() int64 {
	var verbrauch syscall.Rusage
	_ = syscall.Getrusage(syscall.RUSAGE_SELF, &verbrauch) //nolint:errcheck // siehe oben

	return (verbrauch.Utime.Nano() + verbrauch.Stime.Nano()) / 1e6
}
