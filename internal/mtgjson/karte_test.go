package mtgjson_test

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// probeSet liest das selbst geschriebene Beispielset.
func probeSet(t *testing.T) *mtgjson.Set {
	t.Helper()

	quelle, err := mtgjson.Lies(filepath.Join("..", "..", "testdaten", "mtgjson", "probe.json"), mtgjson.ZuordnungMagic)
	if err != nil {
		t.Fatalf("Lies: %v", err)
	}

	if len(quelle.Sets) != 1 {
		t.Fatalf("Sets: %d, erwartet 1", len(quelle.Sets))
	}

	return quelle.Sets[0]
}

// Die Zuordnung bildet MTG-JSON-Felder auf die Kartenfelder von `magic` ab.
//
// Die Zielnamen sind **belegt**, nicht erfunden: sie stehen in `card_fields`
// des Pakets `magic.mse-game` (docs/FORMATE.md). Wo MSE ein Feld ableitet
// (`save value: false`), wird nicht dorthin geschrieben, sondern in die Felder,
// aus denen es entsteht.
func TestZuordnungMagicBildetKartenfelderAb(t *testing.T) {
	set := probeSet(t)

	faelle := map[string]struct {
		karte    int
		erwartet []domain.Value
	}{
		"eine Karte ohne Kampfwerte": {
			karte: 0,
			erwartet: []domain.Value{
				{Field: "name", Text: "Flammenball"},
				{Field: "casting cost", Text: "2R"},
				{Field: "super type", Text: "Instant"},
				{Field: "rule text", Text: "Flammenball fügt einem Ziel 3 Schaden zu."},
				{Field: "flavor text", Text: "Es begann mit einem Funken."},
				{Field: "rarity", Text: "common"},
				{Field: "illustrator", Text: "Selbst gezeichnet"},
			},
		},
		"eine Kreatur mit Ober- und Untertypen": {
			karte: 1,
			erwartet: []domain.Value{
				{Field: "name", Text: "Eiserner Wächter"},
				{Field: "casting cost", Text: "3WW"},
				{Field: "super type", Text: "Legendary Creature"},
				{Field: "sub type", Text: "Golem Soldier"},
				{Field: "rule text", Text: "Wachsamkeit"},
				{Field: "power", Text: "4"},
				{Field: "toughness", Text: "6"},
				{Field: "rarity", Text: "mythic rare"},
				{Field: "illustrator", Text: "Selbst gezeichnet"},
			},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			gefunden := set.Karten[fall.karte].Werte

			if len(gefunden) != len(fall.erwartet) {
				t.Fatalf("Werte: %v, erwartet %v", gefunden, fall.erwartet)
			}

			for i, wert := range fall.erwartet {
				if gefunden[i] != wert {
					t.Errorf("Wert %d: %+v, erwartet %+v", i, gefunden[i], wert)
				}
			}
		})
	}
}
