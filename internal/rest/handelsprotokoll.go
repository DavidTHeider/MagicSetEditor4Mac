package rest

// **Wie ein Handel dorthin kam** — TRADE-03.AK-03
// (#1717).
//
//	GET /handel/{vorgang}/protokoll   die Zustandswechsel, ältester zuerst

import "net/http"

// Handelsprotokollzeile ist ein Zustandswechsel, wie der Dienst ihn nennt.
type Handelsprotokollzeile struct {
	Revision int `json:"revision"`

	// Von ist leer, wenn die Revision neu ist.
	Von  string `json:"von"`
	Nach string `json:"nach"`
	Zeit string `json:"zeit"`
}

// handelsprotokollZeigen beantwortet `GET /handel/{vorgang}/protokoll`.
func (d *Dienst) handelsprotokollZeigen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	zeilen, fehler := d.naehte.Leitung.HandelProtokoll(anfrage.Context(),
		anfrage.PathValue("vorgang"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Eine leere Liste und kein `null`** — ein Vorgang ohne Wechsel ist
	// eine Antwort.
	if zeilen == nil {
		zeilen = []Handelsprotokollzeile{}
	}

	ja(schreiben, http.StatusOK, map[string]any{"zeilen": zeilen})
}
