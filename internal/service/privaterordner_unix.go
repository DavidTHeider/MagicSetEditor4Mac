//go:build unix

package service

import (
	"io/fs"
	"os"
	"syscall"
)

// ich ist der Nutzer dieses Prozesses.
var ich = uint32(os.Getuid()) //nolint:gosec // G115: eine Nutzerkennung ist nie negativ

// ortVon liest Besitzer und Rechte eines Ordners — **über `syscall.Stat`**, so
// kommt `Stat_t` ohne Typzusicherung.
//
// **Ein Fehler gibt den Nullwert und den Fehler** — der Aufrufer weist beides
// in **einem** Zweig ab ([wegPruefen]); scheitern kann `Stat` hier nur, wenn
// ein Vorfahr zwischen dem Auflösen und diesem Ruf verschwindet.
func ortVon(pfad string) (ort, error) {
	var stand syscall.Stat_t
	fehler := syscall.Stat(pfad, &stand)

	return ort{
		besitzer: stand.Uid,
		offen:    stand.Mode&0o022 != 0,
		klebend:  stand.Mode&syscall.S_ISVTX != 0,
	}, fehler //nolint:wrapcheck // der Aufrufer nennt den Weg
}

// ordnerOeffnen öffnet den Ordner selbst — **ein Verweis an seiner Stelle
// scheitert** (`O_NOFOLLOW`), eine Datei auch (`O_DIRECTORY`).
func ordnerOeffnen(pfad string) (*os.File, error) {
	// #nosec G304 -- der Ordner, dessen Weg [wegPruefen] eben geprüft hat
	return os.OpenFile(pfad, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_DIRECTORY, 0) //nolint:gosec,wrapcheck // dito
}

// eigenesVon liest Besitzer und Rechte **am geöffneten Ordner**.
func eigenesVon(datei *os.File) (besitzer uint32, rechte fs.FileMode, fehler error) {
	var stand syscall.Stat_t
	fehler = syscall.Fstat(int(datei.Fd()), &stand) //nolint:gosec // G115: ein Dateideskriptor passt in int

	return stand.Uid, fs.FileMode(stand.Mode) & fs.ModePerm, fehler //nolint:wrapcheck // der Aufrufer nennt den Ordner
}
