package storage

// **Die eigenen Übersetzungen** — LOC-02
// (#1688).
//
// **Ohne Profil**: ein Sprachpaket gehört dem Rechner wie ein installiertes
// MSE-Paket (Wanderung 0092). Wer es benutzt, sagt seine Einstellung.

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sprachpaketablage liest und schreibt die eigenen Übersetzungen.
type Sprachpaketablage struct{ db *sql.DB }

// NeueSprachpaketablage gibt die Ablage über diesem Bestand.
func NeueSprachpaketablage(db *sql.DB) *Sprachpaketablage {
	return &Sprachpaketablage{db: db}
}

// sprachpaketspalten sind die Spalten in der Reihenfolge von [sprachpaketLesen].
const sprachpaketspalten = `kennung, name, basis, stand, lizenz, herkunft, texte, angelegt, geaendert`

// Sprachpakete gibt alle — nach der Kennung geordnet.
func (a *Sprachpaketablage) Sprachpakete(ctx context.Context) ([]domain.Sprachpaket, error) {
	zeilen, err := a.db.QueryContext(ctx,
		`SELECT `+sprachpaketspalten+` FROM sprachpakete ORDER BY kennung`)
	if err != nil {
		return nil, uebersetze(err, "die Sprachpakete lesen")
	}

	defer func() { _ = zeilen.Close() }()

	pakete := []domain.Sprachpaket{}

	for zeilen.Next() {
		paket, lesefehler := sprachpaketLesen(zeilen)
		if lesefehler != nil {
			return nil, lesefehler
		}

		pakete = append(pakete, paket)
	}

	return pakete, uebersetze(zeilen.Err(), "die Sprachpakete lesen")
}

// Sprachpaket gibt eines.
func (a *Sprachpaketablage) Sprachpaket(ctx context.Context, kennung string) (domain.Sprachpaket, error) {
	paket, err := sprachpaketLesen(a.db.QueryRowContext(ctx,
		`SELECT `+sprachpaketspalten+` FROM sprachpakete WHERE kennung = ?`, kennung))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Sprachpaket{}, fehltSprachpaket(kennung)
	}

	return paket, err
}

// sprachpaketLesen liest eine Zeile.
//
// **Die Zeiten schreibt der Bestand selbst** (`strftime`), und ein Texte-Feld,
// das kein Objekt ist, lässt die Bedingung der Wanderung nicht hinein —
// deshalb gehen beide Umwandlungen ohne eigenen Fehlerweg.
func sprachpaketLesen(zeile zeilenleser) (domain.Sprachpaket, error) {
	var (
		paket               domain.Sprachpaket
		stand, texte        string
		angelegt, geaendert string
	)

	if err := zeile.Scan(&paket.Kennung, &paket.Name, &paket.Basis, &stand, &paket.Lizenz,
		&paket.Herkunft, &texte, &angelegt, &geaendert); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Sprachpaket{}, err
		}

		return domain.Sprachpaket{}, uebersetze(err, "ein Sprachpaket lesen")
	}

	paket.Stand = domain.Sprachstand(stand)
	paket.Texte = map[string]domain.Uebersetzungstext{}
	_ = json.Unmarshal([]byte(texte), &paket.Texte)              //nolint:errcheck // siehe oben
	paket.Angelegt, _ = time.Parse(time.RFC3339Nano, angelegt)   //nolint:errcheck // dito
	paket.Geaendert, _ = time.Parse(time.RFC3339Nano, geaendert) //nolint:errcheck // dito

	return paket, nil
}

// SprachpaketAnlegen legt ein neues an — eine vergebene Kennung ist doppelt.
func (a *Sprachpaketablage) SprachpaketAnlegen(ctx context.Context, paket domain.Sprachpaket) error {
	_, err := a.db.ExecContext(ctx,
		`INSERT INTO sprachpakete (kennung, name, basis, stand, lizenz, herkunft, texte)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		paket.Kennung, paket.Name, paket.Basis, string(paket.Stand), paket.Lizenz,
		paket.Herkunft, texteAlsJSON(paket.Texte))

	return uebersetze(err, "das Sprachpaket %q anlegen", paket.Kennung)
}

// SprachpaketSchreiben ersetzt Angaben und Sätze eines vorhandenen.
func (a *Sprachpaketablage) SprachpaketSchreiben(ctx context.Context, paket domain.Sprachpaket) error {
	ergebnis, err := a.db.ExecContext(ctx,
		`UPDATE sprachpakete
		    SET name = ?, basis = ?, stand = ?, lizenz = ?, herkunft = ?, texte = ?,
		        geaendert = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		  WHERE kennung = ?`,
		paket.Name, paket.Basis, string(paket.Stand), paket.Lizenz, paket.Herkunft,
		texteAlsJSON(paket.Texte), paket.Kennung)

	return sprachpaketGetroffen(ergebnis, uebersetze(err, "das Sprachpaket %q schreiben", paket.Kennung),
		paket.Kennung)
}

// SprachpaketAendern liest ein Paket, lässt es ändern und schreibt es — **in
// einem Zug**. Die Sätze stehen als ein Objekt in einer Zeile; wer liest und
// später zurückschreibt, überschriebe, was ein zweiter Prozess inzwischen
// geschrieben hat (`TestZweiProzesseAendernDieselbeUebersetzungOhneVerlust`).
// Lehnt `aendere` ab, wird nichts geschrieben.
func (a *Sprachpaketablage) SprachpaketAendern(
	ctx context.Context, kennung string, aendere func(*domain.Sprachpaket) error,
) (fehler error) {
	zug, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "das Sprachpaket %q ändern", kennung)
	}

	// **Auch nach einem Zusammenbruch in `aendere`** (BK-04, #2129).
	defer func() { fehler = zurueckgenommen(zug, fehler) }()

	paket, err := sprachpaketLesen(zug.QueryRowContext(ctx,
		`SELECT `+sprachpaketspalten+` FROM sprachpakete WHERE kennung = ?`, kennung))
	if errors.Is(err, sql.ErrNoRows) {
		return fehltSprachpaket(kennung)
	}

	if err != nil {
		return err
	}

	if err := aendere(&paket); err != nil {
		return err
	}

	if _, err := zug.ExecContext(ctx,
		`UPDATE sprachpakete
		    SET name = ?, basis = ?, stand = ?, lizenz = ?, herkunft = ?, texte = ?,
		        geaendert = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		  WHERE kennung = ?`,
		paket.Name, paket.Basis, string(paket.Stand), paket.Lizenz, paket.Herkunft,
		texteAlsJSON(paket.Texte), kennung); err != nil {
		return uebersetze(err, "das Sprachpaket %q ändern", kennung)
	}

	return uebersetze(zug.Commit(), "das Sprachpaket %q ändern", kennung)
}

// SprachpaketLoeschen entfernt eines.
func (a *Sprachpaketablage) SprachpaketLoeschen(ctx context.Context, kennung string) error {
	ergebnis, err := a.db.ExecContext(ctx, `DELETE FROM sprachpakete WHERE kennung = ?`, kennung)

	return sprachpaketGetroffen(ergebnis, uebersetze(err, "das Sprachpaket %q löschen", kennung), kennung)
}

// sprachpaketGetroffen sagt „gibt es nicht", wenn eine Änderung keine Zeile traf.
func sprachpaketGetroffen(ergebnis sql.Result, fehler error, kennung string) error {
	if fehler != nil {
		return fehler
	}

	if anzahl, _ := ergebnis.RowsAffected(); anzahl == 0 { //nolint:errcheck // SQLite zählt immer
		return fehltSprachpaket(kennung)
	}

	return nil
}

// fehltSprachpaket ist der Befund für eine unbekannte Kennung.
func fehltSprachpaket(kennung string) error {
	return fmt.Errorf("%w: das Sprachpaket %q gibt es nicht", domain.ErrNotFound, kennung)
}

// texteAlsJSON schreibt die Sätze als Objekt — auch ohne einen Satz.
func texteAlsJSON(texte map[string]domain.Uebersetzungstext) string {
	if texte == nil {
		texte = map[string]domain.Uebersetzungstext{}
	}

	roh, _ := json.Marshal(texte) //nolint:errcheck // Zeichenketten lassen sich immer schreiben

	return string(roh)
}
