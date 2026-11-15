// **Freigaben im Bestand** — Werkstück 11 von Phase 4.12
// (#1351).

package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Freigabenablage führt die Freigaben an Profilen.
type Freigabenablage struct{ db *sql.DB }

// NeueFreigabenablage baut sie über einer Datenquelle.
//
// **Ohne einsetzbare Uhr, anders als die Tokenverwaltung.** Dort entscheidet
// ein Zeitpunkt, **ob** etwas gilt (der Ablauf), und ein Lauf muss ihn
// festlegen können. Hier ist er eine Auskunft — „seit wann hat dieser Mensch
// Zugang" —, und eine Naht dafür wäre eine ohne Aufrufer (§10).
func NeueFreigabenablage(db *sql.DB) *Freigabenablage {
	return &Freigabenablage{db: db}
}

// freigabenspalten sind die Spalten, die [domain.Profilfreigabe] füllt.
const freigabenspalten = `profil_id, nutzer_id, rolle, angelegt`

// Freigaben gibt die Freigaben **eines Profils**.
func (a *Freigabenablage) Freigaben(ctx context.Context, profilID int64) (
	[]domain.Profilfreigabe, error,
) {
	return a.lies(ctx,
		`SELECT `+freigabenspalten+` FROM profilfreigaben
		 WHERE profil_id = ? ORDER BY nutzer_id`, profilID)
}

// FreigabenFuer gibt die Freigaben **eines Nutzers** — über alle Profile.
//
// **Die andere Richtung derselben Tabelle.** Die Übersicht fragt so: „welche
// fremden Profile darf ich sehen?"
func (a *Freigabenablage) FreigabenFuer(ctx context.Context, nutzerID int64) (
	[]domain.Profilfreigabe, error,
) {
	return a.lies(ctx,
		`SELECT `+freigabenspalten+` FROM profilfreigaben
		 WHERE nutzer_id = ? ORDER BY profil_id`, nutzerID)
}

// lies führt eine der beiden Abfragen aus.
func (a *Freigabenablage) lies(ctx context.Context, abfrage string, kennung int64) (
	[]domain.Profilfreigabe, error,
) {
	zeilen, err := a.db.QueryContext(ctx, abfrage, kennung)
	if err != nil {
		return nil, uebersetze(err, "die Freigaben lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // nur gelesen

	gefunden := []domain.Profilfreigabe{}

	for zeilen.Next() {
		var eine domain.Profilfreigabe

		var rolle string

		if err := zeilen.Scan(&eine.ProfilID, &eine.NutzerID, &rolle,
			&eine.Angelegt); err != nil {
			return nil, uebersetze(err, "eine Freigabezeile lesen")
		}

		eine.Rolle = domain.Rolle(rolle)
		gefunden = append(gefunden, eine)
	}

	if err := zeilen.Err(); err != nil {
		return nil, uebersetze(err, "die Freigaben durchgehen")
	}

	return gefunden, nil
}

// FreigabeSetzen legt eine Freigabe an oder schreibt ihre Rolle fort.
//
// **Ein Weg für beides.** „Gib diesem Nutzer die Rolle X" ist eine Aussage
// über den Endzustand; zwei Verfahren dafür zwängen jeden Aufrufer, vorher zu
// fragen, ob es schon eine gibt.
func (a *Freigabenablage) FreigabeSetzen(
	ctx context.Context, freigabe domain.Profilfreigabe,
) error {
	if fehler := freigabe.Validate(); fehler != nil {
		return fehler
	}

	angelegt := time.Now().UTC().Format(time.RFC3339)

	// **`ON CONFLICT` und nicht erst fragen**: zwischen Frage und Antwort
	// könnte jemand anders schreiben, und dann wäre die Antwort falsch.
	//
	// **`angelegt` bleibt beim Fortschreiben stehen** — es sagt, seit wann
	// dieser Mensch Zugang hat, und eine Rollenänderung fängt das nicht neu an.
	if _, err := a.db.ExecContext(ctx,
		`INSERT INTO profilfreigaben (profil_id, nutzer_id, rolle, angelegt)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT (profil_id, nutzer_id) DO UPDATE SET rolle = excluded.rolle`,
		freigabe.ProfilID, freigabe.NutzerID, string(freigabe.Rolle),
		angelegt); err != nil {
		return uebersetze(err, "die Freigabe setzen")
	}

	return nil
}

// FreigabeLoeschen zieht eine Freigabe zurück.
func (a *Freigabenablage) FreigabeLoeschen(
	ctx context.Context, profilID, nutzerID int64,
) error {
	ergebnis, err := a.db.ExecContext(ctx,
		`DELETE FROM profilfreigaben WHERE profil_id = ? AND nutzer_id = ?`,
		profilID, nutzerID)
	if err != nil {
		return uebersetze(err, "die Freigabe zurückziehen")
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return zaehlfehler
	}

	// **Null Zeilen sind ein Nein und kein stilles Ja.** Wer eine Freigabe
	// zurückzieht, die es nicht gibt, hat sich vertan — und ein „gut" ließe
	// ihn glauben, er habe jemandem den Zugang genommen (§8).
	if zeilen == 0 {
		return fmt.Errorf(
			"%w: für den Nutzer %d gibt es an diesem Profil keine Freigabe",
			domain.ErrNotFound, nutzerID)
	}

	return nil
}
