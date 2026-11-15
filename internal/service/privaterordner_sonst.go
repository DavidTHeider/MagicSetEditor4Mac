//go:build !unix

package service

import (
	"io/fs"
	"os"
)

// **Auf den beschriebenen Plattformen ohne `unix`** (`docs/PORTIERUNG.md`)
// gibt es weder `Stat_t` noch `O_NOFOLLOW`; geprüft wird dort nur, dass es den
// Ordner gibt. Beschrieben, nicht gebaut (§12).

var ich uint32

func ortVon(pfad string) (ort, error) {
	_, fehler := os.Stat(pfad)

	return ort{}, fehler //nolint:wrapcheck // der Aufrufer nennt den Weg
}

func ordnerOeffnen(pfad string) (*os.File, error) {
	return os.Open(pfad) //nolint:gosec,wrapcheck // der geprüfte Ordner
}

func eigenesVon(datei *os.File) (uint32, fs.FileMode, error) {
	angaben, fehler := datei.Stat()
	if fehler != nil {
		return ich, 0, fehler //nolint:wrapcheck // der Aufrufer nennt den Ordner
	}

	return ich, angaben.Mode().Perm(), nil
}
