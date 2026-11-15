package main

// **Die Meta-Auskunft des Dienstes** — Werkstück O9 von Phase 5.2
// (#1433).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// metaauskunft reicht Fassung und Plattform an den Dienst.
type metaauskunft struct{}

func (metaauskunft) Meta(ctx context.Context) (any, error) {
	return app.NeueApp("").Meta(ctx), nil
}
