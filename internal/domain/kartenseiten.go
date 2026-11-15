package domain

// **Wann eine Karte zwei öffentliche Seiten hat** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Die Regel und wo sie herkommt
//
// **Eine Karte hat eine zweite Seite, wenn sie ein nichtleeres `name 2`
// führt.** Das ist MSEs Schreibweise für die Felder der zweiten Seite: an den
// Feldnamen tritt ` 2` — mit Leerzeichen, so wie `card_fields` es schreibt.
//
// **Nachgesehen und nicht erschlossen** (§15): `magic-m15-mainframe-dfc.mse-style`
// erklärt zu jedem Feld ein zweites — `name` und `name 2`, `casting cost` und
// `casting cost 2`, `image` und `image 2` — und setzt sie **nebeneinander**:
// `name` beginnt bei 58 bis 79, `name 2` bei 435 bis 456, und die Karte ist
// 752 Punkte breit statt 375. Gemessen am 2026-09-15 über `M15-Magic-Pack`.
//
// # Warum am Namen und nicht an der Zahl der `2`-Felder
//
// **Weil ein Planeswalker `loyalty cost 2` führt und trotzdem eine Seite
// hat.** Der Satz stand schon in `internal/mtgjson`, wo diese Regel zuerst
// gebraucht wurde; er steht jetzt hier, weil die Frage „hat diese Karte zwei
// Seiten?" eine Frage an die **Karte** ist und nicht an ein Ausfuhrformat
// (§2: fachliche Regeln über eine Entität leben in `domain`).

import "strings"

// ZweiteSeite hängt an einen Feldnamen den Zusatz der zweiten Seite.
//
// **` 2` und nicht `2`** — mit Leerzeichen, wie `card_fields` es schreibt.
func ZweiteSeite(feld string) string { return feld + " 2" }

// HatZweiteSeite sagt, ob diese Karte eine zweite öffentliche Seite führt.
//
// **Zweite öffentliche Seite und nicht Rückseite.** Was jeder sieht, wenn
// eine Karte verdeckt liegt, ist der **Rücken** — für jede Karte derselbe,
// und er steht in keinem Feld. Die zweite Seite ist eine Karte, die jemand
// umdreht; sie hat einen Namen, Kosten und ein Bild.
func (c *CardRow) HatZweiteSeite() bool {
	if c == nil {
		return false
	}

	return WerteHabenZweiteSeite(c.Values)
}

// WerteHabenZweiteSeite ist dieselbe Frage an eine lose Werteliste.
//
// **Für die Wege, die noch keine Zeile haben** — die Ausfuhr nach MTG JSON
// arbeitet auf den Werten, bevor daraus eine [CardRow] wird.
func WerteHabenZweiteSeite(werte []Value) bool {
	for _, wert := range werte {
		if wert.Field == ZweiteSeite("name") &&
			strings.TrimSpace(wert.Text) != "" {
			return true
		}
	}

	return false
}
