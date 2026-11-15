//go:build darwin

package schriften

import (
	"os"
	"path/filepath"
)

// Orte sind die Verzeichnisse, in denen macOS Schriften führt.
//
// Die Reihenfolge ist die der Vorrangigkeit: was jemand selbst installiert hat,
// gilt vor dem, was mitgeliefert wird. `Supplemental` steht eigens da — dort
// liegen seit macOS 10.15 die Schriften, die nicht zum Grundbestand gehören,
// und genau dort liegen die meisten, die eine Kartenvorlage nennt.
//
// **Ohne `.ttc` und `.dfont`** (siehe `Endungen`): ein Teil des Grundbestands
// ist so abgelegt und damit hier unerreichbar. Das ist eine benannte Lücke, kein
// Versehen — `docs/PORTIERUNG.md` führt sie.
func Orte() []string {
	orte := []string{
		"/System/Library/Fonts",
		"/System/Library/Fonts/Supplemental",
		"/Library/Fonts",
	}

	if zuhause, err := os.UserHomeDir(); err == nil {
		// Das eigene Verzeichnis zuerst: eine selbst installierte Schrift ist
		// die, die jemand meint.
		orte = append([]string{filepath.Join(zuhause, "Library", "Fonts")}, orte...)
	}

	return orte
}
