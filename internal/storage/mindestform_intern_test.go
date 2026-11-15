package storage

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Fehlerwege der Mindestformprüfung** —
// #963.
//
// **Paketintern**, weil beide von außen nicht zu erreichen sind: die
// Migrationsquelle ist eingebettet und immer in Ordnung, und eine geschlossene
// Verbindung entsteht im Betrieb nicht. Genau deshalb stehen sie hier — ein
// Zweig, den kein Lauf betritt, ist eine Behauptung (§3).
func TestDieMindestformMeldetIhreFehlerwege(t *testing.T) {
	t.Parallel()

	t.Run("eine Migration, die es nicht gibt", func(t *testing.T) {
		t.Parallel()

		db := offeneProbe(t)

		err := mindestformPruefen(probekontext(t), db, fstest.MapFS{}, nil, nil,
			map[string]string{"0002_setbestand.sql": ""})

		if !errors.Is(err, domain.ErrSicherung) {
			t.Errorf("der Fehler ist %v", err)
		}
	})

	t.Run("eine geschlossene Verbindung", func(t *testing.T) {
		t.Parallel()

		db := offeneProbe(t)
		if err := db.Close(); err != nil {
			t.Fatalf("schließen: %v", err)
		}

		quelle := fstest.MapFS{
			"migrationen/0002_setbestand.sql": {Data: []byte("CREATE TABLE sets (id INTEGER);")},
		}

		err := mindestformPruefen(probekontext(t), db, quelle, nil, nil,
			map[string]string{"0002_setbestand.sql": ""})

		if !errors.Is(err, domain.ErrSicherung) {
			t.Errorf("der Fehler ist %v", err)
		}
	})

	// **Ohne verbuchte Migration gibt es nichts zu verlangen.**
	//
	// Das ist der Fall der sehr alten Sicherung: sie kennt Tabellen nicht, die
	// es damals nicht gab, und wird migriert.
	t.Run("ohne Buchführung", func(t *testing.T) {
		t.Parallel()

		if err := mindestformPruefen(probekontext(t), offeneProbe(t), fstest.MapFS{}, nil, nil,
			map[string]string{}); err != nil {
			t.Errorf("eine leere Buchführung gab %v", err)
		}
	})
}

// offeneProbe gibt eine leere, offene Datenbank.
func offeneProbe(t *testing.T) *sql.DB {
	t.Helper()

	db, err := FuerLaeufe("file:" + filepath.Join(t.TempDir(), "probe.db"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}

// **Eine umbenannte Tabelle wird unter ihrem neuen Namen verlangt** — und
// nicht unter dem alten. Der Mutant der Zufallsmutation an dieser Stelle
// (`angelegt[neu] = false`, B05, #2224) überlebt zu Recht: `angelegt` wird
// nur über seine Schlüssel gelesen. Dieser Lauf hält fest, was ein anderer
// Eingriff dort brechen könnte.
func TestEineUmbenannteTabelleWirdUnterIhremNeuenNamenVerlangt(t *testing.T) {
	t.Parallel()

	quelle := fstest.MapFS{
		"migrationen/0001_anlage.sql":      {Data: []byte("CREATE TABLE alt (id INTEGER);")},
		"migrationen/0002_umbenennung.sql": {Data: []byte("ALTER TABLE alt RENAME TO neu;")},
	}

	namen, err := tabellenAusMigrationen(quelle, nil, nil, map[string]string{"0001_anlage.sql": "", "0002_umbenennung.sql": ""})
	if err != nil {
		t.Fatal(err)
	}

	if len(namen) != 1 || namen[0] != "neu" {
		t.Errorf("verlangt werden %v, erwartet [neu]", namen)
	}
}

// **Eine weggenommene Tabelle wird nicht verlangt** — angelegt minus
// weggenommen. Bis zur Grundlinie betraten die echten Wanderungen diesen Weg
// mit jedem Tabellenumbau; ihre Tabellendateien enthalten kein `DROP`, und
// deshalb steht er hier mit einer eigenen Quelle.
func TestEineWeggenommeneTabelleWirdNichtVerlangt(t *testing.T) {
	t.Parallel()

	quelle := fstest.MapFS{
		"migrationen/0001_anlage.sql":   {Data: []byte("CREATE TABLE bleibt (x); CREATE TABLE geht (x);")},
		"migrationen/0002_wegnahme.sql": {Data: []byte("DROP TABLE geht;")},
	}

	namen, err := tabellenAusMigrationen(quelle, nil, nil, map[string]string{"0001_anlage.sql": "", "0002_wegnahme.sql": ""})
	if err != nil {
		t.Fatal(err)
	}

	if len(namen) != 1 || namen[0] != "bleibt" {
		t.Errorf("verlangt werden %v, erwartet nur bleibt", namen)
	}
}
