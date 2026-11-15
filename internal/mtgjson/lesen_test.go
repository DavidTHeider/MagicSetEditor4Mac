package mtgjson_test

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// Eine Datei im Format *Individual Set* wird gelesen, samt ihrer Karten.
//
// Die Datei trägt den Umschlag aus `meta` und `data`, den MTGJSON seit Fassung
// 5 schreibt. Dass er da ist, ist belegt: das Spielpaket `magic.mse-game`
// nennt in `json_paths` den Pfad `data:cards:` (docs/FORMATE.md).
func TestIndividualSetWirdGelesen(t *testing.T) {
	quelle, err := mtgjson.Lies(filepath.Join("..", "..", "testdaten", "mtgjson", "probe.json"), mtgjson.ZuordnungMagic)
	if err != nil {
		t.Fatalf("Lies: %v", err)
	}

	if len(quelle.Sets) != 1 {
		t.Fatalf("Sets: %d, erwartet 1", len(quelle.Sets))
	}

	set := quelle.Sets[0]
	if set.Code != "PRB" {
		t.Errorf("Code: %q, erwartet %q", set.Code, "PRB")
	}

	if set.Name != "Probeset" {
		t.Errorf("Name: %q, erwartet %q", set.Name, "Probeset")
	}

	if len(set.Karten) != 2 {
		t.Fatalf("Karten: %d, erwartet 2", len(set.Karten))
	}
}
