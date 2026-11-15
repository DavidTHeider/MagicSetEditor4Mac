package storage

// **Die Freigaben eines Kontakts** — Werkstück N1 von Phase 5.4
// (#1454).

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Erlauben gibt einem Kontakt einen Zweck frei.
//
// **Ein bestätigtes Gegenüber vorausgesetzt.** Die Regel steht in
// [domain.Peerkontakt.Validate] und wird hier gegen den **gespeicherten**
// Stand geprüft — in derselben Transaktion, in der geschrieben wird. Ein Blick
// davor wäre eine Auskunft und keine Bedingung.
func (k *Kontaktablage) Erlauben(
	ctx context.Context, kennung string, zweck domain.Zweck, jetzt time.Time,
) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	if fehler := domain.PruefeZweck(zweck); fehler != nil {
		return fehler
	}

	return k.inTransaktion(ctx, func(tx *sql.Tx) error {
		id, stand, fehler := kontaktZeile(ctx, tx, profilID, kennung)
		if fehler != nil {
			return fehler
		}

		if fehler := (domain.Peerkontakt{
			Kennung: kennung,
			Stand:   stand,
			Rechte:  []domain.Zweck{zweck},
		}).Validate(); fehler != nil {
			return fehler
		}

		if _, err := tx.ExecContext(ctx, `
INSERT INTO peerrechte (kontakt_id, zweck, seit)
VALUES (?, ?, ?)
ON CONFLICT (kontakt_id, zweck) DO NOTHING`,
			id, string(zweck), jetzt.Format(time.RFC3339Nano)); err != nil {
			return uebersetze(err, "einen Zweck freigeben")
		}

		return nil
	})
}

// Entziehen nimmt eine Freigabe zurück.
//
// **Ein Zweck, der gar nicht freigegeben war, ist kein Fehler.** Das Ziel ist
// erreicht — und wer einen Entzug zweimal anstößt, soll nicht beim zweiten Mal
// eine Meldung bekommen.
func (k *Kontaktablage) Entziehen(
	ctx context.Context, kennung string, zweck domain.Zweck,
) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	return k.inTransaktion(ctx, func(tx *sql.Tx) error {
		id, _, fehler := kontaktZeile(ctx, tx, profilID, kennung)
		if fehler != nil {
			return fehler
		}

		if _, err := tx.ExecContext(ctx,
			`DELETE FROM peerrechte WHERE kontakt_id = ? AND zweck = ?`,
			id, string(zweck)); err != nil {
			return uebersetze(err, "einen Zweck entziehen")
		}

		return nil
	})
}

// AllesZurueck nimmt jede Bestätigung und jedes Recht dieses Profils zurück.
//
// **Für die Rotation.** Die Abnahme des Issues verlangt sie wörtlich:
// „Rotation erzwingt Neuprüfung bekannter Kontakte." Wer rotiert, hat für
// jeden Kontakt einen anderen Fingerabdruck als den, den er einmal vorgelesen
// hat.
//
// **Eine Sperre bleibt.** Wer gesperrt hat, hat entschieden; eine Rotation ist
// kein Grund, jemanden wieder hereinzulassen.
func (k *Kontaktablage) AllesZurueck(ctx context.Context) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	return k.inTransaktion(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
DELETE FROM peerrechte
 WHERE kontakt_id IN (SELECT id FROM peerkontakte WHERE profil_id = ?)`,
			profilID); err != nil {
			return uebersetze(err, "die Rechte zurücknehmen")
		}

		if _, err := tx.ExecContext(ctx, `
UPDATE peerkontakte
   SET stand = ?, fassung = fassung + 1
 WHERE profil_id = ? AND stand = ?`,
			string(domain.KontaktUnbestaetigt), profilID,
			string(domain.KontaktBestaetigt)); err != nil {
			return uebersetze(err, "die Bestätigungen zurücknehmen")
		}

		return nil
	})
}

// rechte liest die Freigaben eines Kontakts.
func (k *Kontaktablage) rechte(
	ctx context.Context, id int64,
) ([]domain.Zweck, error) {
	zeilen, err := k.db.QueryContext(ctx,
		`SELECT zweck FROM peerrechte WHERE kontakt_id = ? ORDER BY zweck`,
		id)
	if err != nil {
		return nil, uebersetze(err, "die Rechte eines Kontakts lesen")
	}

	defer func() { _ = zeilen.Close() }()

	rechte := make([]domain.Zweck, 0)

	for zeilen.Next() {
		var zweck string

		if err := zeilen.Scan(&zweck); err != nil {
			return nil, uebersetze(err, "ein Recht lesen")
		}

		rechte = append(rechte, domain.Zweck(zweck))
	}

	if err := zeilen.Err(); err != nil {
		return nil, uebersetze(err, "die Rechte eines Kontakts lesen")
	}

	return rechte, nil
}

// kontaktZeile gibt Zeilenkennung und Stand eines Kontakts dieses Profils.
//
// # Warum der Stand hier mitkommt
//
// **Weil ein zweiter Blick ein zweiter Stand wäre.** [Kontaktablage.Erlauben]
// muss wissen, ob das Gegenüber bestätigt ist — und zwar in **derselben**
// Transaktion, in der es schreibt. Eine eigene Abfrage dafür wäre nicht nur
// eine mehr, sondern eine zweite Gelegenheit, etwas anderes zu lesen.
func kontaktZeile(
	ctx context.Context, tx *sql.Tx, profilID int64, kennung string,
) (int64, domain.Kontaktstand, error) {
	var (
		id    int64
		stand string
	)

	err := tx.QueryRowContext(ctx,
		`SELECT id, stand FROM peerkontakte WHERE profil_id = ? AND kennung = ?`,
		profilID, kennung).Scan(&id, &stand)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return 0, "", fmt.Errorf(
			"%w: diesem Profil ist %s noch nicht begegnet",
			domain.ErrNotFound, kennung)
	case err != nil:
		return 0, "", uebersetze(err, "einen Kontakt suchen")
	}

	return id, domain.Kontaktstand(stand), nil
}
