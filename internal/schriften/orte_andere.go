//go:build !darwin

package schriften

// Orte gibt für andere Plattformen die dort üblichen Verzeichnisse.
//
// **Beschrieben, nicht erprobt** (AGENTS.md §12): gebaut wird für macOS, und
// diese Liste steht da, damit der Bau für Windows und Linux nicht bricht — nicht
// als Zusage, dass die Suche dort funktioniert. Was fehlt und was es kostet,
// steht in `docs/PORTIERUNG.md`.
func Orte() []string {
	return []string{
		// Windows
		`C:\Windows\Fonts`,

		// Linux und BSD
		"/usr/share/fonts",
		"/usr/local/share/fonts",
	}
}
