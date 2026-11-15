// **Vorgaben, die auf Zeilen zeigen — im Bestand** — Werkstück 15 von
// Phase 4.12
// (#1351),
// Punkte 4 bis 7.
//
// # Warum Setzen ersetzt und nicht ergänzt
//
// **Weil eine Vorgabe eine Wahl ist und keine Sammlung.** Wer drei
// Präferenzdecks nennt, hat drei gewählt — und wer danach eines nennt, eines.
// Eine Naht, die nur ergänzen könnte, ließe sich nicht leeren; eine, die
// beides kann, bräuchte je Ruf eine Angabe, was gemeint ist.
//
// **Und deshalb geht es in einer Transaktion**: Räumen und Schreiben sind ein
// Vorgang, und ein Abbruch dazwischen ließe ein Profil ohne Vorgaben zurück,
// das vorher welche hatte.

package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Vorgabenablage führt die Profilvorgaben.
type Vorgabenablage struct{ db *sql.DB }

// NeueVorgabenablage baut sie über einer Datenquelle.
func NeueVorgabenablage(db *sql.DB) *Vorgabenablage { return &Vorgabenablage{db: db} }

// Vorgaben gibt alle Vorgaben eines Profils, nach Art.
//
// **Eine Tabelle und keine Liste**: die Fläche fragt nach einer Art, und eine
// Liste zwänge sie zum Durchsuchen. **Arten ohne Wert fehlen darin** — das ist
// der Normalfall, und ein leerer Eintrag wäre die teurere Art, das zu sagen.
func (a *Vorgabenablage) Vorgaben(ctx context.Context, profilID int64) (
	map[domain.Vorgabeart][]string, error,
) {
	// **Ein Profil ohne Vorgaben und eines, das es nicht gibt, sind
	// zweierlei** (§8). Ohne diese Frage bekäme ein Tippfehler in der Kennung
	// fünf leere Arten zurück — und die Seite sähe aus, als wäre nichts
	// gesetzt, statt zu sagen, dass sie ins Leere sieht.
	//
	// **Der Preis ist eine zweite Abfrage je Lesen**, und er ist derselbe, den
	// `GET /sets/{id}/karten` zahlt.
	if fehler := a.profilGibtEs(ctx, profilID); fehler != nil {
		return nil, fehler
	}

	zeilen, err := a.db.QueryContext(ctx,
		`SELECT art, wert FROM profilvorgaben
		 WHERE profil_id = ? ORDER BY art, ordnung, wert`, profilID)
	if err != nil {
		return nil, uebersetze(err, "die Vorgaben des Profils %d lesen", profilID)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // nur gelesen

	gefunden := map[domain.Vorgabeart][]string{}

	for zeilen.Next() {
		var art, wert string

		if err := zeilen.Scan(&art, &wert); err != nil {
			return nil, uebersetze(err, "eine Vorgabezeile lesen")
		}

		gefunden[domain.Vorgabeart(art)] = append(
			gefunden[domain.Vorgabeart(art)], wert)
	}

	if err := zeilen.Err(); err != nil {
		return nil, uebersetze(err, "die Vorgaben durchgehen")
	}

	return gefunden, nil
}

// VorgabeSetzen ersetzt die Werte **einer** Art.
//
// **Die anderen Arten bleiben stehen.** Ein Setzen, das alles abräumte, wäre
// bei fünf Arten ein Zurücksetzen mit vier ungewollten Nebenwirkungen.
func (a *Vorgabenablage) VorgabeSetzen(
	ctx context.Context, profilID int64, vorgabe domain.Profilvorgabe,
) error {
	if fehler := vorgabe.Validate(); fehler != nil {
		return fehler
	}

	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "die Vorgaben setzen")
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach Commit wirkungslos

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM profilvorgaben WHERE profil_id = ? AND art = ?`,
		profilID, string(vorgabe.Art)); err != nil {
		return uebersetze(err, "die bisherigen Vorgaben räumen")
	}

	for stelle, wert := range vorgabe.Werte {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO profilvorgaben (profil_id, art, wert, ordnung)
			 VALUES (?, ?, ?, ?)`,
			profilID, string(vorgabe.Art), wert, stelle); err != nil {
			return uebersetze(err, "die Vorgabe %q eintragen", wert)
		}
	}

	if err := tx.Commit(); err != nil {
		return uebersetze(err, "die Vorgaben festschreiben")
	}

	return nil
}

// profilGibtEs weist eine Kennung ab, zu der kein Profil gehört.
func (a *Vorgabenablage) profilGibtEs(ctx context.Context, profilID int64) error {
	var zahl int

	if err := a.db.QueryRowContext(ctx,
		`SELECT count(*) FROM profile WHERE id = ?`, profilID).Scan(&zahl); err != nil {
		return uebersetze(err, "das Profil %d nachschlagen", profilID)
	}

	if zahl == 0 {
		return fmt.Errorf("%w: das Profil %d gibt es nicht",
			domain.ErrNotFound, profilID)
	}

	return nil
}
