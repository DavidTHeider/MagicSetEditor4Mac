package storage

// **Wer wessen Ergänzungen später sehen darf** — Werkstück G5 von Phase 5.5
// (#1462).
//
// **Eine eigene Datei, weil es eine eigene Frage ist.** `partien.go` hält den
// Zustand, `partienstrom.go` den Verlauf; hier steht die Einwilligung, und sie
// gehört zu keinem von beiden — sie wirkt erst, wenn die Partie vorbei ist.

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Freigaben gibt die Einwilligungen einer Partie.
//
// **Eine je Sitz und nicht eine je Zeile.** Die Tabelle hält Paare; was der
// Aufrufer braucht, ist „wer hat wem etwas freigegeben" — und das ist eine
// Zusammenfassung, keine Abfrage.
func (a *Partieablage) Freigaben(
	ctx context.Context, partieID string,
) ([]domain.Offenlegung, error) {
	// **Nur aus dem gezeigten Profil** (#1815).
	if fehler := partieImProfil(ctx, a.db, partieID); fehler != nil {
		return nil, fehler
	}

	zeilen, fehler := a.db.QueryContext(ctx,
		`SELECT von, an FROM partiefreigaben
		  WHERE partie_id = ? ORDER BY von, an`, partieID)
	if fehler != nil {
		return nil, uebersetze(fehler, "die Freigaben lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird über Err()

	gesammelt := []domain.Offenlegung{}
	stelle := map[string]int{}

	for zeilen.Next() {
		var von, an string
		if lesefehler := zeilen.Scan(&von, &an); lesefehler != nil {
			return nil, uebersetze(lesefehler, "eine Freigabe lesen")
		}

		wo, schon := stelle[von]
		if !schon {
			gesammelt = append(gesammelt, domain.Offenlegung{Von: von})
			wo = len(gesammelt) - 1
			stelle[von] = wo
		}

		gesammelt[wo].An = append(gesammelt[wo].An, an)
	}

	if fehler := zeilen.Err(); fehler != nil {
		return nil, uebersetze(fehler, "die Freigaben lesen")
	}

	return gesammelt, nil
}

// Freigeben schreibt die Einwilligung **eines** Sitzes — und zwar ganz.
//
// **Erst weg, dann hin, in einem Geschäft.** Wer nur einfügte, hätte nach dem
// zweiten Ruf die Vereinigung beider Listen — und ein Sitz, der einen
// Empfänger herausnimmt, hätte ihn nicht herausgenommen.
func (a *Partieablage) Freigeben(
	ctx context.Context, partieID string, gabe domain.Offenlegung,
) error {
	return a.imFreigabengeschaeft(ctx, partieID, gabe.Von,
		func(geschaeft *sql.Tx) error {
			for _, an := range gabe.An {
				if _, fehler := geschaeft.ExecContext(ctx,
					`INSERT INTO partiefreigaben (partie_id, von, an)
					 VALUES (?, ?, ?)`, partieID, gabe.Von, an); fehler != nil {
					return uebersetze(fehler, "die Freigabe eintragen")
				}
			}

			return nil
		})
}

// Freigabeloeschen nimmt die Einwilligung eines Sitzes zurück.
func (a *Partieablage) Freigabeloeschen(
	ctx context.Context, partieID, von string,
) error {
	return a.imFreigabengeschaeft(ctx, partieID, von,
		func(*sql.Tx) error { return nil })
}

// imFreigabengeschaeft räumt die Zeilen eines Sitzes weg und lässt dann
// schreiben.
func (a *Partieablage) imFreigabengeschaeft(
	ctx context.Context, partieID, von string, dann func(*sql.Tx) error,
) error {
	geschaeft, fehler := a.db.BeginTx(ctx, nil)
	if fehler != nil {
		return uebersetze(fehler, "die Freigabe schreiben")
	}

	defer func() { _ = geschaeft.Rollback() }() //nolint:errcheck // nach Commit wirkungslos

	// **Nur an einer Partie des gezeigten Profils** — beide Wege gehen durch
	// dieses Geschäft (#1815).
	if fehler := partieImProfil(ctx, geschaeft, partieID); fehler != nil {
		return fehler
	}

	if _, fehler := geschaeft.ExecContext(ctx,
		`DELETE FROM partiefreigaben WHERE partie_id = ? AND von = ?`,
		partieID, von); fehler != nil {
		return uebersetze(fehler, "die bisherige Freigabe entfernen")
	}

	if fehler := dann(geschaeft); fehler != nil {
		return fehler
	}

	if fehler := geschaeft.Commit(); fehler != nil {
		return uebersetze(fehler, "die Freigabe schreiben")
	}

	return nil
}

// Vergessen löscht eine Partie samt Strom, Befehlen und Freigaben.
//
// # Der Löschweg aus dem Issue
//
// **#1462 verlangt einen „Aufbewahrungs-/Löschweg".** Dies ist seine
// ausdrückliche Hälfte: jemand sagt, dass dieses Protokoll weg soll. Die
// andere ist die Frist, und sie steht in der Datenpflege —
// `storage/datenpflege.go`, Stelle `alteParties`.
//
// **Die Fremdschlüssel räumen mit** (`ON DELETE CASCADE` in 0064 und 0065),
// und `PRAGMA foreign_keys` steht in diesem Bestand an. Ein zweites `DELETE`
// daneben wäre eine zweite Zusage über dieselbe Sache — und die erste, die
// vergessen wird, merkt niemand.
//
// **Was es nicht verspricht**: dass es woanders keine Kopie gibt. Wer einen
// Verlauf ausgegeben hat, hat ihn ausgegeben; das Issue sagt denselben Satz
// für die Deckleihe, und er gilt hier genauso.
func (a *Partieablage) Vergessen(ctx context.Context, partieID string) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	ergebnis, fehler := a.db.ExecContext(ctx,
		`DELETE FROM partien WHERE id = ? AND profil_id = ?`,
		partieID, profilID)
	if fehler != nil {
		return uebersetze(fehler, "die Partie vergessen")
	}

	betroffen, fehler := ergebnis.RowsAffected()
	if fehler != nil {
		return uebersetze(fehler, "die Partie vergessen")
	}

	if betroffen == 0 {
		return fmt.Errorf("%w: die Partie %q gibt es in diesem Profil nicht",
			domain.ErrNotFound, partieID)
	}

	return nil
}
