package storage

// **Die Fehlerwege der Handelskette** — Werkstück `T1` von Phase 5.6
// (#1466).
//
// **Sie sind mit einer echten SQLite-Datenbank nicht herzustellen**: ein
// Lesefehler **nach** der Schleife entsteht an einer abgebrochenen Verbindung
// oder einer vollen Platte, nicht an einer Abfrage, die durchkam. Dafür gibt
// es die Stolpertreiber — dieselben wie bei der Massenänderung.

import (
	"database/sql"
	"testing"
)

func TestDieHandelsketteMeldetIhreLesefehler(t *testing.T) {
	for _, treiber := range []string{"stolper-umwandeln", "stolper-lesen"} {
		t.Run(treiber, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			if _, fehler := NeueHandelsablage(db).
				Revisionen(probekontext(t), "h1"); fehler == nil {
				t.Error("der Lesefehler ging verloren")
			}

			// **Und das Protokoll** (TRADE-03.AK-03, #1717).
			if _, fehler := NeueHandelsablage(db).
				Protokoll(probekontext(t), "h1"); fehler == nil {
				t.Error("der Lesefehler des Protokolls ging verloren")
			}
		})
	}
}
