package storage

import (
	"database/sql"
	"errors"
)

// zurueckgenommen rollt `tx` zurück, **gleich wie** die Funktion endet, und
// gibt den Fehler zurück, mit dem sie endet — als
// `defer func() { fehler = zurueckgenommen(tx, fehler) }()` gleich nach dem
// Beginn.
//
// **Auch bei einem `panic`** (BK-04, #2129). Hier stand bis dahin
// `if fehler != nil { … }`, und ein Zusammenbruch setzt den benannten Fehler
// nicht: die Transaktion blieb offen, hielt die einzige Verbindung des Pools
// und die Schreibsperre der Datei, bis das Programm endete. Nach einem
// geglückten `Commit` ist das Zurückrollen ein `ErrTxDone` und wirkungslos
// ([verwerfen]). Ein Fehler beim Zurückrollen wird nur an einen schon
// vorhandenen gehängt — ohne ihn gab es nichts zurückzunehmen.
func zurueckgenommen(tx *sql.Tx, fehler error) error {
	zurueck := verwerfen(tx)
	if fehler == nil {
		return nil
	}

	return errors.Join(fehler, zurueck)
}

// verwerfen rollt `tx` zurück; nach einem Commit ist das ein `ErrTxDone` und
// keine Nachricht wert.
func verwerfen(tx *sql.Tx) error {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return uebersetze(err, "zurückrollen")
	}

	return nil
}
