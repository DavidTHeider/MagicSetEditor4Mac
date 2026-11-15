// **Eine Karte aus einem Namen** —
// #1032.

package rest

import (
	"context"
	"net/http"
)

// Fremdanlage ist, was der Anlegeendpunkt braucht.
type Fremdanlage interface {
	Anlegen(ctx context.Context, setID int64, name string) (Anlage, error)
}

// Anlage ist, was ein Anlegen zurückmeldet.
type Anlage struct {
	KarteID      int64    `json:"karteID"`
	Felder       []string `json:"felder"`
	Befunde      []string `json:"befunde"`
	Herkunft     string   `json:"herkunft"`
	AlterStunden float64  `json:"alterStunden"`
}

// scryfallKarte legt eine Karte aus einem Scryfall-Namen an.
//
// # Warum `POST /sets/{id}/karten/scryfall` und nicht an einer Kartenadresse
//
// **Weil die Karte noch nicht existiert.** Eine Adresse mit `karteID` nennte
// eine Kennung, die der Aufrufer nicht haben kann; angelegt wird **im Set**,
// und das Set ist die Sammlung, an die etwas angehängt wird.
//
// # Was mit 201 zurückkommt
//
// **Die Kennung der angelegten Karte** — auch wenn das Füllen scheitert. Eine
// Karte, die entstanden ist, verschwindet nicht (§7, Regel 3); wer sie nicht
// nennt, lässt sie im Bestand zurück, ohne dass jemand davon weiß.
func (d *Dienst) scryfallKarte(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gutesSet := kennung(schreiben, anfrage, "setID")
	if !gutesSet {
		return
	}

	var rumpf struct {
		Name string `json:"name"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	angelegt, fehler := d.naehte.Anlage.Anlegen(anfrage.Context(), setID, rumpf.Name)
	if fehler != nil {
		// **Was schon geschrieben ist, kommt mit** (#1041): eine angelegte
		// Karte ist da, und wer nur den Fehler bekäme, fände sie nicht.
		fehlerhaftMit(schreiben, fehler, teilstand(angelegt))

		return
	}

	ja(schreiben, http.StatusCreated, angelegt)
}

// teilstand gibt, was schon geschrieben wurde — oder nichts.
//
// **Nichts heißt: es steht kein Feld in der Fehlerantwort.** Ein Aufruf, der
// vor der Anlage scheitert, hat keinen halben Stand, und ein leeres Feld wäre
// eine Frage, die niemand stellt.
func teilstand(angelegt Anlage) any {
	if angelegt.KarteID == 0 {
		return nil
	}

	return angelegt
}
