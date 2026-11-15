// **Die Fragen an eine geordnete Sammlung, über den Dienst** — Werkstück 5
// von Phase 4.5 (#1292).
//
// Einordnung nach §16: **erweiternd** — vier Wege kommen hinzu.

package rest

import (
	"context"
	"net/http"
	"strconv"
)

// Sammlungssichtenwerkstatt ist, was der Dienst für die vier Fragen braucht.
type Sammlungssichtenwerkstatt interface {
	SammlungBesitz(ctx context.Context, printing string) (any, error)
	SammlungLagerorte(ctx context.Context) (any, error)
	SammlungDubletten(ctx context.Context, behalten int) (any, error)
	SammlungWuensche(ctx context.Context) (any, error)
}

// sammlungBesitz sagt, was von einem Druck da ist.
func (d *Dienst) sammlungBesitz(schreiben http.ResponseWriter, anfrage *http.Request) {
	stand, fehler := d.naehte.Sammlungssichten.SammlungBesitz(
		anfrage.Context(), anfrage.PathValue("printing"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, stand)
}

// sammlungLagerorte zählt die Lagerorte auf.
func (d *Dienst) sammlungLagerorte(schreiben http.ResponseWriter, anfrage *http.Request) {
	orte, fehler := d.naehte.Sammlungssichten.SammlungLagerorte(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"lagerorte": orte})
}

// sammlungDubletten zeigt, wovon mehr da ist als gewollt.
//
// **Die Zahl steht in der Abfrage** — sie wählt die Antwort aus und ändert
// nichts. Ohne sie gilt 1: das ist die Zahl, die die meisten meinen.
func (d *Dienst) sammlungDubletten(schreiben http.ResponseWriter, anfrage *http.Request) {
	behalten := 1

	if roh := anfrage.URL.Query().Get("behalten"); roh != "" {
		gelesen, fehler := strconv.Atoi(roh)
		if fehler != nil {
			nein(schreiben, http.StatusBadRequest, "validierung",
				"»behalten« ist keine Zahl")

			return
		}

		behalten = gelesen
	}

	dubletten, fehler := d.naehte.Sammlungssichten.SammlungDubletten(
		anfrage.Context(), behalten)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"dubletten": dubletten})
}

// sammlungWuensche gibt die Wunschliste.
func (d *Dienst) sammlungWuensche(schreiben http.ResponseWriter, anfrage *http.Request) {
	wuensche, fehler := d.naehte.Sammlungssichten.SammlungWuensche(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"wuensche": wuensche})
}
