//go:build !windows

package app

import (
	"errors"
	"syscall"
)

// prozessLebt sagt, ob ein Prozess mit dieser Kennung läuft — Signal 0
// prüft, ohne zu senden. **`EPERM` heißt: er läuft**, nur gehört er jemand
// anderem; dann ist seine Marke nicht unsere Sache.
func prozessLebt(pid int) bool {
	err := syscall.Kill(pid, 0)

	return err == nil || errors.Is(err, syscall.EPERM)
}
