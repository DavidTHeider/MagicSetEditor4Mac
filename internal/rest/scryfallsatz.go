// **Ein ganzer Satz von Scryfall** —
// #1033.

package rest

import (
	"context"
	"net/http"
)

// Fremdsaetze ist, was der Satzendpunkt braucht.
type Fremdsaetze interface {
	Einlesen(ctx context.Context, code, spiel string) (Satz, error)
}

// Satz ist, was ein Einlesen zurückmeldet.
type Satz struct {
	SetID      int64    `json:"setID"`
	Karten     int      `json:"karten"`
	Befunde    []string `json:"befunde"`
	SchonOffen bool     `json:"schonOffen"`
}

// scryfallSatz liest einen ganzen Satz als neues Set ein.
//
// # Warum `POST /sets/scryfall`
//
// **Weil ein Set entsteht.** `POST /sets` fährt ein Setpaket von der Platte
// ein; dies ist derselbe Vorgang aus einer anderen Quelle, und die Quelle
// steht im Weg statt im Rumpf.
//
// # Warum 200 und nicht 201, wenn der Satz schon liegt
//
// **Weil nichts entstanden ist** — „eine Quelle, ein Set" (#886). Ein 201
// behauptete eine Anlage, die es nicht gab.
func (d *Dienst) scryfallSatz(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Code  string `json:"code"`
		Spiel string `json:"spiel"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	eingelesen, fehler := d.naehte.Saetze.Einlesen(anfrage.Context(), rumpf.Code, rumpf.Spiel)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	if eingelesen.SchonOffen {
		ja(schreiben, http.StatusOK, eingelesen)

		return
	}

	ja(schreiben, http.StatusCreated, eingelesen)
}
