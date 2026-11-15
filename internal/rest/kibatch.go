package rest

// **Ein Batch von Vorschlägen über REST** — Werkstück A13.1 von Phase 5.3
// (#1452).
//
// # Er steht nicht unter einer Karte
//
// **Weil er mehrere meint.** `POST /karten/{id}/ki-vorschlag` nennt sein Ziel
// im Weg; ein Batch nennt seine im Rumpf, und ein Weg unter **einer** Karte
// behauptete eine Zugehörigkeit, die es nicht gibt.
//
// # Und er schreibt an keiner Karte
//
// **Übernommen wird je Vorschlag** — dort steht die Entscheidung, die am
// Fenster der zweite Klick ist. Ein Cloudanbieter braucht dazu die Freigabe
// eines Menschen, und die gibt es nur am Fenster
// ([0105](../../docs/entscheidungen/0105-eine-freigabe-fuer-einen-batch.md)).

import (
	"net/http"
)

// KIBatchbezug ist, was für einen Batch gebraucht wird.
type KIBatchbezug struct {
	KarteIDs []int64 `json:"karteIds"`
	Anbieter string  `json:"anbieter"`
	Alias    string  `json:"alias"`
	Auftrag  string  `json:"auftrag"`
}

// KIBatchlagensicht ist, was aus **einem** Ziel geworden ist.
type KIBatchlagensicht struct {
	KarteID int64  `json:"karteId"`
	Kennung string `json:"kennung"`
	Modell  string `json:"modell"`

	Felder  []KIFeldwertsicht `json:"felder"`
	Befunde []string          `json:"befunde"`

	Ablauf string `json:"ablauf"`

	// Fehler ist der Grund **an diesem Ziel** — leer, wo es geklappt hat.
	Fehler string `json:"fehler"`
}

// KIBatchsicht trägt die Lagen aller Ziele.
type KIBatchsicht struct {
	Kennung  string              `json:"kennung"`
	Anbieter string              `json:"anbieter"`
	Ziele    []KIBatchlagensicht `json:"ziele"`
}

func (d *Dienst) kibatch(schreiben http.ResponseWriter, anfrage *http.Request) {
	var bezug KIBatchbezug

	if !gelesen(schreiben, anfrage, &bezug) {
		return
	}

	sicht, rufefehler := d.naehte.Preise.KIBatch(anfrage.Context(), bezug)
	if rufefehler != nil {
		fehlerhaft(schreiben, rufefehler)

		return
	}

	ja(schreiben, http.StatusOK, sicht)
}
