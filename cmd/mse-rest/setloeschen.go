// **Das Löschen eines Sets** —
// #1218.
//
// **Über `app.SetloeschnahtUeber`**, wie jede andere Naht ohne Fenster: was
// zusammengesetzt wird, gehört an eine Stelle (§2).
package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

type setloescher struct{ db *sql.DB }

func (s setloescher) SetLoeschen(ctx context.Context, setID int64, name string) (
	[]string, error,
) {
	return app.SetloeschnahtUeber(s.db).SetLoeschen(ctx, setID, name)
}
