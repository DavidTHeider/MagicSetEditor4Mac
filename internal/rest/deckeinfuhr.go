package rest

// **Die Decklisteneinfuhr im Dienst** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// # Zwei Wege, weil es zwei Schritte sind
//
// `POST …/liste/vorschau` liest und **schreibt nichts**; es nennt den Stand,
// an den es gebunden ist. `POST …/liste` legt, was bestätigt ist, und
// verlangt denselben Stand zurück.
//
// **`POST` für die Vorschau und nicht `GET`**: eine Deckliste steht im Rumpf
// und nicht in der Adresse — sie ist dreihundert Zeilen lang, und was jemand
// sammelt, gehört nicht in ein Protokoll (§11).

import (
	"context"
	"net/http"
)

// Deckeinfuhrnaht ist, was der Dienst für eine Deckliste braucht.
//
// **Die Methoden heißen `DecklisteVorschau` und `DecklisteUebernehmen`** und
// nicht bloß `Vorschau` und `Uebernehmen` — derselbe Grund wie bei
// [Listeneinfuhrwerkstatt]: in diesem Paket erfüllt **ein** Doppel jede Naht,
// und beide kurzen Namen sind schon vergeben.
//
// **Sie geben `any` zurück**, wie die Listeneinfuhr daneben: die Gestalt
// kommt aus `internal/app` und wird hier nicht noch einmal geschrieben. Zwei
// Abbildungen desselben Verbunds liefen auseinander.
type Deckeinfuhrnaht interface {
	DecklisteVorschau(ctx context.Context, deckID int64, text string) (any, error)
	DecklisteUebernehmen(ctx context.Context, deckID int64, stand string,
		zeilen []Deckeinfuhrzeile) (any, error)
}

// Deckeinfuhrzeile ist eine Zeile, für die entschieden ist.
type Deckeinfuhrzeile struct {
	Printing string `json:"printing"`
	Menge    int    `json:"menge"`
	Bereich  string `json:"bereich"`
}

// decklisteVorschau zeigt, was aus einer Deckliste würde.
func (d *Dienst) decklisteVorschau(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	deckID, gut := kennung(schreiben, anfrage, "deckID")
	if !gut {
		return
	}

	var wunsch struct {
		Text string `json:"text"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	vorschau, fehler := d.naehte.Deckeinfuhr.DecklisteVorschau(anfrage.Context(),
		deckID, wunsch.Text)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, vorschau)
}

// decklisteUebernehmen legt die bestätigten Zeilen ins Deck.
func (d *Dienst) decklisteUebernehmen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	deckID, gut := kennung(schreiben, anfrage, "deckID")
	if !gut {
		return
	}

	var wunsch struct {
		Stand  string             `json:"stand"`
		Zeilen []Deckeinfuhrzeile `json:"zeilen"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	bericht, fehler := d.naehte.Deckeinfuhr.DecklisteUebernehmen(anfrage.Context(),
		deckID, wunsch.Stand, wunsch.Zeilen)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, bericht)
}
