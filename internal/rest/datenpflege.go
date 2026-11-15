package rest

// **Die Datenpflege am Dienst** — Werkstück O8 von Phase 5.2
// (#1432).
//
// **Hinter der Marke**, wie jeder andere Weg: die Antwort sagt, was in diesem
// Bestand herumliegt, und das ist eine Auskunft über ihn.

import (
	"context"
	"net/http"
)

// Pflegeauskunft ist die Naht zur Anwendung.
type Pflegeauskunft interface {
	Pflegevorschau(ctx context.Context) (any, error)
	Pflegen(ctx context.Context, erstellt string, wahl []Pflegewahl) (any, error)
}

// Pflegewahl ist eine gewählte Maßnahme samt gesehener Revision.
type Pflegewahl struct {
	Art      string `json:"art"`
	Revision string `json:"revision"`
}

// pflegevorschau beantwortet `GET /pflege/vorschau`.
func (d *Dienst) pflegevorschau(schreiben http.ResponseWriter, anfrage *http.Request) {
	vorschau, fehler := d.naehte.Pflege.Pflegevorschau(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, vorschau)
}

// pflegen beantwortet `POST /pflege`.
func (d *Dienst) pflegen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var eingabe struct {
		Erstellt string       `json:"erstellt"`
		Wahl     []Pflegewahl `json:"wahl"`
	}

	if !gelesen(schreiben, anfrage, &eingabe) {
		return
	}

	if eingabe.Erstellt == "" {
		nein(schreiben, http.StatusBadRequest, "validierung",
			"erstellt fehlt — nenne den Zeitpunkt der Vorschau")

		return
	}

	bilanz, fehler := d.naehte.Pflege.Pflegen(
		anfrage.Context(), eingabe.Erstellt, eingabe.Wahl)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, bilanz)
}
