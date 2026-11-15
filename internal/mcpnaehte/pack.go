// Das Zufallspack für den MCP-Server
// (#889).

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// pack führt die zwei Werkzeuge an die Naht dieser Schicht.
type pack struct{ db *sql.DB }

func (p pack) Packarten(ctx context.Context, setID int64) (mcp.Packangebot, error) {
	weg, err := app.PackUeber(p.db)
	if err != nil {
		return mcp.Packangebot{}, err
	}

	antwort, lesefehler := weg.Arten(ctx, setID)
	if lesefehler != nil {
		return mcp.Packangebot{}, lesefehler
	}

	arten := make([]mcp.Packart, 0, len(antwort.Arten))
	for _, art := range antwort.Arten {
		arten = append(arten,
			mcp.Packart{Name: art.Name, Waehlbar: art.Waehlbar, Karten: art.Karten})
	}

	return mcp.Packangebot{Arten: arten, Befunde: antwort.Befunde}, nil
}

func (p pack) PackZiehen(ctx context.Context, setID int64, art, same string) (
	mcp.Packziehung, error,
) {
	weg, err := app.PackUeber(p.db)
	if err != nil {
		return mcp.Packziehung{}, err
	}

	antwort, ziehfehler := weg.Ziehen(ctx, setID, art, same)
	if ziehfehler != nil {
		return mcp.Packziehung{}, ziehfehler
	}

	karten := make([]mcp.Packkarte, 0, len(antwort.Karten))
	for _, karte := range antwort.Karten {
		karten = append(karten, mcp.Packkarte{ID: karte.ID, Name: packname(karte)})
	}

	return mcp.Packziehung{Karten: karten, Befunde: antwort.Befunde}, nil
}

// packname gibt den Namen einer gezogenen Karte.
//
// **Leer ist kein Fehler**: eine Karte ohne `name` gibt es, und ein
// Platzhalter verschleierte, dass sie so im Bestand steht.
func packname(karte app.KarteDTO) string {
	for _, wert := range karte.Werte {
		if wert.Feld == "name" {
			return wert.Wert
		}
	}

	return ""
}
