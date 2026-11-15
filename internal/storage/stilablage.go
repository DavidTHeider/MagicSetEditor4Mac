package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Stilablage führt Stil und Stiloptionen eines Sets.
//
// # Warum eine eigene Ablage und nicht der SetLeser
//
// Sie schreibt **quer**: `sets.stil`, `sets.spiel` und `styling_values` gehören
// zu einem Vorgang, und der ist eine Transaktion. Am Leser hinge damit ein
// Schreibweg, und die Trennung, die `SetLeser` und `SetSpeicher` seit Phase 2
// haben, wäre für ein Feld aufgegeben.
//
// **`styling_values` steht seit Phase 2 im Schema und war leer** — dieselbe
// unerfüllte Zusage wie bei `packages` (#54).
type Stilablage struct {
	db *sql.DB
}

// NeueStilablage baut die Ablage über einer Datenbank.
func NeueStilablage(db *sql.DB) *Stilablage { return &Stilablage{db: db} }

// Stilwerte gibt die Stiloptionen eines Sets in ihrer Reihenfolge.
func (s *Stilablage) Stilwerte(ctx context.Context, setID int64) ([]domain.Value, error) {
	// **Nur aus dem gezeigten Profil** (#1815).
	if fehler := setImProfil(ctx, s.db, setID); fehler != nil {
		return nil, fehler
	}

	zeilen, err := s.db.QueryContext(ctx,
		`SELECT feld, wert FROM styling_values WHERE set_id = ? ORDER BY reihenfolge`, setID)
	if err != nil {
		return nil, uebersetze(err, "die Stiloptionen des Sets %d lesen", setID)
	}
	defer func() { _ = zeilen.Close() }() //nolint:errcheck // der Fehler kommt über zeilen.Err

	werte := make([]domain.Value, 0)

	for zeilen.Next() {
		var wert domain.Value

		if err := zeilen.Scan(&wert.Field, &wert.Text); err != nil {
			return nil, uebersetze(err, "eine Stiloption lesen")
		}

		werte = append(werte, wert)
	}

	if err := zeilen.Err(); err != nil {
		return nil, uebersetze(err, "die Stiloptionen des Sets %d lesen", setID)
	}

	return werte, nil
}

// StilSetzen schreibt Spiel, Stil und die bleibenden Stiloptionen eines Sets.
//
// **In einer Transaktion, und das ist der Punkt.** Ein Set, dessen Stil schon
// der neue ist und dessen Optionen noch die alten sind, sieht wie ein Set aus —
// und der nächste Lauf rechnete mit Werten, die zu keinem Stil gehören.
//
// **Die Werte werden ersetzt, nicht ergänzt.** Was der Aufrufer übergibt, ist
// der vollständige neue Stand; was er weglässt, kennt der neue Stil nicht.
// Welche das sind, entscheidet `service` — es kennt beide Stile.
func (s *Stilablage) StilSetzen(
	ctx context.Context, setID int64, spiel, stil string, werte []domain.Value,
) error {
	// **Jede Wurzelabfrage klemmt auf ein Profil** (#1351, Vermerk 0066).
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "Transaktion beginnen")
	}
	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach dem Commit ist sie zu

	ergebnis, fehler := tx.ExecContext(ctx,
		`UPDATE sets SET spiel = ?, stil = ? WHERE id = ? AND profil_id = ?`,
		spiel, stil, setID, profilID)
	if fehler != nil {
		return uebersetze(fehler, "den Stil des Sets %d setzen", setID)
	}

	betroffen, zaehlfehler := ergebnis.RowsAffected()
	if zaehlfehler != nil {
		return uebersetze(zaehlfehler, "den Stil des Sets %d setzen", setID)
	}

	// **Ein Set, das es nicht gibt, ist ein `ErrNotFound`** — und hier fällt das
	// Urteil, weil hier geschrieben wird. Ohne diese Zeile ginge ein Wechsel an
	// einem gelöschten Set als geglückt durch, und die Oberfläche zeigte einen
	// Stil, den niemand trägt.
	if betroffen == 0 {
		return fmt.Errorf("%w: das Set %d führt der Bestand nicht", domain.ErrNotFound, setID)
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM styling_values WHERE set_id = ?`, setID); err != nil {
		return uebersetze(err, "die alten Stiloptionen räumen")
	}

	for stelle, wert := range werte {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO styling_values (set_id, feld, wert, reihenfolge) VALUES (?, ?, ?, ?)`,
			setID, wert.Field, wert.Text, stelle); err != nil {
			return uebersetze(err, "die Stiloption %q schreiben", wert.Field)
		}
	}

	if abschluss := tx.Commit(); abschluss != nil {
		return uebersetze(abschluss, "den Stilwechsel abschließen")
	}

	return nil
}

// StiloptionSetzen schreibt **eine** Stiloption, ohne die Nachbarn und ohne
// `sets.stil` zu berühren.
//
// # Warum es das gibt
//
// Bis zum 2026-08-22 ging jede Optionsänderung über [Stilablage.StilSetzen]:
// `service` las die vollständige Liste, änderte einen Eintrag und schrieb alles
// zurück — samt Spiel und Stil. Der Zweitblick über Phase L hat zwei Verluste
// daran angestrichen, und beide sind im Zwei-Schreiber-Betrieb der Normalfall
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)):
//
//  1. **Zwei Optionsschreibungen verlieren einander.** Beide lesen denselben
//     Stand; die zuletzt fertige ersetzt die Änderung der ersten.
//  2. **Ein alter Optionslauf dreht einen Stilwechsel zurück**, weil er den
//     Stil mitträgt, den er beim Lesen sah.
//
// **`styling_values` trägt `PRIMARY KEY (set_id, feld)`** — eine einzelne
// Option lässt sich damit in **einer** Anweisung setzen. Kein Lesen davor, also
// nichts, was zwischen Lesen und Schreiben veralten könnte.
//
// # Die Reihenfolge einer neuen Option
//
// Sie kommt ans Ende. `Stilwerte` ordnet danach, und eine Option, die es vorher
// nicht gab, hat keine Stelle, an die sie gehörte — anders als ein Kartenfeld,
// das an seinen Platz zurückkehrt (#160).
func (s *Stilablage) StiloptionSetzen(ctx context.Context, setID int64, feld, wert string) error {
	// **Ein Set, das es nicht gibt oder das einem anderen Profil gehört, ist
	// ein `ErrNotFound`** — gefragt im selben Zug wie geschrieben (#1815).
	return imGeschaeft(ctx, s.db, "eine Stiloption setzen", func(tx *sql.Tx) error {
		if fehler := setImProfil(ctx, tx, setID); fehler != nil {
			return fehler
		}

		_, fehler := tx.ExecContext(ctx, `
			INSERT INTO styling_values (set_id, feld, wert, reihenfolge)
			VALUES (?, ?, ?,
				(SELECT COALESCE(MAX(reihenfolge), -1) + 1 FROM styling_values WHERE set_id = ?))
			ON CONFLICT (set_id, feld) DO UPDATE SET wert = excluded.wert`,
			setID, feld, wert, setID)

		return uebersetze(fehler, "die Stiloption %q des Sets %d setzen", feld, setID)
	})
}

// StiloptionEntfernen nimmt **eine** Stiloption aus dem Bestand.
//
// **Eine Option, die gar nicht dasteht, ist kein Fehler** — es ist bereits der
// Zustand, den der Aufrufer wollte
// (#764).
func (s *Stilablage) StiloptionEntfernen(ctx context.Context, setID int64, feld string) error {
	return imGeschaeft(ctx, s.db, "eine Stiloption entfernen", func(tx *sql.Tx) error {
		if fehler := setImProfil(ctx, tx, setID); fehler != nil {
			return fehler
		}

		_, fehler := tx.ExecContext(ctx,
			`DELETE FROM styling_values WHERE set_id = ? AND feld = ?`, setID, feld)

		return uebersetze(fehler, "die Stiloption %q des Sets %d entfernen", feld, setID)
	})
}
