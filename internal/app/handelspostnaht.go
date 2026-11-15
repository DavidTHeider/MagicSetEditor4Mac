package app

// **Die Naht des Handelsausgangs** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// **Weil nur hier `internal/storage` stehen darf** (§2) — derselbe Grund wie
// bei [freigabenaht] und [tauschnaht]. Nebenan steht, **was** ein Poststand
// ist; hier steht, **woraus** er gelesen wird.

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// handelsausgang baut die Ablage für diesen Ruf.
func (a *App) handelsausgang() *storage.Ausgangsablage {
	return storage.NeueAusgangsablage(a.laufendeVerbindung())
}
