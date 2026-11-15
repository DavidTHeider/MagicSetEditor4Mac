package app

import (
	"os"
	"runtime/debug"
)

// setzeAbsturzausgabe ist die echte Spur: Go schreibt eine unbehandelte
// Panik oder einen `fatal error` zusätzlich in diese Datei (F04, #2343).
func setzeAbsturzausgabe(datei *os.File) error {
	return debug.SetCrashOutput(datei, debug.CrashOptions{})
}
