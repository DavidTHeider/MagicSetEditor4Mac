// **Die Einstellungen eines Menschen und seine Auskunft im Dienst** —
// Werkstück 8 von Phase 4.12
// (#1351).

package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// Nutzereinstellungen gibt die Einstellungen eines Menschen.
func (n nutzer) Nutzereinstellungen(ctx context.Context, nutzerID string) (
	[]rest.Einstellungswahl, error,
) {
	return alsEinstellungswahlen(
		app.NutzerflaecheKopflos(n.db).Nutzereinstellungen(ctx, nutzerID))
}

// NutzereinstellungSetzen schreibt eine fort und gibt alle zurück.
func (n nutzer) NutzereinstellungSetzen(
	ctx context.Context, nutzerID, schluessel, wert string,
) ([]rest.Einstellungswahl, error) {
	return alsEinstellungswahlen(app.NutzerflaecheKopflos(n.db).
		NutzereinstellungSetzen(ctx, nutzerID, schluessel, wert))
}

// Personendaten gibt die Auskunft, wie sie geschrieben wird.
func (n nutzer) Personendaten(ctx context.Context, nutzerID string) (
	json.RawMessage, error,
) {
	antwort := app.NutzerflaecheKopflos(n.db).Personendaten(ctx, nutzerID)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	roh, fehler := json.Marshal(antwort.Daten)
	if fehler != nil {
		return nil, fmt.Errorf("die Auskunft packen: %w", fehler)
	}

	return roh, nil
}

// alsEinstellungswahlen übersetzt eine Antwort der Fassade.
func alsEinstellungswahlen(antwort app.NutzereinstellungenAntwort) (
	[]rest.Einstellungswahl, error,
) {
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	wahlen := make([]rest.Einstellungswahl, 0, len(antwort.Einstellungen))

	for _, eine := range antwort.Einstellungen {
		wahlen = append(wahlen, rest.Einstellungswahl{
			Schluessel: eine.Schluessel, Wert: eine.Wert, Werte: eine.Werte,
		})
	}

	return wahlen, nil
}
