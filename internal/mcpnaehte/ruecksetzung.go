package mcpnaehte

// **Die abgestufte Rücksetzung über MCP** — Werkstück O4 von Phase 5.2
// (#1428).
//
// **Nur die Vorschau**, und der Grund steht im Kopf von `mcp/ruecksetzung.go`.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// ruecksetzungsauskunft reicht die Vorschau an das Werkzeug.
type ruecksetzungsauskunft struct {
	db  *sql.DB
	ort string
}

func (r ruecksetzungsauskunft) Ruecksetzungsvorschau(
	ctx context.Context, art, profilID, nutzerID string,
) (mcp.Ruecksetzungsvorschau, error) {
	antwort := app.RuecksetzungsAnwendungUeber(r.db, r.ort).
		Ruecksetzungsvorschau(ctx, art, profilID, nutzerID)
	if antwort.Fehler != nil {
		return mcp.Ruecksetzungsvorschau{}, antwort.Fehler.AlsFehler()
	}

	vorschau := mcp.Ruecksetzungsvorschau{
		Art:                antwort.Art,
		Betroffen:          make([]mcp.Betroffenes, 0, len(antwort.Betroffen)),
		Hindernisse:        make([]mcp.Hindernis, 0, len(antwort.Hindernisse)),
		Bereit:             antwort.Bereit,
		NameNoetig:         antwort.NameNoetig,
		NachweisNoetig:     antwort.NachweisNoetig,
		SicherungEmpfohlen: antwort.SicherungEmpfohlen,
	}

	for _, gruppe := range antwort.Betroffen {
		vorschau.Betroffen = append(vorschau.Betroffen, mcp.Betroffenes{
			Was: gruppe.Was, Anzahl: gruppe.Anzahl, Namen: gruppe.Namen,
		})
	}

	for _, hindernis := range antwort.Hindernisse {
		vorschau.Hindernisse = append(vorschau.Hindernisse, mcp.Hindernis{
			Art: hindernis.Art, Stelle: hindernis.Stelle,
		})
	}

	return vorschau, nil
}
