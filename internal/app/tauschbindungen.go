package app

import "context"

// **Der Tauschabgleich am Fenster** — Werkstück `T4` von Phase 5.6
// (#1469).
//
// **Eine Bindung und nicht vier.** Kandidaten und Vorschlag entstehen aus
// **einer** Rechnung über dieselben zwei Listen; zwei Rufe dafür ließen das
// Fenster zwei Stände nebeneinanderlegen, von denen einer der ältere ist.

// Tauschabgleich sagt, wer was von wem bekommen könnte.
//
// **`behalten` ist die Zahl, die bleiben soll**, `hoechstzahl` der Deckel je
// Seite — **0 heißt die Vorgabe** (zwanzig).
func (b *Bindungen) Tauschabgleich(
	ctx context.Context,
	kennung string, stand FremdstandDTO, behalten, hoechstzahl int,
) TauschabgleichAntwort {
	return b.app(ctx).Tauschabgleich(b.ruf(ctx), kennung, stand, behalten, hoechstzahl)
}
