package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Kartenverweise ist die Ablage der Verweise auf echte Drucke
// (#1167).
//
// **Eine eigene Ablage und kein Feld am Kartenschreiber**: ein Verweis ist
// kein Feldwert, geht in keine `.mse-set` und steht in keinem Verlauf der
// Feldänderungen. Wer ihn dort führte, schriebe ihn in jede weitergegebene
// Datei — und verlöre ihn beim Rundlauf durch MSE2.
type Kartenverweise struct {
	db *sql.DB
}

// NeueKartenverweise gibt die Ablage über dieser Datenbank.
func NeueKartenverweise(db *sql.DB) *Kartenverweise {
	return &Kartenverweise{db: db}
}

// Setzen legt den Verweis einer Karte an oder ersetzt ihn.
//
// **Ersetzen und nicht danebenlegen**: eine Karte zeigt auf einen Druck.
func (k *Kartenverweise) Setzen(
	ctx context.Context, karteID int64, verweis domain.Kartenverweis,
) error {
	verweis.Normalisieren()

	// **In einem Zug mit der Zeile des Drucks** (A2, #1441).
	tx, err := k.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "den Verweis der Karte %d setzen", karteID)
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach einem Commit ist sie wirkungslos

	// **Nur an einer Karte des gezeigten Profils** (#1815).
	if fehler := karteImProfil(ctx, tx, karteID); fehler != nil {
		return fehler
	}

	druck, druckfehler := druckZeile(ctx, tx, druckbezug{
		druckID: verweis.DruckID, printing: verweis.Printing,
		name: verweis.Name, setcode: verweis.Setcode, nummer: verweis.Nummer,
	})
	if druckfehler != nil {
		return druckfehler
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO karten_verweis (card_id, printing, druck_id, name, setcode, adresse, nummer, gesetzt)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT (card_id) DO UPDATE SET
		   printing = excluded.printing, druck_id = excluded.druck_id,
		   name     = excluded.name,     setcode  = excluded.setcode,
		   nummer   = excluded.nummer,   adresse  = excluded.adresse,
		   gesetzt  = excluded.gesetzt`,
		karteID, verweis.Printing, druck, verweis.Name, verweis.Setcode,
		verweis.Adresse, verweis.Nummer, verweis.Gesetzt.Unix()); err != nil {
		return uebersetze(err, "den Verweis der Karte %d setzen", karteID)
	}

	// **Ein Satz und kein Zweig** — dieselbe Form wie in `decks.go`.
	return uebersetze(tx.Commit(), "den Verweis der Karte %d setzen", karteID)
}

// Fuer gibt den Verweis einer Karte — `nil`, wenn sie keinen hat.
//
// **`nil` und kein Fehler**: die meisten Karten haben keinen. Ein Fehler
// dafür machte den Normalfall zur Ausnahme.
func (k *Kartenverweise) Fuer(ctx context.Context, karteID int64) (
	*domain.Kartenverweis, error,
) {
	if fehler := karteImProfil(ctx, k.db, karteID); fehler != nil {
		return nil, fehler
	}

	var (
		verweis domain.Kartenverweis
		druck   sql.NullInt64
		gesetzt int64
	)

	err := k.db.QueryRowContext(ctx,
		`SELECT printing, druck_id, name, setcode, nummer, adresse, gesetzt
		   FROM karten_verweis WHERE card_id = ?`, karteID).
		Scan(&verweis.Printing, &druck, &verweis.Name, &verweis.Setcode,
			&verweis.Nummer, &verweis.Adresse, &gesetzt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil //nolint:nilnil // „keiner" ist kein Fehler, siehe oben
	}

	if err != nil {
		return nil, uebersetze(err, "den Verweis der Karte %d lesen", karteID)
	}

	verweis.Gesetzt = time.Unix(gesetzt, 0).UTC()
	verweis.DruckID = druck.Int64

	return &verweis, nil
}

// Loesen entfernt den Verweis einer Karte und sagt, ob es einen gab.
//
// **Die Auskunft ist Pflicht**: ohne sie meldete eine Fläche Erfolg, wo nichts
// geschah — und niemand bemerkte, dass er die falsche Karte erwischt hat.
func (k *Kartenverweise) Loesen(ctx context.Context, karteID int64) (bool, error) {
	var gab bool

	fehler := imGeschaeft(ctx, k.db, "den Verweis lösen", func(tx *sql.Tx) error {
		if fehler := karteImProfil(ctx, tx, karteID); fehler != nil {
			return fehler
		}

		ergebnis, err := tx.ExecContext(ctx,
			`DELETE FROM karten_verweis WHERE card_id = ?`, karteID)
		if err != nil {
			return uebersetze(err, "den Verweis der Karte %d lösen", karteID)
		}

		betroffen, err := ergebnis.RowsAffected()
		if err != nil {
			return uebersetze(err, "den Verweis der Karte %d lösen", karteID)
		}

		gab = betroffen > 0

		return nil
	})

	return gab, fehler
}

// JeSet gibt die Verweise aller Karten eines Sets, nach Kartenkennung.
//
// **In einem Zug und nicht je Karte**: die Kartenliste zeigt sie an jeder
// Zeile, und ein Set hat leicht dreihundert.
//
// **Karten ohne Verweis stehen nicht darin.** Eine Karte mit leerem Verweis
// wäre von einer mit Verweis auf eine namenlose Karte nicht zu unterscheiden.
func (k *Kartenverweise) JeSet(ctx context.Context, setID int64) (
	map[int64]domain.Kartenverweis, error,
) {
	if fehler := setImProfil(ctx, k.db, setID); fehler != nil {
		return nil, fehler
	}

	zeilen, err := k.db.QueryContext(ctx,
		`SELECT karten_verweis.card_id, printing, druck_id, name, setcode, nummer, adresse, gesetzt
		   FROM karten_verweis
		   JOIN cards ON cards.id = karten_verweis.card_id
		  WHERE cards.set_id = ?`, setID)
	if err != nil {
		return nil, uebersetze(err, "die Verweise des Sets %d lesen", setID)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // der Fehler kommt über zeilen.Err

	gefunden := make(map[int64]domain.Kartenverweis)

	for zeilen.Next() {
		var (
			karteID int64
			verweis domain.Kartenverweis
			druck   sql.NullInt64
			gesetzt int64
		)

		if err := zeilen.Scan(&karteID, &verweis.Printing, &druck, &verweis.Name,
			&verweis.Setcode, &verweis.Nummer, &verweis.Adresse, &gesetzt); err != nil {
			return nil, uebersetze(err, "einen Verweis lesen")
		}

		verweis.DruckID = druck.Int64

		verweis.Gesetzt = time.Unix(gesetzt, 0).UTC()
		gefunden[karteID] = verweis
	}

	if err := zeilen.Err(); err != nil {
		return nil, uebersetze(err, "die Verweise des Sets %d lesen", setID)
	}

	return gefunden, nil
}
