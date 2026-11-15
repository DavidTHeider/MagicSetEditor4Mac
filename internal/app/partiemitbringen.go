package app

// **Was ein eingeladener Gast an den Tisch mitbringt** — P9b
// (#2377).
//
// Ein eigenes Deck geht als Abzug mit, wie am eigenen Tisch eingefroren; eine
// Leihe nur bei ihrer Kennung — den Abzug hält der Wirt, der sie ausgestellt
// hat. **Ob es genau eines ist, sagt der Kern** (`domain.Mitgebracht`); hier
// wird nur übersetzt.

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// mitbringen macht aus dem Wunsch des Fensters, was mitgeht — `nil`, wenn
// nichts gewählt ist.
func (a *App) mitbringen(
	ctx context.Context, mit PartiemitbringDTO,
) (*domain.Mitgebracht, error) {
	if mit.DeckID == "" && mit.LeiheID == "" {
		return nil, nil //nolint:nilnil // nichts mitzubringen ist kein Fehler
	}

	gebracht := domain.Mitgebracht{Leihe: mit.LeiheID}

	if mit.DeckID != "" {
		deckID, fehler := strconv.ParseInt(mit.DeckID, 10, 64)
		if fehler != nil {
			return nil, domain.MitGrund(domain.ErrValidation, fehler,
				"das mitgebrachte Deck lesen")
		}

		eigenes, fehler := a.partieaufbau().Mitbringen(ctx, deckID)
		if fehler != nil {
			return nil, fehler
		}

		gebracht.Deck, gebracht.Spiel = eigenes.Deck, eigenes.Spiel
	}

	if fehler := gebracht.Pruefe(); fehler != nil {
		return nil, fehler
	}

	return &gebracht, nil
}

// mitgebrachtMerken trägt ein, womit sich der Gast dieses Beitritts setzt.
func (a *App) mitgebrachtMerken(mit *domain.Mitgebracht) {
	a.partie.beitritt.schloss.Lock()
	defer a.partie.beitritt.schloss.Unlock()

	a.partie.beitritt.mitgebracht = mit
}

// mitgebracht gibt es zurück.
func (a *App) mitgebracht() *domain.Mitgebracht {
	a.partie.beitritt.schloss.Lock()
	defer a.partie.beitritt.schloss.Unlock()

	return a.partie.beitritt.mitgebracht
}
