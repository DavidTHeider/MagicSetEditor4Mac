package rest

// **Die Kartenerkennung über den Dienst** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Warum sie nicht unter einer Karte steht
//
// **Weil es noch keine gibt.** Ein Foto wird gelesen, bevor irgendeine Karte
// feststeht — und welche es wird, entscheidet ein Mensch. Ein Weg unter
// `/karten/{id}` behauptete eine Zugehörigkeit, die der ganze Punkt dieses
// Werkstücks ist: sie besteht **nicht**.

import (
	"net/http"
)

// KIErkennungsbezug ist, was für eine Erkennung gebraucht wird.
type KIErkennungsbezug struct {
	Pfad     string `json:"pfad"`
	Anbieter string `json:"anbieter"`
	Alias    string `json:"alias"`
}

// **Die Antwort ist `any`** — dieselbe Wahl wie bei der Einschätzung nebenan:
// die Gestalt steht in `internal/app` und ist dort einmal beschrieben.

func (d *Dienst) kierkennung(schreiben http.ResponseWriter, anfrage *http.Request) {
	var bezug KIErkennungsbezug

	if !gelesen(schreiben, anfrage, &bezug) {
		return
	}

	sicht, rufefehler := d.naehte.Preise.KIErkennung(anfrage.Context(), bezug)
	if rufefehler != nil {
		fehlerhaft(schreiben, rufefehler)

		return
	}

	ja(schreiben, http.StatusOK, sicht)
}
