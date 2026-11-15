// **Die Sicherung des ganzen Bestands am MCP-Server** —
// #1169.
//
// **Ohne das Zurückspielen** — warum, steht in
// `internal/mcp/bestandssicherung.go`.

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// bestandssicherung erfüllt `mcp.Bestandssicherung` über der Naht von `app`.
type bestandssicherung struct {
	db  *sql.DB
	ort string
}

func (b bestandssicherung) BestandSichern(ctx context.Context, ziel string, klartext bool) error {
	db := b.db

	return app.BestandsnahtUeber(&db, b.ort).Sichern(app.MitSicherungswunsch(ctx, domain.FlaecheMCP, "", klartext), ziel)
}

func (b bestandssicherung) SicherungSchauen(ctx context.Context, pfad string) (
	mcp.Sicherungsangaben, error,
) {
	db := b.db

	schau, fehler := app.BestandsnahtUeber(&db, b.ort).Beschauen(ctx, pfad)
	if fehler != nil {
		return mcp.Sicherungsangaben{}, fehler
	}

	return mcp.Sicherungsangaben{
		AngelegtAm:  schau.AngelegtAm,
		Schemastand: schau.Schemastand,
		Sets:        schau.Sets,
		Karten:      schau.Karten,
		Bilder:      schau.Bilder,

		Verschluesselt: schau.Verschluesselt,

		PaketeEnthalten: schau.PaketeEnthalten,
		PaketeFehlen:    schau.PaketeFehlen,
	}, nil
}

// Sicherungsstand nennt das Alter der letzten Sicherung.
func (b bestandssicherung) Sicherungsstand(ctx context.Context) (any, error) {
	db := b.db

	return app.BestandsnahtUeber(&db, b.ort).Sicherungsstand(ctx)
}

// Verschluesselungsstand sagt, ob der Bestand verschlüsselt daliegt.
func (b bestandssicherung) Verschluesselungsstand() (bool, error) {
	db := b.db

	return app.BestandsnahtUeber(&db, b.ort).Verschluesselungsstand()
}
