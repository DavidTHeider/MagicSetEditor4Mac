package rest

// **Kann die neueste Fassung diesen Bestand öffnen?** — Werkstück O6 von
// Phase 5.2 (#1430).
//
// **Hinter der Marke**, wie jeder andere Weg: die Antwort sagt, welche
// Wanderungen dieser Bestand verbucht hat, und das ist eine Auskunft über ihn.

import (
	"context"
	"net/http"
)

// Vertraeglichkeitsauskunft ist die Naht zur Anwendung.
type Vertraeglichkeitsauskunft interface {
	Vertraeglichkeit(ctx context.Context) (any, error)
}

// vertraeglichkeit beantwortet `GET /vertraeglichkeit`.
func (d *Dienst) vertraeglichkeit(schreiben http.ResponseWriter, anfrage *http.Request) {
	urteil, fehler := d.naehte.Vertraeglichkeit.Vertraeglichkeit(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, urteil)
}
