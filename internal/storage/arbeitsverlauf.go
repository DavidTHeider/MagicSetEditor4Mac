package storage

// **Die Ablage des Arbeitsverlaufs** — Werkstück U11 von Phase 5.1
// (#1422).
//
// Je Profil: was geöffnet, eingeführt, ausgegeben wurde, und was angeheftet
// ist. **Das Profil kommt aus dem Kontext** wie bei der Sammlung; ohne
// Profil gibt es keinen Verlauf — weder zum Lesen noch zum Schreiben.

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Arbeitsverlauf ist die Ablage der Schritte und des Angehefteten.
type Arbeitsverlauf struct{ db *sql.DB }

// NeuerArbeitsverlauf baut die Ablage über einer Datenbank.
func NeuerArbeitsverlauf(db *sql.DB) *Arbeitsverlauf { return &Arbeitsverlauf{db: db} }

// Vermerken hält einen Schritt fest — je Objekt und Vorgang der jüngste.
func (a *Arbeitsverlauf) Vermerken(ctx context.Context, schritt domain.Arbeitsschritt) error {
	if befunde := schritt.Validate(); len(befunde) > 0 {
		return befunde[0]
	}

	profilID, err := ProfilAus(ctx)
	if err != nil {
		return err
	}

	if _, err := a.db.ExecContext(ctx,
		`INSERT INTO arbeitsverlauf (profil_id, art, kennung, name, vorgang, zeitpunkt)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT (profil_id, art, kennung, vorgang) DO UPDATE SET
		   name = excluded.name, zeitpunkt = excluded.zeitpunkt`,
		profilID, string(schritt.Objekt.Art), schritt.Objekt.Kennung, schritt.Objekt.Name,
		string(schritt.Vorgang), schritt.Zeitpunkt.UTC().Format(time.RFC3339Nano)); err != nil {
		return uebersetze(err, "den Schritt vermerken")
	}

	return nil
}

// Schritte gibt die Schritte des Profils, der jüngste zuerst.
func (a *Arbeitsverlauf) Schritte(ctx context.Context) ([]domain.Arbeitsschritt, error) {
	profilID, err := ProfilAus(ctx)
	if err != nil {
		return nil, err
	}

	zeilen, err := a.db.QueryContext(ctx,
		`SELECT art, kennung, name, vorgang, zeitpunkt FROM arbeitsverlauf
		 WHERE profil_id = ? ORDER BY zeitpunkt DESC, id DESC`, profilID)
	if err != nil {
		return nil, uebersetze(err, "den Arbeitsverlauf lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird über Err()

	schritte := []domain.Arbeitsschritt{}

	for zeilen.Next() {
		var art, kennung, name, vorgang, zeitpunkt string
		if err := zeilen.Scan(&art, &kennung, &name, &vorgang, &zeitpunkt); err != nil {
			return nil, uebersetze(err, "einen Schritt lesen")
		}

		wann, err := time.Parse(time.RFC3339Nano, zeitpunkt)
		if err != nil {
			return nil, uebersetze(err, "den Zeitpunkt %q lesen", zeitpunkt)
		}

		schritte = append(schritte, domain.Arbeitsschritt{
			Objekt:    domain.Arbeitsobjekt{Art: domain.Arbeitsart(art), Kennung: kennung, Name: name},
			Vorgang:   domain.Arbeitsvorgang(vorgang),
			Zeitpunkt: wann,
		})
	}

	if err := zeilen.Err(); err != nil {
		return nil, uebersetze(err, "den Arbeitsverlauf lesen")
	}

	return schritte, nil
}

// Anheften hält ein Objekt oben — ein zweites Mal aktualisiert den Namen.
func (a *Arbeitsverlauf) Anheften(ctx context.Context, objekt domain.Arbeitsobjekt) error {
	if befunde := objekt.Validate(); len(befunde) > 0 {
		return befunde[0]
	}

	profilID, err := ProfilAus(ctx)
	if err != nil {
		return err
	}

	if _, err := a.db.ExecContext(ctx,
		`INSERT INTO angeheftet (profil_id, art, kennung, name, zeitpunkt) VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT (profil_id, art, kennung) DO UPDATE SET name = excluded.name`,
		profilID, string(objekt.Art), objekt.Kennung, objekt.Name,
		time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return uebersetze(err, "das Objekt anheften")
	}

	return nil
}

// Loesen nimmt ein Objekt von oben — und sagt, ob es dort war.
//
// **Über `RETURNING` statt über die Zeilenzahl**: ob eine Zeile fiel, sagt
// die Zeile selbst, und ein Treiber, der nicht zählen kann, ist kein Fall,
// den es hier zu prüfen gäbe.
func (a *Arbeitsverlauf) Loesen(ctx context.Context, art domain.Arbeitsart, kennung string) (bool, error) {
	profilID, err := ProfilAus(ctx)
	if err != nil {
		return false, err
	}

	var id int64

	err = a.db.QueryRowContext(ctx,
		`DELETE FROM angeheftet WHERE profil_id = ? AND art = ? AND kennung = ? RETURNING id`,
		profilID, string(art), kennung).Scan(&id)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return false, nil
	case err != nil:
		return false, uebersetze(err, "das Objekt lösen")
	}

	return true, nil
}

// Angeheftete gibt die angehefteten Objekte des Profils, das jüngste zuerst.
func (a *Arbeitsverlauf) Angeheftete(ctx context.Context) ([]domain.Arbeitsobjekt, error) {
	profilID, err := ProfilAus(ctx)
	if err != nil {
		return nil, err
	}

	zeilen, err := a.db.QueryContext(ctx,
		`SELECT art, kennung, name FROM angeheftet WHERE profil_id = ? ORDER BY zeitpunkt DESC, id DESC`, profilID)
	if err != nil {
		return nil, uebersetze(err, "das Angeheftete lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird über Err()

	objekte := []domain.Arbeitsobjekt{}

	for zeilen.Next() {
		var art, kennung, name string
		if err := zeilen.Scan(&art, &kennung, &name); err != nil {
			return nil, uebersetze(err, "ein angeheftetes Objekt lesen")
		}

		objekte = append(objekte, domain.Arbeitsobjekt{Art: domain.Arbeitsart(art), Kennung: kennung, Name: name})
	}

	if err := zeilen.Err(); err != nil {
		return nil, uebersetze(err, "das Angeheftete lesen")
	}

	return objekte, nil
}

// Vergessen räumt jede Zeile eines Objekts — wenn es das Objekt nicht mehr gibt.
func (a *Arbeitsverlauf) Vergessen(ctx context.Context, art domain.Arbeitsart, kennung string) error {
	profilID, err := ProfilAus(ctx)
	if err != nil {
		return err
	}

	// Zwei feste Befehle statt eines zusammengesetzten: der Tabellenname ist
	// hier zwar keine Eingabe, aber ein Wächter kann das nicht wissen.
	befehle := []struct{ tabelle, befehl string }{
		{"arbeitsverlauf", `DELETE FROM arbeitsverlauf WHERE profil_id = ? AND art = ? AND kennung = ?`},
		{"angeheftet", `DELETE FROM angeheftet WHERE profil_id = ? AND art = ? AND kennung = ?`},
	}

	for _, schritt := range befehle {
		if _, err := a.db.ExecContext(ctx, schritt.befehl, profilID, string(art), kennung); err != nil {
			return uebersetze(err, "das Objekt aus %s vergessen", schritt.tabelle)
		}
	}

	return nil
}
