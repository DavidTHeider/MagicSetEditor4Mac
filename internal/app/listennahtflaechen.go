// **Die Kartenliste für die Flächen ohne Fenster** — Werkstück 3 von
// Phase 3.7
// (#1168).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort. **Auch ohne Fenster gibt
// es Spalten**: eine Kommandozeile, die eine Kartenliste ausgibt, hat
// dieselbe Frage — welche Felder stehen darin und in welcher Reihenfolge.

package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Fremdliste ist die Kartenliste für eine Fläche ohne Fenster.
type Fremdliste struct{ quelle Listenquelle }

// Listenflaechennaht baut sie über einem Bestand.
func Listenflaechennaht(db *sql.DB) (*Fremdliste, error) {
	pakete, err := Paketverzeichnis()
	if err != nil {
		return nil, err
	}

	leser := storage.NeuerSetLeser(db)

	waehler := stilwaehlerMit(db, pakete)

	// **Und die Marken je Karte** — wie an der Wurzel des Fensters
	// ([App.listenbau]). Sie fehlten hier: `ListenzeileDTO.Marken` gibt es an
	// jeder Fläche, und über Kommandozeile, MCP und REST kam das Feld leer
	// zurück, gleich wie viele Marken an der Karte hingen (§16, #1376).
	return &Fremdliste{quelle: service.NeuerKartenlistenbau(
		leser, leser, waehler, storage.NeueListenablage(db), waehler).
		MitMarken(storage.NeueMarken(db))}, nil
}

// MitListenquelle setzt die Naht — für Tests.
func (f *Fremdliste) MitListenquelle(quelle Listenquelle) *Fremdliste {
	f.quelle = quelle

	return f
}

// Aufbau gibt Spalten, Zeilen und die wirksame Ordnung.
func (f *Fremdliste) Aufbau(ctx context.Context, setID int64) (
	domain.Kartenlistenaufbau, error,
) {
	return f.quelle.Aufbau(ctx, setID)
}

// SpaltenSetzen schreibt die Wahl fort.
func (f *Fremdliste) SpaltenSetzen(
	ctx context.Context, setID int64, wahl []domain.Spaltenwahl,
) error {
	return f.quelle.SpaltenSetzen(ctx, setID, wahl)
}

// SortierungSetzen schreibt die Ordnung fort.
func (f *Fremdliste) SortierungSetzen(
	ctx context.Context, setID int64, ordnung domain.Kartenordnung,
) error {
	return f.quelle.SortierungSetzen(ctx, setID, ordnung)
}
