// **Ein Deck aus einem fremden Profil übernehmen, über MCP** — Werkstück 12
// von Phase 4.12
// (#1351).
//
// **Ein Werkzeug, erweiternd** (§16, 0033).
//
// # Warum es nicht idempotent ist
//
// **Weil ein zweiter Ruf ein zweites Deck anlegt.** Die Übernahme ist eine
// Kopie und keine Zuordnung; sie hat kein „so soll es sein", auf das man
// zweimal zulaufen könnte.

package mcp

import (
	"context"
	"encoding/json"
)

// Deckuebernahmepflege ist die Naht zur Übernahme.
type Deckuebernahmepflege interface {
	DeckUebernehmen(ctx context.Context, quellProfil, quellDeck string) (
		Uebernommenes, error)

	// DeckeZurUebernahme nennt die Decks eines fremden Profils — hinter
	// derselben Erlaubnis (UX-08.AK-04, #1680).
	DeckeZurUebernahme(ctx context.Context, quellProfil string) (any, error)
}

// Uebernommenes ist das übernommene Deck samt dem, was nicht mitkam.
type Uebernommenes struct {
	DeckID string `json:"deckID"`
	Name   string `json:"name"`
	Spiel  string `json:"spiel"`

	// Plaetze ist die Zahl der übernommenen Plätze.
	Plaetze int `json:"plaetze"`

	// Befunde sagt, was **nicht** mitkam — eine Auskunft, keine Absage.
	Befunde []string `json:"befunde"`
}

// deckuebernahmewerkzeug gibt die lesende und die schreibende Hälfte —
// getrennt, weil die lesenden in der Liste vor den schreibenden stehen.
//
// **Ohne Naht keines** — dieselbe Zusage wie bei den Decks: ein Prozess, der
// sie nicht hat, bietet das Werkzeug gar nicht erst an, statt es mit einem
// Absturz anzubieten.
func deckuebernahmewerkzeug(pflege Deckuebernahmepflege) (lesend, schreibend []Werkzeug) {
	if pflege == nil {
		return nil, nil
	}

	return []Werkzeug{deckwahlwerkzeug(pflege)}, []Werkzeug{{
		Name: "schreiben.deck.uebernehmen", Art: Schreibend,
		Erklaerung: "ÄNDERT den Bestand: HOLT ein Deck aus dem Profil eines " +
			"ANDEREN in das eigene und schreibt auf, woher es kam. Holen " +
			"darf, wer in das fremde Profil wechseln darf: hat dessen Eigner " +
			"einen zweiten Faktor, braucht der Mensch, für den dieser Server " +
			"läuft, dort eine Freigabe — »leser« genügt, siehe " +
			"»lesen.freigaben«. NICHT mit: ein Platz, der auf eine EIGENE " +
			"Karte des anderen Profils zeigt, und das Format; beides sind " +
			"Zeilen dort drüben. Was nicht mitkam, steht unter »befunde« — " +
			"das Deck ist trotzdem da. Aus dem EIGENEN Profil lässt sich " +
			"nichts übernehmen.",
		Schema: objekt(map[string]any{
			"quellProfil": text(
				"Die Kennung des fremden Profils aus »lesen.nutzer«."),
			"quellDeck": text("Die Kennung des Decks in jenem Profil."),
		}, "quellProfil", "quellDeck"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				QuellProfil string `json:"quellProfil"`
				QuellDeck   string `json:"quellDeck"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			uebernommen, err := pflege.DeckUebernehmen(
				ctx, argumente.QuellProfil, argumente.QuellDeck)
			if err != nil {
				return "", err
			}

			return alsJSON(uebernommen)
		},
	}}
}

// deckwahlwerkzeug nennt die Decks eines fremden Profils (UX-08.AK-04,
// #1680).
//
// **Die lesende Hälfte der Übernahme**: ohne sie war die Kennung eines Decks
// von drüben nirgends zu erfahren — ein Modell musste sie raten.
func deckwahlwerkzeug(pflege Deckuebernahmepflege) Werkzeug {
	return Werkzeug{
		Name: "lesen.decks.uebernehmbar", Art: Lesend,
		Erklaerung: "Nennt die Decks im Profil eines ANDEREN — zur Wahl für " +
			"»schreiben.deck.uebernehmen«. Dieselbe Erlaubnis wie die " +
			"Übernahme, geprüft BEVOR gelesen wird: wer nicht hinein darf, " +
			"erfährt nicht, welche Decks es dort gibt. Ohne Plätze. " +
			"ÄNDERT NICHTS.",
		Schema: objekt(map[string]any{
			"quellProfil": text(
				"Die Kennung des fremden Profils aus »lesen.nutzer«."),
		}, "quellProfil"),
		Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
			var argumente struct {
				QuellProfil string `json:"quellProfil"`
			}

			if err := lies(roh, &argumente); err != nil {
				return "", err
			}

			decks, err := pflege.DeckeZurUebernahme(ctx, argumente.QuellProfil)
			if err != nil {
				return "", err
			}

			return alsJSON(decks)
		},
	}
}
