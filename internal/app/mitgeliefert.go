package app

import (
	"os"
	"path/filepath"
)

// **Was das Bündel mitbringt** —
// #702.
//
// # Warum es das gibt
//
// Bis hierher startete eine frisch geladene Anwendung ins Leere: ohne bezogene
// Pakete zeichnet keine Karte, und ohne Set gibt es nichts zu öffnen. Der erste
// Eindruck war eine Einstiegsseite mit drei Knöpfen und einem Download.
//
// Seit die Vorlagen mitgeliefert werden dürfen
// ([0025](../../docs/entscheidungen/0025-die-vorlagen-duerfen-mitgeliefert-werden.md))
// legt der Freigabeweg zwei weitere Verzeichnisse neben die Saat:
//
//   - `Resources/vorlagen` — die Pakete, die eine M15-Karte braucht.
//   - `Resources/beispielsets` — ein echtes Set zum Aufschlagen.
//
// # Warum neben dem Programm und nicht neben der Datenbank
//
// Dieselbe Begründung wie bei [Saatverzeichnis]: sie sind **Teil der
// Anwendung**, kein Bestand der Nutzenden. Sie kommen mit jeder Fassung neu,
// und wer sie ändert, ändert die Anwendung. Bezogene Pakete liegen unter
// `~/Library/Application Support/` und überleben eine Neuinstallation.
//
// # Was das nicht leistet
//
// **Ein Bau ohne Freigabeweg findet nichts.** `scripts/mitliefern.sh` läuft
// **nicht** in `scripts/bauen.sh` — 167 MB bei jedem `make dev` wären
// unbenutzbar. Wer die Anwendung entwickelt, hat die Saat und seine bezogenen
// Pakete; wer sie lädt, hat beides mitgeliefert.

// Vorlagenverzeichnis gibt den Ort der mitgelieferten Pakete — oder nichts.
func Vorlagenverzeichnis() string { return vorlagenNeben(os.Executable) }

// Beispielsetverzeichnis gibt den Ort der mitgelieferten Sets — oder nichts.
func Beispielsetverzeichnis() string { return beispielsetsNeben(os.Executable) }

func vorlagenNeben(programm func() (string, error)) string {
	return nebenDemProgramm(programm, "vorlagen")
}

func beispielsetsNeben(programm func() (string, error)) string {
	return nebenDemProgramm(programm, "beispielsets")
}

// nebenDemProgramm rechnet vom Programmpfad auf ein Verzeichnis im Bündel.
//
// Die Naht ist das Argument — aus demselben Grund wie bei [saatNeben]: ein Test
// hat kein `.app`, und ein Lauf gegen das echte Bündel dieses Rechners wäre auf
// einer zweiten Maschine grün oder rot, je nachdem, wer dort gebaut hat
// ([0007](../../docs/entscheidungen/0007-kein-entfernter-lauf.md)).
func nebenDemProgramm(programm func() (string, error), name string) string {
	pfad, err := programm()
	if err != nil {
		return ""
	}

	// …/Contents/MacOS/Programm → …/Contents/Resources/<name>
	ort := filepath.Join(filepath.Dir(filepath.Dir(pfad)), "Resources", name)

	angaben, fehler := os.Stat(ort)
	if fehler != nil || !angaben.IsDir() {
		return ""
	}

	return ort
}
