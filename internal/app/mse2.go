package app

import (
	"os"
	"path/filepath"
)

// MSE2Orte nennt die üblichen Orte einer MSE2-Installation.
//
// # Warum eine Liste und keine Suche
//
// **Vorgeschlagen, nicht ungefragt gelesen** — die Zusage aus
// #55. Ein
// Programm, das die Platte nach Vorlagen durchsucht, liest Verzeichnisse, die
// niemand ihm genannt hat. Diese Funktion **stat**et die genannten Orte und
// liest keinen einzigen; was darin liegt, erfährt die Anwendung erst, wenn
// jemand einen davon wählt.
//
// # Woher die Orte kommen
//
// MSE2 ist ein Windows-Programm; unter macOS läuft es über Wine oder
// CrossOver, und beide legen ein Windows-Dateisystem unter dem
// Heimatverzeichnis ab. Die Pfade sind **Vermutungen** über verbreitete
// Installationen — belegt ist keiner, und deshalb steht daneben immer der Weg
// „Verzeichnis wählen …".
func MSE2Orte() []string {
	// **Ohne Heimatverzeichnis keine Vermutung.** `os.UserHomeDir` scheitert,
	// wenn `HOME` fehlt — im Betrieb selten, in einem Dienst ohne Umgebung
	// nicht. Dann bleibt der Weg „Verzeichnis wählen …", und das ist die
	// ehrlichere Antwort als eine Liste relativer Pfade.
	heim, err := os.UserHomeDir()
	if err != nil {
		return []string{}
	}

	moegliche := []string{
		filepath.Join(heim, ".wine", "drive_c", "Program Files (x86)",
			"Magic Set Editor 2", "data"),
		filepath.Join(heim, ".wine", "drive_c", "Program Files",
			"Magic Set Editor 2", "data"),
		filepath.Join(heim, "Library", "Application Support", "CrossOver", "Bottles",
			"Magic Set Editor", "drive_c", "Program Files (x86)", "Magic Set Editor 2", "data"),
		filepath.Join(heim, "Applications", "Magic Set Editor 2", "data"),
		"/Applications/Magic Set Editor 2/data",
	}

	vorhandene := make([]string, 0, len(moegliche))

	for _, ort := range moegliche {
		angaben, fehler := os.Stat(ort)
		if fehler == nil && angaben.IsDir() {
			vorhandene = append(vorhandene, ort)
		}
	}

	return vorhandene
}
