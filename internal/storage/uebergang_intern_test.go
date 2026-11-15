// Die Kaskade eines gelöschten Sets bis in den Volltextindex.
//
// Hier standen bis zum 2026-10-12 auch zwei Übergänge von einem Bestand aus der
// Zeit vor `0004_volltext_mit_inhalt.sql`. Seit es nur die Grundlinie gibt
// (Vermerk 0189), entsteht ein solcher Bestand nicht mehr — er wird mit Rat
// abgewiesen, und das hält `grundlinienstand_intern_test.go`.
package storage

import (
	"database/sql"
	"testing"
)

// **Die Kaskade reicht bis in den Index.**
//
// Der bestehende Synchronitätstest löscht `card_values` unmittelbar. Im Betrieb
// löscht niemand Feldwerte — es wird ein **Set** gelöscht, und die Kaskade geht
// über `cards` und `card_values` bis zum Trigger. Ob sie dort ankommt, stand
// nirgends.
func TestEinGeloeschtesSetVerschwindetAusDemIndex(t *testing.T) {
	db := offeneDatenbank(t)

	if err := Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	fuelleBestand(t, db)

	if zahl := treffer(t, db, `"Feuerball"*`); zahl != 1 {
		t.Fatalf("%d Treffer vor dem Löschen", zahl)
	}

	if _, err := db.ExecContext(probekontext(t), `DELETE FROM sets WHERE id = 1`); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	if zahl := treffer(t, db, `"Feuerball"*`); zahl != 0 {
		t.Errorf("%d Treffer nach dem Löschen des Sets", zahl)
	}
}

// fuelleBestand legt ein Set mit einer Karte und einem Feldwert an.
func fuelleBestand(t *testing.T, db *sql.DB) {
	t.Helper()

	befehle := []string{
		`INSERT INTO sets (id, name, spiel, geoeffnet, profil_id)
		   VALUES (1, 'Probe', 'probe', '2026-08-08T10:00:00.000Z', 1)`,
		`INSERT INTO cards (id, set_id, reihenfolge) VALUES (1, 1, 0)`,
		`INSERT INTO card_values VALUES (1, 'name', 'Feuerball', 0)`,
	}

	for _, befehl := range befehle {
		if _, err := db.ExecContext(probekontext(t), befehl); err != nil {
			t.Fatalf("%s: %v", befehl, err)
		}
	}
}

// treffer zählt die Treffer einer FTS-Abfrage.
func treffer(t *testing.T, db *sql.DB, muster string) int {
	t.Helper()

	var zahl int

	if err := db.QueryRowContext(probekontext(t),
		`SELECT count(*) FROM card_suche WHERE card_suche MATCH ?`, muster).Scan(&zahl); err != nil {
		t.Fatalf("zählen: %v", err)
	}

	return zahl
}
