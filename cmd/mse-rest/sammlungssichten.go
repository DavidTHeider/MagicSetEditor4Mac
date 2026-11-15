// **Die Sammlungssichten des Dienstes** —
// #1292.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// sammlungssichten führt die vier Fragen an die Naht von `internal/rest`.
//
// **Die Methoden tragen `Sammlung` im Namen**: in `internal/rest` erfüllt ein
// Doppel jede Naht, und `Besitz` oder `Wuensche` allein wären dort früher oder
// später doppelt vergeben.
type sammlungssichten struct{ db *sql.DB }

func (s sammlungssichten) SammlungBesitz(ctx context.Context, printing string) (
	any, error,
) {
	return app.SichtenUeber(s.db).Besitz(ctx, printing)
}

func (s sammlungssichten) SammlungLagerorte(ctx context.Context) (any, error) {
	return app.SichtenUeber(s.db).Lagerorte(ctx)
}

func (s sammlungssichten) SammlungDubletten(ctx context.Context, behalten int) (
	any, error,
) {
	return app.SichtenUeber(s.db).Dubletten(ctx, behalten)
}

func (s sammlungssichten) SammlungWuensche(ctx context.Context) (any, error) {
	return app.SichtenUeber(s.db).Wuensche(ctx)
}
