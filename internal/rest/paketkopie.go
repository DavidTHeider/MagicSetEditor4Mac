// **Eine eigene Kopie eines Pakets** — FMT-02.AK-04
// (#1691).
//
// Einordnung nach §16: **erweiternd** — ein Endpunkt kommt hinzu.

package rest

import "net/http"

// paketKopieren legt eine eigene Kopie an.
//
// **Am Paket, das kopiert wird** — `POST /pakete/{art}/{name}/kopie`, wie
// `POST /symbole/{symbolID}/kopie`. Der Rumpf nennt höchstens den neuen
// Namen; ohne ihn heißt die Kopie wie das Original mit `-kopie`.
func (d *Dienst) paketKopieren(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Neu string `json:"neu"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	paket, fehler := d.naehte.Pakete.PaketKopieren(
		anfrage.Context(), anfrage.PathValue("art"), anfrage.PathValue("name"), rumpf.Neu)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{"paket": paket})
}
