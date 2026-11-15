// **Die Einstellungen eines Menschen und seine Auskunft für den MCP-Server** —
// Werkstück 8 von Phase 4.12
// (#1351).

package mcpnaehte

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// Nutzereinstellungen gibt die Einstellungen eines Menschen.
func (n nutzer) Nutzereinstellungen(ctx context.Context, nutzerID string) (
	[]mcp.Nutzereinstellungszeile, error,
) {
	return alsEinstellungszeilen(
		app.NutzerflaecheKopflos(n.db).Nutzereinstellungen(ctx, nutzerID))
}

// NutzereinstellungSetzen schreibt eine fort und gibt alle zurück.
func (n nutzer) NutzereinstellungSetzen(
	ctx context.Context, nutzerID, schluessel, wert string,
) ([]mcp.Nutzereinstellungszeile, error) {
	return alsEinstellungszeilen(app.NutzerflaecheKopflos(n.db).
		NutzereinstellungSetzen(ctx, nutzerID, schluessel, wert))
}

// Personendaten gibt die Auskunft als eingerücktes JSON.
//
// **Eingerückt, weil ein Mensch sie liest** — sie ist die Antwort auf „was
// wisst ihr über mich", und eine Zeile mit zwanzig Feldern beantwortet die
// Frage nicht.
func (n nutzer) Personendaten(ctx context.Context, nutzerID string) (string, error) {
	antwort := app.NutzerflaecheKopflos(n.db).Personendaten(ctx, nutzerID)
	if antwort.Fehler != nil {
		return "", antwort.Fehler.AlsFehler()
	}

	roh, fehler := json.MarshalIndent(antwort.Daten, "", "  ")
	if fehler != nil {
		return "", fmt.Errorf("die Auskunft packen: %w", fehler)
	}

	return string(roh), nil
}

// alsEinstellungszeilen übersetzt eine Antwort der Fassade.
func alsEinstellungszeilen(antwort app.NutzereinstellungenAntwort) (
	[]mcp.Nutzereinstellungszeile, error,
) {
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	zeilen := make([]mcp.Nutzereinstellungszeile, 0, len(antwort.Einstellungen))

	for _, eine := range antwort.Einstellungen {
		zeilen = append(zeilen, mcp.Nutzereinstellungszeile{
			Schluessel: eine.Schluessel, Wert: eine.Wert, Werte: eine.Werte,
		})
	}

	return zeilen, nil
}
