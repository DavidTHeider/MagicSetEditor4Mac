// **Die Parameterhilfen der Schlüsselwörter, über den Dienst** — Werkstück D
// von Phase 4.11
// (#1342).
//
// Einordnung nach §16: **erweiternd** — ein Weg kommt hinzu, `GET`.

package rest

import (
	"context"
	"net/http"
)

// Parameterhilfenwerkstatt ist, was der Dienst dafür braucht.
type Parameterhilfenwerkstatt interface {
	Parameterhilfen(ctx context.Context, setID, wortID int64) (any, error)
}

// parameterhilfen nennt beide Menüs zu einem Schlüsselwort.
func (d *Dienst) parameterhilfen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	wortID, gutesWort := kennung(schreiben, anfrage, "wortID")
	if !gutesWort {
		return
	}

	hilfen, fehler := d.naehte.Parameterhilfen.Parameterhilfen(
		anfrage.Context(), setID, wortID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, hilfen)
}
