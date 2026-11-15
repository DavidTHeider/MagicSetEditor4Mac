package storage

// **Die Symbolbibliothek im Bestand** — DOM-02
// (#1697).
//
// Je Profil, ein Name einmal (ohne Rücksicht auf Groß und Klein), und das
// Profil steht in **jeder** Frage — ein Symbol eines anderen Profils ist hier
// eines, das es nicht gibt (`ErrNotFound`), kein verbotenes.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Symbolbibliothek schreibt und liest die Symbole eines Profils.
type Symbolbibliothek struct{ db *sql.DB }

// NeueSymbolbibliothek baut die Ablage über dem Bestand.
func NeueSymbolbibliothek(db *sql.DB) *Symbolbibliothek { return &Symbolbibliothek{db: db} }

// jetzt ist der Zeitpunkt, wie die Tabelle ihn schreibt.
const symboljetzt = `strftime('%Y-%m-%dT%H:%M:%fZ', 'now')`

// Symbole gibt die Symbole des Profils, nach Namen geordnet.
func (b *Symbolbibliothek) Symbole(ctx context.Context) ([]domain.Bibliothekssymbol, error) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return nil, fehler
	}

	zeilen, err := b.db.QueryContext(ctx,
		`SELECT id, name, inhalt, abdruck, vorher, angelegt, geaendert
		   FROM symbolbibliothek WHERE profil_id = ? ORDER BY name, id`, profilID)
	if err != nil {
		return nil, uebersetze(err, "die Symbole lesen")
	}

	defer func() { _ = zeilen.Close() }()

	symbole := []domain.Bibliothekssymbol{}

	for zeilen.Next() {
		symbol, lesefehler := symbolLesen(zeilen)
		if lesefehler != nil {
			return nil, lesefehler
		}

		symbole = append(symbole, symbol)
	}

	return symbole, uebersetze(zeilen.Err(), "die Symbole lesen")
}

// Symbol gibt ein Symbol des Profils.
func (b *Symbolbibliothek) Symbol(ctx context.Context, id int64) (domain.Bibliothekssymbol, error) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return domain.Bibliothekssymbol{}, fehler
	}

	symbol, err := symbolLesen(b.db.QueryRowContext(ctx,
		`SELECT id, name, inhalt, abdruck, vorher, angelegt, geaendert
		   FROM symbolbibliothek WHERE id = ? AND profil_id = ?`, id, profilID))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Bibliothekssymbol{}, fmt.Errorf("%w: das Symbol %d gibt es in diesem Profil nicht",
			domain.ErrNotFound, id)
	}

	if err != nil {
		return domain.Bibliothekssymbol{}, fmt.Errorf("das Symbol %d: %w", id, err)
	}

	return symbol, nil
}

// symbolLesen liest eine Zeile.
func symbolLesen(zeile zeilenleser) (domain.Bibliothekssymbol, error) {
	var (
		symbol              domain.Bibliothekssymbol
		angelegt, geaendert string
	)

	if err := zeile.Scan(&symbol.ID, &symbol.Name, &symbol.Inhalt, &symbol.Abdruck,
		&symbol.Vorher, &angelegt, &geaendert); err != nil {
		return domain.Bibliothekssymbol{}, uebersetze(err, "ein Symbol lesen")
	}

	// **Den Zeitpunkt schreibt die Tabelle selbst** — ein Fehler beim Lesen
	// hieße, dass ihn jemand von Hand geändert hat, und der Nullwert sagt
	// dann „unbekannt" statt einer erfundenen Zeit.
	symbol.Angelegt, _ = time.Parse(time.RFC3339Nano, angelegt)   //nolint:errcheck // siehe oben
	symbol.Geaendert, _ = time.Parse(time.RFC3339Nano, geaendert) //nolint:errcheck // dito

	return symbol, nil
}

// SymbolAnlegen legt ein Symbol an und gibt es mit Kennung und Zeiten.
func (b *Symbolbibliothek) SymbolAnlegen(
	ctx context.Context, symbol domain.Bibliothekssymbol,
) (domain.Bibliothekssymbol, error) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return domain.Bibliothekssymbol{}, fehler
	}

	ergebnis, err := b.db.ExecContext(ctx,
		`INSERT INTO symbolbibliothek (profil_id, name, inhalt, abdruck) VALUES (?, ?, ?, ?)`,
		profilID, symbol.Name, symbol.Inhalt, symbol.Abdruck)
	if err != nil {
		return domain.Bibliothekssymbol{}, uebersetze(err, "das Symbol %q anlegen", symbol.Name)
	}

	id, _ := ergebnis.LastInsertId() //nolint:errcheck // SQLite gibt sie immer

	return b.Symbol(ctx, id)
}

// SymbolUmbenennen gibt einem Symbol einen anderen Namen.
func (b *Symbolbibliothek) SymbolUmbenennen(ctx context.Context, id int64, name string) error {
	return b.aendere(ctx, id, "umbenennen",
		`UPDATE symbolbibliothek SET name = ?, geaendert = `+symboljetzt+` WHERE id = ? AND profil_id = ?`, name)
}

// SymbolInhaltSetzen legt einen neuen Inhalt ein; der bisherige wird zum
// vorigen.
func (b *Symbolbibliothek) SymbolInhaltSetzen(ctx context.Context, id int64, inhalt, abdruck string) error {
	return b.aendere(ctx, id, "ändern",
		`UPDATE symbolbibliothek SET vorher = inhalt, inhalt = ?, abdruck = ?, geaendert = `+symboljetzt+`
		  WHERE id = ? AND profil_id = ?`, inhalt, abdruck)
}

// SymbolLoeschen nimmt ein Symbol aus der Bibliothek; seine Marken gehen mit
// (Auslöser `objektmarken_symbol_weg`).
func (b *Symbolbibliothek) SymbolLoeschen(ctx context.Context, id int64) error {
	return b.aendere(ctx, id, "löschen",
		`DELETE FROM symbolbibliothek WHERE id = ? AND profil_id = ?`)
}

// aendere führt eine Änderung an einem Symbol des Profils aus — die Werte
// zuerst, dann Kennung und Profil — und meldet eines, das es nicht gibt.
func (b *Symbolbibliothek) aendere(ctx context.Context, id int64, was, frage string, werte ...any) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	ergebnis, err := b.db.ExecContext(ctx, frage, append(werte, id, profilID)...)
	if err != nil {
		return uebersetze(err, "das Symbol %d %s", id, was)
	}

	if anzahl, _ := ergebnis.RowsAffected(); anzahl == 0 { //nolint:errcheck // SQLite zählt immer
		return fmt.Errorf("%w: das Symbol %d gibt es in diesem Profil nicht", domain.ErrNotFound, id)
	}

	return nil
}
