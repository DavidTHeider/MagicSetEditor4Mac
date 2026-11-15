package app

import "context"

// **Die Peeridentität am Fenster** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Drei Bindungen und nicht eine
//
// **Weil sie drei verschiedene Dinge tun**, und zwei davon schreiben: fragen,
// anlegen, ablösen. Ein Ruf, der alles dreies je nach Lage täte, legte still
// eine Identität an — und genau das schließt das Issue aus.

// Peeridentitaet sagt, womit sich dieses Profil ausweist — **rein lesend**.
func (b *Bindungen) Peeridentitaet(ctx context.Context) PeeridentitaetAntwort {
	return b.app(ctx).Peeridentitaet(b.ruf(ctx))
}

// PeeridentitaetSicherstellen legt eine an, wenn es noch keine gibt.
//
// **Bei einem Verlust oder einem Widerspruch antwortet sie mit einem
// Konflikt** und sagt, was zu tun ist.
func (b *Bindungen) PeeridentitaetSicherstellen(ctx context.Context) PeeridentitaetAntwort {
	return b.app(ctx).PeeridentitaetSicherstellen(b.ruf(ctx))
}

// PeeridentitaetRotieren legt ausdrücklich eine neue an.
//
// **Wer sie ruft, verlangt von jedem bekannten Kontakt eine neue Prüfung** —
// die Antwort sagt es, und die alte Kennung steht dabei.
func (b *Bindungen) PeeridentitaetRotieren(ctx context.Context) PeeridentitaetAntwort {
	return b.app(ctx).PeeridentitaetRotieren(b.ruf(ctx))
}
