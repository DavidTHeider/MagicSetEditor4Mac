// **Anbieterkonten im Bestand** — Werkstück 4 von Phase 4.12
// (#1351).
//
// # Kein Profil an einem Konto
//
// **Ein Anbieterkonto gehört einem Nutzer.** Wer sich anmeldet, ist eine
// Person; in welchem seiner Profile er danach arbeitet, sagt der Zeiger. Diese
// Abfragen klemmen deshalb **nicht** auf ein Profil — und sie stehen mit
// diesem Satz in `internal/werkzeuge/profilklausel_test.go`.

package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// KontoVerbinden hängt eine fremde Kennung an einen Nutzer.
func (a *Nutzerablage) KontoVerbinden(ctx context.Context, neu domain.Anbieterkonto) (
	domain.Anbieterkonto, error,
) {
	if err := neu.Validate(); err != nil {
		return domain.Anbieterkonto{}, err
	}

	// **Den Nutzer gibt es, oder es ist ein Nein.** SQLite meldete sonst nur
	// einen Fremdschlüsselbruch, und wer das liest, sucht an der falschen
	// Stelle (§8) — dieselbe Form wie in [Nutzerablage.ProfilAnlegen].
	var da int

	if err := a.db.QueryRowContext(ctx,
		`SELECT count(*) FROM nutzer WHERE id = ?`, neu.NutzerID).Scan(&da); err != nil {
		return domain.Anbieterkonto{}, uebersetze(err, "den Nutzer suchen")
	}

	if da == 0 {
		return domain.Anbieterkonto{}, fmt.Errorf(
			"%w: den Nutzer %d führt der Bestand nicht", domain.ErrNotFound, neu.NutzerID)
	}

	ergebnis, err := a.db.ExecContext(ctx,
		`INSERT INTO anbieterkonten (nutzer_id, anbieter, subjekt, anzeigename)
		 VALUES (?, ?, ?, ?)`,
		neu.NutzerID, neu.Anbieter, neu.Subjekt, neu.Anzeigename)
	if err != nil {
		return domain.Anbieterkonto{}, uebersetze(err,
			"das Konto bei %s verbinden", neu.Anbieter)
	}

	kennung, kennungsfehler := ergebnis.LastInsertId()
	if kennungsfehler != nil {
		return domain.Anbieterkonto{}, uebersetze(kennungsfehler,
			"die Kennung des Kontos lesen")
	}

	neu.ID = kennung

	return neu, nil
}

// NutzerZuKonto gibt den Nutzer, dem diese fremde Kennung gehört.
//
// **Und „gibt es nicht" ist der Normalfall**, nicht die Panne: wer sich zum
// ersten Mal über einen Anbieter anmeldet, hat noch kein Konto. Der Aufrufer
// legt dann eines an — deshalb muss er die Fälle unterscheiden können (§8).
func (a *Nutzerablage) NutzerZuKonto(ctx context.Context, anbieter, subjekt string) (
	int64, error,
) {
	var nutzerID int64

	err := a.db.QueryRowContext(ctx,
		`SELECT nutzer_id FROM anbieterkonten WHERE anbieter = ? AND subjekt = ?`,
		anbieter, subjekt).Scan(&nutzerID)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return 0, fmt.Errorf("%w: zu dieser Kennung bei %s gibt es keinen Nutzer",
			domain.ErrNotFound, anbieter)

	case err != nil:
		return 0, uebersetze(err, "den Nutzer zur Kennung bei %s suchen", anbieter)
	}

	return nutzerID, nil
}

// Konten gibt die Anbieterkonten eines Nutzers.
func (a *Nutzerablage) Konten(ctx context.Context, nutzerID int64) (
	[]domain.Anbieterkonto, error,
) {
	zeilen, err := a.db.QueryContext(ctx,
		`SELECT id, nutzer_id, anbieter, subjekt, anzeigename
		   FROM anbieterkonten
		  WHERE nutzer_id = ?
		  ORDER BY anbieter`, nutzerID)
	if err != nil {
		return nil, uebersetze(err, "die Konten des Nutzers %d lesen", nutzerID)
	}

	defer func() { _ = zeilen.Close() }()

	konten := []domain.Anbieterkonto{}

	for zeilen.Next() {
		var eines domain.Anbieterkonto

		if err := zeilen.Scan(&eines.ID, &eines.NutzerID, &eines.Anbieter,
			&eines.Subjekt, &eines.Anzeigename); err != nil {
			return nil, uebersetze(err, "ein Konto lesen")
		}

		konten = append(konten, eines)
	}

	return konten, uebersetze(zeilen.Err(), "die Konten lesen")
}

// KontoLoesen nimmt eine fremde Kennung von einem Nutzer.
//
// # Der Rückweg gehört dazu
//
// Das Issue sagt es ausdrücklich: „ein Nutzer, dessen Anbieterkonto wegfällt,
// darf nicht ausgesperrt sein." Ohne diesen Weg wäre die Anmeldung eine
// Einbahnstraße — und wer sein Konto beim Anbieter schließt, käme an seine
// eigenen Daten nicht mehr heran.
func (a *Nutzerablage) KontoLoesen(ctx context.Context, nutzerID int64, anbieter string) error {
	ergebnis, err := a.db.ExecContext(ctx,
		`DELETE FROM anbieterkonten WHERE nutzer_id = ? AND anbieter = ?`,
		nutzerID, anbieter)
	if err != nil {
		return uebersetze(err, "das Konto bei %s lösen", anbieter)
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return zaehlfehler
	}

	if zeilen == 0 {
		return fmt.Errorf("%w: der Nutzer %d hat kein Konto bei %s",
			domain.ErrNotFound, nutzerID, anbieter)
	}

	return nil
}
