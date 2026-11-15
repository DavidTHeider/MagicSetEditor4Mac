package rest

// **Was die Seiten eines Tauschs ungefähr wert sind** — TRADE-02b
// (#1951).
//
//	GET /handel/{vorgang}/anhalt   Preis je Zeile, Summe je Seite, Differenz
//
// **Ein Anhalt und keine Bewertung**: die Preise gehen nicht in den Vertrag
// und nicht über die Leitung. Die Gestalt steht an der App-Naht
// (`app.PreisanhaltDTO`) und wird hier nur durchgereicht.

import "net/http"

// handelsanhaltZeigen beantwortet `GET /handel/{vorgang}/anhalt`.
func (d *Dienst) handelsanhaltZeigen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	anhalt, fehler := d.naehte.Leitung.HandelPreisanhalt(anfrage.Context(),
		anfrage.PathValue("vorgang"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, anhalt)
}
