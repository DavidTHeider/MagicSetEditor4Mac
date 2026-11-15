// **Ein Deck aus einem fremden Profil übernehmen, über REST** — Werkstück 12
// von Phase 4.12
// (#1351).
//
// **Ein Weg, erweiternd** (§16, 0033).
//
// # Warum `POST` und nicht `PUT`
//
// **Weil ein zweiter Ruf ein zweites Deck anlegt.** Ein `PUT` sagt „so soll es
// sein"; die Übernahme sagt „hol es her". Sie hat keinen Ort, auf den man
// zweimal zulaufen könnte — dieselbe Überlegung wie bei `POST /decks`.
//
// # Warum unter `/decks` und nicht unter `/profile`
//
// **Weil das Ergebnis ein Deck im eigenen Profil ist.** Das fremde Profil ist
// die Quelle und steht deshalb im Rumpf, nicht im Weg: der Weg nennt, wo
// etwas entsteht.

package rest

import (
	"context"
	"net/http"
)

// Uebernommenes ist das übernommene Deck samt dem, was nicht mitkam.
type Uebernommenes struct {
	DeckID  string   `json:"deckID"`
	Name    string   `json:"name"`
	Spiel   string   `json:"spiel"`
	Plaetze int      `json:"plaetze"`
	Befunde []string `json:"befunde"`
}

// Uebernahmewunsch ist der Rumpf: woher.
type Uebernahmewunsch struct {
	QuellProfil string `json:"quellProfil"`
	QuellDeck   string `json:"quellDeck"`
}

// Fremduebernahme holt ein Deck aus einem fremden Profil — und nennt die
// Decks dort zur Wahl (UX-08.AK-04, #1680).
type Fremduebernahme interface {
	DeckUebernehmen(ctx context.Context, quellProfil, quellDeck string) (
		Uebernommenes, error)
	DeckeZurUebernahme(ctx context.Context, quellProfil string) (any, error)
}

// deckUebernehmen holt es.
//
// **201 und nicht 200**: es ist etwas entstanden, und die Antwort nennt seine
// Kennung — dieselbe Zusage wie bei `POST /decks`.
func (d *Dienst) deckUebernehmen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf Uebernahmewunsch

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	uebernommen, fehler := d.naehte.Deckuebernahme.DeckUebernehmen(
		anfrage.Context(), rumpf.QuellProfil, rumpf.QuellDeck)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, uebernommen)
}

// deckwahl nennt die Decks eines fremden Profils — zur Wahl für die
// Übernahme.
//
// **Unter `/profile`**, anders als die Übernahme selbst: gelesen wird
// **dort**, und der Weg nennt, wo. Dieselbe Erlaubnis wie die Übernahme,
// geprüft, bevor gelesen wird (Vermerk 0071).
func (d *Dienst) deckwahl(schreiben http.ResponseWriter, anfrage *http.Request) {
	decks, fehler := d.naehte.Deckuebernahme.DeckeZurUebernahme(
		anfrage.Context(), anfrage.PathValue("profilID"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, decks)
}
