// **Abschluss und Protokoll eines Turniers** — DOM-04.AK-04
// (#1699),
// Vermerk 0139.
//
// Einordnung nach §16: **erweiternd** — drei Wege kommen hinzu.

package rest

import (
	"context"
	"net/http"
)

// turnierAbschliessen setzt den Abschluss — `204`, es entsteht nichts Neues.
func (d *Dienst) turnierAbschliessen(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.amAbschluss(schreiben, anfrage, d.naehte.Turniere.TurnierAbschliessen)
}

// turnierWiedereroeffnen nimmt ihn zurück.
func (d *Dienst) turnierWiedereroeffnen(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.amAbschluss(schreiben, anfrage, d.naehte.Turniere.TurnierWiedereroeffnen)
}

// amAbschluss bedient beide Richtungen.
func (d *Dienst) amAbschluss(
	schreiben http.ResponseWriter, anfrage *http.Request,
	weg func(context.Context, int64) error,
) {
	turnierID, gut := kennung(schreiben, anfrage, "turnierID")
	if !gut {
		return
	}

	if fehler := weg(anfrage.Context(), turnierID); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	schreiben.WriteHeader(http.StatusNoContent)
}

// turnierProtokoll gibt, was an einem Turnier geschah.
func (d *Dienst) turnierProtokoll(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.zumTurnier(schreiben, anfrage, d.naehte.Turniere.TurnierProtokoll)
}
