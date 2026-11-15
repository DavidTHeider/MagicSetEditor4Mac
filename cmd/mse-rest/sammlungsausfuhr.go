// **Die Sammlungsausfuhr des Dienstes** —
// #1292.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// sammlungsausfuhr führt die Ausfuhr an die Naht von `internal/rest`.
//
// **Die Methoden heißen `SammlungAusgeben` und `SammlungCSVEinlesen`**: in
// `internal/rest` erfüllt ein Doppel jede Naht, und ein Name je Fähigkeit ist
// dort die Bedingung dafür, dass sie nebeneinander bestehen.
type sammlungsausfuhr struct{ db *sql.DB }

func (s sammlungsausfuhr) SammlungAusgeben(ctx context.Context, form string) (
	string, error,
) {
	return app.AusfuhrUeber(s.db).Ausgeben(ctx, form)
}

func (s sammlungsausfuhr) SammlungCSVEinlesen(ctx context.Context, text, quelle, zusammenfuehrung string) (
	any, error,
) {
	return app.AusfuhrUeber(s.db).CSVEinlesen(ctx, text, quelle, zusammenfuehrung)
}

// SammlungCSVVorschauen sagt, was das Einlesen täte — DOM-03.AK-01 (#1698).
func (s sammlungsausfuhr) SammlungCSVVorschauen(ctx context.Context, text string) (any, error) {
	return app.AusfuhrUeber(s.db).CSVVorschauen(ctx, text)
}
