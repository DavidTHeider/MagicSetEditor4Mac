package app

import "context"

// **Der Weg von der Partie zum Handel** — Werkstück `T7` von Phase 5.6
// (#1472).
//
// **Rein lesend** (`docs/SCHREIBWEGE.md`): es sagt, wer am Tisch saß, und
// ändert nichts. Was daraus wird, entscheidet ein Mensch.

// Handelspartner sagt, welcher bekannte Kontakt auf einem Sitz saß.
func (b *Bindungen) Handelspartner(ctx context.Context, sitz string) HandelspartnerAntwort {
	return b.app(ctx).Handelspartner(b.ruf(ctx), sitz)
}
