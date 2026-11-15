// Die Fehlerpfade der Schlüsselwortpflege
// (#1062).
//
// **Paketintern**, weil [geschlosseneDatenbank] es ist — aus dem Grund, der in
// `fehlerpfade_test.go` steht.
package storage

import (
	"database/sql"
	"testing"
)

// **Jeder Weg meldet, wenn die Datenbank nicht antwortet.**
//
// Ein Fehlerzweig, den kein Test betritt, ist eine Behauptung darüber, was im
// Ernstfall geschieht (§3).
func TestDieSchluesselwortwegeMeldenEineTauteDatenbank(t *testing.T) {
	db := geschlosseneDatenbank(t)
	pflege := NeueSchluesselwortablage(db)

	if _, err := pflege.SchluesselwortAnlegen(probekontext(t), 1, "custom"); err == nil {
		t.Error("das Anlegen kam durch")
	}

	if err := pflege.SchluesselwortSetzen(
		probekontext(t), 1, 1, "wort", "X"); err == nil {
		t.Error("das Setzen kam durch")
	}

	if err := pflege.SchluesselwortLoeschen(probekontext(t), 1, 1); err == nil {
		t.Error("das Löschen kam durch")
	}

	if _, err := NeuerSetLeser(db).Schluesselwoerter(probekontext(t), 1); err == nil {
		t.Error("das Lesen kam durch")
	}
}

// **Und die beiden Wege, die nur ein Treiber herbeiführt.**
//
// Ein `RowsAffected`, das scheitert, und eine Zeile mit der falschen
// Spaltenzahl: beides kann der SQLite-Treiber nicht, und beides entscheidet im
// Ernstfall, ob die Anwendung es meldet (§3).
func TestDieSchluesselwortwegeMeldenIhreTreiberfehler(t *testing.T) {
	faelle := map[string]func(*sql.DB) error{
		"stolper-zaehlen": func(db *sql.DB) error {
			return NeueSchluesselwortablage(db).SchluesselwortSetzen(
				probekontext(t), 1, 1, "wort", "X")
		},
		"stolper-umwandeln": func(db *sql.DB) error {
			_, fehler := NeuerSetLeser(db).Schluesselwoerter(probekontext(t), 1)

			return fehler
		},
	}

	for treiber, weg := range faelle {
		t.Run(treiber, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			if fehler := weg(db); fehler == nil {
				t.Errorf("%s: kein Fehler", treiber)
			}
		})
	}
}
