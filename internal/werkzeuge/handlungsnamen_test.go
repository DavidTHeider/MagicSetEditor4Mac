package werkzeuge_test

// **Jede Art, die der Kern stellt oder verkündet, hat ein Wort im Fenster.**
//
// # Warum es diesen Wächter gibt
//
// **Gefunden am nativen Fenster** (`netzpartie09`, P9c, #2377): unter
// „Zuletzt geschehen“ stand `Gast: sitz.platznehmen` — die Ereignisart aus
// P9b hatte keinen Namen im Katalog, und der Rückfall in `partiewoerter.ts`
// zeigte die Kennung. Der Rückfall ist für fremde Daten da (§8), nicht für
// die eigenen.
//
// # Was er hält
//
// Jede Konstante `Befehl…` oder `Ereignis…` unter `internal/service`, deren
// Wert die Gestalt `bereich.art` hat, steht als Schlüssel in
// `handlungsname` des deutschen Katalogs. Dass die anderen Sprachen
// dieselben Schlüssel führen, hält der Typ der Kataloge.
//
// # Was er nicht hält
//
// **Ob das Wort trifft** — das liest ein Mensch. Und eine Art, die nicht als
// solche Konstante dasteht, sieht er nicht.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var kernart = regexp.MustCompile(`(?m)^\s*(?:const\s+)?(?:Befehl|Ereignis)[A-Za-z]*\s*=\s*"([a-z]+\.[a-z]+)"`)

func TestJedeKernartHatEinWortImFenster(t *testing.T) {
	t.Parallel()

	wurzel := filepath.Join("..", "..")
	gosrc := quelltextUnter(t, filepath.Join(wurzel, "internal", "service"), ".go")

	katalog, fehler := os.ReadFile(filepath.Join(wurzel, "frontend", "src", "sprache", "de", "seitenhandlungen.ts")) //nolint:gosec // ein fester Ort im Baum
	if fehler != nil {
		t.Fatal(fehler)
	}

	arten := namenIn(kernart, gosrc)
	if len(arten) < 20 {
		t.Fatalf("nur %d Arten gefunden — liest der Wächter die richtigen Dateien?", len(arten))
	}

	for _, art := range arten {
		if !strings.Contains(string(katalog), "'"+art+"':") {
			t.Errorf("die Art %q hat kein Wort in handlungsname", art)
		}
	}
}
