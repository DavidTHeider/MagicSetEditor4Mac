// Die Setangaben, die Vorbehalte und die mitgelieferten Sets für den
// MCP-Server
// (#1107).

package mcpnaehte

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// setangaben führt die vier Werkzeuge an die Naht dieser Schicht.
type setangaben struct{ db *sql.DB }

func (s setangaben) Angaben(ctx context.Context, setID int64) (
	[]mcp.Angabenfeld, []mcp.Angabenwert, error,
) {
	weg, err := app.SetangabennahtUeber(s.db, domain.UrheberServer)
	if err != nil {
		return nil, nil, err
	}

	felder, werte, lesefehler := weg.Angaben(ctx, setID)
	if lesefehler != nil {
		return nil, nil, lesefehler
	}

	beschreibungen := make([]mcp.Angabenfeld, 0, len(felder))
	for _, feld := range felder {
		beschreibungen = append(beschreibungen, mcp.Angabenfeld{
			Name: feld.Name, Aufschrift: feld.Aufschrift, Art: feld.Art,
		})
	}

	belegt := make([]mcp.Angabenwert, 0, len(werte))
	for _, wert := range werte {
		belegt = append(belegt, mcp.Angabenwert{Feld: wert.Feld, Wert: wert.Wert})
	}

	return beschreibungen, belegt, nil
}

func (s setangaben) Setzen(ctx context.Context, setID int64, feld, wert string) (
	vorher string, hatte bool, fehler error,
) {
	// **Der Feldname wird geprüft, der Wert nicht** — dieselbe Regel wie im
	// Fenster: ein leerer Name ergäbe eine Zeile, die niemand wiederfindet.
	if befund := domain.PruefeFeldname(feld); befund != nil {
		return "", false, befund
	}

	weg, err := app.SetangabennahtUeber(s.db, domain.UrheberServer)
	if err != nil {
		return "", false, err
	}

	return weg.Setzen(ctx, setID, feld, wert)
}

func (s setangaben) Vorbehalte(ctx context.Context, setID int64) ([]mcp.Vorbehalt, error) {
	weg, err := app.SetangabennahtUeber(s.db, domain.UrheberServer)
	if err != nil {
		return nil, err
	}

	bloecke, lesefehler := weg.Vorbehalte(ctx, setID)
	if lesefehler != nil {
		return nil, lesefehler
	}

	vorbehalte := make([]mcp.Vorbehalt, 0, len(bloecke))
	for _, block := range bloecke {
		vorbehalte = append(vorbehalte, mcp.Vorbehalt{
			Datei: block.Datei, Zeile: block.Zeile, Inhalt: block.Inhalt,
		})
	}

	return vorbehalte, nil
}

// Beispielsets nennt die mitgelieferten Sets — **ohne Bestand**.
func (s setangaben) Beispielsets() ([]mcp.Beispielset, error) {
	sets, befund := app.Beispielsetliste()
	if befund != nil {
		return nil, errors.New(befund.Meldung)
	}

	genannt := make([]mcp.Beispielset, 0, len(sets))
	for _, set := range sets {
		genannt = append(genannt, mcp.Beispielset{Name: set.Name, Pfad: set.Pfad})
	}

	return genannt, nil
}
