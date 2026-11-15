package app

import "context"

// **Die Bindungen für den Weg über fremde Netze** — Werkstück N3 von Phase
// 5.4 (#1456).

// FernHorchStarten gibt eine Einladung aus und nimmt Verbindungen entgegen.
//
// **Schreibend** (`docs/SCHREIBWEGE.md`): es öffnet einen Tunnel über ein
// fremdes Relais und macht diesen Rechner für jeden erreichbar, der die
// Einladung bekommt.
func (b *Bindungen) FernHorchStarten(ctx context.Context) Fernantwort {
	return b.app(ctx).FernHorchStarten(b.ruf(ctx))
}

// FernHorchBeenden macht dieses Profil über fremde Netze unerreichbar.
//
// **Und entwertet jede ausgegebene Einladung** — der Serverschlüssel wird
// nicht aufbewahrt.
func (b *Bindungen) FernHorchBeenden(ctx context.Context) Fernantwort {
	return b.app(ctx).FernHorchBeenden(b.ruf(ctx))
}

// FernHorchStand sagt, ob und unter welcher Einladung gehorcht wird —
// **rein lesend**.
func (b *Bindungen) FernHorchStand(ctx context.Context) Fernantwort {
	return b.app(ctx).FernHorchStand(b.ruf(ctx))
}
