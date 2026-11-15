package app

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// probekontext gibt einen Kontext mit dem ersten Profil.
//
// **Weil jede Wurzelabfrage darauf klemmt** (#1351). Im Betrieb hängt
// [App.imProfil] es an jede Bindung, gespeist aus dem Zeiger, den
// [App.Starten] gelesen hat. Eine Vorrichtung, die eine Naht **unmittelbar**
// ruft, geht daran vorbei und muss es selbst tun.
//
// **Profil 1 und nicht der Zeiger**: die Wanderung legt es an, jede
// Vorrichtung kennt es, und ein Test, der erst nachschlägt, prüfte den
// Nachschlag mit.
func probekontext(t *testing.T) context.Context {
	t.Helper()

	return storage.MitProfil(t.Context(), 1)
}
