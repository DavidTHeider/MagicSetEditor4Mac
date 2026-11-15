package domain

// **Der Stand eines Decks als ein Wort** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// # Wozu
//
// **Damit eine Vorschau an den Stand gebunden werden kann, aus dem sie
// entstand.** Eine Decklisteneinfuhr liest, zeigt, wartet auf ein Ja und
// schreibt dann — und zwischen Lesen und Schreiben liegt ein Mensch. Wer in
// dieser Zeit eine Karte hineinlegt, hat ein anderes Deck vor sich als das,
// über das entschieden wurde (§7, Regel 2).
//
// # Warum ein Abdruck und nicht ein Zähler
//
// **Weil es keinen gibt.** [Deck.Fassung] steigt beim Ändern des
// Deckdatensatzes und ausdrücklich **nicht** beim Legen eines Platzes; die
// Plätze liegen in einer eigenen Tabelle mit eigener Eindeutigkeit. Einen
// zweiten Zähler einzuführen hieße, jede Stelle zu finden, die einen Platz
// schreibt — der Abdruck kommt ohne das aus, weil er den Stand **liest**
// statt ihn fortzuschreiben.
//
// # Und warum er die Bezeichnungen auslässt
//
// **Weil sie von außen kommen** (§16, Regel 5): Name, Manakosten, Typzeile
// und Manawert holt der Dienst bei Scryfall, und ohne Netz bleiben sie leer.
// Ein Abdruck, der sie mitzählte, wäre bei jedem zweiten Aufruf ein anderer —
// und die Einfuhr scheiterte daran, dass das Netz gerade weg war.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// Deckstand gibt einen Abdruck über Fassung und Plätze eines Decks.
//
// **Er ist keine Zusage über Geheimhaltung.** Er unterscheidet Stände; wer
// aus ihm den Inhalt zurückrechnen wollte, hätte die Deckliste ohnehin.
func Deckstand(deck Deck) string {
	zeilen := make([]string, 0, len(deck.Plaetze))

	for _, platz := range deck.Plaetze {
		zeilen = append(zeilen, fmt.Sprintf("%d|%d|%s|%d|%s",
			platz.ID, platz.KarteID, platz.Printing, platz.Menge,
			platz.Bereich))
	}

	// **Geordnet, weil die Reihenfolge der Abfrage gehört und nicht dem
	// Deck.** Wer sie mitzählte, ließe eine Einfuhr scheitern, weil jemand
	// anders sortiert hat.
	sort.Strings(zeilen)

	summe := sha256.Sum256([]byte(fmt.Sprintf("%d\n%s",
		deck.Fassung, strings.Join(zeilen, "\n"))))

	return hex.EncodeToString(summe[:])
}
