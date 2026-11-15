package storage

// **Ein Zug, an einer Stelle** — Werkstück O12 von Phase 5.2
// (#1439).
//
// # Warum das eine eigene Funktion ist
//
// **Weil `BeginTx`, `Rollback` und `Commit` drei Zeilen sind, die immer
// zusammen auftreten — und drei Fehlerwege, die sonst je Aufrufer noch einmal
// dastehen.** Das Kostenjournal hat drei schreibende Wege; ohne diese Klammer
// wären es neun Zweige für dieselben drei Fragen, und jeder einzelne wollte
// eigens geprüft werden (§3).
//
// **Der `Rollback` steht im `defer` und wird nicht geprüft**: nach einem
// geglückten `Commit` ist er wirkungslos, und das ist der Normalfall.

import (
	"context"
	"database/sql"
)

// imGeschaeft führt `tun` in einer Transaktion aus.
//
// **`was` beschreibt den Vorgang** und steht in beiden Fehlermeldungen — wer
// „das Journal öffnen" liest, weiß, welcher Weg gescheitert ist.
func imGeschaeft(ctx context.Context, db *sql.DB, was string, tun func(*sql.Tx) error) error {
	geschaeft, fehler := db.BeginTx(ctx, nil)
	if fehler != nil {
		return uebersetze(fehler, "%s beginnen", was)
	}

	defer func() { _ = geschaeft.Rollback() }() //nolint:errcheck // nach einem Commit ohne Wirkung

	if fehler := tun(geschaeft); fehler != nil {
		return fehler
	}

	return uebersetze(geschaeft.Commit(), "%s abschließen", was)
}
