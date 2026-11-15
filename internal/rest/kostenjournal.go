package rest

// **Das Kostenjournal am Dienst** — Werkstück O12 von Phase 5.2
// (#1439).

import (
	"context"
	"net/http"
)

// Kostenauskunft ist die Naht zur Anwendung.
type Kostenauskunft interface {
	Kostenstand(ctx context.Context, konto, zeitraum string) (any, error)
	Kostenposten(ctx context.Context, konto, zeitraum string) (any, error)
	KostenbudgetSetzen(ctx context.Context, konto, zeitraum string,
		wert int64, waehrung string) (any, error)
}

// kostenstand beantwortet `GET /kosten/{konto}`.
func (d *Dienst) kostenstand(schreiben http.ResponseWriter, anfrage *http.Request) {
	stand, fehler := d.naehte.Kosten.Kostenstand(anfrage.Context(),
		anfrage.PathValue("konto"), anfrage.URL.Query().Get("zeitraum"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, stand)
}

// kostenposten beantwortet `GET /kosten/{konto}/posten`.
func (d *Dienst) kostenposten(schreiben http.ResponseWriter, anfrage *http.Request) {
	posten, fehler := d.naehte.Kosten.Kostenposten(anfrage.Context(),
		anfrage.PathValue("konto"), anfrage.URL.Query().Get("zeitraum"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, posten)
}

// kostenbudget beantwortet `PUT /kosten/{konto}/budget`.
//
// **Der Betrag steht als Zahl und Währung im Rumpf** — nicht als „12,50 €":
// wer eine Formatierung zurückrechnet, rundet.
func (d *Dienst) kostenbudget(schreiben http.ResponseWriter, anfrage *http.Request) {
	var eingabe struct {
		Zeitraum string `json:"zeitraum"`
		Wert     int64  `json:"wert"`
		Waehrung string `json:"waehrung"`
	}

	if !gelesen(schreiben, anfrage, &eingabe) {
		return
	}

	if eingabe.Waehrung == "" {
		nein(schreiben, http.StatusBadRequest, "validierung",
			"waehrung fehlt — ein Betrag ohne Währung ist keiner")

		return
	}

	antwort, fehler := d.naehte.Kosten.KostenbudgetSetzen(anfrage.Context(),
		anfrage.PathValue("konto"), eingabe.Zeitraum, eingabe.Wert, eingabe.Waehrung)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, antwort)
}
