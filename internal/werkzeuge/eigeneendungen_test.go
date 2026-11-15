package werkzeuge

// **Eine eigene Endung ist im Bündel erklärt** — FMT-04.AK-02
// (#1693).
//
// # Was er hält
//
// Jede Endung, die dieses Programm **selbst** erfindet — heute der Umschlag
// (`austausch.Endung`) —, steht in `build/darwin/Info.plist` als
// ausgeführter Typ mit eigener Beschreibung. Sonst zeigt der Finder eine
// namenlose Datei, und ihr Name allein (`.msepackage`) sähe aus wie ein
// MSE-Paket, das MSE2 lesen könnte (Vermerk 0144).
//
// # Was er nicht hält
//
// Ob der Typ richtig beschrieben ist, und ob eine neue Endung hier überhaupt
// eingetragen wird — die Liste darunter ist eine Aufzählung, keine Suche.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
)

func TestJedeEigeneEndungIstImBuendelErklaert(t *testing.T) {
	roh, err := os.ReadFile(filepath.Join(wurzel, "build", "darwin", "Info.plist")) //nolint:gosec // eine Datei im Repositorium
	if err != nil {
		t.Fatal(err)
	}

	liste := string(roh)
	anfang := strings.Index(liste, "<key>UTExportedTypeDeclarations</key>")

	if anfang < 0 {
		t.Fatal("Info.plist erklärt keinen eigenen Typ (UTExportedTypeDeclarations fehlt)")
	}

	for _, endung := range []string{austausch.Endung} {
		if !strings.Contains(liste[anfang:], "<string>"+strings.TrimPrefix(endung, ".")+"</string>") {
			t.Errorf("die Endung %s ist im Bündel nicht erklärt", endung)
		}
	}
}
