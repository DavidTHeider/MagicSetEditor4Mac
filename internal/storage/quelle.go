package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Quelle gibt die Datei, aus der ein Set gelesen wurde — oder nil.
//
// **Nil ist kein Fehler**: ein Set, das in der Anwendung entsteht, hat keine
// Quelle ([0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)).
func (l *SetLeser) Quelle(ctx context.Context, setID int64) (*domain.Quelle, error) {
	var quelle domain.Quelle

	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	err := l.db.QueryRowContext(ctx,
		`SELECT q.datei, q.inhalt, s.herkunft
		   FROM set_quellen q
		   JOIN sets s ON s.id = q.set_id AND s.profil_id = ?
		  WHERE q.set_id = ?`, profilID, setID).
		Scan(&quelle.Datei, &quelle.Inhalt, &quelle.Herkunft)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil //nolint:nilnil // „keine Quelle" ist kein Fehler, siehe oben

	case err != nil:
		return nil, uebersetze(err, "die Quelle des Sets %d lesen", setID)
	}

	return &quelle, nil
}

// HerkunftSetzen schreibt den Arbeitsort eines Sets.
//
// **Nach „Sichern unter"** (#226):
// bis dahin blieb der Ort der alte, und das folgende ⌘S ging wieder gegen die
// alte Herkunft.
//
// **Am Leser, und das ist die Ausnahme.** `SetLeser` liest; hier schreibt er
// **eine** Spalte, weil `Quelle` daneben steht und beide dieselbe Zeile meinen.
// Ein eigener Speicher für eine Spalte wäre eine zweite Verbindung auf
// dasselbe.
func (l *SetLeser) HerkunftSetzen(ctx context.Context, setID int64, ort string) error {
	// **Jede Wurzelabfrage klemmt auf ein Profil** (#1351, Vermerk 0066).
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	ergebnis, err := l.db.ExecContext(ctx,
		`UPDATE sets SET herkunft = ? WHERE id = ? AND profil_id = ?`,
		ort, setID, profilID)
	if err != nil {
		return uebersetze(err, "den Arbeitsort des Sets %d setzen", setID)
	}

	betroffen, zaehlfehler := ergebnis.RowsAffected()
	if zaehlfehler != nil {
		return uebersetze(zaehlfehler, "den Arbeitsort des Sets %d setzen", setID)
	}

	if betroffen == 0 {
		return fmt.Errorf("%w: das Set %d führt der Bestand nicht", domain.ErrNotFound, setID)
	}

	return nil
}
