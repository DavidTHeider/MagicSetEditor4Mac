package service

import (
	"os"
	"path/filepath"
)

// wegwerfpaket legt einen **eigenen** Wegwerfort an und darin das
// Paketverzeichnis unter dem Namen, den es tragen soll — `Installieren` liest
// ihn aus dem letzten Pfadglied.
//
// **Eindeutig und nicht fest** (FMT-02, #1691): Kopie und Neuanlage bauten
// unter `mse4-kopie-<ziel>` bzw. `mse4-anlage-<name>` im gemeinsamen
// Temp-Verzeichnis. Zwei Prozesse mit demselben Ziel — Fenster und
// Kommandozeile, oder drei Flächen in einem Testlauf — räumten einander den
// Ordner weg, und das aufgeschobene Aufräumen löschte, was der andere gerade
// baute. `os.MkdirTemp` gibt jedem Bau seinen eigenen Ort.
func wegwerfpaket(bauort, praefix, paketname string) (bau, ort string, fehler error) {
	bau, fehler = os.MkdirTemp(bauort, praefix)
	if fehler != nil {
		return "", "", fehler
	}

	ort = filepath.Join(bau, paketname)

	return bau, ort, os.Mkdir(ort, 0o750)
}
