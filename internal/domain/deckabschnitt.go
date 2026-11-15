package domain

// **Welche Überschrift welchen Bereich meint** — Werkstück `T8` von Phase
// 5.6 (#1473).
//
// # Warum das hier steht und nicht im Zerteiler
//
// **Weil `internal/sammelliste` nicht weiß, wofür gelesen wird.** Es liest
// eine Liste in Zeilen — für die Sammlung wie fürs Deck; Bereiche hat nur das
// Deck. Die Zuordnung ist eine fachliche Regel über **eine** Entität, und
// damit gehört sie hierher (§2).

// nichtImDeck sind die Überschriften, unter denen keine Deckkarte steht.
//
// **Eine Aufzählung und keine Regel.** Wer sie herleiten wollte, müsste
// wissen, was ein Deckbaukasten sonst noch schreibt — und eine unbekannte
// Überschrift ist im Zweifel das Deck und keine Merkliste.
var nichtImDeck = map[string]bool{
	"maybeboard": true,

	// **Spielmarken sind keine Karten des Decks** — sie entstehen im Spiel,
	// und kein Format zählt sie mit.
	"tokens": true,

	// **`about` trägt den Namen des Decks und den seines Erbauers** — Arena
	// schreibt es so. Was darunter steht, ist Text.
	"about": true,
}

// DeckbereichAus sagt, wohin eine Zeile unter dieser Überschrift gehört.
//
// **Der zweite Wert ist das Nein.** Eine Merkliste hat keinen Bereich, und
// ein leerer Bereich wäre eine stille Zuordnung ins Hauptdeck.
func DeckbereichAus(abschnitt string) (Deckbereich, bool) {
	if nichtImDeck[abschnitt] {
		return "", false
	}

	if abschnitt == string(BereichSideboard) {
		return BereichSideboard, true
	}

	return BereichHauptdeck, true
}
