package dokumentsymbol_test

// **Der Wächter über das Dokumentsymbol** —
// #891.
//
// # Was er hält
//
// Die eingecheckte `build/dokument.png` ist das, was [dokumentsymbol.Zeichnen]
// zeichnet. Ein Bild, das jemand von Hand ändert, läuft mit dem Code
// auseinander — und niemand sieht es, weil ein PNG im Diff nichts sagt.
//
// # Was er **nicht** hält
//
// Ob das Symbol gut aussieht. Das entscheidet ein Blick in den Finder (§3), und
// er steht in `docs/RELEASE.md` bei der Handarbeit.
//
// # Wie das Bild neu entsteht
//
//	MSE_SYMBOL_SCHREIBEN=1 go test ./internal/werkzeuge/dokumentsymbol/
//
// **Ein Schalter und kein eigenes Programm.** Ein `cmd/`-Verzeichnis für ein
// Bild, das sich alle paar Jahre ändert, wäre eine Fläche mehr; und `internal`
// darf ein Programm außerhalb des Moduls ohnehin nicht rufen.

import (
	"bytes"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/werkzeuge/dokumentsymbol"
)

// ort ist der Pfad des eingecheckten Bildes, vom Paket aus gesehen.
var ort = filepath.Join("..", "..", "..", "build", "dokument.png")

func TestDasDokumentsymbolIstDasGezeichnete(t *testing.T) {
	var gezeichnet bytes.Buffer

	if err := png.Encode(&gezeichnet, dokumentsymbol.Zeichnen()); err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if os.Getenv("MSE_SYMBOL_SCHREIBEN") != "" {
		if err := os.WriteFile(ort, gezeichnet.Bytes(), 0o600); err != nil {
			t.Fatalf("schreiben: %v", err)
		}

		t.Log("build/dokument.png neu geschrieben")

		return
	}

	eingecheckt, err := os.ReadFile(ort) // #nosec G304 -- fester Pfad im Baum
	if err != nil {
		t.Fatalf("die eingecheckte Datei lesen: %v", err)
	}

	if !bytes.Equal(eingecheckt, gezeichnet.Bytes()) {
		t.Errorf("build/dokument.png ist nicht das, was `Zeichnen` zeichnet — "+
			"neu schreiben mit MSE_SYMBOL_SCHREIBEN=1 (%d gegen %d Bytes)",
			len(eingecheckt), gezeichnet.Len())
	}
}
