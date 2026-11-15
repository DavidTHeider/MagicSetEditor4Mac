package rest

// **Einen Deckplatz in den anderen Bereich verschieben** — Befund 7 des
// Zweitblicks über Phase 5.1
// (#1405).
//
// **Eine eigene Datei**, damit `decks.go` unter der Zeilengrenze aus §2
// bleibt.

import (
	"net/http"
)

// deckplatzVerschieben legt einen Platz in den anderen Bereich — in einem Zug.
func (d *Dienst) deckplatzVerschieben(schreiben http.ResponseWriter, anfrage *http.Request) {
	deckID, gutesDeck := kennung(schreiben, anfrage, "deckID")
	platzID, guterPlatz := kennung(schreiben, anfrage, "platzID")

	if !gutesDeck || !guterPlatz {
		return
	}

	var wunsch struct {
		Bereich string `json:"bereich"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	platz, fehler := d.naehte.Decks.DeckplatzVerschieben(
		anfrage.Context(), deckID, platzID, wunsch.Bereich)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, platz)
}
