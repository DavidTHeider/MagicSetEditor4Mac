package storage

// **Der Zählfehler der Autoersetzungsablage** — Werkstück E von Phase 4.11
// (#1341).
//
// **Ein interner Lauf**, weil er den Stolpertreiber braucht: `RowsAffected`
// scheitert beim SQLite-Treiber nie, und im Ernstfall entscheidet es, ob die
// Anwendung „die Regel stand nicht da" von „das Zählen ging schief"
// unterscheidet.

import (
	"database/sql"
	"testing"
)

func TestDieAutoersetzungsablageMeldetIhrenZaehlfehler(t *testing.T) {
	db, err := sql.Open("stolper-zaehlen", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	zweiterSchreibzaehler = 0

	if fehler := NeueAutoersetzungsablage(db).
		Loeschen(probekontext(t), "magic", "(W)"); fehler == nil {
		t.Error("der Zählfehler ging verloren")
	}
}
