package storage

// **Ausgestellt und eingelöst** — NET-01b
// (#1954).
//
// # Warum eine Transaktion
//
// **Weil eine Einlösung zwei Zeilen schreibt**: den Kontakt und den Vermerk,
// dass die Nonce verbraucht ist. Stünde das eine ohne das andere da, wäre die
// Einladung verbraucht, ohne gewirkt zu haben — oder sie wirkte ein zweites
// Mal.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ausgestelltGnade ist, wie lange eine abgelaufene Einladung noch als
// „verbraucht" oder „abgelaufen" erkannt wird, bevor sie „fremd" heißt.
const ausgestelltGnade = 7 * 24 * time.Hour

// ablaufform ist die Form, in der `ablauf` steht — Sekunden, UTC.
func ablaufform(zeit time.Time) string {
	return zeit.UTC().Truncate(time.Second).Format(time.RFC3339)
}

// Ausgestellt vermerkt eine Einladung, die dieses Profil ausgegeben hat — und
// räumt, was seit einer Woche abgelaufen ist.
func (k *Kontaktablage) Ausgestellt(
	ctx context.Context, nonce string, zweck domain.Zweck, ablauf, jetzt time.Time,
) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	return k.inTransaktion(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
DELETE FROM ausgestellte_einladungen WHERE profil_id = ? AND ablauf < ?`,
			profilID, ablaufform(jetzt.Add(-ausgestelltGnade))); err != nil {
			return uebersetze(err, "alte Einladungen räumen")
		}

		if _, err := tx.ExecContext(ctx, `
INSERT INTO ausgestellte_einladungen (profil_id, nonce, zweck, ablauf)
VALUES (?, ?, ?, ?)`, profilID, nonce, string(zweck), ablaufform(ablauf)); err != nil {
			return uebersetze(err, "eine Einladung vermerken")
		}

		return nil
	})
}

// Einloesen nimmt auf, wer eine Einladung dieses Profils einlöst — ganz
// oder gar nicht.
//
// **Unbestätigt und ohne Recht** (NET-01b.AK-01): wer schon bestätigt ist,
// bleibt es, und wer gesperrt ist, löst nichts ein. Die Nonce bleibt dann
// unverbraucht, damit nach dem Entsperren nicht eine tote Einladung übrig
// ist.
func (k *Kontaktablage) Einloesen(
	ctx context.Context, nonce string, kontakt domain.Peerkontakt, jetzt time.Time,
) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	if fehler := kontakt.Validate(); fehler != nil {
		return fehler
	}

	zeit := jetzt.Format(time.RFC3339Nano)

	return k.inTransaktion(ctx, func(tx *sql.Tx) error {
		if fehler := einloesbar(ctx, tx, profilID, nonce, jetzt); fehler != nil {
			return fehler
		}

		ergebnis, err := tx.ExecContext(ctx, `
INSERT INTO peerkontakte
       (profil_id, kennung, fingerabdruck, name, eigener_name, stand, gesehen,
        erstmals, adresse)
VALUES (?, ?, ?, ?, '', 'unbestaetigt', ?, ?, ?)
ON CONFLICT (profil_id, kennung) DO UPDATE
   SET fingerabdruck = excluded.fingerabdruck,
       name          = excluded.name,
       gesehen       = excluded.gesehen,
       adresse       = iif(excluded.adresse = '',
                           peerkontakte.adresse, excluded.adresse),
       fassung       = peerkontakte.fassung + 1
 WHERE peerkontakte.stand <> 'gesperrt'`,
			profilID, kontakt.Kennung, kontakt.Fingerabdruck,
			kontakt.SelbstgewaehlterName, zeit, zeit, kontakt.Adresse)
		if err != nil {
			return uebersetze(err, "einen Einlösenden aufnehmen")
		}

		if geaendert, err := ergebnis.RowsAffected(); err != nil || geaendert == 0 {
			return domain.ErrEinloesungGesperrt
		}

		if _, err := tx.ExecContext(ctx, `
UPDATE ausgestellte_einladungen SET eingeloest = ?, kennung = ?
 WHERE profil_id = ? AND nonce = ?`, zeit, kontakt.Kennung, profilID, nonce); err != nil {
			return uebersetze(err, "eine Einlösung vermerken")
		}

		return nil
	})
}

// einloesbar sagt, ob die Nonce hier ausgestellt, unverbraucht und gültig
// ist — in dieser Reihenfolge, weil „verbraucht" mehr sagt als „abgelaufen".
func einloesbar(
	ctx context.Context, tx *sql.Tx, profilID int64, nonce string, jetzt time.Time,
) error {
	var ablauf, eingeloest string

	err := tx.QueryRowContext(ctx, `
SELECT ablauf, eingeloest FROM ausgestellte_einladungen
 WHERE profil_id = ? AND nonce = ?`, profilID, nonce).Scan(&ablauf, &eingeloest)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return domain.ErrEinladungFremd
	case err != nil:
		return uebersetze(err, "eine Einladung nachschlagen")
	case eingeloest != "":
		return domain.ErrEinladungVerbraucht
	case ablaufform(jetzt) >= ablauf:
		return fmt.Errorf("%w (seit %s)", domain.ErrEinladungAbgelaufen, ablauf)
	}

	return nil
}
