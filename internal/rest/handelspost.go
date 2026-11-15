package rest

// **Der Poststand im Dienst** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// **`GET`, weil nichts entsteht**: der Stand ist eine Auskunft über das, was
// schon dasteht. Und er trägt Zahlen und keine Nachrichten — die Diagnose
// sagt, ob etwas klemmt, und nicht, worum es geht.

import (
	"context"
	"net/http"
)

// Postnaht ist, was der Dienst für den Poststand braucht.
type Postnaht interface {
	Handelspoststand(ctx context.Context, kennung string) (Poststand, error)
}

// Poststand sagt, was zu einem Gegenüber noch aussteht.
type Poststand struct {
	Kennung          string `json:"kennung"`
	Offen            int    `json:"offen"`
	HoechsteVersuche int    `json:"hoechsteVersuche"`
}

// poststandZeigen gibt den Stand zu einem Gegenüber.
func (d *Dienst) poststandZeigen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	stand, fehler := d.naehte.Post.Handelspoststand(anfrage.Context(),
		anfrage.PathValue("kennung"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, stand)
}
