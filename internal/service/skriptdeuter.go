package service

// **Der Deuter, den die Verrechnung ruft** —
// #1625.
//
// # Warum er hier steht und nicht in der Domäne
//
// **Weil `internal/domain` nichts aus dem Projekt importiert** (§2) — und ein
// Auswerter ist ein Sprachwerkzeug: `internal/msescript` kennt nicht einmal
// `domain`. Die Schnittstelle steht dort, wo sie gebraucht wird
// ([domain.Wirkungsdeuter]), erfüllt wird sie hier.
//
// # Und warum er je Partie eingesetzt wird und nicht je Ruf
//
// **Weil die Verrechnung mitten in `Passen` geschieht.** Wer den Deuter je
// Ruf reichte, müsste ihn durch `Passen`, `Weiter` und jeden ihrer Rufer
// ziehen; eingesetzt am Zustand ist er da, wo er gebraucht wird — und
// [domain.Partie.MitDeuter] sagt, dass er nicht mit in den Bestand reist.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// skriptdeuter liest ein Kartenskript über [Kartenskript].
type skriptdeuter struct{ skript *Kartenskript }

// Wirkungen gibt, was das Skript bewirken will — oder einen Befund.
//
// **Ein gescheitertes Skript ist kein Abbruch** (§8): der Zauber verrechnet,
// er bewirkt nichts, und der Strom sagt warum.
//
// **Der Kontext entsteht hier**, weil die Domäne keinen führt: sie rechnet,
// sie ruft nicht hinaus. Die **Frist** bringt [Kartenskript] mit, und sie ist
// das, was diesen Ruf begrenzt.
func (d skriptdeuter) Wirkungen(
	skript string, lage domain.Skriptlage,
) (wirkungen []domain.Spielwirkung, befund string) {
	gedeutet, gescheitert := d.skript.Deuten(
		context.Background(), skript, Skriptumfeld{
			Beherrscher: lage.Beherrscher,
			Objekt:      lage.Objekt,
			Ziele:       lage.Ziele,
		})
	if gescheitert != nil {
		return nil, SkriptBefund(gescheitert)
	}

	return gedeutet, ""
}
