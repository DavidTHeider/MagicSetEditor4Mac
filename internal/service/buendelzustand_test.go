package service_test

// **Ein Bündel ist eigen, wenn es das sagt — oder fort ist** — Nachlauf über
// Phase 4, Befund 9: ein fehlendes `Info.plist` galt wie ein fehlendes
// Bündel, und ein vorhandenes fremdes `MagicSetEditor4Mac Fremd.app` ohne
// diese Datei ging als eigenes durch.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinBuendelIstEigenWennEsDasSagtOderFortIst(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	buendel := func(name, kennung string) string {
		programm := filepath.Join(ort, name, "Contents", "MacOS", "mse-mcp")
		if err := os.MkdirAll(filepath.Dir(programm), 0o700); err != nil {
			t.Fatal(err)
		}

		if kennung != "" {
			schreibeDatei(t, filepath.Join(ort, name, "Contents", "Info.plist"),
				"<plist><dict><key>CFBundleIdentifier</key>\n\t<string>"+kennung+"</string></dict></plist>")
		}

		return programm
	}

	for name, fall := range map[string]struct {
		programm string
		eigen    bool
	}{
		"eigenes":             {buendel("MagicSetEditor4Mac.app", "io.github.davidtheider.magicseteditor4mac"), true},
		"eigenes, umbenannt":  {buendel("Umbenannt.app", "io.github.davidtheider.magicseteditor4mac"), true},
		"fremdes":             {buendel("MagicSetEditor4Mac Fremd.app", "com.fremd.x"), false},
		"da, ohne Info.plist": {buendel("MagicSetEditor4Mac Leer.app", ""), false},
		"fort":                {filepath.Join(ort, "Fort.app", "Contents", "MacOS", "mse-mcp"), true},
		"in keinem Bündel":    {filepath.Join(ort, "bin", "mse-mcp"), false},
	} {
		if eigen := service.EigenesBuendel(fall.programm); eigen != fall.eigen {
			t.Errorf("%s: eigen %t, erwartet %t", name, eigen, fall.eigen)
		}
	}
}
