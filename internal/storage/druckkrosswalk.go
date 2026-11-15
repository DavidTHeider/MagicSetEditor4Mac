package storage

// **Die Karte von MTGJSONs Kennungen auf die Drucke dieses Bestands** —
// Werkstück A3 von Phase 5.3
// (#1442).
//
// # Wofür es das braucht
//
// Die Preisdatei von MTGJSON ist nach **seiner** Kennung geordnet und führt
// jede Karte, die es gibt — über hunderttausend, über ein Gigabyte. Der
// Bestand führt die, die jemand besitzt.
//
// **Was dazwischen liegt, ist diese Karte**: `internal/mtgjsonpreise` liest
// nur, was darin steht, und überspringt den Rest, **ohne ihn zu lesen**.
//
// # Warum sie hier steht und nicht im Blatt
//
// Weil sie eine Frage an den **Bestand** ist. `internal/mtgjsonpreise` kennt
// `domain` und `netz` und sonst nichts aus diesem Projekt (§2); wer ihm den
// Bestand gäbe, machte aus einem Formatleser einen Anwendungsfall.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Druckkrosswalk gibt je Druck seine MTGJSON-Kennung und seinen Scryfall-Alias.
//
// # Nur Drucke, die beide haben
//
// **Der Scryfall-Alias ist Pflicht an einem Preisstand** ([domain.Preisstand]):
// ein Druck ohne ihn hat keine Zeile, an der ein Preis hängen könnte. Und ohne
// MTGJSON-Kennung ist in der Preisdatei nichts zu finden.
//
// # Nur die Semantik »druck«
//
// **Eine Oracle-ID nennt das Konzept**, das dreißig Drucke teilen
// ([0086](../../docs/entscheidungen/0086-sieben-begriffe-fuer-eine-karte.md)).
// Wer sie hier führte, ordnete denselben Preis dreißig Karten zu.
func Druckkrosswalk(ctx context.Context, db *sql.DB) (map[string]string, error) {
	zeilen, fehler := db.QueryContext(ctx,
		`SELECT fremd.wert, eigen.wert
		   FROM druckkennungen fremd
		   JOIN druckkennungen eigen ON eigen.druck_id = fremd.druck_id
		  WHERE fremd.system = ? AND fremd.semantik = ?
		    AND eigen.system = ? AND eigen.semantik = ?`,
		string(domain.SystemMTGJSON), string(domain.SemantikDruck),
		string(domain.SystemScryfall), string(domain.SemantikDruck))
	if fehler != nil {
		return nil, uebersetze(fehler, "die Druckkennungen lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird über Err()

	karte := map[string]string{}

	for zeilen.Next() {
		var fremd, eigen string
		if lesefehler := zeilen.Scan(&fremd, &eigen); lesefehler != nil {
			return nil, uebersetze(lesefehler, "eine Druckkennung lesen")
		}

		karte[fremd] = eigen
	}

	return karte, uebersetze(zeilen.Err(), "die Druckkennungen lesen")
}
