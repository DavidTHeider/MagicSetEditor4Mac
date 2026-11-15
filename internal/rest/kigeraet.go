package rest

import "net/http"

// kigeraetsitzung schaltet Apples Gerätemodell für diesen Dienst ab oder an
// — AI-05.AK-04 (#2042).
// **Sitzung heißt: so lange dieser Dienst läuft**; danach gilt wieder die
// Einstellung `ki.geraetemodell`.
func (d *Dienst) kigeraetsitzung(schreiben http.ResponseWriter, anfrage *http.Request) {
	var wunsch struct {
		An bool `json:"an"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	an, rufefehler := d.naehte.Preise.KIGeraetemodellFuerSitzung(anfrage.Context(), wunsch.An)
	if rufefehler != nil {
		fehlerhaft(schreiben, rufefehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"an": an})
}
