package app

// **Der Beitritt am Fenster** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Vier Bindungen, und die vierte ist die eigentliche
//
// **`PartieWuenschen` ist der Zug.** Die anderen drei setzen sich hin, stehen
// auf und sagen, wo man sitzt — gespielt wird mit ihr. Und sie gibt **kein
// Ergebnis** zurück: was aus einem Wunsch wird, entscheidet der Wirt, und es
// kommt über den Meldeweg (`partietischmelder.go`) und nicht als Antwort.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// PartieBeitreten setzt sich an den Tisch eines Gegenübers — mit dem, was
// ein eingeladener Gast mitbringt (P9b, #2377).
func (b *Bindungen) PartieBeitreten(
	ctx context.Context,
	gegenueber, partie string, mit PartiemitbringDTO,
) PartiebeitrittAntwort {
	return b.app(ctx).PartieBeitreten(b.ruf(ctx), gegenueber, partie, mit)
}

// PartieBeitrittBeenden steht vom Tisch auf.
func (b *Bindungen) PartieBeitrittBeenden(ctx context.Context) PartiebeitrittAntwort {
	return b.app(ctx).PartieBeitrittBeenden(b.ruf(ctx))
}

// PartieBeitrittStand sagt, an welchem Tisch dieses Programm sitzt.
func (b *Bindungen) PartieBeitrittStand(ctx context.Context) PartiebeitrittAntwort {
	return b.app(ctx).PartieBeitrittStand(b.ruf(ctx))
}

// PartieWuenschen schickt einen Spielbefehl an den Wirt.
//
// **Dieselbe Gestalt wie `PartieBefehl`** (G1) — bis auf die Antwort: dort
// kommt zurück, was geschehen ist, hier nur, ob der Wunsch hinausging.
func (b *Bindungen) PartieWuenschen(
	ctx context.Context,
	befehlID, art string, nutzlast map[string]any,
) PartiebeitrittAntwort {
	return b.app(ctx).PartieWuenschen(b.ruf(ctx), domain.Spielbefehl{
		Kennung: befehlID, Art: art, Nutzlast: nutzlast,
	})
}
