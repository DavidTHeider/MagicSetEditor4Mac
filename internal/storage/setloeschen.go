// **Ein Set wieder loswerden** —
// #1218.
//
// # Warum der Name mitkommt
//
// `sets` führt keine `fassung`, und für einen unumkehrbaren Vorgang wäre sie
// auch das schwächere Mittel. Der **Name** tut hier zweierlei zugleich: er ist
// die Bestätigung — wer ihn nennt, hat das Set vor sich gehabt — und er ist
// der Wächter gegen einen überholten Stand (§7, Regel 2). Wurde das Set
// zwischen Lesen und Löschen umbenannt, stimmt er nicht mehr, und es bleibt
// stehen.
//
// # Warum eine Transaktion für ein einziges `DELETE`
//
// Nicht wegen der Kaskaden — die sind in einer Anweisung ohnehin ganz oder
// gar nicht. Sondern wegen der **Prüfung davor**: Name lesen und löschen sind
// zwei Anweisungen, und wer sie nicht klammert, löscht ein Set, das zwischen
// beiden umbenannt wurde. Genau das soll der Name verhindern.

package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// SetLoeschen nimmt ein Set samt allem, was an ihm hängt, aus dem Bestand.
//
// **Die Kaskaden des Schemas räumen den Anhang**, und der Volltextindex geht
// mit: `card_suche` hängt an einem `AFTER DELETE`-Auslöser auf `card_values`,
// und der feuert auch bei einer **kaskadierten** Löschung.
//
// **Gemessen am 2026-08-27** und nicht angenommen: bei `recursive_triggers = 0`
// — dem Wert, den diese Anwendung mitbringt — bleiben nach dem Löschen eines
// von zwei Sets genau die zwei Indexzeilen des anderen stehen. Eine
// ausdrückliche Räumung fände nie eine Zeile (§3).
//
// **Die Sammlung bleibt unberührt** ([0052](../../docs/entscheidungen/0052-die-sammlung-ist-nicht-das-set.md)):
// ein Exemplar hängt an einem Druck, nicht an einem Set. Wer sein Set löscht,
// hat die Karten nicht verkauft.
func (s *SetSpeicher) SetLoeschen(ctx context.Context, setID int64, name string) error {
	// **Jede Wurzelabfrage klemmt auf ein Profil** (#1351, Vermerk 0066).
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "das Löschen beginnen")
	}
	defer func() { _ = tx.Rollback() }()

	if pruefung := setnamePruefen(ctx, tx, setID, profilID, name); pruefung != nil {
		return pruefung
	}

	if _, loeschfehler := tx.ExecContext(ctx,
		`DELETE FROM sets WHERE id = ? AND profil_id = ?`, setID, profilID); loeschfehler != nil {
		return uebersetze(loeschfehler, "das Set %d löschen", setID)
	}

	if abschluss := tx.Commit(); abschluss != nil {
		return uebersetze(abschluss, "das Löschen abschließen")
	}

	return nil
}

// setnamePruefen verlangt, dass das Set da ist und so heißt.
//
// **Zwei Antworten auf denselben Fehlschlag**, und der Unterschied ist der
// halbe Zweck: „gibt es nicht" heißt, die Liste ist veraltet; „heißt anders"
// heißt, jemand hat umbenannt (§8).
func setnamePruefen(
	ctx context.Context, tx *sql.Tx, setID, profilID int64, name string,
) error {
	var vorhanden string

	err := tx.QueryRowContext(ctx,
		`SELECT name FROM sets WHERE id = ? AND profil_id = ?`, setID, profilID).
		Scan(&vorhanden)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return fmt.Errorf("%w: das Set %d gibt es nicht", domain.ErrNotFound, setID)

	case err != nil:
		return uebersetze(err, "den Namen des Sets %d lesen", setID)
	}

	if vorhanden != name {
		return fmt.Errorf(
			"%w: das Set %d heißt %q und nicht %q — zum Löschen muss der Name genau stimmen",
			domain.ErrConflict, setID, vorhanden, name)
	}

	return nil
}
