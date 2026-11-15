package storage

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// **Jeder Auslöser steht im Katalog, mit der einen Invariante, die er hält**
// — ENG-05.AK-02 (#1730).
//
// # Warum
//
// **Ein Auslöser ist die Stelle, an der Geschäftslogik unbemerkt ins Schema
// wandert.** AK-02 lässt ihn nur für „klar benannte datenbanklokale
// Invarianten" zu. Benannt ist er erst, wenn er in `docs/INVARIANTEN.md`
// steht — mit seiner Tafel. Ein neuer Auslöser ohne Zeile dort macht diesen
// Lauf rot, und eine Zeile ohne Auslöser ebenso.
//
// # Was er nicht hält
//
// **Ob die Invariante in der Zeile die ist, die der Auslöser hält.** Das
// liest ein Mensch.
func TestJederAusloeserStehtImKatalog(t *testing.T) {
	t.Parallel()

	bild := schemaAus(t)

	katalog, fehler := os.ReadFile(filepath.Join("..", "..", "docs", "INVARIANTEN.md"))
	if fehler != nil {
		t.Fatalf("den Katalog lesen: %v", fehler)
	}

	_, abschnitt, gefunden := strings.Cut(string(katalog), "## Die Auslöser")
	if !gefunden {
		t.Fatal("INVARIANTEN.md hat keinen Abschnitt »Die Auslöser«")
	}

	abschnitt, _, _ = strings.Cut(abschnitt, "\n## ")

	zeile := regexp.MustCompile("(?m)^\\| `([a-z_]+)` \\| `([a-z_]+)` \\| .+ \\|$")

	genannt := map[string]string{}
	for _, fund := range zeile.FindAllStringSubmatch(abschnitt, -1) {
		genannt[fund[1]] = fund[2]
	}

	for name, tafel := range bild.tafelVon {
		switch zugeordnet, steht := genannt[name]; {
		case !steht:
			t.Errorf("der Auslöser %s steht nicht in INVARIANTEN.md", name)
		case zugeordnet != tafel:
			t.Errorf("der Auslöser %s hängt an %s, der Katalog sagt %s", name, tafel, zugeordnet)
		}
	}

	for name := range genannt {
		if _, gibtEs := bild.tafelVon[name]; !gibtEs {
			t.Errorf("INVARIANTEN.md nennt den Auslöser %s, den es nicht gibt", name)
		}
	}

	if len(genannt) == 0 {
		t.Error("der Katalog nennt keinen Auslöser — das Muster greift daneben")
	}
}
