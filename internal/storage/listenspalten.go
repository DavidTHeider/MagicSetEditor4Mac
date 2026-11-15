package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Spaltenwahl und die Sortierung eines Spiels** — Werkstück 3 von
// Phase 3.7
// (#1168).

// Listenablage liest und schreibt, was die Kartenliste zeigt.
type Listenablage struct{ db *sql.DB }

// NeueListenablage gibt die Ablage über diesem Bestand.
func NeueListenablage(db *sql.DB) *Listenablage { return &Listenablage{db: db} }

// Spalten gibt die Wahl eines Spiels — **nach ihrer Stelle geordnet**.
//
// **Geordnet vom Bestand und nicht vom Aufrufer**: sonst sortierte jede
// Fläche für sich, und zwei von ihnen liefen auseinander.
//
// **Eine leere Liste heißt „niemand hat gewählt"** — dann gilt, was das
// Spielpaket vorschlägt. Das ist etwas anderes als „alles abgewählt", und der
// Unterschied trägt: jenes wäre eine Liste ohne Spalten.
func (a *Listenablage) Spalten(ctx context.Context, spiel string) (
	[]domain.Spaltenwahl, error,
) {
	if err := pruefeSpiel(spiel); err != nil {
		return nil, err
	}

	// **Jede Wurzelabfrage klemmt auf ein Profil** (#1351, Vermerk 0066).
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, err := a.db.QueryContext(ctx,
		`SELECT feld, sichtbar, stelle FROM spiel_listenspalten
		  WHERE spiel = ? AND profil_id = ?
		  ORDER BY stelle, feld`, spiel, profilID)
	if err != nil {
		return nil, uebersetze(err, "die Spaltenwahl lesen")
	}

	defer func() { _ = zeilen.Close() }()

	wahl := []domain.Spaltenwahl{}

	for zeilen.Next() {
		var eine domain.Spaltenwahl
		if err := zeilen.Scan(&eine.Feld, &eine.Sichtbar, &eine.Stelle); err != nil {
			return nil, uebersetze(err, "eine Spaltenzeile lesen")
		}

		wahl = append(wahl, eine)
	}

	return wahl, uebersetze(zeilen.Err(), "die Spaltenwahl lesen")
}

// SpaltenSetzen ersetzt die Wahl eines Spiels **vollständig**.
//
// **Vollständig und nicht zusammengeführt**: wer eine Spalte abwählt, nimmt
// sie weg. Ein `INSERT … ON CONFLICT` allein ließe sie stehen, und niemand
// könnte je eine Spalte loswerden.
func (a *Listenablage) SpaltenSetzen(
	ctx context.Context, spiel string, wahl []domain.Spaltenwahl,
) error {
	if err := pruefeSpiel(spiel); err != nil {
		return err
	}

	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	transaktion, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "die Spaltenwahl beginnen")
	}

	defer func() { _ = transaktion.Rollback() }()

	if _, err := transaktion.ExecContext(ctx,
		`DELETE FROM spiel_listenspalten WHERE spiel = ? AND profil_id = ?`,
		spiel, profilID); err != nil {
		return uebersetze(err, "die alte Spaltenwahl räumen")
	}

	for _, eine := range wahl {
		if _, err := transaktion.ExecContext(ctx,
			`INSERT INTO spiel_listenspalten (spiel, feld, sichtbar, stelle, profil_id)
			 VALUES (?, ?, ?, ?, ?)`,
			spiel, eine.Feld, eine.Sichtbar, eine.Stelle, profilID); err != nil {
			return uebersetze(err, "die Spalte %q schreiben", eine.Feld)
		}
	}

	return uebersetze(transaktion.Commit(), "die Spaltenwahl abschließen")
}

// Sortierung gibt, wonach ein Spiel ordnet — ohne Wahl die Vorgabe.
func (a *Listenablage) Sortierung(ctx context.Context, spiel string) (
	domain.Kartenordnung, error,
) {
	if err := pruefeSpiel(spiel); err != nil {
		return domain.Kartenordnung{}, err
	}

	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.Kartenordnung{}, profilfehler
	}

	ordnung := domain.StandardKartenordnung()

	err := a.db.QueryRowContext(ctx,
		`SELECT feld, aufsteigend FROM spiel_sortierung
		  WHERE spiel = ? AND profil_id = ?`, spiel, profilID).
		Scan(&ordnung.Feld, &ordnung.Aufsteigend)

	// **Keine Zeile ist keine Fehlermeldung**, sondern die Vorgabe: kein Feld,
	// aufsteigend.
	if errors.Is(err, sql.ErrNoRows) {
		return domain.StandardKartenordnung(), nil
	}

	if err != nil {
		return domain.Kartenordnung{}, uebersetze(err, "die Sortierung lesen")
	}

	return ordnung, nil
}

// SortierungSetzen schreibt sie fort.
func (a *Listenablage) SortierungSetzen(
	ctx context.Context, spiel string, ordnung domain.Kartenordnung,
) error {
	if err := pruefeSpiel(spiel); err != nil {
		return err
	}

	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	// **Der Konflikt nennt jetzt beide Spalten** — der Schlüssel ist
	// `(profil_id, spiel)`, und ein `ON CONFLICT(spiel)` träfe ihn nicht mehr.
	_, err := a.db.ExecContext(ctx,
		`INSERT INTO spiel_sortierung (spiel, feld, aufsteigend, profil_id)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(profil_id, spiel)
		   DO UPDATE SET feld = excluded.feld,
		                 aufsteigend = excluded.aufsteigend`,
		spiel, ordnung.Feld, ordnung.Aufsteigend, profilID)

	return uebersetze(err, "die Sortierung schreiben")
}

// pruefeSpiel weist eine Wahl ohne Spiel ab — sie gehörte niemandem.
func pruefeSpiel(spiel string) error {
	if strings.TrimSpace(spiel) == "" {
		return fmt.Errorf("%w: ohne Spiel gehört die Wahl niemandem",
			domain.ErrValidation)
	}

	return nil
}
