package storage

// **Die Ablage der Drucke** — Werkstück A2 von Phase 5.3
// (#1441), 0086.
//
// # Sichern ist ein Wiedersehen
//
// Wer einen Druck sichert, der über eine **eindeutige** Kennung schon
// daliegt, bekommt dessen Zeile — Name, Setcode und Nummer werden
// nachgetragen, wo sie fehlten, und neue Kennungen hängen sich an. Eine
// eindeutige Kennung, die schon **einem anderen** Druck gehört, ist ein
// Konflikt (`ErrConflict`): kein stilles Umhängen, kein Automerge.
//
// # Warum der Konflikt vor dem Einfügen geprüft wird
//
// Der teilweise eindeutige Index (`druckkennungen_eindeutig`) würde ihn auch
// melden — aber als nackte `UNIQUE`-Verletzung, ohne zu sagen, **welcher**
// Druck die Kennung schon trägt. Der Satz nennt beide.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Drucke ist die Ablage der Drucke und ihrer Kennungen.
type Drucke struct{ db *sql.DB }

// NeueDrucke baut die Ablage über einer Datenbank.
func NeueDrucke(db *sql.DB) *Drucke { return &Drucke{db: db} }

// Sichern legt einen Druck ab oder findet ihn über eine Kennung wieder.
func (d *Drucke) Sichern(ctx context.Context, druck *domain.Druck) error {
	druck.Normalisieren()

	if befunde := druck.Validate(); len(befunde) > 0 {
		return befunde[0]
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "den Druck sichern")
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach einem Commit ist sie wirkungslos

	if druck.ID == 0 {
		if err := d.wiedersehen(ctx, tx, druck); err != nil {
			return err
		}
	}

	if err := d.zeileSchreiben(ctx, tx, druck); err != nil {
		return err
	}

	for _, kennung := range druck.Kennungen {
		if err := d.kennungAnhaengen(ctx, tx, druck.ID, kennung); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return uebersetze(err, "den Druck sichern")
	}

	return nil
}

// wiedersehen setzt die Kennung des Drucks, wenn eine eindeutige Kennung
// schon daliegt.
func (d *Drucke) wiedersehen(ctx context.Context, tx *sql.Tx, druck *domain.Druck) error {
	for _, kennung := range druck.Kennungen {
		if !kennung.Eindeutig() {
			continue
		}

		id, gefunden, err := besitzer(ctx, tx, kennung)
		if err != nil {
			return err
		}

		if gefunden {
			druck.ID = id

			return nil
		}
	}

	return nil
}

// besitzer sagt, welcher Druck eine eindeutige Kennung trägt.
func besitzer(ctx context.Context, q abfrager, kennung domain.Druckkennung) (id int64, gefunden bool, err error) {
	err = q.QueryRowContext(ctx,
		`SELECT druck_id FROM druckkennungen WHERE system = ? AND semantik = ? AND wert = ?`,
		string(kennung.System), string(domain.SemantikDruck), kennung.Wert).Scan(&id)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return 0, false, nil
	case err != nil:
		return 0, false, uebersetze(err, "die Kennung %s/%s nachschlagen", kennung.System, kennung.Wert)
	}

	return id, true, nil
}

// zeileSchreiben legt die Zeile an oder trägt nach, was fehlte.
func (d *Drucke) zeileSchreiben(ctx context.Context, tx *sql.Tx, druck *domain.Druck) error {
	if druck.ID == 0 {
		ergebnis, err := tx.ExecContext(ctx,
			`INSERT INTO drucke (name, setcode, nummer, sprache) VALUES (?, ?, ?, ?)`,
			druck.Name, druck.Setcode, druck.Nummer, druck.Sprache)
		if err != nil {
			return uebersetze(err, "den Druck anlegen")
		}

		id, err := ergebnis.LastInsertId()
		if err != nil {
			return uebersetze(err, "die Kennung des Drucks lesen")
		}

		druck.ID = id

		return nil
	}

	// **Was der Aufrufer mitbringt, gewinnt; was er leer lässt, bleibt.**
	if _, err := tx.ExecContext(ctx,
		`UPDATE drucke SET
		   name = COALESCE(NULLIF(?, ''), name), setcode = COALESCE(NULLIF(?, ''), setcode),
		   nummer = COALESCE(NULLIF(?, ''), nummer), sprache = COALESCE(NULLIF(?, ''), sprache)
		 WHERE id = ?`,
		druck.Name, druck.Setcode, druck.Nummer, druck.Sprache, druck.ID); err != nil {
		return uebersetze(err, "den Druck %d fortschreiben", druck.ID)
	}

	return nil
}

// kennungAnhaengen hängt eine Kennung an — und weist eine fremde eindeutige ab.
func (d *Drucke) kennungAnhaengen(
	ctx context.Context, tx *sql.Tx, druckID int64, kennung domain.Druckkennung,
) error {
	if kennung.Eindeutig() {
		andere, gefunden, err := besitzer(ctx, tx, kennung)
		if err != nil {
			return err
		}

		if gefunden && andere != druckID {
			return fmt.Errorf("%w: die Kennung %s/%s gehört schon dem Druck %d, nicht %d — "+
				"ein Druck je Kennung, und ein Widerspruch ist ein Befund, kein Umhängen",
				domain.ErrConflict, kennung.System, kennung.Wert, andere, druckID)
		}
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO druckkennungen (druck_id, system, semantik, wert) VALUES (?, ?, ?, ?)
		 ON CONFLICT (druck_id, system, semantik, wert) DO NOTHING`,
		druckID, string(kennung.System), string(kennung.Semantik), kennung.Wert); err != nil {
		return uebersetze(err, "die Kennung %s/%s anhängen", kennung.System, kennung.Wert)
	}

	return nil
}

// Fuer liest einen Druck samt Kennungen.
func (d *Drucke) Fuer(ctx context.Context, id int64) (*domain.Druck, error) {
	drucke, err := d.lesen(ctx, druckspalten+` WHERE d.id = ? ORDER BY d.id`, id)
	if err != nil {
		return nil, err
	}

	if len(drucke) == 0 {
		return nil, fmt.Errorf("%w: den Druck %d gibt es nicht", domain.ErrNotFound, id)
	}

	return &drucke[0], nil
}

// NachKennung gibt jeden Druck, der die Kennung trägt — einen bei »druck«,
// mehrere bei einem Konzept oder Produkt.
func (d *Drucke) NachKennung(ctx context.Context, kennung domain.Druckkennung) ([]domain.Druck, error) {
	kennung.Normalisieren()

	return d.lesen(ctx,
		druckspalten+`
		  WHERE d.id IN (SELECT druck_id FROM druckkennungen
		                  WHERE system = ? AND semantik = ? AND wert = ?)
		  ORDER BY d.id`,
		string(kennung.System), string(kennung.Semantik), kennung.Wert)
}

// NachScryfall gibt den Druck zu einem Scryfall-Alias — oder nil.
func (d *Drucke) NachScryfall(ctx context.Context, printing string) (*domain.Druck, error) {
	drucke, err := d.NachKennung(ctx, domain.Druckkennung{
		System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: printing,
	})
	if err != nil || len(drucke) == 0 {
		return nil, err
	}

	return &drucke[0], nil
}

// Kandidaten gibt die Drucke zu Name, Setcode und Nummer — in jeder Sprache.
func (d *Drucke) Kandidaten(ctx context.Context, name, setcode, nummer string) ([]domain.Druck, error) {
	return d.lesen(ctx,
		druckspalten+`
		  WHERE d.name = ? COLLATE NOCASE AND d.setcode = ? COLLATE NOCASE
		    AND d.nummer = ? COLLATE NOCASE
		  ORDER BY d.id`,
		strings.TrimSpace(name), strings.ToLower(strings.TrimSpace(setcode)), strings.TrimSpace(nummer))
}

// Alle gibt jeden Druck — für Sichten und Läufe.
func (d *Drucke) Alle(ctx context.Context) ([]domain.Druck, error) {
	return d.lesen(ctx, druckspalten+` ORDER BY d.id`)
}

// druckspalten ist der Anfang jeder Abfrage über Drucke.
//
// **Jeder Aufrufer schreibt seine Abfrage ganz**: eine Bedingung, die erst in
// `lesen` angehängt würde, wäre eine Verkettung um eine Zeichenkette dieses
// Pakets — harmlos, und gosec kann das nicht wissen (G202). Eine
// Unterdrückung wäre eine Behauptung, die niemand mehr prüft; eine
// vollständige Abfrage je Weg ist keine.
const druckspalten = `SELECT d.id, d.name, d.setcode, d.nummer, d.sprache FROM drucke d`

// lesen holt Drucke samt Kennungen in der Reihenfolge ihrer Anlage.
func (d *Drucke) lesen(ctx context.Context, abfrage string, werte ...any) ([]domain.Druck, error) {
	zeilen, err := d.db.QueryContext(ctx, abfrage, werte...)
	if err != nil {
		return nil, uebersetze(err, "die Drucke lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird über Err()

	var drucke []domain.Druck

	for zeilen.Next() {
		var druck domain.Druck
		if err := zeilen.Scan(&druck.ID, &druck.Name, &druck.Setcode, &druck.Nummer, &druck.Sprache); err != nil {
			return nil, uebersetze(err, "einen Druck lesen")
		}

		drucke = append(drucke, druck)
	}

	if err := zeilen.Err(); err != nil {
		return nil, uebersetze(err, "die Drucke lesen")
	}

	for i := range drucke {
		kennungen, err := d.kennungen(ctx, drucke[i].ID)
		if err != nil {
			return nil, err
		}

		drucke[i].Kennungen = kennungen
	}

	return drucke, nil
}

// kennungen liest die Kennungen eines Drucks in der Reihenfolge ihrer Aufnahme.
func (d *Drucke) kennungen(ctx context.Context, druckID int64) ([]domain.Druckkennung, error) {
	zeilen, err := d.db.QueryContext(ctx,
		`SELECT system, semantik, wert FROM druckkennungen WHERE druck_id = ? ORDER BY rowid`, druckID)
	if err != nil {
		return nil, uebersetze(err, "die Kennungen des Drucks %d lesen", druckID)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird über Err()

	var kennungen []domain.Druckkennung

	for zeilen.Next() {
		var system, semantik, wert string
		if err := zeilen.Scan(&system, &semantik, &wert); err != nil {
			return nil, uebersetze(err, "eine Kennung lesen")
		}

		kennungen = append(kennungen, domain.Druckkennung{
			System: domain.Kennungssystem(system), Semantik: domain.Semantik(semantik), Wert: wert,
		})
	}

	if err := zeilen.Err(); err != nil {
		return nil, uebersetze(err, "die Kennungen des Drucks %d lesen", druckID)
	}

	return kennungen, nil
}
