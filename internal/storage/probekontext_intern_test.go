package storage

// **Der Probekontext für die internen Läufe** — Werkstück 2 von Phase 4.12
// (#1351).
//
// **Zweimal derselbe Helfer, und das ist kein Versehen**: die einen Läufe
// liegen in `package storage_test`, die anderen in `package storage`, und ein
// Paket kann den Helfer des anderen nicht sehen. Er steht deshalb hier noch
// einmal — vier Zeilen, die sich nicht teilen lassen.

import (
	"context"
	"testing"
)

// probekontext gibt einen Kontext, der auf das **erste** Profil zeigt.
//
// **Die Eins ist keine Annahme, sondern die Wanderung**: `0034` legt einen
// Nutzer mit einem Profil an, und beide bekommen die erste Kennung.
func probekontext(t *testing.T) context.Context {
	t.Helper()

	return MitProfil(t.Context(), 1)
}
