package app

// **Die Marken an jedem Fachobjekt über einem Bestand** — UX-12
// (#1684).
// Ein Composition Root wie `markennaht.go`: Fenster, Kommandozeile, Dienst
// und Werkzeugserver bekommen dieselbe Naht.

import (
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// ObjektmarkenUeber baut sie über einem Bestand.
func ObjektmarkenUeber(db *sql.DB) *Objektmarkennaht {
	return &Objektmarkennaht{marken: service.NeueObjektmarken(storage.NeueObjektmarkenablage(db))}
}
