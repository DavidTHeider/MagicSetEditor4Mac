package app

import "context"

// **Teilen am Fenster** — NET-04 (#1708). Der Grund, warum es nur hier steht,
// steht im Kopf von `teilen.go`.

// TeilenVorbereiten legt eine Kopie ab und sagt, was darin steht.
func (b *Bindungen) TeilenVorbereiten(ctx context.Context, art, bezug string) TeilenAntwort {
	return b.app(ctx).TeilenVorbereiten(b.ruf(ctx), art, bezug)
}

// TeilenAnbieten öffnet den Teilen-Dialog von macOS.
func (b *Bindungen) TeilenAnbieten(ctx context.Context, kennung string) TeilenAntwort {
	return b.app(ctx).TeilenAnbieten(b.ruf(ctx), kennung)
}

// TeilenVerwerfen räumt eine Teilung weg.
func (b *Bindungen) TeilenVerwerfen(ctx context.Context, kennung string) TeilenAntwort {
	return b.app(ctx).TeilenVerwerfen(b.ruf(ctx), kennung)
}
