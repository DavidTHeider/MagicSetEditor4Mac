package storage

// **Die Ablage gemessener Modellfähigkeiten** — Werkstück A9 von Phase 5.3
// (#1448).
//
// # Warum eine Messung liegen bleibt
//
// **Weil sie Geld kostet.** Eine Probe schickt einen Auftrag an einen
// Anbieter, der je Zeichen abrechnet; ein Ergebnis, das beim Schließen des
// Fensters verschwindet, müsste beim nächsten Start bezahlt werden
// ([0100](../../docs/entscheidungen/0100-was-openai-nicht-sagt.md)).
//
// # Je Profil und nicht je Bestand
//
// **Weil der Schlüssel je Profil liegt.** Wer mit einem anderen Zugang
// fragt, fragt vielleicht ein anderes Konto mit anderen Modellen — und eine
// Messung, die über Profile hinweg gälte, spräche dann über etwas anderes.

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Modellprobenablage hält die Messungen.
type Modellprobenablage struct{ db *sql.DB }

// NeueModellprobenablage baut sie über einer Datenbank.
func NeueModellprobenablage(db *sql.DB) *Modellprobenablage {
	return &Modellprobenablage{db: db}
}

// Merken schreibt eine Messung — und ersetzt die vorige.
//
// **Ersetzen und nicht anhängen**: was gestern galt, sagt über heute nichts,
// und zwei Daten nebeneinander ließen offen, welches gilt.
func (a *Modellprobenablage) Merken(
	ctx context.Context, profilID int64, probe domain.KIModellprobe,
) error {
	if befunde := probe.Validate(); len(befunde) > 0 {
		return fmt.Errorf("%w: die Messung ist unvollständig: %s",
			domain.ErrValidation, befunde[0].Error())
	}

	_, fehler := a.db.ExecContext(ctx, `
		INSERT INTO ki_modellproben
			(profil_id, anbieter, alias, faehigkeit, datum, beleg, laufzeit, modell)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (profil_id, anbieter, alias, faehigkeit) DO UPDATE SET
			datum = excluded.datum, beleg = excluded.beleg,
			laufzeit = excluded.laufzeit, modell = excluded.modell`,
		profilID, probe.Anbieter, probe.Alias, string(probe.Faehigkeit),
		probe.Datum, probe.Beleg, probe.Laufzeit, probe.Modell)
	if fehler != nil {
		return uebersetze(fehler, "die Messung merken")
	}

	return nil
}

// Alle gibt die Messungen eines Profils — in fester Reihenfolge.
//
// **Fest, damit zwei Läufe dieselbe Antwort geben**: der Katalog zeigt sie,
// und eine wechselnde Reihenfolge sähe aus wie eine Änderung.
func (a *Modellprobenablage) Alle(
	ctx context.Context, profilID int64,
) ([]domain.KIModellprobe, error) {
	zeilen, fehler := a.db.QueryContext(ctx, `
		SELECT anbieter, alias, faehigkeit, datum, beleg, laufzeit, modell
		FROM ki_modellproben WHERE profil_id = ?
		ORDER BY anbieter, alias, faehigkeit`, profilID)
	if fehler != nil {
		return nil, uebersetze(fehler, "die Messungen lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // nur lesend

	proben := make([]domain.KIModellprobe, 0)

	for zeilen.Next() {
		var probe domain.KIModellprobe

		var faehigkeit string

		if fehler := zeilen.Scan(&probe.Anbieter, &probe.Alias, &faehigkeit,
			&probe.Datum, &probe.Beleg, &probe.Laufzeit,
			&probe.Modell); fehler != nil {
			return nil, uebersetze(fehler, "eine Messung lesen")
		}

		probe.Faehigkeit = domain.KIFaehigkeit(faehigkeit)
		proben = append(proben, probe)
	}

	if fehler := zeilen.Err(); fehler != nil {
		return nil, uebersetze(fehler, "die Messungen lesen")
	}

	return proben, nil
}

// Vergessen nimmt eine Messung zurück.
//
// **Weil eine Messung veralten kann, ohne falsch zu werden.** Wer ein Modell
// neu prüfen will, soll die alte Auskunft wegnehmen können, statt sie
// stillschweigend überschrieben zu bekommen — und wer einen Anbieter
// abschaltet, will seine Messungen nicht behalten.
func (a *Modellprobenablage) Vergessen(
	ctx context.Context, profilID int64, anbieter, alias string,
) error {
	_, fehler := a.db.ExecContext(ctx, `
		DELETE FROM ki_modellproben
		WHERE profil_id = ? AND anbieter = ? AND alias = ?`,
		profilID, anbieter, alias)
	if fehler != nil {
		return uebersetze(fehler, "die Messung vergessen")
	}

	return nil
}
