package rest_test

// **Die willigen Antworten der Decklisteneinfuhr** — Werkstück `T8` von
// Phase 5.6
// (#1473).
//
// **Eine eigene Datei**, weil `willigerdoppel_test.go` mit ihnen über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort die Antworten
// auf alles Übrige, hier die auf die Deckliste.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// DecklisteVorschau gibt eine leere Vorschau samt Stand.
//
// **Mit Stand und mit leeren Listen**: der Dienst reicht die Gestalt durch,
// und ein `null` statt `[]` wäre eine Vorrichtung, die großzügiger ist als
// der Betrieb (§3).
func (williger) DecklisteVorschau(context.Context, int64, string) (any, error) {
	return map[string]any{
		"deckID": 1, "stand": "abc", "zeilen": []any{}, "befunde": []string{},
	}, nil
}

// DecklisteUebernehmen meldet einen gelegten Platz.
func (williger) DecklisteUebernehmen(
	context.Context, int64, string, []rest.Deckeinfuhrzeile,
) (any, error) {
	return map[string]any{"gelegt": 1, "stueck": 4, "stand": "cde"}, nil
}
