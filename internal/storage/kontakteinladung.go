package storage

// **Eine Einladung annehmen — ganz oder gar nicht** (NET-01,
// #1705).
//
// # Warum Nonce, Kontakt und Rechte in einer Transaktion stehen
//
// **Weil AK-04 keinen halben Stand zulässt.** Eine verbrauchte Nonce ohne
// Kontakt hätte eine Einladung gekostet, ohne zu wirken; ein Kontakt ohne
// verbrauchte Nonce ließe dieselbe Einladung zweimal zu — beim zweiten Mal
// vielleicht mit einer Freigabe, die beim ersten niemand gab.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Aufnehmen schreibt eine angenommene Einladung.
//
// # Was an einem vorhandenen Kontakt geschieht
//
// **Eine Bestätigung wird nicht zurückgenommen** — wer bestätigt ist, bleibt
// es, auch wenn diesmal nicht verglichen wurde. **Eine Sperre wird nicht
// übergangen**: die Bedingung steht im `WHERE`, und greift sie, geschieht
// nichts, auch nicht mit der Nonce. **Ein leerer eigener Name, eine leere
// Adresse und ein leerer Vorgänger** ersetzen keinen vorhandenen.
func (k *Kontaktablage) Aufnehmen(
	ctx context.Context, aufnahme domain.Kontaktaufnahme, jetzt time.Time,
) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	kontakt := aufnahme.Kontakt
	if fehler := kontakt.Validate(); fehler != nil {
		return fehler
	}

	zeit := jetzt.Format(time.RFC3339Nano)

	return k.inTransaktion(ctx, func(tx *sql.Tx) error {
		if fehler := nonceVerbrauchen(
			ctx, tx, profilID, aufnahme.Nonce, kontakt.Kennung, zeit); fehler != nil {
			return fehler
		}

		ergebnis, err := tx.ExecContext(ctx, `
INSERT INTO peerkontakte
       (profil_id, kennung, fingerabdruck, name, eigener_name, stand, gesehen,
        erstmals, adresse, vorgaenger)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (profil_id, kennung) DO UPDATE
   SET fingerabdruck = excluded.fingerabdruck,
       name          = excluded.name,
       eigener_name  = iif(excluded.eigener_name = '',
                           peerkontakte.eigener_name, excluded.eigener_name),
       adresse       = iif(excluded.adresse = '',
                           peerkontakte.adresse, excluded.adresse),
       vorgaenger    = iif(excluded.vorgaenger = '',
                           peerkontakte.vorgaenger, excluded.vorgaenger),
       stand         = iif(peerkontakte.stand = 'bestaetigt',
                           peerkontakte.stand, excluded.stand),
       fassung       = peerkontakte.fassung + 1
 WHERE peerkontakte.stand <> 'gesperrt'`,
			profilID, kontakt.Kennung, kontakt.Fingerabdruck,
			kontakt.SelbstgewaehlterName, kontakt.EigenerName,
			string(kontakt.Stand), zeit, zeit, kontakt.Adresse,
			kontakt.Vorgaenger)
		if err != nil {
			return uebersetze(err, "eine Einladung annehmen")
		}

		if geaendert, err := ergebnis.RowsAffected(); err != nil || geaendert == 0 {
			return fmt.Errorf(
				"%w: dieses Gegenüber ist gesperrt — erst entsperren, dann "+
					"die Einladung noch einmal annehmen", domain.ErrConflict)
		}

		return rechteGeben(ctx, tx, profilID, kontakt, zeit)
	})
}

// Angenommen sagt, ob dieses Profil eine Einladung schon angenommen hat.
func (k *Kontaktablage) Angenommen(
	ctx context.Context, nonce string,
) (bool, error) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return false, fehler
	}

	var eins int

	err := k.db.QueryRowContext(ctx, `
SELECT 1 FROM kontakteinladungen WHERE profil_id = ? AND nonce = ?`,
		profilID, nonce).Scan(&eins)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return false, nil
	case err != nil:
		return false, uebersetze(err, "eine Einladung nachschlagen")
	}

	return true, nil
}

// nonceVerbrauchen vermerkt die Einladung — oder sagt, dass sie es schon ist.
func nonceVerbrauchen(
	ctx context.Context, tx *sql.Tx, profilID int64, nonce, kennung,
	zeit string,
) error {
	ergebnis, err := tx.ExecContext(ctx, `
INSERT INTO kontakteinladungen (profil_id, nonce, kennung, angenommen)
VALUES (?, ?, ?, ?)
ON CONFLICT (profil_id, nonce) DO NOTHING`, profilID, nonce, kennung, zeit)
	if err != nil {
		return uebersetze(err, "eine Einladung vermerken")
	}

	if neu, err := ergebnis.RowsAffected(); err != nil || neu == 0 {
		return fmt.Errorf(
			"%w: diese Einladung ist schon angenommen — für eine neue "+
				"Verbindung bitte um eine neue", domain.ErrConflict)
	}

	return nil
}

// rechteGeben schreibt die Freigaben — gegen den **gespeicherten** Stand.
func rechteGeben(
	ctx context.Context, tx *sql.Tx, profilID int64,
	kontakt domain.Peerkontakt, zeit string,
) error {
	if len(kontakt.Rechte) == 0 {
		return nil
	}

	id, stand, fehler := kontaktZeile(ctx, tx, profilID, kontakt.Kennung)
	if fehler != nil {
		return fehler
	}

	if fehler := (domain.Peerkontakt{
		Kennung: kontakt.Kennung, Stand: stand, Rechte: kontakt.Rechte,
	}).Validate(); fehler != nil {
		return fehler
	}

	for _, zweck := range kontakt.Rechte {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO peerrechte (kontakt_id, zweck, seit)
VALUES (?, ?, ?)
ON CONFLICT (kontakt_id, zweck) DO NOTHING`, id, string(zweck), zeit); err != nil {
			return uebersetze(err, "einen Zweck freigeben")
		}
	}

	return nil
}
