package storage

// **Der Vergleich mit der Vorschau gehört ins Schreibgeschäft** — Befund 9 des
// Nachlaufs zum Zweitblick über Phase 5.2 (#1435).
//
// Er lag davor: `service.Ruecksetzung` zählte, verglich und rief dann das
// Löschen. Zwischen dem Zählen und dem Löschen liegt eine Lücke, und wer sie
// trifft, verliert etwas, das in keiner Vorschau stand — dieselbe Klasse wie
// Befund 11 des Zweitblicks, wo die Revision der Einstellungen aus demselben
// Grund ins Geschäft gezogen wurde.
//
// **Hier zählt derselbe Zähler noch einmal**, auf der Verbindung des
// Geschäfts, und der Aufrufer entscheidet, ob das noch sein Stand ist.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gebundenerUmfang zählt im Geschäft nach und legt es dem Aufrufer vor.
//
// **`bindung` ist wahlfrei, und das ist kein Schlupfloch**: die Wege, die ohne
// sie rufen, löschen einen benannten Datensatz ohne Vorschau — dort gibt es
// keinen Stand, gegen den zu binden wäre. Wer eine Vorschau gezeigt hat,
// reicht sie durch; das ist die Zusage, und `service` löst sie ein.
//
// **`profile` nennt die Kennungen, deren Umfang zusammengenommen den Stand
// ergibt** — bei einem Profil eine, bei einem Menschen alle seine. Gezählt
// wird je Profil und **nicht** zusammengerechnet: was eine Gruppe ist und wie
// zwei sich addieren, weiß `service` (§2).
func gebundenerUmfang(
	ctx context.Context, tx *sql.Tx, kennungen string, wovon int64,
	bindung []func(profile []domain.Profil, umfaenge [][]domain.Betroffenes) error,
	zusaetze ...domain.Betroffenes,
) error {
	if len(bindung) == 0 || bindung[0] == nil {
		return nil
	}

	profile, lesefehler := profileImGeschaeft(ctx, tx, kennungen, wovon)
	if lesefehler != nil {
		return lesefehler
	}

	umfaenge := make([][]domain.Betroffenes, 0, len(profile))

	for _, profil := range profile {
		umfang, umfangfehler := Profilumfang(ctx, tx, profil.ID)
		if umfangfehler != nil {
			return umfangfehler
		}

		umfaenge = append(umfaenge, umfang)
	}

	// **Was nicht in einem der Profile liegt, geht als eigener Umfang mit** —
	// die fremden Zugänge eines Menschen (#2287). Leere Gruppen nicht, wie im
	// [Profilumfang]: die Vorschau rechnet dieselbe Regel.
	for _, zusatz := range zusaetze {
		if zusatz.Anzahl > 0 {
			umfaenge = append(umfaenge, []domain.Betroffenes{zusatz})
		}
	}

	return bindung[0](profile, umfaenge)
}

// profileImGeschaeft liest Kennung und Namen der betroffenen Profile.
func profileImGeschaeft(
	ctx context.Context, tx *sql.Tx, kennungen string, wovon int64,
) ([]domain.Profil, error) {
	zeilen, fehler := tx.QueryContext(ctx, kennungen, wovon)
	if fehler != nil {
		return nil, uebersetze(fehler, "die betroffenen Profile lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird über Err()

	profile := []domain.Profil{}

	for zeilen.Next() {
		var profil domain.Profil
		if lesefehler := zeilen.Scan(&profil.ID, &profil.Name); lesefehler != nil {
			return nil, uebersetze(lesefehler, "eine Profilkennung lesen")
		}

		profile = append(profile, profil)
	}

	if schlussfehler := zeilen.Err(); schlussfehler != nil {
		return nil, uebersetze(schlussfehler, "die betroffenen Profile lesen")
	}

	return profile, nil
}
