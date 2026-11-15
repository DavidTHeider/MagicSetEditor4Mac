// **Der Dateibaum durch die echte Naht** — FMT-02.AK-01
// (#1691):
// anlegen, einfügen, umbenennen, entfernen gegen ein Paket im Zuhause des
// Tests, so wie Kommandozeile, MCP und Dienst die Naht rufen.
package app_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestDieWerkstattnahtFuehrtDenDateibaum(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paket := paketImZuhause(t)

	naht, fehler := app.WerkstattnahtUeber()
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	ctx := probekontext(t)

	stand, err := naht.Anlegen(ctx, paket, "neu", "# neu\n")
	if err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	if err = naht.Umbenennen(ctx, paket, "neu", "alt", stand); err != nil {
		t.Fatalf("Umbenennen: %v", err)
	}

	bild := filepath.Join(t.TempDir(), "maske.png")
	if err = os.WriteFile(bild, []byte("\x89PNG\r\n\x1a\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err = naht.Einfuegen(ctx, paket, bild, "maske.png"); err != nil {
		t.Fatalf("Einfügen: %v", err)
	}

	dateien, beiwerk, err := naht.DateienUndBeiwerk(ctx, paket)
	if err != nil || !slices.Contains(dateien, "alt") || !slices.Equal(beiwerk, []string{"maske.png"}) {
		t.Fatalf("Dateien %v, Beiwerk %v (%v)", dateien, beiwerk, err)
	}

	if err = naht.Entfernen(ctx, paket, "maske.png", ""); err != nil {
		t.Fatalf("Entfernen: %v", err)
	}

	if _, beiwerk, err = naht.DateienUndBeiwerk(ctx, paket); err != nil || len(beiwerk) != 0 {
		t.Errorf("das Beiwerk steht noch da: %v (%v)", beiwerk, err)
	}
}
