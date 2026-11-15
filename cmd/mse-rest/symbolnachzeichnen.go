// **Ein Bild als Symbol nachzeichnen** — Werkstück F von Phase 4.11
// (#1340).

package main

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// symbolnachzeichnung erfüllt `rest.Symbolwerkstatt`.
type symbolnachzeichnung struct{ db *sql.DB }

func (s symbolnachzeichnung) SymbolNachzeichnen(
	ctx context.Context, setID int64, bildpfad string,
) (any, error) {
	naht, fehler := app.SymbolnachzeichnungUeber(s.db)
	if fehler != nil {
		return nil, fehler
	}

	return naht.Nachzeichnen(ctx, setID, bildpfad)
}

// **Und der Editor** — seit dem 2026-09-02
// (#1391).
// Dieselbe Werkstatt, weil beide am selben Gegenstand arbeiten.

// Lesen holt das abgelegte Symbol eines Sets — im Profil der Anfrage (#1815).
func (s symbolnachzeichnung) SymbolLesen(ctx context.Context, setID int64, name string) (any, error) {
	naht, err := app.SymbolwerkstattUeber(s.db)
	if err != nil {
		return nil, err
	}

	return naht.Lesen(ctx, setID, name)
}

// Ablegen schreibt eine Zeichnung neben das Set.
func (s symbolnachzeichnung) SymbolAblegen(
	ctx context.Context, setID int64, teile json.RawMessage, unter string,
) (string, error) {
	naht, err := app.SymbolwerkstattUeber(s.db)
	if err != nil {
		return "", err
	}

	var gelesen []app.SymbolteilDTO
	if err := json.Unmarshal(teile, &gelesen); err != nil {
		return "", app.SymbolteileUnlesbar(err)
	}

	return naht.Ablegen(ctx, setID, gelesen, unter)
}

// Rechnen führt eine der Rechnungen aus — **ohne Naht und ohne Bestand**.
func (symbolnachzeichnung) SymbolRechnen(was string, argumente json.RawMessage) (any, error) {
	return app.SymbolRechnen(was, argumente)
}

// Palette gibt, was ein Kartenfeld zum Einfügen anbietet — ohne die Bildchen.
//
// **Mit dem Kontext der Anfrage und nicht mit `context.Background()`**
// (#1815): nur er trägt das Profil. Ohne ihn las die Palette die Karte jedes
// Profils — und seit die Ablage nach dem Profil fragt, gar keine mehr.
func (s symbolnachzeichnung) Palette(ctx context.Context, karteID int64, feld string) (any, error) {
	werk, err := app.Zeichennaht(s.db)
	if err != nil {
		return nil, err
	}

	return werk.Palette(ctx, karteID, feld)
}
