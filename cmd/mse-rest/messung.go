package main

// **Jeder Ruf an den Dienst wird gemessen** — OPS-05
// (#1722):
// Verteilung immer, Spur nur, wenn sie eingeschaltet ist. Eingerichtet wird
// die Spur beim Start nach den Einstellungen des Geräts, wie das Protokoll am
// Fenster.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// rufMessen gibt eine Messung des Dienstes an die Anwendung.
func rufMessen(ctx context.Context, messung rest.Messung) {
	app.ErfasseRuf(ctx, app.Rufmessung{
		Flaeche: domain.ModusDienst, Weg: messung.Weg, Schreibend: messung.Schreibend,
		Dauer: messung.Dauer, Ausgang: messung.Ausgang,
	})
}
