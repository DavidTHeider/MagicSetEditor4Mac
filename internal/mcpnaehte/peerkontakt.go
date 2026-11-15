package mcpnaehte

// **Die bekannten Gegenüber am MCP-Server** — Werkstück N1 von Phase 5.4
// (#1454).
//
// **Lesen und Sperren.** Warum das Freigeben hier nicht steht, sagt der Kopf
// von `internal/mcp/peerkontakt.go`.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// kontaktauskunft reicht die Kontakte an die Werkzeuge.
type kontaktauskunft struct{ db *sql.DB }

func (k kontaktauskunft) Peerkontakte(
	ctx context.Context,
) (mcp.Peerkontakte, error) {
	return alsMCPkontakte(app.PeerflaecheUeber(k.db).Peerkontakte(ctx))
}

func (k kontaktauskunft) PeerkontaktSperren(
	ctx context.Context, kennung string,
) (mcp.Peerkontakte, error) {
	return alsMCPkontakte(
		app.PeerflaecheUeber(k.db).PeerkontaktSperren(ctx, kennung))
}

// alsMCPkontakte macht aus der Antwort der Anwendung die des Werkzeugs.
func alsMCPkontakte(
	antwort app.PeerkontakteAntwort,
) (mcp.Peerkontakte, error) {
	if antwort.Fehler != nil {
		return mcp.Peerkontakte{}, antwort.Fehler.AlsFehler()
	}

	kontakte := make([]mcp.Peerkontakt, 0, len(antwort.Kontakte))

	for _, kontakt := range antwort.Kontakte {
		kontakte = append(kontakte, mcp.Peerkontakt{
			Kennung:              kontakt.Kennung,
			Fingerabdruck:        kontakt.Fingerabdruck,
			SelbstgewaehlterName: kontakt.SelbstgewaehlterName,
			EigenerName:          kontakt.EigenerName,
			Stand:                kontakt.Stand,
			Gesehen:              kontakt.Gesehen,
			Rechte:               kontakt.Rechte,
		})
	}

	return mcp.Peerkontakte{
		Kontakte: kontakte, Zwecke: antwort.Zwecke,
		Nachladefehler: alsNachladefehler(antwort.Nachladefehler),
	}, nil
}

// Kontakte gibt die Naht für den Werkzeugsatz.
func Kontakte(db *sql.DB) mcp.Kontaktauskunft {
	return kontaktauskunft{db: db}
}
