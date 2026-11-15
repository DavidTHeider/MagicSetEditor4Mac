package app

// **Fortschritt für die Wurzeln ohne Fenster** — FMT-03.AK-02
// (#1692).
//
// `cmd/mse`, `cmd/mse-mcp` und `cmd/mse-rest` dürfen `internal/service`
// nicht kennen (§2); wo der Melder im Kontext liegt, weiß `service`. Dieselbe
// Gestalt wie [MitVorgang].

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// MitFortschritt hängt einen Melder an den Kontext eines Rufs.
func MitFortschritt(ctx context.Context, melder domain.Fortschrittsmelder) context.Context {
	return service.MitFortschritt(ctx, melder)
}
