package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// **Die Saat liegt im Bündel neben dem Programm.**
//
// `scripts/saat.sh` legt sie beim Bauen unter `Contents/Resources/saat` ab
// (#58); hier
// steht der Weg, auf dem die laufende Anwendung sie wiederfindet.
func TestDieSaatLiegtNebenDemProgramm(t *testing.T) {
	buendel := t.TempDir()
	programm := filepath.Join(buendel, "Probe.app", "Contents", "MacOS", "Probe")
	saat := filepath.Join(buendel, "Probe.app", "Contents", "Resources", "saat")

	if err := os.MkdirAll(saat, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if gefunden := saatNeben(func() (string, error) { return programm, nil }); gefunden != saat {
		t.Errorf("gefunden %q, erwartet %q", gefunden, saat)
	}
}

// **Was es nicht gibt, wird nicht behauptet.**
//
// Eine Wurzel auf ein Verzeichnis, das im Bündel fehlt, wäre eine Zusage ohne
// Deckung: `Paketwurzeln` liefe darüber ins Leere, und die Vorschau meldete ein
// fehlendes Stilpaket statt „keine Saat im Bündel".
func TestOhneSaatImBuendelKeineWurzel(t *testing.T) {
	buendel := t.TempDir()
	inhalt := filepath.Join(buendel, "Probe.app", "Contents")

	if err := os.MkdirAll(filepath.Join(inhalt, "MacOS"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// **Eine Datei ist kein Verzeichnis**, und der Unterschied zählt: `Oeffne`
	// liest ein `.mse-set` auch als Archiv, aber eine *Wurzel* muss ein
	// Verzeichnis sein — der Grund steht bei [msepaket.Paketwurzeln].
	datei := filepath.Join(inhalt, "Resources")
	if err := os.MkdirAll(datei, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(datei, "saat"), []byte("kein Verzeichnis"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	faelle := map[string]struct {
		pfad   string
		fehler error
	}{
		"ohne Resources/saat":  {pfad: filepath.Join(buendel, "Anders.app", "Contents", "MacOS", "X")},
		"saat ist eine Datei":  {pfad: filepath.Join(inhalt, "MacOS", "Probe")},
		"kein Pfad zum Selbst": {fehler: errors.New("kein Programmpfad")},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			gefunden := saatNeben(func() (string, error) { return fall.pfad, fall.fehler })
			if gefunden != "" {
				t.Errorf("gefunden %q, erwartet war nichts", gefunden)
			}
		})
	}
}
