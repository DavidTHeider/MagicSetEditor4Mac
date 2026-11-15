// **Die Einstellungen eines Menschen und seine Auskunft im Dienst** —
// Werkstück 8 von Phase 4.12
// (#1351),
// Punkte 9, 10 und 11.
//
// # Warum der Schlüssel im Weg steht und der Wert im Rumpf
//
// **Weil der Schlüssel der Ort ist.** `PUT /…/einstellungen/{schluessel}`
// benennt genau eine Einstellung; derselbe Ruf zweimal führt zum selben
// Zustand. Dieselbe Bauform wie bei den Setangaben.

package rest

import "net/http"

// nutzereinstellungen gibt die Einstellungen eines Menschen.
func (d *Dienst) nutzereinstellungen(schreiben http.ResponseWriter, anfrage *http.Request) {
	gelesen, fehler := d.naehte.Nutzer.Nutzereinstellungen(anfrage.Context(),
		anfrage.PathValue("nutzerID"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"einstellungen": gelesen})
}

// nutzereinstellungSetzen schreibt eine fort und gibt alle zurück.
func (d *Dienst) nutzereinstellungSetzen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	var rumpf struct {
		Wert string `json:"wert"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	geaendert, fehler := d.naehte.Nutzer.NutzereinstellungSetzen(anfrage.Context(),
		anfrage.PathValue("nutzerID"), anfrage.PathValue("schluessel"), rumpf.Wert)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"einstellungen": geaendert})
}

// personendaten gibt heraus, was über einen Menschen gespeichert ist.
//
// **Roh und nicht umgepackt.** Was ankommt, ist die Auskunft selbst; sie in
// einen Umschlag zu setzen hieße, sie zweimal zu erklären.
func (d *Dienst) personendaten(schreiben http.ResponseWriter, anfrage *http.Request) {
	auskunft, fehler := d.naehte.Nutzer.Personendaten(anfrage.Context(),
		anfrage.PathValue("nutzerID"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, auskunft)
}
