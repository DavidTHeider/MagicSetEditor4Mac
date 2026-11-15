// **Die Revision der Rechte** — Werkstück O1 von Phase 5.2
// (#1425),
// nach [Vermerk 0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md).
//
// # Warum hier nur gelesen wird
//
// **Weil das Steigen in der Wanderung steht** — als Auslöser an `bestand`
// und `profilfreigaben` (`0047_freigabenachweise.sql`). Sieben Wege ändern,
// wer wo was darf, und einer davon ist die Kaskade beim Löschen eines
// Profils, die kein Aufrufer sieht. Ein Zähler, den jeder Weg selbst erhöhen
// müsste, wäre an dem Tag falsch, an dem einer es vergisst — und ein Nachweis
// gälte dann still weiter.

package storage

import "context"

// Revision gibt den Stand der Rechte — eine Zahl, die mit jedem Wechsel und
// jeder Freigabe steigt.
func (a *Nutzerablage) Revision(ctx context.Context) (int64, error) {
	var revision int64

	if err := a.db.QueryRowContext(ctx,
		`SELECT revision FROM freigabenrevision WHERE id = 1`).Scan(&revision); err != nil {
		return 0, uebersetze(err, "die Revision der Rechte lesen")
	}

	return revision, nil
}
