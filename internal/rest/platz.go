// **Der Platz über REST** — Werkstück 4 von Phase 4.13
// (#1352).
//
// **Zwei Wege, beide erweiternd** (§16).
//
// # Warum `DELETE` auf den Posten und nicht `POST /platz/freigeben`
//
// **Weil der Posten der Ort ist** und das Wegräumen eine Löschung. Derselbe
// Ruf zweimal führt zum selben Zustand: nach dem ersten bleibt die jüngste
// Sicherung, nach dem zweiten dieselbe.

package rest

import (
	"context"
	"net/http"
)

// Platzstand ist, was wie viel belegt.
type Platzstand struct {
	Posten []Platzzeile `json:"posten"`

	// Nachladefehler: weggeräumt **ist**, nur die Zählung danach war nicht
	// zu machen (§7, Regel 3; #1970). Erweiternd (0033).
	Nachladefehler *Befund `json:"nachladefehler,omitempty"`
}

// Platzzeile ist ein Posten samt seinem Gewicht.
type Platzzeile struct {
	Posten   string `json:"posten"`
	Dateien  int    `json:"dateien"`
	Bytes    int64  `json:"bytes"`
	Raeumbar bool   `json:"raeumbar"`
}

// Fremdplatz sagt, wie viel Platz was belegt.
type Fremdplatz interface {
	Platz(ctx context.Context) (Platzstand, error)
	PlatzFreigeben(ctx context.Context, posten string) (Platzstand, error)
}

// platz zählt, was wie viel belegt.
func (d *Dienst) platz(schreiben http.ResponseWriter, anfrage *http.Request) {
	stand, fehler := d.naehte.Platz.Platz(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, stand)
}

// platzFreigeben räumt einen Posten weg.
func (d *Dienst) platzFreigeben(schreiben http.ResponseWriter, anfrage *http.Request) {
	stand, fehler := d.naehte.Platz.PlatzFreigeben(
		anfrage.Context(), anfrage.PathValue("posten"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, stand)
}
