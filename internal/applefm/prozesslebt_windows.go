//go:build windows

package applefm

// prozessLebt sagt auf Windows immer ja — `fm` gibt es dort nicht, und ein
// Socket, den nichts angelegt hat, ist nichts zum Aufräumen
// (`docs/PORTIERUNG.md`).
func prozessLebt(int) bool { return true }
