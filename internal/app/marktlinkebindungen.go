package app

import "context"

// **Die Marktlinks am Fenster** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// **Zwei Bindungen**: eine, die nichts tut außer Wege zu nennen, und eine,
// die dieses Programm verlässt.

// Marktlinks gibt je Anbieter den genauesten Weg zu einer Karte.
//
// **Rein lesend** (`docs/SCHREIBWEGE.md`) — es entsteht nichts, und es geht
// nichts hinaus.
func (b *Bindungen) Marktlinks(ctx context.Context, printing, name string) MarktlinkAntwort {
	return b.app(ctx).Marktlinks(b.ruf(ctx), printing, name)
}

// MarktlinkOeffnen macht einen Weg im Standardbrowser auf.
//
// **Schreibend im Sinne von `docs/SCHREIBWEGE.md`** — nicht im Bestand,
// sondern **nach außen**: dieser Ruf verlässt das Programm, und er geschieht
// nur auf einen Klick.
func (b *Bindungen) MarktlinkOeffnen(ctx context.Context, anbieter, adresse string) SchreibAntwort {
	return b.app(ctx).MarktlinkOeffnen(b.ruf(ctx), anbieter, adresse)
}
