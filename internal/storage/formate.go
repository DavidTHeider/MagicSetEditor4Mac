package storage

// **Der Formatspeicher** — Werkstück 4 von Phase 4.6
// (#1293),
// nach Vermerk 0061: „Daten, nicht Code".
//
// Er wird **leer** ausgeliefert. Warum, steht in der Wanderung 0029: die
// Namen der Formate lassen sich messen, ihre Zahlen nicht.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// formatspalten ist die Spaltenliste — **einmal**, damit Lesen und Auflisten
// nicht auseinanderlaufen.
const formatspalten = `id, fassung, name, spiel, mindestgroesse, hoechstgroesse,
	sideboardgroesse, sideboard_offen, hoechstzahl_gleich, scryfallname`

// Formatablage liest und schreibt Formate.
type Formatablage struct{ db *sql.DB }

// NeueFormatablage gibt die Ablage über diesem Bestand.
func NeueFormatablage(db *sql.DB) *Formatablage {
	return &Formatablage{db: db}
}

// Anlegen legt ein Format an und gibt es mit seiner Kennung zurück.
func (a *Formatablage) Anlegen(ctx context.Context, format domain.Format) (
	domain.Format, error,
) {
	// **Geprüft wird im Dienst** (§2); was der Bestand selbst hält, steht als
	// `CHECK` in der Wanderung.
	format.Normalisieren()

	// **Jede Wurzelabfrage klemmt auf ein Profil** (#1351, Vermerk 0066).
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.Format{}, profilfehler
	}

	ergebnis, err := a.db.ExecContext(ctx,
		`INSERT INTO formate (name, spiel, mindestgroesse, hoechstgroesse,
		                      sideboardgroesse, sideboard_offen,
		                      hoechstzahl_gleich, scryfallname, profil_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		format.Name, format.Spiel, format.Mindestgroesse, format.Hoechstgroesse,
		format.Sideboardgroesse, format.SideboardOffen,
		format.HoechstzahlGleich, format.Scryfallname, profilID)
	if err != nil {
		return domain.Format{}, uebersetze(err, "das Format anlegen")
	}

	kennung, idfehler := ergebnis.LastInsertId()

	format.ID = kennung
	format.Fassung = 1

	return format, uebersetze(idfehler, "die Kennung des Formats lesen")
}

// Format gibt ein Format.
func (a *Formatablage) Format(ctx context.Context, kennung int64) (
	domain.Format, error,
) {
	var format domain.Format

	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.Format{}, profilfehler
	}

	err := a.db.QueryRowContext(ctx,
		`SELECT `+formatspalten+` FROM formate WHERE id = ? AND profil_id = ?`,
		kennung, profilID).
		Scan(&format.ID, &format.Fassung, &format.Name, &format.Spiel,
			&format.Mindestgroesse, &format.Hoechstgroesse,
			&format.Sideboardgroesse, &format.SideboardOffen,
			&format.HoechstzahlGleich, &format.Scryfallname)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return domain.Format{}, fmt.Errorf(
			"%w: ein Format mit dieser Kennung gibt es nicht", domain.ErrNotFound)

	case err != nil:
		return domain.Format{}, uebersetze(err, "das Format lesen")
	}

	return format, nil
}

// Formate gibt die Formate eines Spiels — **ein leeres Spiel heißt: alle**.
//
// **Geordnet vom Bestand und nicht vom Aufrufer**: sonst sortierte jede Fläche
// für sich, und zwei von ihnen liefen auseinander.
func (a *Formatablage) Formate(ctx context.Context, spiel string) (
	[]domain.Format, error,
) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, err := a.db.QueryContext(ctx,
		`SELECT `+formatspalten+` FROM formate
		  WHERE profil_id = ? AND (? = '' OR spiel = ?)
		  ORDER BY spiel, name`, profilID, spiel, spiel)
	if err != nil {
		return nil, uebersetze(err, "die Formate lesen")
	}

	defer zeilen.Close() //nolint:errcheck // der Fehler steht in zeilen.Err()

	formate := []domain.Format{}

	for zeilen.Next() {
		var format domain.Format

		if err := zeilen.Scan(&format.ID, &format.Fassung, &format.Name,
			&format.Spiel, &format.Mindestgroesse, &format.Hoechstgroesse,
			&format.Sideboardgroesse, &format.SideboardOffen,
			&format.HoechstzahlGleich, &format.Scryfallname); err != nil {
			return nil, uebersetze(err, "ein Format lesen")
		}

		formate = append(formate, format)
	}

	return formate, uebersetze(zeilen.Err(), "die Formate lesen")
}

// Aendern schreibt ein Format fort — **gegen seine Fassung** (§7, Regel 2).
func (a *Formatablage) Aendern(ctx context.Context, format domain.Format) (
	domain.Format, error,
) {
	format.Normalisieren()

	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.Format{}, profilfehler
	}

	ergebnis, err := a.db.ExecContext(ctx,
		`UPDATE formate
		    SET name = ?, spiel = ?, mindestgroesse = ?, hoechstgroesse = ?,
		        sideboardgroesse = ?, sideboard_offen = ?,
		        hoechstzahl_gleich = ?, scryfallname = ?,
		        fassung = fassung + 1, geaendert = datetime('now')
		  WHERE id = ? AND fassung = ? AND profil_id = ?`,
		format.Name, format.Spiel, format.Mindestgroesse, format.Hoechstgroesse,
		format.Sideboardgroesse, format.SideboardOffen,
		format.HoechstzahlGleich, format.Scryfallname,
		format.ID, format.Fassung, profilID)
	if err != nil {
		return domain.Format{}, uebersetze(err, "das Format ändern")
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return domain.Format{}, zaehlfehler
	}

	if zeilen == 0 {
		return domain.Format{}, a.warumNicht(ctx, format.ID, profilID)
	}

	format.Fassung++

	return format, nil
}

// warumNicht sagt, ob das Format fehlt oder seine Fassung veraltet ist.
//
// **Das Profil kommt vom Rufer.** Es hier ein zweites Mal aus dem Kontext zu
// holen wäre ein Zweig, den kein Aufruf betreten kann: der Rufer hat es schon
// gelesen und wäre sonst gar nicht bis hierher gekommen (§3).
func (a *Formatablage) warumNicht(
	ctx context.Context, kennung, profilID int64,
) error {
	var vorhanden int64

	err := a.db.QueryRowContext(ctx,
		`SELECT count(*) FROM formate WHERE id = ? AND profil_id = ?`,
		kennung, profilID).Scan(&vorhanden)
	if err != nil {
		return uebersetze(err, "das Format nachschlagen")
	}

	if vorhanden == 0 {
		return fmt.Errorf("%w: das Format %d gibt es nicht", domain.ErrNotFound, kennung)
	}

	return fmt.Errorf(
		"%w — das Format %d hat sich seit dem Lesen geändert", domain.ErrUeberholt, kennung)
}

// Loeschen nimmt ein Format heraus — **und lässt die Decks stehen**.
//
// Das erledigt `ON DELETE SET NULL`; hier steht es, damit niemand es für einen
// Zufall hält. Ein Deck ohne Format ist **ungeprüft**, nicht verschwunden.
func (a *Formatablage) Loeschen(ctx context.Context, kennung int64) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	ergebnis, err := a.db.ExecContext(ctx,
		`DELETE FROM formate WHERE id = ? AND profil_id = ?`, kennung, profilID)
	if err != nil {
		return uebersetze(err, "das Format löschen")
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return zaehlfehler
	}

	if zeilen == 0 {
		return fmt.Errorf("%w: dieses Format gibt es nicht", domain.ErrNotFound)
	}

	return nil
}
