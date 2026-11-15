//go:build fremdpakete

package service

// **Der Symbolrasterer an echten `.mse-symbol`-Dateien**
// (#323).
//
// Die Einheitstests in `internal/render` prüfen Bereiche, Kombinationsarten
// und Symmetrien an **gebauten** Formen. Hier läuft dieselbe Naht über die
// Vektorzeichnungen, die `Full-Magic-Pack` mitbringt: `msepaket` liest,
// `service` wandelt, `render` rastert.
//
// **Was er prüft**: dass jede gelesen wird und eine Maske ergibt, in der etwas
// steht.
//
// **Was er nicht prüft**: ob die Form stimmt. Dafür bräuchte es ein Goldbild
// von MSE2, und MSE2 rastert ein Symbol nur im Rahmen einer Karte, deren Set
// es führt — ein solches Set liegt nicht vor. Das steht auch im Issue und in
// `docs/FORMATE.md` (§10).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestDerSymbolrastererLiestJedeEchteZeichnung(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	eintraege, err := os.ReadDir(wurzel)
	if err != nil {
		t.Fatalf("Wurzel lesen: %v", err)
	}

	var gelesen, gezeichnet int

	for _, eintrag := range eintraege {
		if !eintrag.IsDir() || !strings.HasSuffix(eintrag.Name(), ".mse-style") {
			continue
		}

		paket := filepath.Join(wurzel, eintrag.Name())

		dateien, fehler := os.ReadDir(paket)
		if fehler != nil {
			continue
		}

		for _, datei := range dateien {
			if !strings.HasSuffix(datei.Name(), ".mse-symbol") {
				continue
			}

			name := eintrag.Name() + "/" + datei.Name()

			zeichnung, _, fehler := msepaket.LiesSymbolzeichnung(
				filepath.Join(paket, datei.Name()))
			if fehler != nil {
				t.Errorf("%s lesen: %v", name, fehler)

				continue
			}

			gelesen++

			if len(zeichnung.Teile) == 0 {
				t.Errorf("%s führt keine Teile", name)

				continue
			}

			maske := render.RastereSymbol(alsZeichenteile(zeichnung.Teile), 40, 40, 0.05, false)

			if zurForm(maske) == 0 {
				t.Errorf("%s ergibt eine leere Maske", name)

				continue
			}

			gezeichnet++
		}
	}

	t.Logf("%d Zeichnungen gelesen, %d gerastert", gelesen, gezeichnet)

	// **Die Gegenprobe zur Reichweite** (§10): findet der Lauf nichts, sagt er
	// nichts — und wäre trotzdem grün. Im Pack liegen 21.
	if gelesen < 21 {
		t.Errorf("nur %d Zeichnungen gefunden, erwartet mindestens 21", gelesen)
	}
}

// zurForm zählt die Bildpunkte, die zur Form gehören.
func zurForm(maske render.Symbolmaske) int {
	var zahl int

	for y := range 40 {
		for x := range 40 {
			if maske.Bei(x, y) != render.SymbolAussen {
				zahl++
			}
		}
	}

	return zahl
}
