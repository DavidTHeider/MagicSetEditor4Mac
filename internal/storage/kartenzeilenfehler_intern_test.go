package storage

import (
	"database/sql"
	"testing"
)

// **Die drei Fehlerwege der Kartenliste**
// (#891).
//
// Eine echte SQLite-Datenbank gibt sie nicht her: die Abfrage steht fest, der
// Scan passt, und eine vollständige Schleife meldet nichts. Beides kommt im
// Betrieb vor — eine volle Platte, eine abgebrochene Verbindung —, und dafür
// gibt es den Stolpertreiber (`stolpertreiber_test.go`).
//
// # Was hier nicht geprüft wird
//
// **Woran der Scan scheitert.** `krummezeile` gibt acht Spalten; diese beiden
// Abfragen lesen vier und zwei, also scheitert er schon an der **Anzahl** und
// nicht an der Umwandlung. Geprüft ist damit, dass ein gescheiterter Scan
// gemeldet wird — nicht, welcher Wert ihn scheitern ließe.
func TestDieKartenlisteMeldetJedenFehlerweg(t *testing.T) {
	// Die drei Treiber, und was sie in dieser Reihenfolge treffen: die Abfrage
	// selbst, den Scan einer Zeile, und den Abschluss nach der Schleife.
	for _, treiber := range []string{
		"stolper-fragen", "stolper-umwandeln", "stolper-lesen",
	} {
		t.Run(treiber, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			leser := NeuerSetLeser(db)

			if _, fehler := leser.Kartenzeilen(probekontext(t), 1); fehler == nil {
				t.Error("Kartenzeilen: kein Fehler")
			}

			if _, fehler := leser.Kartenfeldwerte(probekontext(t), 1, "name"); fehler == nil {
				t.Error("Kartenfeldwerte: kein Fehler")
			}
		})
	}
}
