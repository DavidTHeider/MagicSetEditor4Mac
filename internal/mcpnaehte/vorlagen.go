// **Die Exportvorlagen des MCP-Servers** —
// #1291.
//
// **Die Wache steht hier**, wie bei jedem Schreiben auf die Platte: ein
// Werkzeug schreibt nur, was es nicht überschreibt
// ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)).

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// vorlagen führt die Exportvorlagen an die Naht von `internal/mcp`.
type vorlagen struct{ db *sql.DB }

// Vorlagen nennt die Exportvorlagen, die zu einem Set passen.
func (v vorlagen) Vorlagen(ctx context.Context, setID int64) (
	[]mcp.Vorlagenzeile, []string, error,
) {
	weg, err := app.VorlagenwegUeber(v.db, domain.UrheberServer)
	if err != nil {
		return nil, nil, err
	}

	gelesen, befunde, lesefehler := weg.Vorlagen(ctx, setID)
	if lesefehler != nil {
		return nil, befunde, lesefehler
	}

	zeilen := make([]mcp.Vorlagenzeile, 0, len(gelesen))

	for _, vorlage := range gelesen {
		optionen := make([]mcp.Vorlagenoption, 0, len(vorlage.Optionen))
		for _, option := range vorlage.Optionen {
			optionen = append(optionen, mcp.Vorlagenoption{
				Name: option.Name, Art: option.Art, Wert: option.Wert,
			})
		}

		zeilen = append(zeilen, mcp.Vorlagenzeile{
			Paket: vorlage.Paket, Name: vorlage.Name,
			Dateityp: vorlage.Dateityp, Optionen: optionen,
		})
	}

	return zeilen, befunde, nil
}

// Ausfuehren schreibt das Ergebnis — nur dorthin, wo nichts liegt.
func (v vorlagen) Ausfuehren(
	ctx context.Context, setID int64, paket, ziel string, wahl map[string]string,
) ([]string, error) {
	aufraeumen, err := reserviereOrt(ziel)
	if err != nil {
		return nil, err
	}

	// **Die Reservierung fällt, wenn nichts daraus wird** (#1343):
	// eine Null-Byte-Datei sähe aus wie ein geglücktes Schreiben.
	geschrieben := false

	defer func() {
		if !geschrieben {
			aufraeumen()
		}
	}()

	weg, err := app.VorlagenwegUeber(v.db, domain.UrheberServer)
	if err != nil {
		return nil, err
	}

	lauf, laufFehler := weg.Ausfuehren(ctx, setID, paket, ziel, wahl)
	if laufFehler != nil {
		return lauf.Befunde, laufFehler
	}

	geschrieben = true

	return lauf.Befunde, nil
}

// OptionSetzen behält den Wert einer Vorlagenoption.
func (v vorlagen) OptionSetzen(ctx context.Context, vorlage, feld, wert string) error {
	weg, err := app.VorlagenwegUeber(v.db, domain.UrheberServer)
	if err != nil {
		return err
	}

	return weg.OptionSetzen(ctx, vorlage, feld, wert)
}
