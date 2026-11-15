package storage

// **Die Fehlerwege des Vermerks** —
// #891.
//
// Er zählt die betroffenen Zeilen, um „das Set gibt es nicht" von
// „geschrieben" zu trennen; beide Wege dahinter sind mit einer echten
// Datenbank nicht herzustellen. Dafür gibt es den Stolpertreiber — derselbe
// Weg wie bei `HerkunftSetzen` (#226).

import (
	"database/sql"
	"errors"
	"testing"
)

func TestSetBenutztMeldetDieFehlerpfade(t *testing.T) {
	faelle := map[string]string{
		"das Schreiben scheitert": "stolper-schreiben",
		"die Zählung scheitert":   "stolper-zaehlen",
	}

	for name, treiber := range faelle {
		t.Run(name, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			if fehler := NeuerSetSpeicher(db).SetBenutzt(probekontext(t), 1); !errors.Is(
				fehler, errGestolpert) {
				t.Errorf("Fehler %v, erwartet war der Stolperfehler", fehler)
			}
		})
	}
}
