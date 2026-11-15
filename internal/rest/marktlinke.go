package rest

// **Die Wege nach draußen im Dienst** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// # Warum der Dienst Wege nennt und keinen öffnet
//
// **Weil er auf der Rückschleife lauscht und der Browser auf dem Bildschirm
// steht.** Wer über einen Port anfragt, kann nicht wollen, dass sich auf
// **diesem** Rechner ein Fenster öffnet — und wer es doch will, sitzt davor
// und hat das Fenster. Das Issue sagt es: „Standardbrowser durch injizierten
// Plattformöffner und **direkte Nutzeraktion**."
//
// **Die Adressen selbst kommen heraus**, und sie sind geprüft: wer sie
// weiterreicht, reicht etwas weiter, das durch die Positivliste gegangen ist.

import (
	"context"
	"net/http"
)

// Linknaht ist, was der Dienst für die Wege braucht.
type Linknaht interface {
	Marktlinks(ctx context.Context, printing, name string) ([]Marktlink, error)
}

// Marktlink ist ein Weg nach draußen.
type Marktlink struct {
	Anbieter string `json:"anbieter"`

	// Art ist `produkt`, `suche` oder `startseite` — **die Genauigkeit
	// hängt daran**, und wer sie verschweigt, verspricht zu viel.
	Art string `json:"art"`

	// Herkunft ist `kennung`, `name` oder `keine`.
	Herkunft string `json:"herkunft"`

	Adresse string `json:"adresse"`
}

// marktlinkeZeigen gibt die Wege zu einer Karte.
func (d *Dienst) marktlinkeZeigen(
	schreiben http.ResponseWriter, anfrage *http.Request,
) {
	links, fehler := d.naehte.Links.Marktlinks(anfrage.Context(),
		anfrage.PathValue("printing"), anfrage.URL.Query().Get("name"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, links)
}
