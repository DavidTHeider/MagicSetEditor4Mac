package storage

// **Ein `RowsAffected`, das scheitert, wird beim Einlösen gemeldet** —
// Werkstück O1 von Phase 5.2
// (#1425).
//
// **Ein interner Lauf**, weil er den Stolpertreiber braucht: SQLite zählt
// immer, und der Weg, den niemand herbeiführen kann, ist der, dessen Meldung
// unbemerkt verloren ginge (§3).

import "testing"

func TestDieNachweisablageMeldetIhrenZaehlfehler(t *testing.T) {
	db := stolperbestand(t, "stolper-zaehlen")

	zweiterSchreibzaehler = 0

	if fehler := NeueNachweisablage(db).NachweisEinloesen(probekontext(t), "n-1"); fehler == nil {
		t.Error("der Zählfehler ging verloren")
	}
}
