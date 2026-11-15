package storage

import (
	"database/sql"
	"testing"
)

// **Die Fehlerwege der Verweisablage** (#1167).
//
// Paketintern wie `exportoptionsfehler_intern_test.go` und aus demselben
// Grund: mit einer funktionierenden Datenbank sind sie nicht erreichbar, weil
// sie funktioniert. Ein Fehlerzweig, den kein Lauf betritt, ist eine
// Behauptung darüber, was im Ernstfall geschieht (§3).
func TestDieVerweisablageMeldetIhreFehlerwege(t *testing.T) {
	geschlossen := geschlosseneDatenbank(t)
	ablage := NeueKartenverweise(geschlossen)

	if _, fehler := ablage.Fuer(probekontext(t), 1); fehler == nil {
		t.Error("das Lesen kam durch")
	}

	if _, fehler := ablage.Loesen(probekontext(t), 1); fehler == nil {
		t.Error("das Lösen kam durch")
	}

	if _, fehler := ablage.JeSet(probekontext(t), 1); fehler == nil {
		t.Error("die Setliste kam durch")
	}
}

// **Und die drei Wege, die nur ein Treiber herbeiführt.**
//
// Ein `RowsAffected`, das scheitert, eine Zeile mit der falschen Spaltenzahl
// und ein Zeilenende, das einen Fehler nachreicht: keines kann der
// SQLite-Treiber, und jedes entscheidet im Ernstfall, ob die Anwendung es
// meldet.
func TestDieVerweisablageMeldetIhreTreiberfehler(t *testing.T) {
	faelle := map[string]func(*sql.DB) error{
		"stolper-zaehlen": func(db *sql.DB) error {
			_, fehler := NeueKartenverweise(db).Loesen(probekontext(t), 1)

			return fehler
		},
		"stolper-umwandeln": func(db *sql.DB) error {
			_, fehler := NeueKartenverweise(db).JeSet(probekontext(t), 1)

			return fehler
		},
		"stolper-lesen": func(db *sql.DB) error {
			_, fehler := NeueKartenverweise(db).JeSet(probekontext(t), 1)

			return fehler
		},
	}

	for treiber, weg := range faelle {
		t.Run(treiber, func(t *testing.T) {
			zweiterSchreibzaehler = 0

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
