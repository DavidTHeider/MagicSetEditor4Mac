package storage

// **Ein Kennwort wechselt** — ID-02
// (#1701).
//
// **Nur die gehütete Form, und nur diese Spalte**: ein Kennwortwechsel ändert
// weder Rolle noch Freigaben noch Profile (ID-02.AK-04). Die Fassung des
// Nutzers steigt nicht — sie hütet die Angaben, die ein Formular zeigt, und
// ein Kennwort zeigt keines.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// KennwortSetzen legt die gehütete Form eines neuen Kennworts ab.
func (a *Nutzerablage) KennwortSetzen(ctx context.Context, nutzerID int64, gehuetet string) error {
	return kennwortSetzen(ctx, a.db, nutzerID, gehuetet)
}

// KennwortMitCodeSetzen verbraucht einen Code vom Zettel, setzt das
// Kennwort und hält den Zustand `danach` fest — **in einem Geschäft**
// (Zweitblick über ID-02, Befund 5; Nachlauf, Befund 2). Ein schon
// verbrauchter oder fremder Code setzt nichts, ein Kennwort, das sich nicht
// setzen lässt, verbraucht keinen Code, und eine Sperre, die sich nicht lösen
// lässt, lässt beides stehen.
func (a *Nutzerablage) KennwortMitCodeSetzen(
	ctx context.Context, nutzerID, codeID int64, gehuetet string, danach domain.Sitzungszustand,
) error {
	return imGeschaeft(ctx, a.db, "das Kennwort über den Zettel setzen", func(geschaeft *sql.Tx) error {
		ergebnis, fehler := geschaeft.ExecContext(ctx,
			`UPDATE wiederherstellungscodes SET gebraucht_am = datetime('now')
			 WHERE id = ? AND nutzer_id = ? AND gebraucht_am = ''`, codeID, nutzerID)
		if fehler != nil {
			return uebersetze(fehler, "den Code abstempeln")
		}

		if zeilen, zaehlfehler := betroffen(ergebnis); zaehlfehler != nil || zeilen == 0 {
			return errors.Join(zaehlfehler,
				fmt.Errorf("%w: diesen Wiederherstellungscode gibt es nicht mehr", domain.ErrNotFound))
		}

		if fehler := kennwortSetzen(ctx, geschaeft, nutzerID, gehuetet); fehler != nil {
			return fehler
		}

		return sitzungFesthaltenIn(ctx, geschaeft, danach)
	})
}

// ausfuehrer ist, was ein UPDATE ausführen kann — die Datenbank oder ein
// Geschäft.
type ausfuehrer interface {
	ExecContext(ctx context.Context, frage string, werte ...any) (sql.Result, error)
}

func kennwortSetzen(ctx context.Context, ausfuehren ausfuehrer, nutzerID int64, gehuetet string) error {
	ergebnis, fehler := ausfuehren.ExecContext(ctx, `UPDATE nutzer SET kennwort = ? WHERE id = ?`, gehuetet, nutzerID)
	if fehler != nil {
		return uebersetze(fehler, "das Kennwort des Nutzers %d setzen", nutzerID)
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return zaehlfehler
	}

	if zeilen == 0 {
		return fmt.Errorf("%w: den Nutzer %d gibt es nicht", domain.ErrNotFound, nutzerID)
	}

	return nil
}
