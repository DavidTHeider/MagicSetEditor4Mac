// **Die Exportvorlagen über den Dienst** —
// #1291.

package rest

import "net/http"

// vorlagen nennt die Exportvorlagen, die zu einem Set passen.
func (d *Dienst) vorlagen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	gelesen, befunde, fehler := d.naehte.Vorlagen.Vorlagen(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"vorlagen": ohneLeereVorlagen(gelesen), "befunde": ohneLeeretexte(befunde),
	})
}

// vorlageAusfuehren führt eine Vorlage aus und schreibt das Ergebnis.
//
// Einordnung nach §16: **erweiternd** — ein Weg kommt hinzu, keiner ändert
// seine Bedeutung.
func (d *Dienst) vorlageAusfuehren(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	var rumpf struct {
		Paket string            `json:"paket"`
		Ziel  string            `json:"ziel"`
		Wahl  map[string]string `json:"wahl"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	befunde, fehler := d.naehte.Vorlagen.Ausfuehren(
		anfrage.Context(), setID, rumpf.Paket, rumpf.Ziel, rumpf.Wahl)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{
		"ziel": rumpf.Ziel, "befunde": ohneLeeretexte(befunde),
	})
}

// vorlagenoptionSetzen behält den Wert einer Vorlagenoption.
func (d *Dienst) vorlagenoptionSetzen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Wert string `json:"wert"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	vorlage := anfrage.PathValue("vorlage")
	feld := anfrage.PathValue("feld")

	if fehler := d.naehte.Vorlagen.VorlagenoptionSetzen(
		anfrage.Context(), vorlage, feld, rumpf.Wert); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{
		"vorlage": vorlage, "feld": feld, "wert": rumpf.Wert,
	})
}

// ohneLeereVorlagen macht aus `nil` eine leere Liste (§3).
func ohneLeereVorlagen(vorlagen []Vorlagenzeile) []Vorlagenzeile {
	if vorlagen == nil {
		return []Vorlagenzeile{}
	}

	return vorlagen
}
