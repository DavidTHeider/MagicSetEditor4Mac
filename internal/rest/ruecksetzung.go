package rest

// **Die abgestufte Rücksetzung am Dienst** — Werkstück O4 von Phase 5.2
// (#1428).
//
// **Hinter der Marke**, wie jeder andere Weg — und der schreibende zusätzlich
// hinter dem, was der Fachkern prüft: wem das Ziel gehört, und ob für die
// ganze Anwendung ein Nachweis vorliegt.

import (
	"context"
	"net/http"
)

// Ruecksetzungsauskunft ist die Naht zur Anwendung.
type Ruecksetzungsauskunft interface {
	Ruecksetzungsarten(ctx context.Context) (any, error)
	Ruecksetzungsvorschau(ctx context.Context, art, profilID, nutzerID string) (any, error)
	Zuruecksetzen(ctx context.Context, art, profilID, nutzerID, name, nachweis,
		revision string) (
		any, error)
}

// ruecksetzungsarten beantwortet `GET /ruecksetzung/arten`.
func (d *Dienst) ruecksetzungsarten(schreiben http.ResponseWriter, anfrage *http.Request) {
	arten, fehler := d.naehte.Ruecksetzung.Ruecksetzungsarten(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, arten)
}

// ruecksetzungsvorschau beantwortet `GET /ruecksetzung/{art}/vorschau`.
//
// **Die Kennungen stehen in der Abfrage und nicht im Weg**: sie sind
// wahlweise — leer heißt „meines" —, und ein Weg mit optionalen Stellen wäre
// zwei Wege.
func (d *Dienst) ruecksetzungsvorschau(schreiben http.ResponseWriter, anfrage *http.Request) {
	vorschau, fehler := d.naehte.Ruecksetzung.Ruecksetzungsvorschau(
		anfrage.Context(), anfrage.PathValue("art"),
		anfrage.URL.Query().Get("profilID"), anfrage.URL.Query().Get("nutzerID"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, vorschau)
}

// zuruecksetzen beantwortet `POST /ruecksetzung/{art}`.
func (d *Dienst) zuruecksetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var eingabe struct {
		ProfilID string `json:"profilID"`
		NutzerID string `json:"nutzerID"`
		Name     string `json:"name"`
		Nachweis string `json:"nachweis"`

		// Revision ist die Marke aus der Vorschau — sie bindet die Ausführung
		// an das, was jemand gesehen hat (Befund 12 des Zweitblicks über
		// Phase 5.2). **Leer geht nur, wo nichts gelöscht wird** — Befund 9
		// des Nachlaufs: für Profil, Nutzer und die ganze Anwendung ist sie
		// Pflicht, sonst wäre die Bindung eine Zusage mit offener Tür.
		Revision string `json:"revision"`
	}

	if !gelesen(schreiben, anfrage, &eingabe) {
		return
	}

	bilanz, fehler := d.naehte.Ruecksetzung.Zuruecksetzen(
		anfrage.Context(), anfrage.PathValue("art"),
		eingabe.ProfilID, eingabe.NutzerID, eingabe.Name, eingabe.Nachweis,
		eingabe.Revision)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, bilanz)
}
