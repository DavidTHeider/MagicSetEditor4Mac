package app

import "context"

// **Die Freigabe am Fenster** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Drei Bindungen und kein Schalter
//
// **Dieselbe Gestalt wie beim Horchposten** (N2): Freigeben braucht eine
// Partie, einen Sitz und die Gastsitze, Beenden braucht nichts, und der Stand
// ist eine Frage. Ein Wahrheitswert nähme die Sitze auch beim Ausschalten
// entgegen — und niemand könnte sagen, was er damit tut.

// PartieFreigeben öffnet eine Partie für Gäste.
func (b *Bindungen) PartieFreigeben(
	ctx context.Context,
	kennung, fuer string, gastsitze []string,
) PartiefreigabeAntwort {
	return b.app(ctx).PartieFreigeben(b.ruf(ctx), kennung, fuer, gastsitze)
}

// PartieFreigabeBeenden schließt sie wieder.
func (b *Bindungen) PartieFreigabeBeenden(ctx context.Context) PartiefreigabeAntwort {
	return b.app(ctx).PartieFreigabeBeenden(b.ruf(ctx))
}

// PartieFreigabeStand sagt, was gerade offensteht.
func (b *Bindungen) PartieFreigabeStand(ctx context.Context) PartiefreigabeAntwort {
	return b.app(ctx).PartieFreigabeStand(b.ruf(ctx))
}

// Partien nennt die Partien dieses Profils —
// #1600.
//
// **Ohne Argument**, weil die Frage keines hat: welche Partien es gibt, hängt
// am Profil, und das steht im Kontext.
func (b *Bindungen) Partien(ctx context.Context) PartienAntwort {
	return b.app(ctx).Partien(b.ruf(ctx))
}
