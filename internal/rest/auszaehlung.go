// **Auszählen nach einer erklärten Dimension, über den Dienst** — Werkstück C
// von Phase 4.11 (#1338).
//
// Einordnung nach §16: **erweiternd** — zwei Wege kommen hinzu, beide `GET`.

package rest

import (
	"context"
	"net/http"
)

// Auszaehlwerkstatt ist, was der Dienst für die Auszählung braucht.
type Auszaehlwerkstatt interface {
	Auszaehldimensionen(ctx context.Context, setID int64) (any, error)
	Auszaehlen(ctx context.Context, setID int64, name string) (any, error)
}

// auszaehldimensionen nennt die Dimensionen eines Sets.
func (d *Dienst) auszaehldimensionen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	dimensionen, fehler := d.naehte.Auszaehlung.Auszaehldimensionen(
		anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"dimensionen": dimensionen})
}

// auszaehlung zählt ein Set nach einer Dimension aus.
//
// **`GET` und der Name im Weg**: der Ruf ändert nichts und gibt für dieselbe
// Frage dieselbe Antwort — er ist zwischenspeicherbar, und das steht ihm ins
// Gesicht geschrieben.
func (d *Dienst) auszaehlung(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	gezaehlt, fehler := d.naehte.Auszaehlung.Auszaehlen(
		anfrage.Context(), setID, anfrage.PathValue("name"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, gezaehlt)
}
