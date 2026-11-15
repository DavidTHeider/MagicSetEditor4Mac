package domain_test

// **Ein Erwerbsdatum ist ein Tag in kanonischer Form** — UX-06.AK-02
// (#1678).
//
// # Warum es diese Läufe gibt
//
// **Das Fenster zeigt ein Datumsfeld, die anderen Flächen nahmen alles.**
// `mse`, REST und MCP reichten jeden Text durch, und `Normalisieren` kürzte
// nur die Ränder: „gestern" und „31.02.2026" standen als Erwerbsdatum im
// Bestand — und die Sammlung wird nach genau diesem Text geordnet
// (`storage/sammlungsfenster.go`). Eine Ordnung über „gestern" ist keine.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinErwerbsdatumIstEinTag(t *testing.T) {
	for _, fall := range []struct {
		name    string
		eingabe string
		gilt    bool
	}{
		{"leer — niemand weiß es", "", true},
		{"Jahr-Monat-Tag", "2026-09-24", true},
		{"mit Rändern, die Normalisieren kürzt", " 2026-09-24 ", true},
		{"ein Tag, den es nicht gibt", "2026-02-30", false},
		{"die deutsche Schreibweise", "24.09.2026", false},
		{"ohne führende Null", "2026-9-24", false},
		{"ein Wort", "gestern", false},
		{"mit Uhrzeit", "2026-09-24T10:00:00Z", false},
	} {
		t.Run(fall.name, func(t *testing.T) {
			exemplar := domain.Exemplar{
				Printing: "abc", Menge: 1, Sprache: "en", Erwerbsdatum: fall.eingabe,
			}
			exemplar.Normalisieren()

			befund := befundAm(exemplar.Validate(), "erwerbsdatum")
			if fall.gilt != (befund == nil) {
				t.Errorf("%q: gilt %v, Befund %v", fall.eingabe, fall.gilt, befund)
			}
		})
	}
}
