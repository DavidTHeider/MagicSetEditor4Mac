// **Ein Turnier als Datei und als Bericht** — DOM-04.AK-03
// (#1699),
// Vermerk 0140.
//
// Einordnung nach §16: **erweiternd** — drei Wege kommen hinzu. **Als Text
// im Rumpf und nicht als Datei**: ein Dienst steht an keiner Platte des
// Rufers.

package rest

import "net/http"

// turnierAusgabe gibt ein Turnier als Datei im Text.
func (d *Dienst) turnierAusgabe(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.zumTurnier(schreiben, anfrage, d.naehte.Turniere.TurnierAusgabe)
}

// turnierBericht gibt den Abschlussbericht als Markdown im Feld `inhalt`.
func (d *Dienst) turnierBericht(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.zumTurnier(schreiben, anfrage, d.naehte.Turniere.TurnierBericht)
}

// turnierEinlesen legt ein Turnier aus dem Text einer Ausgabe an — `201`,
// denn es entsteht eines.
func (d *Dienst) turnierEinlesen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Inhalt string `json:"inhalt"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	inhalt, gut := pflichttext(schreiben, "inhalt", rumpf.Inhalt)
	if !gut {
		return
	}

	angaben, fehler := d.naehte.Turniere.TurnierEinlesen(anfrage.Context(), inhalt)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, angaben)
}
