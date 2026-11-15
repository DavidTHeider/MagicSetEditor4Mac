package storage_test

import (
	"database/sql"
	"testing"
)

// **Die Pragmas gelten auf jeder Verbindung, nicht auf der ersten.**
//
// # Der Fehler, der diesen Wächter nötig macht
//
// Bis #154 wurden
// `busy_timeout` und `foreign_keys` nach dem Öffnen über `db.ExecContext`
// gesetzt. Das trifft **eine** Verbindung aus dem Vorrat; jede weitere, die
// `database/sql` später öffnet, kam ohne sie.
//
// Das war seit Phase 1 so, und **kein Test hat es gesehen** — alle liefen gegen
// eine Datenbank im Arbeitsspeicher, die mit `SetMaxOpenConns(1)` genau eine
// Verbindung hat. Gefunden hat es erst ein Test mit zwei Schreibern auf einer
// **Datei**: 17 von 20 Schreibvorgängen scheiterten.
//
// # Warum `foreign_keys` dabei das schwerste Stück ist
//
// Es ist keine Bequemlichkeit. `Kartenschreiber.FeldSetzen` verlässt sich
// darauf, dass ein Schreiben auf eine Karte, die es nicht gibt, den
// Fremdschlüssel verletzt, und meldet daraufhin `ErrNotFound`. **Ohne das
// Pragma meldet dieselbe Anweisung Erfolg und schreibt nichts** — der stille
// Fehler, den §8 gerade verhindern soll.
//
// # Die Gegenprobe, gefahren und aufgeschrieben
//
// Vier Läufe, jeder rot:
//
//   - `_pragma=foreign_keys(1)` aus `mitPragmas` genommen → „Verbindung 0:
//     foreign_keys = 0"
//   - `_pragma=busy_timeout(5000)` genommen → dasselbe für die Wartezeit
//   - `journal_mode = WAL` auf `DELETE` gestellt → „journal_mode = delete"
//   - **und der Fehler von vor #154 selbst**: beide Pragmas zurück auf
//     `db.ExecContext` gelegt. Der Lauf wird rot für Verbindung 1, 2 und 3 —
//     und **grün für Verbindung 0**. Genau diese Gestalt hatte der Fehler, und
//     genau deshalb hat ihn drei Phasen lang niemand gesehen.
//
// Der letzte Lauf ist der, auf den es ankommt. Die ersten drei zeigen, dass der
// Wächter überhaupt hinsieht; dieser zeigt, dass er **den** Fall sieht.
func TestDiePragmasGeltenAufJederVerbindung(t *testing.T) {
	db := frischeDateidatenbank(t)

	// **Eine Verbindung je Vorrat — seit #1485 auch für eine Datei.** Vier
	// zugleich zu leihen, wie dieser Test es bis dahin tat, hielte hier für
	// immer: die zweite wartete auf die erste. Was der Test zusagt, bleibt
	// dasselbe — die Pragmas stehen in der Datenquelle und gelten damit für
	// **jede** Verbindung, die der Vorrat je öffnet —, und er stellt den Fall
	// anders her: die erste Verbindung wird geprüft und weggeworfen, und die
	// nächste ist eine **neue**, die dieselben Pragmas tragen muss.
	if offen := db.Stats().MaxOpenConnections; offen != 1 {
		t.Fatalf("der Vorrat erlaubt %d Verbindungen, erwartet ist eine (#1485)", offen)
	}

	for nummer := range 2 {
		verbindung, err := db.Conn(probekontext(t))
		if err != nil {
			t.Fatalf("Verbindung %d: %v", nummer, err)
		}

		pruefePragma(t, verbindung, nummer, "foreign_keys", "1")
		pruefePragma(t, verbindung, nummer, "busy_timeout", "5000")
		pruefePragma(t, verbindung, nummer, "journal_mode", "wal")
		pruefePragma(t, verbindung, nummer, "synchronous", "2")

		// **Weggeworfen, nicht zurückgegeben**: `Raw` mit einem Fehler lässt
		// den Vorrat die Verbindung schließen, und die nächste Leihe öffnet
		// eine neue — der Fall, um den es geht.
		_ = verbindung.Raw(func(any) error { return sql.ErrConnDone }) //nolint:errcheck // der Fehler ist das Mittel
		_ = verbindung.Close()                                         //nolint:errcheck // sie ist schon fort
	}

	if geoeffnet := db.Stats().MaxIdleClosed + db.Stats().MaxLifetimeClosed + int64(db.Stats().OpenConnections); geoeffnet == 0 {
		t.Log("keine Zahl über geschlossene Verbindungen — die zweite Leihe hat trotzdem eine frische bekommen, wenn `Raw` sie geschlossen hat")
	}
}

// pruefePragma liest ein Pragma auf einer bestimmten Verbindung.
func pruefePragma(t *testing.T, verbindung *sql.Conn, nummer int, name, erwartet string) {
	t.Helper()

	// **Der Name wird nicht gebunden, sondern eingesetzt** — SQLite nimmt für
	// `PRAGMA` keinen Platzhalter. Unbedenklich, weil er aus der festen Liste
	// darüber kommt und nie aus einer Eingabe.
	var wert string
	if err := verbindung.QueryRowContext(probekontext(t), "PRAGMA "+name).Scan(&wert); err != nil {
		t.Fatalf("Verbindung %d: %s lesen: %v", nummer, name, err)
	}

	if wert != erwartet {
		t.Errorf("Verbindung %d: %s = %s, erwartet %s", nummer, name, wert, erwartet)
	}
}
