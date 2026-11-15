package app

import "context"

// **Die Freigaben am Fenster** — Werkstück `T3` von Phase 5.6
// (#1468).
//
// **Eine eigene Datei**, weil `peerkontaktbindungen.go` eine andere Frage
// beantwortet: dort steht, **wer** etwas darf, hier, **was** er dabei sieht
// (§2).

// Sammlungsfreigaben gibt, was einem Gegenüber erteilt ist — **rein lesend**.
func (b *Bindungen) Sammlungsfreigaben(
	ctx context.Context,
	kennung string,
) SammlungsfreigabenAntwort {
	return b.app(ctx).Sammlungsfreigaben(b.ruf(ctx), kennung)
}

// SammlungFreigeben erteilt einem Gegenüber eine Sicht.
func (b *Bindungen) SammlungFreigeben(
	ctx context.Context,
	freigabe SammlungsfreigabeDTO,
) SammlungsfreigabenAntwort {
	return b.app(ctx).SammlungFreigeben(b.ruf(ctx), freigabe)
}

// SammlungsfreigabeWiderrufen nimmt eine Sicht zurück.
func (b *Bindungen) SammlungsfreigabeWiderrufen(
	ctx context.Context,
	kennung, zweck, umfang string,
) SammlungsfreigabenAntwort {
	return b.app(ctx).SammlungsfreigabeWiderrufen(
		b.ruf(ctx), kennung, zweck, umfang)
}

// Freigabesicht zeigt, was ein Gegenüber wirklich sähe — **rein lesend**.
//
// **Über denselben Weg wie die Antwort an ein Gegenüber.** Eine zweite
// Rechnung daneben zeigte, was sie selbst ausrechnet, und nicht, was
// hinausgeht.
func (b *Bindungen) Freigabesicht(
	ctx context.Context,
	kennung, zweck string,
) FreigabesichtAntwort {
	return b.app(ctx).Freigabesicht(b.ruf(ctx), kennung, zweck)
}
