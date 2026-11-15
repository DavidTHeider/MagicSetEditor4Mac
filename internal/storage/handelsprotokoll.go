package storage

// **Das Protokoll eines Handels lesen** — TRADE-03.AK-03
// (#1717).
// Geschrieben wird es von den Auslösern aus 0095, nicht von hier.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Protokoll gibt die Zustandswechsel eines Vorgangs, ältester zuerst.
func (a *Handelsablage) Protokoll(
	ctx context.Context, vorgang string,
) ([]domain.Handelsprotokollzeile, error) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return nil, fehler
	}

	zeilen, err := a.db.QueryContext(ctx,
		`SELECT revision, von, nach, zeit FROM handelsprotokoll
		  WHERE profil_id = ? AND vorgang = ?
		  ORDER BY id`, profilID, vorgang)
	if err != nil {
		return nil, uebersetze(err, "das Protokoll eines Handels lesen")
	}

	defer func() { _ = zeilen.Close() }()

	protokoll := make([]domain.Handelsprotokollzeile, 0)

	for zeilen.Next() {
		var (
			zeile         domain.Handelsprotokollzeile
			von, nach, zu string
		)

		if err := zeilen.Scan(&zeile.Revision, &von, &nach, &zu); err != nil {
			return nil, uebersetze(err, "eine Protokollzeile lesen")
		}

		zeit, fehler := time.Parse(time.RFC3339Nano, zu)
		if fehler != nil {
			return nil, domain.MitGrund(domain.ErrValidation, fehler,
				"den Zeitpunkt einer Protokollzeile lesen")
		}

		zeile.Von, zeile.Nach, zeile.Zeit = domain.Handelszustand(von),
			domain.Handelszustand(nach), zeit
		protokoll = append(protokoll, zeile)
	}

	return protokoll, uebersetze(zeilen.Err(), "das Protokoll eines Handels lesen")
}
