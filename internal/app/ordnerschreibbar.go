//go:build !windows

package app

import "syscall"

// schreibbar fragt `access(2)` — ohne eine Datei anzulegen.
func schreibbar(ordner string) bool {
	return syscall.Access(ordner, 0o2) == nil
}
