package app

import "context"

// **Der Poststand am Fenster** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// **Eine eigene Datei**, weil `peerkontaktbindungen.go` eine andere Frage
// beantwortet: dort steht, **wer** etwas darf, hier, **was** noch aussteht.

// Handelspoststand sagt, was zu einem Gegenüber noch aussteht — **rein
// lesend**.
func (b *Bindungen) Handelspoststand(ctx context.Context, kennung string) PoststandAntwort {
	return b.app(ctx).Handelspoststand(b.ruf(ctx), kennung)
}
