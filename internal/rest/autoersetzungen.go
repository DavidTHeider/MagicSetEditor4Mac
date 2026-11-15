// **Die Autoersetzungen verwalten, über den Dienst** — Werkstück E von Phase
// 4.11 (#1341).
//
// Einordnung nach §16: **erweiternd** — vier Wege kommen hinzu.

package rest

import (
	"context"
	"net/http"
)

// Autoersetzungswerkstatt ist, was der Dienst dafür braucht.
type Autoersetzungswerkstatt interface {
	Autoersetzungen(ctx context.Context, setID int64) (any, error)
	AutoersetzungSetzen(ctx context.Context, setID int64, muster, ersatz string,
		ganzesWort, aktiv bool) error
	AutoersetzungLoeschen(ctx context.Context, setID int64, muster string) error
	AutoersetzungenZuruecksetzen(ctx context.Context, setID int64) error
	AutoersetzungSchalten(ctx context.Context, setID int64, an bool) error
}

// autoersetzungen nennt die Regeln, die für ein Set gelten.
func (d *Dienst) autoersetzungen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	regeln, fehler := d.naehte.Autoersetzungen.Autoersetzungen(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, regeln)
}

// autoersetzungSetzen legt eine Regel an oder schreibt sie fort.
//
// **`PUT` und das Muster im Weg**: dasselbe Muster zweimal ist dieselbe
// Regel, und ein zweiter Ruf ändert nichts weiter — er ist idempotent, und
// das steht ihm ins Gesicht geschrieben.
func (d *Dienst) autoersetzungSetzen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	var rumpf struct {
		Ersatz     string `json:"ersatz"`
		GanzesWort bool   `json:"ganzesWort"`
		Aktiv      bool   `json:"aktiv"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	if fehler := d.naehte.Autoersetzungen.AutoersetzungSetzen(anfrage.Context(), setID,
		anfrage.PathValue("muster"), rumpf.Ersatz,
		rumpf.GanzesWort, rumpf.Aktiv); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusNoContent, nil)
}

// autoersetzungLoeschen nimmt eine eigene Regel heraus.
func (d *Dienst) autoersetzungLoeschen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	if fehler := d.naehte.Autoersetzungen.AutoersetzungLoeschen(
		anfrage.Context(), setID, anfrage.PathValue("muster")); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusNoContent, nil)
}

// autoersetzungenZuruecksetzen vergisst alle Abweichungen eines Spiels.
//
// **`DELETE` auf die Sammlung**: es nimmt alle eigenen Regeln heraus, und der
// Schalter geht dabei zurück auf an.
func (d *Dienst) autoersetzungenZuruecksetzen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	if fehler := d.naehte.Autoersetzungen.AutoersetzungenZuruecksetzen(
		anfrage.Context(), setID); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusNoContent, nil)
}

// autoersetzungSchalten legt den globalen Schalter um.
func (d *Dienst) autoersetzungSchalten(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	var rumpf struct {
		An bool `json:"an"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	if fehler := d.naehte.Autoersetzungen.AutoersetzungSchalten(
		anfrage.Context(), setID, rumpf.An); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusNoContent, nil)
}
