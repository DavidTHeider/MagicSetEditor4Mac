// **Ein Turnier nach einer Vorlage** — DOM-04.AK-01 und AK-03
// (#1699).
//
// Einordnung nach §16: **erweiternd** — ein Weg kommt hinzu.

package rest

import "net/http"

// turnierKopieren legt ein neues Turnier nach einer Vorlage an.
//
// **`201` und das neue Turnier**, wie beim Anlegen: es entsteht eines.
func (d *Dienst) turnierKopieren(schreiben http.ResponseWriter, anfrage *http.Request) {
	turnierID, gut := kennung(schreiben, anfrage, "turnierID")
	if !gut {
		return
	}

	var wunsch struct {
		Name             string `json:"name"`
		MitTeilnehmenden bool   `json:"mitTeilnehmenden"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	kopie, fehler := d.naehte.Turniere.TurnierKopieren(
		anfrage.Context(), turnierID, wunsch.Name, wunsch.MitTeilnehmenden)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, kopie)
}
