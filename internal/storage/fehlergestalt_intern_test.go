// Prüfungen an `sqlitecode` — der Frage, **ob** ein Fehler von SQLite kommt.
//
// **Eine eigene Datei, weil es eine eigene Frage ist** (§2). Nebenan steht die
// Zuordnung von Code zu Fehlerart; hier steht, in welcher **Gestalt** der Code
// überhaupt ankommt. Beim Treiberwechsel hat sich das zweite geändert und das
// erste nicht.
package storage

import (
	"errors"
	"fmt"
	"testing"

	sqlite "github.com/ncruces/go-sqlite3"
)

// **Derselbe Fehler kommt in drei Gestalten** — gemessen am 2026-09-06 beim
// Wechsel auf `ncruces/go-sqlite3` (Werkstück 14 von Phase 4.12,
// #1351).
//
// Ein `INSERT` mit `NULL` gibt einen `*sqlite3.Error`; ein `BeginTx` auf einer
// Datei, die keine Datenbank ist, gibt einen **blanken**
// `sqlite3.ExtendedErrorCode` — einen Zahlwert, der `error` erfüllt, ohne Hülle
// darum. Ein `errors.As` auf den Zeiger allein greift an der zweiten daneben.
//
// **Was das kostete, waren zwei Auskünfte, an denen ein Mensch etwas tut**
// (§8): „die Datei ist keine Datenbank" kam als „von einem anderen Programm
// geöffnet" an, und „belegt — gleich nochmal" gar nicht mehr.
func TestJedeGestaltEinesSQLiteFehlersWirdErkannt(t *testing.T) {
	faelle := []struct {
		name      string
		fehler    error
		basis     sqlite.ErrorCode
		erweitert sqlite.ExtendedErrorCode
	}{
		{
			name:      "blanker erweiterter Code",
			fehler:    sqlite.NOTADB.ExtendedCode(),
			basis:     sqlite.NOTADB,
			erweitert: sqlite.NOTADB.ExtendedCode(),
		},
		{
			name:      "blanker Basiscode",
			fehler:    sqlite.BUSY,
			basis:     sqlite.BUSY,
			erweitert: sqlite.BUSY.ExtendedCode(),
		},
		{
			// **Auch umschlossen**, denn `uebersetze` legt eine Lage darum,
			// bevor irgendjemand fragt.
			name:      "umschlossener blanker Code",
			fehler:    fmt.Errorf("etwas tun: %w", sqlite.CONSTRAINT_UNIQUE),
			basis:     sqlite.CONSTRAINT,
			erweitert: sqlite.CONSTRAINT_UNIQUE,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			erweitert, ist := sqlitecode(fall.fehler)

			if !ist {
				t.Fatalf("%v wurde nicht als SQLite-Fehler erkannt", fall.fehler)
			}

			if erweitert.Code() != fall.basis || erweitert != fall.erweitert {
				t.Errorf("Codes %d/%d, erwartet waren %d/%d",
					erweitert.Code(), erweitert, fall.basis, fall.erweitert)
			}
		})
	}
}

// **Und was nicht von SQLite kommt, kommt nicht von SQLite.** Ohne diesen Fall
// bliebe offen, ob die Funktion überhaupt jemals verneint — eine Zusage, die
// immer ja sagt, ordnete jeden Fehler der Domäne zu.
func TestEinFremderFehlerTraegtKeinenSQLiteCode(t *testing.T) {
	// **Der Code wird mitgeprüft und nicht weggeworfen.**
	//
	// Nicht nur, weil `sqlite3.ExtendedErrorCode` ein `error` ist und
	// `errcheck` das Wegwerfen anstreicht — sondern weil es die schärfere
	// Aussage ist: ein Aufrufer, der das `ist` übersähe, bekäme sonst
	// irgendeinen Code statt der Null.
	for name, fehler := range map[string]error{
		"ein fremder Fehler": errors.New("etwas ging schief"),
		"gar kein Fehler":    nil,
	} {
		t.Run(name, func(t *testing.T) {
			code, ist := sqlitecode(fehler)

			if ist {
				t.Errorf("%v wurde als SQLite-Fehler gelesen", fehler)
			}

			if code != 0 {
				t.Errorf("Code %d, erwartet war 0", code)
			}
		})
	}
}
