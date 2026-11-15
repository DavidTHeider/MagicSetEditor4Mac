// **Das eigene Konto des Dienstes** — Werkstück `T11` von Phase 5.6
// (#1476).
//
// **Ohne Wache auf die Platte** (0044): der Weg fasst keine Datei an. Was er
// **nicht** tut, ist schreiben — weder hier noch beim Anbieter.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// kontoweg führt die Kontoübersicht an die Naht von `internal/rest`.
type kontoweg struct{ db *sql.DB }

// Marktkonto gibt die Übersicht über das eigene Konto.
func (k kontoweg) Marktkonto(
	ctx context.Context, quelle string, seite, menge int,
) (any, error) {
	antwort := app.PeerflaecheUeber(k.db).Marktkonto(ctx, quelle, seite, menge)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort.Stand, nil
}
