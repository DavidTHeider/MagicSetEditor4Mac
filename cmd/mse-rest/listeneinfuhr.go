// **Die Listeneinfuhr des Dienstes** —
// #1292.
//
// **Ohne Wache auf die Platte** (0044): beide Wege fassen nur den Bestand an,
// keine Datei. Was der Dienst **nicht** tut, ist entscheiden, welcher von
// mehreren Drucken gemeint ist — das steht in `internal/rest/listeneinfuhr.go`.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// listeneinfuhr führt die Einfuhr an die Naht von `internal/rest`.
type listeneinfuhr struct{ db *sql.DB }

// ListeDeuten liest eine Liste und ordnet jede Zeile zu.
//
// **Der Rückgabetyp ist `any`**, wie die Naht ihn verlangt: `internal/rest`
// kennt die DTOs von `internal/app` nicht, und sie dort noch einmal zu
// schreiben hieße, dieselbe Gestalt an zwei Stellen zu pflegen.
func (l listeneinfuhr) ListeDeuten(ctx context.Context, text string) (any, error) {
	naht, err := app.ListeneinfuhrUeber(l.db)
	if err != nil {
		return nil, err
	}

	return naht.Deuten(ctx, text)
}

// Uebernehmen schreibt die entschiedenen Zeilen.
func (l listeneinfuhr) ListeEinlesen(
	ctx context.Context, zeilen []rest.Listeneinfuhrzeile, zusammenfuehrung string,
) (any, error) {
	naht, err := app.ListeneinfuhrUeber(l.db)
	if err != nil {
		return nil, err
	}

	return naht.Einlesen(ctx, nahtzeilen(zeilen), zusammenfuehrung)
}

// ListeVorschauen sagt, was das Einlesen täte — DOM-03.AK-01 (#1698).
func (l listeneinfuhr) ListeVorschauen(
	ctx context.Context, zeilen []rest.Listeneinfuhrzeile,
) (any, error) {
	naht, err := app.ListeneinfuhrUeber(l.db)
	if err != nil {
		return nil, err
	}

	return naht.Vorschauen(ctx, nahtzeilen(zeilen))
}

// nahtzeilen übersetzt die Zeilen des Dienstes in die der Naht — **eine**
// Stelle für Einlesen und Vorschau.
func nahtzeilen(zeilen []rest.Listeneinfuhrzeile) []app.Listeneinfuhrzeile {
	uebersetzt := make([]app.Listeneinfuhrzeile, 0, len(zeilen))

	for _, zeile := range zeilen {
		uebersetzt = append(uebersetzt, app.Listeneinfuhrzeile{
			Printing: zeile.Printing, Menge: zeile.Menge,
			Zustand: zeile.Zustand, Sprache: zeile.Sprache,
			Finish: app.AusfuehrungAus(zeile.Finish, zeile.Foil),
			Name:   zeile.Name, Setcode: zeile.Setcode, Nummer: zeile.Nummer,
		})
	}

	return uebersetzt
}
