//go:build unix

// Die Sperre über `flock`.
//
// **`flock` und nicht `fcntl`**: es gilt für die **offene Datei** und nicht
// für den Prozess, und es fällt mit dem Griff. `fcntl`-Sperren fallen bereits,
// wenn irgendein Griff auf dieselbe Datei geschlossen wird — in einem
// Programm mit mehreren Wegen zur selben Datei ist das eine Sperre, die
// überraschend verschwindet.

package dateisperre

import (
	"os"
	"syscall"
)

// teilen nimmt die geteilte Sperre, ohne zu warten.
func teilen(datei *os.File) error {
	return syscall.Flock(int(datei.Fd()), syscall.LOCK_SH|syscall.LOCK_NB)
}

// allein nimmt die ausschließliche Sperre, ohne zu warten.
func allein(datei *os.File) error {
	return syscall.Flock(int(datei.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}
