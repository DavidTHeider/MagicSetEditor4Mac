package storage

// **Zwei Abfragen, ein Stand** — Werkstück O11 von Phase 5.2
// (#1435),
// Befund 13 des Zweitblicks.
//
// # Warum
//
// SQLite gibt **jeder Anweisung** ihren eigenen Stand. Wer zweimal fragt — erst
// nach der Ordnung, dann nach den Zeilen —, bekommt ohne gemeinsames Geschäft
// womöglich zwei verschiedene Stände, und die Antwort beschreibt dann eine
// Sammlung, die es so nie gab.

import (
	"context"
	"database/sql"
)

// imLesegeschaeft führt `tun` in einer Transaktion aus, die **nur liest**.
//
// **Ein eigenes neben [imGeschaeft]**, weil der Zweck ein anderer ist: dort
// geht es darum, dass mehrere Schreibvorgänge zusammen gelten; hier darum,
// dass mehrere Abfragen **denselben Stand** sehen. SQLite gibt jeder Anweisung
// sonst ihren eigenen, und zwischen zwei Abfragen kann ein anderer Prozess
// geschrieben haben.
func imLesegeschaeft(ctx context.Context, db *sql.DB, was string, tun func(*sql.Tx) error) error {
	geschaeft, fehler := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if fehler != nil {
		return uebersetze(fehler, "%s beginnen", was)
	}

	defer func() { _ = geschaeft.Rollback() }() //nolint:errcheck // nach einem Commit ohne Wirkung

	if fehler := tun(geschaeft); fehler != nil {
		return fehler
	}

	return uebersetze(geschaeft.Commit(), "%s abschließen", was)
}
