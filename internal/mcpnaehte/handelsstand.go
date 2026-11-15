// **Der Handelsstand des MCP-Servers** —
// #1637.
//
// **Ohne Wache** (0044): der Weg liest und fasst keine Datei an. Was er
// **nicht** kann, ist einen Handel bewegen — die Naht bietet es nicht an.

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// handelsstand erfüllt `mcp.Handelsauskunft` über der Naht von `internal/app`.
type handelsstand struct{ db *sql.DB }

func (h handelsstand) Handelsstand(
	ctx context.Context, vorgang string,
) (mcp.Handelsstand, error) {
	antwort := app.PeerflaecheUeber(h.db).HandelStand(ctx, vorgang)
	if antwort.Fehler != nil {
		return mcp.Handelsstand{}, antwort.Fehler.AlsFehler()
	}

	return alsHandelsstand(antwort.Vorgang), nil
}

// Handelsvorgaenge nennt die Handel mit einem Gegenüber (UX-08.AK-04).
func (h handelsstand) Handelsvorgaenge(
	ctx context.Context, gegenueber string,
) ([]mcp.Handelsstand, error) {
	antwort := app.PeerflaecheUeber(h.db).HandelVorgaenge(ctx, gegenueber)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	staende := make([]mcp.Handelsstand, 0, len(antwort.Vorgaenge))
	for _, vorgang := range antwort.Vorgaenge {
		staende = append(staende, alsHandelsstand(vorgang))
	}

	return staende, nil
}

// alsHandelsstand gibt eine Revision in der Gestalt dieser Fläche.
func alsHandelsstand(stand app.HandelsvorgangDTO) mcp.Handelsstand {
	zeilen := make([]mcp.Handelszeile, 0, len(stand.Positionen))
	for _, position := range stand.Positionen {
		zeilen = append(zeilen, mcp.Handelszeile{
			Von: position.Von, Printing: position.Printing,
			Menge: position.Menge, Guete: position.Guete,
			Sprache: position.Sprache, Finish: position.Finish,
		})
	}

	return mcp.Handelsstand{
		Vorgang: stand.Vorgang, Revision: stand.Revision,
		Abdruck: stand.Abdruck, Art: stand.Art, Initiator: stand.Initiator,
		Parteien: stand.Parteien, Ich: stand.Ich, Zustand: stand.Zustand,
		Positionen: zeilen, Zustimmungen: stand.Zustimmungen,
		Ende: stand.Ende, Beschluss: stand.Beschluss, Lage: stand.Lage,
		Naechster: stand.Naechster, Moegliche: stand.Moegliche,
		Phase: stand.Phase,
	}
}

// Handelszusagen gibt, welche Besitzzeilen festgehalten sind (#1643).
func (h handelsstand) Handelszusagen(
	ctx context.Context, vorgang string,
) ([]mcp.Handelszusage, error) {
	antwort := app.PeerflaecheUeber(h.db).Handelszusagen(ctx, vorgang)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	zusagen := make([]mcp.Handelszusage, 0, len(antwort.Zusagen))

	for _, zusage := range antwort.Zusagen {
		zusagen = append(zusagen, mcp.Handelszusage{
			ExemplarID: zusage.ExemplarID, Menge: zusage.Menge,
			Fassung: zusage.Fassung, Revision: zusage.Revision,
			Frist: zusage.Frist, Uebergeben: zusage.Uebergeben,
		})
	}

	return zusagen, nil
}
