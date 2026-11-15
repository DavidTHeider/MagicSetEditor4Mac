package rest

// **Wie Kommandozeile und Werkzeugserver eingerichtet sind** — OPS-09.AK-03
// (#1907),
// rein lesend: eingerichtet wird am Fenster.

import (
	"context"
	"net/http"
)

// Schnittstellenauskunft ist, was `GET /schnittstellen` braucht.
type Schnittstellenauskunft interface {
	Schnittstellen(ctx context.Context) ([]Schnittstelle, error)
}

// Schnittstelle ist ein Ziel samt Stand.
type Schnittstelle struct {
	Name   string   `json:"name"`
	Stand  string   `json:"stand"`
	Datei  string   `json:"datei"`
	Art    string   `json:"art"`
	Zeilen []string `json:"zeilen"`
}

func (d *Dienst) schnittstellenSchauen(schreiben http.ResponseWriter, anfrage *http.Request) {
	stand, fehler := d.naehte.Schnittstellen.Schnittstellen(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, stand)
}
