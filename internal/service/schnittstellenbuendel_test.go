package service_test

// **N3 des Nachlaufs über Phase 6.12, zu Ende gedacht: ein Bündel ist noch
// nicht das eigene** (#1905).
//
// Ein `mse-mcp` in einem Bündel, das es noch gibt, gehört dem, dessen
// Kennung in seinem `Info.plist` steht. Nur ein Bündel, das es nicht mehr
// gibt — die Anwendung ist umgezogen —, gilt nach seiner Gestalt.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinVorhandenesBuendelZeigtSeineKennung(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name, kennung string
		stand         service.Schnittstellenstand
	}{
		{"ein fremdes Bündel", "org.example.fremd", service.SchnittstelleFremd},
		{"das eigene, umbenannt", "com.wails.MagicSetEditor4Mac", service.SchnittstelleVeraltet},
		{"der Abnahmebau", "com.wails.MagicSetEditor4Mac.abnahme", service.SchnittstelleVeraltet},
		{"eine Kennung, die nur so anfängt", "com.wails.MagicSetEditor4Mac.fremd", service.SchnittstelleFremd},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			programm := einBuendel(t, "Irgendwas.app", `<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
<dict>
	<key>CFBundleName</key>
	<string>com.wails.MagicSetEditor4Mac</string>
	<key>CFBundleIdentifier</key>
	<string>`+fall.kennung+`</string>
</dict>
</plist>
`)

			if stand := standMit(t, programm); stand != fall.stand {
				t.Errorf("Stand %q, erwartet %q", stand, fall.stand)
			}
		})
	}
}

// **Was sich nicht lesen lässt, belegt nichts** — ein `Info.plist`, das ein
// Ordner ist, steht für eine Datei, die da ist und nichts sagt.
func TestEinUnlesbaresInfoPlistIstFremd(t *testing.T) {
	t.Parallel()

	programm := einBuendel(t, "Kaputt.app", "")
	plist := filepath.Join(filepath.Dir(filepath.Dir(programm)), "Info.plist")

	if fehler := os.Remove(plist); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.Mkdir(plist, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	if stand := standMit(t, programm); stand != service.SchnittstelleFremd {
		t.Errorf("Stand %q, erwartet fremd", stand)
	}
}

// einBuendel legt ein Bündel mit `mse-mcp` und diesem `Info.plist` an.
func einBuendel(t *testing.T, name, plist string) string {
	t.Helper()

	inhalt := filepath.Join(t.TempDir(), name, "Contents")
	if fehler := os.MkdirAll(filepath.Join(inhalt, "MacOS"), 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	schreibeDatei(t, filepath.Join(inhalt, "Info.plist"), plist)

	return filepath.Join(inhalt, "MacOS", "mse-mcp")
}

// standMit liest den Stand eines Eintrags, der auf dieses Programm zeigt.
func standMit(t *testing.T, programm string) service.Schnittstellenstand {
	t.Helper()

	ziel, datei := einDesktopziel(t)
	if fehler := os.MkdirAll(filepath.Dir(datei), 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	schreibeDatei(t, datei, `{"mcpServers": {"magicseteditor4mac": {"command": "`+programm+`", "args": []}}}`)

	return service.SchnittstelleLesen(ziel)
}
