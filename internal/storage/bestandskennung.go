// **Die Kennung dieses Bestands** — Werkstück 7 von Phase 4.12
// (#1351).
//
// # Wozu sie da ist
//
// **Weil ein Geheimnis im Schlüsselbund unter einem Schlüssel liegt, und der
// muss sagen, wessen Geheimnis es ist.** Die Nutzerkennung allein reicht
// nicht: jeder Bestand auf diesem Rechner hat einen Nutzer 1.
//
// **Gefunden durch den Flächenlauf und nicht überlegt** — siehe Wanderung
// 0039.

package storage

import (
	"context"
	"database/sql"
)

// Bestandskennung gibt die Kennung dieses Bestands.
//
// **Sie steht seit Wanderung 0039 fest** und ändert sich nie: wer sie ändert,
// verliert jedes Geheimnis, das unter ihr im Schlüsselbund liegt.
func Bestandskennung(ctx context.Context, db *sql.DB) (string, error) {
	var wert string

	fehler := db.QueryRowContext(ctx,
		`SELECT wert FROM bestandskennung WHERE id = 1`).Scan(&wert)

	return wert, uebersetze(fehler, "die Bestandskennung lesen")
}
