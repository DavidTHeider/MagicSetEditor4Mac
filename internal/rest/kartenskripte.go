// **Karten aus einem Skript, über den Dienst** — Werkstück B von Phase 4.11
// (#1337).
//
// Einordnung nach §16: **erweiternd** — zwei Wege kommen hinzu.

package rest

import (
	"context"
	"net/http"
)

// Kartenskriptwerkstatt ist, was der Dienst für die Kartenskripte braucht.
type Kartenskriptwerkstatt interface {
	Kartenskripte(ctx context.Context, setID int64) (any, error)
	KartenskriptAusfuehren(ctx context.Context, setID int64, name string) (any, error)
}

// kartenskripte nennt die Kartenskripte eines Sets.
func (d *Dienst) kartenskripte(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	skripte, fehler := d.naehte.Kartenskripte.Kartenskripte(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"skripte": skripte})
}

// kartenskriptAusfuehren führt ein Skript und hängt seine Karten ans Set.
//
// **`POST` und nicht `PUT`**: der Ruf legt Karten an, und ein zweiter legt
// weitere an — er ist nicht idempotent, und das steht ihm ins Gesicht
// geschrieben.
func (d *Dienst) kartenskriptAusfuehren(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	var rumpf struct {
		Name string `json:"name"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	lauf, fehler := d.naehte.Kartenskripte.KartenskriptAusfuehren(
		anfrage.Context(), setID, rumpf.Name)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **201 und nicht 200**: es sind Karten entstanden, und der Rumpf sagt
	// welche.
	ja(schreiben, http.StatusCreated, lauf)
}
