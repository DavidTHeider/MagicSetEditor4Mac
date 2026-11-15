package service

// **Eine Leihe kommt an den Tisch ihres Verleihers zurück** — P9b
// (#2377).
//
// # Warum der Gast nur die Kennung schickt
//
// **Weil der Abzug hier liegt** — unterschrieben von diesem Rechner, als er
// die Leihe ausstellte. Ein mitgeschickter Abzug, der behauptete, aus diesem
// Bestand zu stammen, ließe den Wirt beliebige Zeilen seines Bestands zeichnen
// und herzeigen. So zeichnet er nur, was er selbst verliehen hat.
//
// # Was geprüft wird
//
// **Dieselbe Prüfung, die der Entleiher beim Vorlegen macht** (`Vorlegen`),
// nur von der anderen Seite: dieser Rechner hat sie ausgestellt, sie ist nicht
// zurückgenommen, und sie gilt **diesem** Gast für **diese** Partie, jetzt.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// VerliehenesDeck gibt den Abzug einer eigenen Leihe an ihrem Tisch.
func (s *Leihstelle) VerliehenesDeck(
	ctx context.Context, kennung, entleiher, partie string,
) (domain.Partiedeck, error) {
	stand, fehler := s.ablage.Leihe(ctx, kennung)
	if fehler != nil {
		return domain.Partiedeck{}, fehler
	}

	switch {
	case !stand.Verliehen:
		return domain.Partiedeck{}, fmt.Errorf(
			"%w: diese Leihe hat nicht dieser Tisch ausgestellt", domain.ErrConflict)
	case stand.Widerrufen:
		return domain.Partiedeck{}, fmt.Errorf(
			"%w: diese Leihe ist zurückgenommen", domain.ErrConflict)
	}

	if fehler := stand.Leihe.Gilt(s.uhr(), entleiher, stand.Abzug.Abdruck,
		domain.Leihbindung{Art: domain.LeiheFuerPartie, Woran: partie},
		domain.RechtSpielen); fehler != nil {
		return domain.Partiedeck{}, fehler
	}

	// **Seine Kartenkennungen sind Zeilen dieses Bestands** (Vermerk 0198).
	abzug := stand.Abzug
	abzug.Quelle = ""

	return abzug, nil
}
