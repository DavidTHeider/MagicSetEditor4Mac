//go:build !windows

package applefm

import (
	"errors"
	"syscall"
)

// prozessLebt fragt das Betriebssystem, ob es einen Prozess gibt — ein
// Signal 0 prüft nur, und „keine Erlaubnis" heißt: es gibt ihn.
func prozessLebt(kennung int) bool {
	fehler := syscall.Kill(kennung, 0)

	return fehler == nil || errors.Is(fehler, syscall.EPERM)
}
