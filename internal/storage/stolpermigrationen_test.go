// **Die zwei Prüfungen über den Stolpertreiber an der Migration.**
//
// **Aus `stolperabfragen_test.go` herausgelöst**, als sie über 400 Zeilen ging
// (§2). Der Schnitt folgt der Frage: dort steht, **was** der Treiber
// antwortet, hier, **was** damit geprüft wird.
package storage

import (
	"database/sql"
	"errors"
	"testing"
)

// **Das Commit steht seit #938
// an der Folge**, nicht an der einzelnen Migration — geprüft wird deshalb
// `anwendenAlle`.
func TestDieMigrationsfolgeMeldetEinGescheitertesCommit(t *testing.T) {
	db, err := sql.Open("stolper-commit", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	fehler := anwendenAlle(probekontext(t), db, einfacheQuelle(),
		[]string{"0001_probe.sql"}, nil, nil, nil, nil)
	if !errors.Is(fehler, errGestolpert) {
		t.Errorf("Fehler %v, erwartet war das gescheiterte Commit", fehler)
	}
}

func TestAngewendeteMigrationenMeldetEinenLesefehlerNachDerSchleife(t *testing.T) {
	db, err := sql.Open("stolper-lesen", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, fehler := AngewendeteMigrationen(probekontext(t), db); !errors.Is(fehler, errGestolpert) {
		t.Errorf("Fehler %v, erwartet war der Lesefehler", fehler)
	}
}
