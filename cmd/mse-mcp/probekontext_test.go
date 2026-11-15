package main

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// probekontext ist ein Kontext im ersten Profil — dieselbe Vorrichtung wie in
// `internal/mcpnaehte`, denn ein Testhelfer lässt sich nicht importieren.
func probekontext(t *testing.T) context.Context {
	t.Helper()

	return storage.MitProfil(t.Context(), 1)
}
