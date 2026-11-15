package rest

// **Die atomare Gruppe über REST** — Werkstück A13.1b von Phase 5.3
// (#1452).
//
// # Warum er nicht unter einer Karte steht
//
// **Weil er mehrere meint** — dieselbe Überlegung wie beim Batch nebenan. Und
// weil das Ganze der Punkt ist: ein Weg je Karte wäre ein Aufruf je Karte, und
// damit wieder N Transaktionen.
//
// # `POST` und nicht `PATCH`
//
// **Weil er mehrere Karten ändert** und nicht eine bestehende Ressource
// fortschreibt. Was er ändert, steht in seinem Rumpf; welche Karten es trifft,
// entscheidet erst der Server aus den Aufträgen.

import (
	"net/http"
)

// KIGruppenwahlsicht ist ein Vorschlag mit den Feldern, die gelten sollen.
type KIGruppenwahlsicht struct {
	Kennung string   `json:"kennung"`
	Felder  []string `json:"felder"`
}

// KIGruppenbezug ist, was für eine Gruppenübernahme gebraucht wird.
type KIGruppenbezug struct {
	Wahl []KIGruppenwahlsicht `json:"wahl"`
}

// KIGruppenkartensicht nennt eine Karte und ihre geschriebenen Felder.
type KIGruppenkartensicht struct {
	KarteID int64    `json:"karteId"`
	Felder  []string `json:"felder"`
}

// KIGruppensicht sagt, was aus der Gruppe wurde.
type KIGruppensicht struct {
	Karten int                    `json:"karten"`
	Felder int                    `json:"felder"`
	Namen  []KIGruppenkartensicht `json:"namen"`
}

func (d *Dienst) kigruppe(schreiben http.ResponseWriter, anfrage *http.Request) {
	var bezug KIGruppenbezug

	if !gelesen(schreiben, anfrage, &bezug) {
		return
	}

	sicht, rufefehler := d.naehte.Preise.KIGruppe(anfrage.Context(), bezug)
	if rufefehler != nil {
		fehlerhaft(schreiben, rufefehler)

		return
	}

	ja(schreiben, http.StatusOK, sicht)
}
