// **Der Scan der Revisionsabfrage** — Befund 11 des Zweitblicks über Phase 5.2
// (#1435).
//
// # Warum ein weißer Lauf
//
// `werteImLauf` liest zwei Spalten. Der Zweig, in dem der `Scan` scheitert,
// lässt sich von außen nicht betreten, solange Abfrage und `Scan`
// zusammenpassen — und genau deshalb fiele es nicht auf, wenn seine Meldung
// verloren ginge. Dieselbe Überlegung wie bei
// `TestEinScanGegenDieFalscheSpaltenzahlMeldetSich`.
package storage

import (
	"database/sql"
	"testing"
)

func TestDerRevisionsscanGegenDieFalscheSpaltenzahlMeldetSich(t *testing.T) {
	db := offeneDatenbank(t)

	geschaeft, fehler := db.BeginTx(probekontext(t), &sql.TxOptions{ReadOnly: true})
	if fehler != nil {
		t.Fatalf("das Geschäft beginnen: %v", fehler)
	}

	defer func() { _ = geschaeft.Rollback() }() //nolint:errcheck // nur gelesen

	if _, err := werteImLauf(probekontext(t), geschaeft, `SELECT 1`); err == nil {
		t.Error("ein Scan gegen eine Spalte kam durch")
	}
}
