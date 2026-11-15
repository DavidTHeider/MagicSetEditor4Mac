package service

// **Ein Deck als Liste** — NET-04
// (#1708).
//
// **Die Umkehrung von [Deckeinfuhr.Vorschau]**: was hier hinausgeht, liest
// die Einfuhr wieder — Hauptdeck und Sideboard unter den Überschriften, die
// `sammelliste.Lies` kennt, jede Zeile wie `sammelliste.SchreibeListe` sie
// schreibt. **Verlustbehaftet wie jede Liste**: Zuordnungen zu Exemplaren,
// Format und Notiz des Decks gehen nicht mit.

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sammelliste"
)

// Deckliste gibt ein Deck als Text in der Gestalt, die die Einfuhr versteht.
func Deckliste(deck domain.Deck) string {
	var gebaut strings.Builder

	for _, abschnitt := range []struct {
		bereich      domain.Deckbereich
		ueberschrift string
	}{
		{domain.BereichHauptdeck, "Deck"},
		{domain.BereichSideboard, "Sideboard"},
	} {
		zeilen := []sammelliste.Ausfuhrzeile{}

		for _, platz := range deck.Plaetze {
			if platz.Bereich != abschnitt.bereich {
				continue
			}

			zeilen = append(zeilen, sammelliste.Ausfuhrzeile{
				Printing: platz.Printing, Menge: platz.Menge, Name: platz.Bezeichnung,
			})
		}

		if len(zeilen) == 0 {
			continue
		}

		// **Eine Leerzeile zwischen den Abschnitten**, wie eine Liste aus
		// einem Deckbaukasten sie hat.
		if gebaut.Len() > 0 {
			gebaut.WriteString("\n")
		}

		gebaut.WriteString(abschnitt.ueberschrift + "\n")
		gebaut.WriteString(sammelliste.SchreibeListe(zeilen))
	}

	return gebaut.String()
}
