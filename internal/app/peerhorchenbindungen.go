package app

import "context"

// **Das Horchen am Fenster** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Sechs Bindungen, und drei davon sind die zweite Entscheidung
//
// **Weil „an" und „aus" verschiedene Angaben brauchen.** Anfangen braucht eine
// Adresse, Aufhören braucht nichts, und der Stand ist eine Frage. Ein Schalter
// mit einem Wahrheitswert nähme die Adresse auch dann entgegen, wenn er
// ausschaltet — und niemand wüsste, welche gilt.
//
// **Und die zweite Entscheidung ist die Ankündigung.** Erreichbar sein und
// auffindbar sein sind zwei Dinge: das erste gilt dem, der die Adresse kennt,
// das zweite **jedem im Netz**. Wer beides in einen Knopf legte, gäbe mehr
// preis, als jemand gewollt hat.

// HorchStarten macht dieses Profil unter einer Adresse erreichbar.
//
// **`:41234` horcht auf jeder Schnittstelle dieses Rechners**,
// `127.0.0.1:41234` nur auf der Rückschleife. Wer `:0` angibt, bekommt einen
// freien Port — welchen, sagt die Antwort.
func (b *Bindungen) HorchStarten(ctx context.Context, adresse string) Horchantwort {
	return b.app(ctx).HorchStarten(b.ruf(ctx), adresse)
}

// HorchBeenden macht dieses Profil wieder unerreichbar.
func (b *Bindungen) HorchBeenden(ctx context.Context) Horchantwort {
	return b.app(ctx).HorchBeenden(b.ruf(ctx))
}

// HorchStand sagt, ob und wo gehorcht wird — **rein lesend**.
func (b *Bindungen) HorchStand(ctx context.Context) Horchantwort {
	return b.app(ctx).HorchStand(b.ruf(ctx))
}

// SichtbarWerden kündigt dieses Profil im selben Netz an.
//
// **Ein zweiter Schritt und nicht die Folge des ersten**: wer horcht, ist
// erreichbar für den, der seine Adresse kennt; wer sich ankündigt, sagt
// **jedem im Netz**, dass es ihn gibt — samt Peerkennung (N2, #1455).
func (b *Bindungen) SichtbarWerden(ctx context.Context, name string) Horchantwort {
	return b.app(ctx).SichtbarWerden(b.ruf(ctx), name)
}

// SichtbarkeitBeenden nimmt die Ankündigung zurück — **das Horchen bleibt**.
func (b *Bindungen) SichtbarkeitBeenden(ctx context.Context) Horchantwort {
	return b.app(ctx).SichtbarkeitBeenden(b.ruf(ctx))
}

// Nachbarn sagt, wer sich im selben Netz meldet — **rein lesend**.
func (b *Bindungen) Nachbarn(ctx context.Context) NachbarnAntwort {
	return b.app(ctx).Nachbarn(b.ruf(ctx))
}
