package storage

// **Wenn das Benennen scheitert, während das Zählen gelang** — Werkstück O4
// von Phase 5.2
// (#1428).
//
// **Ein interner Lauf**, weil er eine [zaehlstelle] braucht, die es im
// Betrieb nicht gibt: dort sind alle Abfragen wohlgeformt. Im Ernstfall
// entscheidet dieser Zweig, ob die Vorschau „drei Sets, keine Namen" sagt —
// oder sagt, dass sie es nicht weiß.

import (
	"testing"
)

func TestEineKrummeNamensabfrageWirdGemeldet(t *testing.T) {
	db, oeffnungsfehler := Oeffnen(t.Context(), ImArbeitsspeicher)
	if oeffnungsfehler != nil {
		t.Fatalf("öffnen: %v", oeffnungsfehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	// **Die Zahl kommt, die Namen nicht.** Genau diese Reihenfolge ist der
	// Zweig: eine leere Namensliste neben einer Zahl größer null sähe aus
	// wie eine Auskunft.
	stelle := zaehlstelle{
		was:   "probe",
		zahl:  `SELECT 3 WHERE ? IS NOT NULL`,
		namen: `SELECT name FROM tabelle_die_es_nicht_gibt WHERE ? IS NOT NULL LIMIT ?`,
	}

	gruppe, fehler := stelle.zaehlen(probekontext(t), db, 1)
	if fehler == nil {
		t.Fatalf("die krumme Abfrage kam durch: %+v", gruppe)
	}

	// **Und die Gruppe trägt die Zahl trotzdem** — sie stand fest, bevor das
	// Benennen scheiterte.
	if gruppe.Anzahl != 3 {
		t.Errorf("die Gruppe ist %+v", gruppe)
	}
}
