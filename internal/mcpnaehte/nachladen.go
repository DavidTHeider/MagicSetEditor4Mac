package mcpnaehte

// **Geschrieben ist geschrieben — an den Werkzeugen** (§7, Regel 3;
// #1960,
// #1970).

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// alsNachladefehler gibt den Nachladefehler in der Gestalt des Werkzeugs —
// `nil`, wenn keiner da ist. **Der Satz ist schon übersetzt** (§8).
func alsNachladefehler(nachladen *app.Fehler) *mcp.Nachladefehler {
	if nachladen == nil {
		return nil
	}

	return &mcp.Nachladefehler{Code: string(nachladen.Code), Meldung: nachladen.Meldung}
}
